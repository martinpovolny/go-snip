package main

import (
	"embed"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"
)

//go:embed static/index.html
var staticFiles embed.FS

func main() {
	addr     := flag.String("addr",     "0.0.0.0:9999", "listen address")
	hosts    := flag.String("hosts",    "/etc/hosts",   "hosts file to resolve peer IPs to names")
	monitors := flag.String("monitors", "",             "monitors config file")
	ttl      := flag.Duration("ttl",    15*time.Second, "cache TTL for WireGuard peers (ping is slow)")
	flag.Parse()

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
		mc := newMonitorCollector(mons, *ttl)
		go mc.Get()

		http.HandleFunc("/monitors-status", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			if err := json.NewEncoder(w).Encode(mc.Get()); err != nil {
				log.Printf("encode: %v", err)
			}
		})
		log.Printf("monitors: loaded %d endpoints from %s", len(mons), *monitors)
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
