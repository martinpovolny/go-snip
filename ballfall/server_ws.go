package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// StartHTTP serves the web UI on addr and the /ws WebSocket endpoint.
func StartHTTP(addr string, h *Hub) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWS(w, r, h)
	})
	mux.Handle("/", http.FileServer(http.FS(webFS())))
	log.Printf("HTTP server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Printf("HTTP server: %v", err)
	}
}

func handleWS(w http.ResponseWriter, r *http.Request, h *Hub) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade: %v", err)
		return
	}
	defer conn.Close()

	c := &Client{send: make(chan []byte, 32)}
	welcome := h.Register(c)
	defer h.Unregister(c)

	// Send welcome.
	if data, err := json.Marshal(welcome); err == nil {
		conn.WriteMessage(websocket.TextMessage, data)
	}

	// Writer goroutine.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for data := range c.send {
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}()

	// Reader.
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var base InMsg
		if err := json.Unmarshal(msg, &base); err != nil {
			continue
		}
		if base.Type == "move" {
			var m MoveMsg
			if err := json.Unmarshal(msg, &m); err == nil {
				h.DispatchMove(c, m)
			}
		}
	}
	<-done
}
