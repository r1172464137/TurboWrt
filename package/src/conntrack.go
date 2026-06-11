package main

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/ti-mo/netfilter"
)

func conntrackWatcher(wg *sync.WaitGroup) {
	defer wg.Done()

	// Subscribe to conntrack NEW events
	conn, err := netfilter.Dial(netfilter.Conntrack, netfilter.GroupsCT)
	if err != nil {
		log.Printf("Conntrack watcher: failed to subscribe: %v", err)
		return
	}
	defer conn.Close()

	log.Println("Conntrack watcher: listening")
	buf := make([]byte, 4096)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("Conntrack read error: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// Parse conntrack messages
		msgs, err := netfilter.Parse(buf[:n])
		if err != nil {
			continue
		}

		for _, msg := range msgs {
			ct := msg.Conntrack()
			if ct == nil {
				continue
			}

			// Only process NEW connections (device just started communicating)
			info := ct.Info()
			if info.State != netfilter.StateNew {
				continue
			}

			// Get source IP (LAN device)
			var srcIP net.IP
			if info.TupleOrig.IP != nil {
				srcIP = info.TupleOrig.IP.SourceAddress
			} else if info.TupleOrig.IPv6 != nil {
				srcIP = info.TupleOrig.IPv6.SourceAddress
			}
			if srcIP == nil {
				continue
			}

			// Check if this IP is routable to WAN interface
			// Only mark active if dst is external
			dstIP := info.TupleOrig.IP.DestinationAddress
			if dstIP == nil && info.TupleOrig.IPv6 != nil {
				dstIP = info.TupleOrig.IPv6.DestinationAddress
			}
			if dstIP == nil || dstIP.IsPrivate() || dstIP.IsLoopback() || dstIP.IsLinkLocalUnicast() {
				// Skip LAN-only traffic
				continue
			}

			markDeviceActive(srcIP.String())
		}
	}
}

func markDeviceActive(ip string) {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now().Unix()
	_, err := db.Exec(`UPDATE devices SET current_ip=?, last_seen=?, online=1 WHERE current_ip=?`,
		ip, now, ip)
	if err != nil {
		log.Printf("markDeviceActive: %v", err)
	}
}

// checkOffline called by speedCollector periodically to detect offline devices
func checkOffline() {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now().Unix()
	threshold := int64(120) // 2 minutes of no activity → offline

	db.Exec(`UPDATE devices SET online=0 WHERE online=1 AND last_seen < ?`, now-threshold)
}
