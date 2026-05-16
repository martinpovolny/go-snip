package main

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:embed static/index.html
var staticFiles embed.FS

// ---------- types ----------

type Peer struct {
	Iface       string   `json:"iface"`
	Name        string   `json:"name"`
	Pubkey      string   `json:"pubkey"`
	PubkeyShort string   `json:"pubkey_short"`
	Endpoint    string   `json:"endpoint"`
	AllowedIPs  string   `json:"allowed_ips"`
	IP          string   `json:"ip"`
	Handshake   string   `json:"handshake"`
	HandshakeTS int64    `json:"handshake_ts"`
	Connected   bool     `json:"connected"`
	RxBytes     int64    `json:"rx_bytes"`
	TxBytes     int64    `json:"tx_bytes"`
	RxHuman     string   `json:"rx"`
	TxHuman     string   `json:"tx"`
	PingAlive   *bool    `json:"ping_alive"`
	PingRTTms   *float64 `json:"ping_rtt_ms"`
}

type Status struct {
	Peers       []Peer  `json:"peers"`
	Error       string  `json:"error,omitempty"`
	GeneratedAt float64 `json:"generated_at"`
}

// ---------- hosts parser ----------

// parseHosts reads /etc/hosts and returns ip → first hostname.
// Loopback addresses and comment lines are skipped.
func parseHosts(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	names := make(map[string]string)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// strip inline comments
		line, _, _ = strings.Cut(line, "#")
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		ip := fields[0]
		if strings.HasPrefix(ip, "127.") || ip == "::1" {
			continue
		}
		// first hostname wins; subsequent aliases ignored
		if _, exists := names[ip]; !exists {
			names[ip] = fields[1]
		}
	}
	return names, sc.Err()
}

// ---------- wg runtime ----------

func wgDump() ([]Peer, error) {
	out, err := exec.Command("wg", "show", "all", "dump").Output()
	if err != nil {
		return nil, fmt.Errorf("wg show: %w", err)
	}
	var peers []Peer
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		parts := strings.Split(sc.Text(), "\t")
		if len(parts) != 9 {
			continue // interface line has 5 fields, skip
		}
		// iface pubkey preshared endpoint allowed_ips latest_handshake rx tx keepalive
		ts, _ := strconv.ParseInt(parts[5], 10, 64)
		rx, _ := strconv.ParseInt(parts[6], 10, 64)
		tx, _ := strconv.ParseInt(parts[7], 10, 64)

		ep := parts[3]
		if ep == "(none)" {
			ep = ""
		}

		ip := ""
		if parts[4] != "(none)" {
			ip = strings.SplitN(parts[4], "/", 2)[0]
		}

		peers = append(peers, Peer{
			Iface:       parts[0],
			Pubkey:      parts[1],
			PubkeyShort: parts[1][:12] + "…",
			Endpoint:    ep,
			AllowedIPs:  parts[4],
			IP:          ip,
			HandshakeTS: ts,
			RxBytes:     rx,
			TxBytes:     tx,
		})
	}
	return peers, nil
}

// ---------- ping ----------

func ping(ip string) (alive bool, rttMs *float64) {
	if ip == "" {
		return false, nil
	}
	out, err := exec.Command("ping", "-c", "2", "-W", "2", ip).Output()
	if err != nil {
		return false, nil
	}
	// parse "rtt min/avg/max/mdev = X/Y/Z/W ms"
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "rtt") {
			// "rtt min/avg/max/mdev = 1.2/2.3/3.4/0.1 ms"
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				vals := strings.Split(parts[3], "/")
				if len(vals) >= 2 {
					avg, err := strconv.ParseFloat(vals[1], 64)
					if err == nil {
						r := math.Round(avg*10) / 10
						return true, &r
					}
				}
			}
		}
	}
	return true, nil
}

// ---------- format helpers ----------

func fmtBytes(n int64) string {
	units := []string{"B", "KiB", "MiB", "GiB", "TiB"}
	v := float64(n)
	for _, u := range units {
		if v < 1024 {
			return fmt.Sprintf("%.1f %s", v, u)
		}
		v /= 1024
	}
	return fmt.Sprintf("%.1f PiB", v)
}

func fmtHandshake(ts int64) string {
	if ts == 0 {
		return "—"
	}
	age := time.Now().Unix() - ts
	switch {
	case age < 60:
		return fmt.Sprintf("%ds ago", age)
	case age < 3600:
		return fmt.Sprintf("%dm ago", age/60)
	case age < 86400:
		return fmt.Sprintf("%dh ago", age/3600)
	default:
		return fmt.Sprintf("%dd ago", age/86400)
	}
}

// ---------- collector ----------

type Collector struct {
	mu        sync.Mutex
	namesPath string
	ttl       time.Duration
	cached    *Status
	fetchedAt time.Time
}

func (c *Collector) Get() *Status {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cached != nil && time.Since(c.fetchedAt) < c.ttl {
		return c.cached
	}
	c.cached = c.collect()
	c.fetchedAt = time.Now()
	return c.cached
}

func (c *Collector) collect() *Status {
	names, err := parseHosts(c.namesPath)
	if err != nil {
		log.Printf("parseHosts: %v", err)
		names = map[string]string{}
	}
	peers, err := wgDump()
	if err != nil {
		return &Status{Error: err.Error(), GeneratedAt: float64(time.Now().UnixMilli()) / 1000}
	}

	// ping concurrently
	type result struct {
		i     int
		alive bool
		rtt   *float64
	}
	ch := make(chan result, len(peers))
	for i, p := range peers {
		go func(i int, ip string) {
			alive, rtt := ping(ip)
			ch <- result{i, alive, rtt}
		}(i, p.IP)
	}
	for range peers {
		r := <-ch
		b := r.alive
		peers[r.i].PingAlive = &b
		peers[r.i].PingRTTms = r.rtt
	}

	now := time.Now().Unix()
	for i := range peers {
		p := &peers[i]
		p.Name = names[p.IP]
		p.Handshake = fmtHandshake(p.HandshakeTS)
		p.Connected = p.HandshakeTS > 0 && (now-p.HandshakeTS) < 180
		p.RxHuman = fmtBytes(p.RxBytes)
		p.TxHuman = fmtBytes(p.TxBytes)
	}

	return &Status{Peers: peers, GeneratedAt: float64(time.Now().UnixMilli()) / 1000}
}

// ---------- HTTP ----------

func main() {
	addr   := flag.String("addr",   "127.0.0.1:9999",        "listen address")
	hosts  := flag.String("hosts",  "/etc/hosts",              "hosts file to resolve peer IPs to names")
	ttl    := flag.Duration("ttl",  15*time.Second,           "cache TTL (ping is slow)")
	flag.Parse()

	col := &Collector{namesPath: *hosts, ttl: *ttl}

	// warm cache on startup so first page load is instant
	go col.Get()

	// /status — JSON API
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if err := json.NewEncoder(w).Encode(col.Get()); err != nil {
			log.Printf("encode: %v", err)
		}
	})

	// / — embedded dashboard
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := staticFiles.ReadFile("static/index.html")
		if err != nil {
			http.Error(w, "index.html not embedded", 500)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	log.Printf("wg-status listening on http://%s  hosts=%s  ttl=%s",
		*addr, *hosts, *ttl)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}
}
