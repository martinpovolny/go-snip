package main

import (
	"embed"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
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
			if store != nil {
				es := NewEventSubscriber(*ntfyURL, *ntfyEventsTopic, *ntfyUser, ntfyPass, store)
				es.Start()
				log.Printf("events: subscribing to %s/%s", *ntfyURL, *ntfyEventsTopic)
			}
		} else {
			log.Printf("ntfy: NTFY_PASSWORD not set, notifications disabled")
		}

		if store != nil {
			http.HandleFunc("/ntfy-events", func(w http.ResponseWriter, r *http.Request) {
				limit := 1000
				if l := r.URL.Query().Get("limit"); l != "" {
					if n, err := strconv.Atoi(l); err == nil && n > 0 {
						limit = n
					}
				}
				evts, err := store.Events(limit)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if evts == nil {
					evts = []EventRecord{}
				}
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Access-Control-Allow-Origin", "*")
				if err := json.NewEncoder(w).Encode(evts); err != nil {
					log.Printf("encode: %v", err)
				}
			})
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
