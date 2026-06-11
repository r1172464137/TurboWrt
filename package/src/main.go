package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DeviceProfile represents a device's identity across MAC changes
type DeviceProfile struct {
	ID          int64  `json:"id"`
	Alias       string `json:"alias"`
	DHCPFP      string `json:"dhcp_fp"`
	MDNSName    string `json:"mdns_name"`
	Hostname    string `json:"hostname"`
	CurrentIP   string `json:"current_ip"`
	CurrentMAC  string `json:"current_mac"`
	IsBlocked   bool   `json:"is_blocked"`
	RateLimit   int    `json:"rate_limit"` // kbps, 0=unlimited
	LastSeen    int64  `json:"last_seen"`
	Online      bool   `json:"online"`
	BytesIn     uint64 `json:"bytes_in"`
	BytesOut    uint64 `json:"bytes_out"`
	SpeedIn     uint64 `json:"speed_in"`  // bps
	SpeedOut    uint64 `json:"speed_out"` // bps
}

// Config holds runtime configuration
type Config struct {
	WANInterface string
	LANBridge    string
	DBPath       string
	APIPort      int
}

var (
	db     *sql.DB
	config Config
	mu     sync.RWMutex
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("devman starting...")

	config = Config{
		WANInterface: getEnv("WAN_IF", "eth0"),
		LANBridge:    getEnv("LAN_IF", "br-lan"),
		DBPath:       getEnv("DB_PATH", "/var/lib/devman.db"),
		APIPort:      9999,
	}

	// Initialize database
	var err error
	db, err = sql.Open("sqlite3", config.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	initDB()

	// Start goroutines
	var wg sync.WaitGroup

	wg.Add(1)
	go dhcpSniffer(&wg) // Step 2

	wg.Add(1)
	go conntrackWatcher(&wg) // Step 3

	wg.Add(1)
	go speedCollector(&wg) // Step 4

	wg.Add(1)
	go ruleEnforcer(&wg) // Step 5

	// HTTP API (rpcd-compatible)
	http.HandleFunc("/api/devices", handleDevices)
	http.HandleFunc("/api/block", handleBlock)
	http.HandleFunc("/api/limit", handleLimit)
	http.HandleFunc("/api/traffic", handleTraffic)

	go http.ListenAndServe(fmt.Sprintf(":%d", config.APIPort), nil)

	// Wait for shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("shutting down...")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func initDB() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS devices (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			alias TEXT DEFAULT '',
			dhcp_fp TEXT DEFAULT '',
			mdns_name TEXT DEFAULT '',
			hostname TEXT DEFAULT '',
			current_ip TEXT DEFAULT '',
			current_mac TEXT DEFAULT '',
			is_blocked INTEGER DEFAULT 0,
			rate_limit INTEGER DEFAULT 0,
			last_seen INTEGER DEFAULT 0,
			first_seen INTEGER DEFAULT 0,
			online INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS mac_history (
			device_id INTEGER,
			mac TEXT,
			ip TEXT,
			seen_at INTEGER
		)`,
		`CREATE TABLE IF NOT EXISTS traffic (
			device_id INTEGER,
			bytes_in INTEGER DEFAULT 0,
			bytes_out INTEGER DEFAULT 0,
			packets_in INTEGER DEFAULT 0,
			packets_out INTEGER DEFAULT 0,
			speed_in INTEGER DEFAULT 0,
			speed_out INTEGER DEFAULT 0,
			recorded_at INTEGER
		)`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			log.Fatal(err)
		}
	}
}

// ---- Stubs (filled in later steps) ----

func dhcpSniffer(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("DHCP sniffer: waiting for implementation")
	// Step 2: gopacket L2 capture on br-lan
}

func conntrackWatcher(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Conntrack watcher: waiting for implementation")
	// Step 3: netlink conntrack events
}

func speedCollector(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("Speed collector: waiting for implementation")
	// Step 4: nft counter polling
}

func ruleEnforcer(wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		enforceRules()
	}
}

func enforceRules() {
	// Step 5: apply nft drop / tc htb based on DB rules
}

// ---- API Handlers ----

func handleDevices(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	rows, err := db.Query(`SELECT id, alias, dhcp_fp, mdns_name, hostname, current_ip, current_mac,
		is_blocked, rate_limit, last_seen, online FROM devices ORDER BY last_seen DESC`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var devices []DeviceProfile
	for rows.Next() {
		var d DeviceProfile
		var blocked, online int
		rows.Scan(&d.ID, &d.Alias, &d.DHCPFP, &d.MDNSName, &d.Hostname,
			&d.CurrentIP, &d.CurrentMAC, &blocked, &d.RateLimit, &d.LastSeen, &online)
		d.IsBlocked = blocked == 1
		d.Online = online == 1

		// Get speed from latest traffic record
		db.QueryRow(`SELECT speed_in, speed_out FROM traffic WHERE device_id=? ORDER BY recorded_at DESC LIMIT 1`, d.ID).
			Scan(&d.SpeedIn, &d.SpeedOut)

		// Get total bytes
		db.QueryRow(`SELECT COALESCE(SUM(bytes_in),0), COALESCE(SUM(bytes_out),0) FROM traffic WHERE device_id=?`, d.ID).
			Scan(&d.BytesIn, &d.BytesOut)

		devices = append(devices, d)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}

func handleBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", 405)
		return
	}
	var req struct {
		DeviceID int64 `json:"device_id"`
		Block    bool  `json:"block"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	mu.Lock()
	defer mu.Unlock()

	val := 0
	if req.Block {
		val = 1
	}
	db.Exec("UPDATE devices SET is_blocked=? WHERE id=?", val, req.DeviceID)

	w.Write([]byte(`{"ok":true}`))
}

func handleLimit(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", 405)
		return
	}
	var req struct {
		DeviceID  int64 `json:"device_id"`
		RateLimit int   `json:"rate_limit"` // kbps, 0=unlimited
	}
	json.NewDecoder(r.Body).Decode(&req)

	mu.Lock()
	defer mu.Unlock()

	db.Exec("UPDATE devices SET rate_limit=? WHERE id=?", req.RateLimit, req.DeviceID)

	w.Write([]byte(`{"ok":true}`))
}

func handleTraffic(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	deviceID := r.URL.Query().Get("device_id")
	rows, err := db.Query(`SELECT speed_in, speed_out, bytes_in, bytes_out, recorded_at 
		FROM traffic WHERE device_id=? ORDER BY recorded_at DESC LIMIT 60`, deviceID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	type TrafficPoint struct {
		SpeedIn    uint64 `json:"speed_in"`
		SpeedOut   uint64 `json:"speed_out"`
		BytesIn    uint64 `json:"bytes_in"`
		BytesOut   uint64 `json:"bytes_out"`
		RecordedAt int64  `json:"recorded_at"`
	}
	var data []TrafficPoint
	for rows.Next() {
		var t TrafficPoint
		rows.Scan(&t.SpeedIn, &t.SpeedOut, &t.BytesIn, &t.BytesOut, &t.RecordedAt)
		data = append(data, t)
	}
	json.NewEncoder(w).Encode(data)
}
