package main

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type counterSnapshot struct {
	IP        string
	BytesIn   uint64
	BytesOut  uint64
	Time      int64
}

var lastCounters = make(map[string]counterSnapshot)

func speedCollector(wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		collectAndCalculate()
		checkOffline()
	}
}

func collectAndCalculate() {
	// Read nft counters: each device IP has a forward counter rule
	// Format: "ip saddr 192.168.1.100 oif eth0 counter packets 12345 bytes 12345678"

	out, err := exec.Command("nft", "-j", "list", "chain", "ip", "devman", "forward").Output()
	if err != nil {
		// Chain might not exist yet — that's fine
		return
	}

	// Parse nft JSON output to get counter values per IP
	// Simple string parsing fallback:
	lines := strings.Split(string(out), "\n")
	current := make(map[string]uint64)

	for _, line := range lines {
		if !strings.Contains(line, "saddr") || !strings.Contains(line, "counter") {
			continue
		}

		ip := extractIP(line)
		bytes := extractBytes(line)

		if ip != "" && bytes > 0 {
			current[ip] = bytes
		}
	}

	now := time.Now().Unix()

	mu.Lock()
	defer mu.Unlock()

	// Calculate speed from counter deltas
	for ip, bytes := range current {
		last, exists := lastCounters[ip]
		if exists {
			delta := bytes - last.BytesOut
			interval := float64(now - last.Time)
			if interval > 0 {
				speed := uint64(float64(delta) / interval * 8) // bps

				// Update traffic table
				db.Exec(`INSERT INTO traffic (device_id, bytes_out, speed_out, recorded_at)
					SELECT id, ?, ?, ? FROM devices WHERE current_ip=?`,
					delta, speed, now, ip)

				// Update device current speed
				db.Exec(`UPDATE devices SET last_seen=? WHERE current_ip=?`,
					now, ip)
			}
		}
		lastCounters[ip] = counterSnapshot{IP: ip, BytesOut: bytes, Time: now}
	}
}

func extractIP(line string) string {
	// "ip saddr 192.168.1.100 oif eth0 ..."
	idx := strings.Index(line, "saddr ")
	if idx == -1 {
		return ""
	}
	rest := line[idx+6:]
	parts := strings.SplitN(rest, " ", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func extractBytes(line string) uint64 {
	idx := strings.Index(line, "bytes ")
	if idx == -1 {
		return 0
	}
	rest := line[idx+6:]
	parts := strings.SplitN(rest, " ", 2)
	if len(parts) > 0 {
		val, _ := strconv.ParseUint(parts[0], 10, 64)
		return val
	}
	return 0
}

// sendSpeedUpdate pushes real-time speed data for API consumption
func sendSpeedUpdate() {}
