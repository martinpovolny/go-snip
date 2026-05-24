package main

import (
	"bufio"
	"embed"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//go:embed static/index.html
var staticFiles embed.FS

func main() {
	addr      := flag.String("addr",       "0.0.0.0:9999",          "listen address")
	hosts     := flag.String("hosts",      "/etc/hosts",             "hosts file to resolve peer IPs to names")
	monitors  := flag.String("monitors",   "",                       "monitors config file")
	dbPath    := flag.String("db",         "",                       "sqlite database path for monitor history")
	ntfyURL         := flag.String("ntfy-url",          "http://localhost:2586", "ntfy server base URL")
	ntfyTopic       := flag.String("ntfy-topic",        "status-dashboard",      "ntfy topic for monitor alerts")
	ntfyEventsTopic := flag.String("ntfy-events-topic", "events",                "ntfy topic for general events feed")
	ntfyUser        := flag.String("ntfy-user",         "agents",                "ntfy username")
	ttl       := flag.Duration("ttl",      15*time.Second,           "cache TTL for WireGuard peers (ping is slow)")
	flag.Parse()

	ntfyPass := os.Getenv("NTFY_PASSWORD")

	// WireGuard peers
	col := &Collector{namesPath: *hosts, ttl: *ttl}
	go col.Get()

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if err := json.NewEncoder(w).Encode(col.Get()); err != nil {
			log.Printf("encode: %v", err)
		}
	})

	// HTTP monitors
	if *monitors != "" {
		mons, err := loadMonitors(*monitors)
		if err != nil {
			log.Fatalf("load monitors: %v", err)
		}
		log.Printf("monitors: loaded %d endpoints from %s", len(mons), *monitors)

		mc := newMonitorCollector(mons, *ttl)

		var store *Store
		if *dbPath != "" {
			store, err = NewStore(*dbPath, defaultKeep)
			if err != nil {
				log.Fatalf("open store: %v", err)
			}
			log.Printf("store: opened %s (keep %d rows/monitor)", *dbPath, defaultKeep)
		}

		var notifier *Notifier
		if ntfyPass != "" {
			notifier = NewNotifier(*ntfyURL, *ntfyTopic, *ntfyUser, ntfyPass, store)
			log.Printf("ntfy: will publish alerts to %s/%s", *ntfyURL, *ntfyTopic)
		} else {
			log.Printf("ntfy: NTFY_PASSWORD not set, notifications disabled")
		}

		if store != nil {
			http.HandleFunc("/notifications", func(w http.ResponseWriter, r *http.Request) {
				limit := 100
				if l := r.URL.Query().Get("limit"); l != "" {
					if n, err := strconv.Atoi(l); err == nil && n > 0 {
						limit = n
					}
				}
				recs, err := store.Notifications(limit)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Access-Control-Allow-Origin", "*")
				if err := json.NewEncoder(w).Encode(recs); err != nil {
					log.Printf("encode: %v", err)
				}
			})
		}

		mc.StartPoller(store, notifier, 5*time.Minute)

		http.HandleFunc("/monitors-status", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			if err := json.NewEncoder(w).Encode(mc.Get()); err != nil {
				log.Printf("encode: %v", err)
			}
		})

		if store != nil {
			http.HandleFunc("/monitors-history", func(w http.ResponseWriter, r *http.Request) {
				monitor := r.URL.Query().Get("monitor")
				if monitor == "" {
					http.Error(w, "missing ?monitor=", http.StatusBadRequest)
					return
				}
				limit := 288 // default: last 24h at 5-min intervals
				if l := r.URL.Query().Get("limit"); l != "" {
					if n, err := strconv.Atoi(l); err == nil && n > 0 {
						limit = n
					}
				}
				history, err := store.History(monitor, limit)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Access-Control-Allow-Origin", "*")
				if err := json.NewEncoder(w).Encode(history); err != nil {
					log.Printf("encode: %v", err)
				}
			})
		}
	}

	// ntfy events feed proxy
	http.HandleFunc("/ntfy-events", func(w http.ResponseWriter, r *http.Request) {
		url := strings.TrimRight(*ntfyURL, "/") + "/" + *ntfyEventsTopic + "/json?since=168h&poll=1"
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if ntfyPass != "" {
			req.SetBasicAuth(*ntfyUser, ntfyPass)
		}
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		type NtfyMsg struct {
			ID       string   `json:"id"`
			Time     int64    `json:"time"`
			Event    string   `json:"event"`
			Message  string   `json:"message"`
			Title    string   `json:"title,omitempty"`
			Tags     []string `json:"tags,omitempty"`
			Priority int      `json:"priority,omitempty"`
		}
		var msgs []NtfyMsg
		sc := bufio.NewScanner(resp.Body)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" {
				continue
			}
			var msg NtfyMsg
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				continue
			}
			if msg.Event == "message" {
				msgs = append(msgs, msg)
			}
		}
		// reverse to newest-first
		for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
			msgs[i], msgs[j] = msgs[j], msgs[i]
		}
		if msgs == nil {
			msgs = []NtfyMsg{}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if err := json.NewEncoder(w).Encode(msgs); err != nil {
			log.Printf("encode: %v", err)
		}
	})

	// Dashboard
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := staticFiles.ReadFile("static/index.html")
		if err != nil {
			http.Error(w, "index.html not embedded", 500)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	log.Printf("status-dashboard listening on http://%s  hosts=%s  ttl=%s", *addr, *hosts, *ttl)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}
}
