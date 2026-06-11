package main

import (
	"encoding/hex"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// dhcpDeviceInfo extracted from DHCP Discover packet
type dhcpDeviceInfo struct {
	MAC      string
	Hostname string
	Option55 string // DHCP Parameter Request List
}

func dhcpSniffer(wg *sync.WaitGroup) {
	defer wg.Done()

	iface := config.LANBridge
	handle, err := pcap.OpenLive(iface, 1600, true, pcap.BlockForever)
	if err != nil {
		log.Printf("DHCP sniffer: failed to open %s: %v", iface, err)
		return
	}
	defer handle.Close()

	// Filter: only DHCP UDP port 67 (client→server) broadcast
	filter := "udp and src port 68 and dst port 67"
	if err := handle.SetBPFFilter(filter); err != nil {
		log.Printf("DHCP sniffer: BPF filter error: %v", err)
		return
	}

	log.Printf("DHCP sniffer: listening on %s", iface)
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for packet := range packetSource.Packets() {
		info := parseDHCP(packet)
		if info == nil {
			continue
		}
		log.Printf("DHCP Discover: MAC=%s hostname=%s opt55=%s", info.MAC, info.Hostname, info.Option55)
		matchDevice(info)
	}
}

func parseDHCP(packet gopacket.Packet) *dhcpDeviceInfo {
	// Must be DHCP (not BOOTP reply, not relay)
	dhcpLayer := packet.Layer(layers.LayerTypeDHCPv4)
	if dhcpLayer == nil {
		return nil
	}
	dhcp, _ := dhcpLayer.(*layers.DHCPv4)

	// Only Discover (message type 1)
	if dhcp.Operation != layers.DHCPOpRequest {
		return nil
	}

	info := &dhcpDeviceInfo{}

	// MAC from Ethernet or DHCP client hardware address
	if ethLayer := packet.Layer(layers.LayerTypeEthernet); ethLayer != nil {
		eth, _ := ethLayer.(*layers.Ethernet)
		info.MAC = eth.SrcMAC.String()
	}

	// Hostname (option 12)
	for _, opt := range dhcp.Options {
		switch opt.Type {
		case layers.DHCPOptHostname:
			info.Hostname = string(opt.Data)
		case layers.DHCPOptParamsRequest:
			// Option 55: Parameter Request List
			info.Option55 = hex.EncodeToString(opt.Data)
		}
	}

	if info.MAC == "" {
		return nil
	}
	return info
}

func matchDevice(info *dhcpDeviceInfo) {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now().Unix()

	// Try to find existing device by DHCP fingerprint + hostname
	var deviceID int64
	var currentIP string

	err := db.QueryRow(`SELECT id, current_ip FROM devices 
		WHERE dhcp_fp = ? AND hostname = ? AND hostname != ''`,
		info.Option55, info.Hostname).Scan(&deviceID, &currentIP)

	if err == nil {
		// Matched existing device
		log.Printf("Device matched: id=%d hostname=%s", deviceID, info.Hostname)

		// Add new MAC to history
		db.Exec(`INSERT INTO mac_history (device_id, mac, seen_at) VALUES (?, ?, ?)`,
			deviceID, info.MAC, now)

		// Update current MAC and last_seen
		db.Exec(`UPDATE devices SET current_mac=?, last_seen=?, online=1 WHERE id=?`,
			info.MAC, now, deviceID)

		// If device was blocked/limited, re-apply rules for new IP
		reapplyRules(deviceID, currentIP)
		return
	}

	// No match — create new device profile
	result, err := db.Exec(`INSERT INTO devices (dhcp_fp, hostname, current_mac, first_seen, last_seen, online)
		VALUES (?, ?, ?, ?, ?, 1)`, info.Option55, info.Hostname, info.MAC, now, now)
	if err != nil {
		log.Printf("Failed to create device: %v", err)
		return
	}
	newID, _ := result.LastInsertId()
	log.Printf("New device created: id=%d hostname=%s mac=%s", newID, info.Hostname, info.MAC)

	db.Exec(`INSERT INTO mac_history (device_id, mac, seen_at) VALUES (?, ?, ?)`,
		newID, info.MAC, now)
}

// reapplyRules re-applies block/limit rules when a device gets a new IP
func reapplyRules(deviceID int64, oldIP string) {
	var blocked int
	var rateLimit int
	db.QueryRow(`SELECT is_blocked, rate_limit FROM devices WHERE id=?`, deviceID).Scan(&blocked, &rateLimit)

	if oldIP != "" && blocked == 1 {
		// Remove old drop rule
		execNft("delete rule ip devman forward ip saddr " + oldIP + " drop 2>/dev/null")
	}
	if oldIP != "" && rateLimit > 0 {
		// Remove old mark rule
		execNft("delete rule ip devman forward ip saddr " + oldIP + " counter 2>/dev/null")
	}
}


