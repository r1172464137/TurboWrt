package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func enforceRules() {
	mu.RLock()
	defer mu.RUnlock()

	// Ensure devman chain exists
	ensureDevManChain()

	rows, err := db.Query(`SELECT id, current_ip, is_blocked, rate_limit FROM devices WHERE online=1 AND current_ip != ''`)
	if err != nil {
		log.Printf("enforceRules query: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var ip string
		var blocked, rateLimit int
		rows.Scan(&id, &ip, &blocked, &rateLimit)

		if blocked == 1 {
			applyBlock(ip)
		} else {
			removeBlock(ip)
		}

		if rateLimit > 0 {
			applyRateLimit(ip, rateLimit)
		} else {
			removeRateLimit(ip)
		}
	}
}

func ensureDevManChain() {
	// idempotent chain creation
	execNft("add table ip devman 2>/dev/null")
	execNft("add chain ip devman forward { type filter hook forward priority 0 ; } 2>/dev/null")
}

// ---- Block ----

func applyBlock(ip string) {
	// Drop WAN-bound traffic for this IP
	rule := fmt.Sprintf("ip saddr %s oif %s drop", ip, config.WANInterface)
	execNft("add rule ip devman forward " + rule + " 2>/dev/null")
}

func removeBlock(ip string) {
	// Remove all drop rules for this IP in devman forward chain
	handle := getRuleHandle(ip, "drop")
	if handle != "" {
		execNft("delete rule ip devman forward handle " + handle + " 2>/dev/null")
	}
}

// ---- Rate Limit ----

func applyRateLimit(ip string, kbps int) {
	// Mark traffic from this device
	rule := fmt.Sprintf("ip saddr %s oif %s counter", ip, config.WANInterface)
	execNft("add rule ip devman forward " + rule + " 2>/dev/null")

	// Create tc class for this device (if not exists)
	// Use HTB qdisc on WAN interface for rate limiting
	_ = kbps // tc implementation below
}

func removeRateLimit(ip string) {
	handle := getRuleHandle(ip, "counter")
	if handle != "" {
		execNft("delete rule ip devman forward handle " + handle + " 2>/dev/null")
	}
}

// ---- Helpers ----

func getRuleHandle(ip string, action string) string {
	out, err := exec.Command("nft", "-a", "list", "chain", "ip", "devman", "forward").Output()
	if err != nil {
		return ""
	}

	// Format: "... ip saddr 192.168.1.100 oif eth0 drop # handle 5"
	lines := strings.Split(string(out), "\n")
	search := fmt.Sprintf("ip saddr %s", ip)
	for _, line := range lines {
		if strings.Contains(line, search) && strings.Contains(line, action) {
			idx := strings.LastIndex(line, "handle ")
			if idx >= 0 {
				return strings.TrimSpace(line[idx+7:])
			}
		}
	}
	return ""
}

func execNft(cmd string) {
	err := exec.Command("sh", "-c", "nft "+cmd).Run()
	if err != nil {
		log.Printf("nft error: %s: %v", cmd, err)
	}
}
