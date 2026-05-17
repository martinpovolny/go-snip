package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
)

// StartTCP listens on a TCP port and accepts connections from players/observers.
func StartTCP(addr string, h *Hub) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("TCP listen %s: %v", addr, err)
		return
	}
	log.Printf("TCP server listening on %s", addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("TCP accept: %v", err)
			continue
		}
		go handleRawConn(conn, h)
	}
}

// StartUnix listens on a Unix domain socket.
func StartUnix(path string, h *Hub) {
	ln, err := net.Listen("unix", path)
	if err != nil {
		log.Printf("Unix listen %s: %v", path, err)
		return
	}
	log.Printf("Unix socket listening on %s", path)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Unix accept: %v", err)
			continue
		}
		go handleRawConn(conn, h)
	}
}

// handleRawConn handles one TCP or Unix connection with newline-delimited JSON.
func handleRawConn(conn net.Conn, h *Hub) {
	defer conn.Close()

	c := &Client{send: make(chan []byte, 32)}
	welcome := h.Register(c)
	defer h.Unregister(c)

	// Send welcome message.
	if data, err := json.Marshal(welcome); err == nil {
		conn.Write(append(data, '\n'))
	}

	// Writer goroutine: drain send channel → conn.
	// Exits when c.send is closed (by Unregister) or on write error.
	go func() {
		for data := range c.send {
			if _, err := conn.Write(data); err != nil {
				// Close the connection so the reader's Scan() returns.
				conn.Close()
				// Drain remaining messages so Unregister's close() doesn't block.
				for range c.send {
				}
				return
			}
		}
	}()

	// Reader: parse incoming JSON lines and dispatch moves/claims.
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		line := scanner.Bytes()
		var base InMsg
		if err := json.Unmarshal(line, &base); err != nil {
			continue
		}
		switch base.Type {
		case "move":
			var m MoveMsg
			if err := json.Unmarshal(line, &m); err == nil {
				h.DispatchMove(c, m)
			}
		case "claim":
			if demoted := h.ClaimPlayer(c); demoted != nil {
				h.SendTo(demoted, RoleMsg{Type: "role", Role: "observer"})
			}
			h.SendTo(c, RoleMsg{Type: "role", Role: "player"})
		case "set_mode":
			var sm SetModeMsg
			if err := json.Unmarshal(line, &sm); err == nil {
				h.DispatchMode(c, sm)
			}
		}
	}
	// Unregister closes c.send, which unblocks the writer goroutine.
}
