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

// Hub manages all connected clients and routes messages to the game.
type Hub struct {
	mu        sync.Mutex
	clients   map[*Client]bool
	player    *Client // at most one player at a time
	lastState []byte  // last broadcast payload; sent to clients on connect
	mode      string  // current game mode: "demo" | "attack"

	// Channels consumed by the game loop; never closed.
	MoveIn chan MoveMsg
	ModeIn chan SetModeMsg
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]bool),
		mode:    "demo",
		MoveIn:  make(chan MoveMsg, 4),
		ModeIn:  make(chan SetModeMsg, 4),
	}
}

// SetMode stores the current mode so new clients receive it in their welcome.
func (h *Hub) SetMode(mode string) {
	h.mu.Lock()
	h.mode = mode
	h.mu.Unlock()
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
	return WelcomeMsg{Type: "welcome", Role: c.role, Width: GridW, Height: GridH, Mode: h.mode}
}

// Unregister removes a client. If it was the player, the next observer is promoted.
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, c)
	if h.player == c {
		h.player = nil
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

// DispatchMove enqueues a move only if the sender is the current player.
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

// DispatchMode enqueues a mode change only if the sender is the current player.
func (h *Hub) DispatchMode(c *Client, m SetModeMsg) {
	h.mu.Lock()
	isPlayer := h.player == c
	h.mu.Unlock()
	if isPlayer {
		select {
		case h.ModeIn <- m:
		default:
		}
	}
}

// SendTo queues a message to a single client (best-effort, drops if slow).
// Appends a newline so TCP scanner clients can detect message boundaries.
func (h *Hub) SendTo(c *Client, msg any) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	data = append(data, '\n')
	select {
	case c.send <- data:
	default:
	}
}

// ClaimPlayer makes c the player unconditionally, demoting the current player
// (if any) to observer. Returns the demoted client so the caller can notify it.
func (h *Hub) ClaimPlayer(c *Client) (demoted *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.player == c {
		return nil
	}
	demoted = h.player
	if demoted != nil {
		demoted.role = "observer"
	}
	h.player = c
	c.role = "player"
	return demoted
}
