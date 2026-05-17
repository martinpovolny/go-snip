package main

import (
	"encoding/json"
	"sync"
)

// Client represents one connected peer (player or observer).
type Client struct {
	send chan []byte // outgoing messages
	role string     // "player" or "observer"
}

// Hub manages all connected clients and routes moves to the game.
type Hub struct {
	mu        sync.Mutex
	clients   map[*Client]bool
	player    *Client // at most one player at a time
	lastState []byte  // last broadcast payload; sent to clients on connect

	// Game calls MoveIn when a player move arrives.
	// The channel is never closed; it is drained by the game loop.
	MoveIn chan MoveMsg
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]bool),
		MoveIn:  make(chan MoveMsg, 4),
	}
}

// Register adds a new client. The first registered client becomes the player;
// subsequent ones are observers. The current board state is queued immediately.
func (h *Hub) Register(c *Client) WelcomeMsg {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = true
	if h.player == nil {
		h.player = c
		c.role = "player"
	} else {
		c.role = "observer"
	}
	if h.lastState != nil {
		select {
		case c.send <- h.lastState:
		default:
		}
	}
	return WelcomeMsg{Type: "welcome", Role: c.role, Width: GridW, Height: GridH}
}

// Unregister removes a client. If it was the player, the next observer is promoted.
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, c)
	if h.player == c {
		h.player = nil
		// promote the first available observer
		for other := range h.clients {
			other.role = "player"
			h.player = other
			break
		}
	}
	close(c.send)
}

// Broadcast serializes msg, caches it as lastState, and queues it to every client.
func (h *Hub) Broadcast(msg any) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	data = append(data, '\n')
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastState = data
	for c := range h.clients {
		select {
		case c.send <- data:
		default: // slow client — drop
		}
	}
}

// DispatchMove is called by a client goroutine; it enqueues a move only if the
// sender is the current player.
func (h *Hub) DispatchMove(c *Client, m MoveMsg) {
	h.mu.Lock()
	isPlayer := h.player == c
	h.mu.Unlock()
	if isPlayer {
		select {
		case h.MoveIn <- m:
		default:
		}
	}
}
