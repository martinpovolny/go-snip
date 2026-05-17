package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	httpAddr := flag.String("http", ":8080", "HTTP/WebSocket listen address")
	tcpAddr  := flag.String("tcp", ":7777", "TCP game socket listen address")
	unixPath := flag.String("unix", "/run/ballfall/ballfall.sock", "Unix domain socket path")
	headless := flag.Bool("headless", false, "run without a display window (server/CI mode)")
	flag.Parse()

	hub  := NewHub()
	game := NewGame(hub)

	go StartTCP(*tcpAddr, hub)

	// Best-effort: create socket directory (systemd RuntimeDirectory does this in prod).
	os.MkdirAll(filepath.Dir(*unixPath), 0o755)
	os.Remove(*unixPath)
	go StartUnix(*unixPath, hub)

	go StartHTTP(*httpAddr, hub)

	// Broadcast initial board so connecting clients see the current state immediately.
	game.mu.Lock()
	hub.Broadcast(game.snapshot("waiting"))
	game.mu.Unlock()

	if *headless {
		log.Printf("headless mode — TCP %s  HTTP %s  Unix %s", *tcpAddr, *httpAddr, *unixPath)
		t := time.NewTicker(time.Second / 60)
		defer t.Stop()
		for range t.C {
			if err := game.Update(); err != nil {
				log.Fatal(err)
			}
		}
		return
	}

	ebiten.SetWindowSize(winW*2, winH*2)
	ebiten.SetWindowTitle("Ball Fall — match 3 to score!")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
