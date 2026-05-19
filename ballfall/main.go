package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	httpAddr := flag.String("http", ":8080", "HTTP/WebSocket listen address")
	tcpAddr  := flag.String("tcp", ":7777", "TCP game socket listen address")
	unixPath := flag.String("unix", "./ballfall.sock", "Unix domain socket path")
	headless := flag.Bool("headless", false, "run without a display window (server mode)")
	flag.Parse()

	hub  := NewHub()
	game := NewGame(hub)

	go StartTCP(*tcpAddr, hub)

	os.MkdirAll(filepath.Dir(*unixPath), 0o755)
	os.Remove(*unixPath)
	go StartUnix(*unixPath, hub)

	go StartHTTP(*httpAddr, hub)

	game.BroadcastInitial()

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

	runGUI(game)
}
