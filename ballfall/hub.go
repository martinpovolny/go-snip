package main

import (
	"encoding/json"
	"sync"
)

// Client represents one connected peer (player or observer).
type Client struct {
	send     chan []byte // outgoing messages
	role     string     // "player" or "observer"
	playerID int        // 1 or 2 in versus mode; 0 otherwise
}

// Hub manages all connected clients and routes messages to the game.
type Hub struct {
	mu          sync.Mutex
	clients     map[*Client]bool
	player      *Client // player 1 (or solo player)
	player2     *Client // player 2 in versus mode; nil otherwise
	lastState   []byte  // solo-mode last broadcast; sent to clients on connect
	lastStateP1 []byte  // versus: last board state for player 1
	lastStateP2 []byte  // versus: last board state for player 2
	mode        string  // current game mode: "demo" | "attack" | "versus"

	// Channels consumed by the game loop; never closed.
	MoveIn  chan MoveMsg    // consumed by game 1 (or the solo game)
	MoveIn2 chan MoveMsg    // consumed by game 2 in versus mode
	ModeIn  chan SetModeMsg
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]bool),
		mode:    "demo",
		MoveIn:  make(chan MoveMsg, 4),
		MoveIn2: make(chan MoveMsg, 4),
		ModeIn:  make(chan SetModeMsg, 4),
	}
}

// SetMode stores the current mode so new clients receive it in their welcome.
func (h *Hub) SetMode(mode string) {
	h.mu.Lock()
	h.mode = mode
	h.mu.Unlock()
}

// Register adds a new client. In solo mode the first client is the player and
// the rest are observers. In versus mode the first two clients are players 1
// and 2; the rest are observers. The current board state is queued immediately.
func (h *Hub) Register(c *Client) WelcomeMsg {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = true
	if h.player == nil {
		h.player = c
		c.role = "player"
		c.playerID = 1
	} else if h.mode == "versus" && h.player2 == nil {
		h.player2 = c
		c.role = "player"
		c.playerID = 2
	} else {
		c.role = "observer"
		c.playerID = 0
	}

	// Send last-known board state(s) so the client doesn't start blank.
	if h.lastStateP1 != nil {
		select { case c.send <- h.lastStateP1: default: }
	}
	if h.lastStateP2 != nil {
		select { case c.send <- h.lastStateP2: default: }
	} else if h.lastState != nil && h.lastStateP1 == nil {
		select { case c.send <- h.lastState: default: }
	}

	return WelcomeMsg{
		Type:     "welcome",
		Role:     c.role,
		Width:    GridW,
		Height:   GridH,
		Mode:     h.mode,
		PlayerID: c.playerID,
	}
}

// Unregister removes a client. If it was the (solo) player, the next observer
// is promoted. In versus mode vacated player slots stay empty until claimed.
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, c)
	if h.player == c {
		h.player = nil
		if h.mode != "versus" {
			// Solo: promote an observer.
			for other := range h.clients {
				other.role = "player"
				other.playerID = 1
				h.player = other
				break
			}
		}
	}
	if h.player2 == c {
		h.player2 = nil
	}
	close(c.send)
}

// BroadcastTagged serializes msg, caches it per-player, and queues it to every client.
// playerID 0 → solo cache; 1 or 2 → versus per-player cache.
func (h *Hub) BroadcastTagged(msg any, playerID int) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	data = append(data, '\n')
	h.mu.Lock()
	defer h.mu.Unlock()
	switch playerID {
	case 1:
		h.lastStateP1 = data
	case 2:
		h.lastStateP2 = data
	default:
		h.lastState = data
	}
	for c := range h.clients {
		select {
		case c.send <- data:
		default: // slow client — drop
		}
	}
}

// Broadcast is a convenience wrapper for solo-mode broadcasts (playerID 0).
func (h *Hub) Broadcast(msg any) {
	h.BroadcastTagged(msg, 0)
}

// DispatchMove enqueues a move to the correct game's move channel based on
// which player sent it.
func (h *Hub) DispatchMove(c *Client, m MoveMsg) {
	h.mu.Lock()
	isP1 := h.player == c
	isP2 := h.player2 == c
	h.mu.Unlock()
	if isP1 {
		select {
		case h.MoveIn <- m:
		default:
		}
	} else if isP2 {
		select {
		case h.MoveIn2 <- m:
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

// HandleClientMsg decodes one incoming JSON message and dispatches it.
// Called by both the TCP and WebSocket handlers so the routing logic lives
// in exactly one place.
func (h *Hub) HandleClientMsg(c *Client, msg []byte) {
	var base InMsg
	if err := json.Unmarshal(msg, &base); err != nil {
		return
	}
	switch base.Type {
	case "move":
		var m MoveMsg
		if err := json.Unmarshal(msg, &m); err == nil {
			h.DispatchMove(c, m)
		}
	case "claim":
		// Optional "player" field selects which slot to claim (default = 1).
		var claimMsg struct {
			Type   string `json:"type"`
			Player int    `json:"player"`
		}
		json.Unmarshal(msg, &claimMsg)
		if claimMsg.Player == 2 {
			demoted := h.ClaimPlayer2(c)
			if demoted != nil {
				h.SendTo(demoted, RoleMsg{Type: "role", Role: "observer"})
			}
			h.SendTo(c, RoleMsg{Type: "role", Role: "player", PlayerID: 2})
		} else {
			if demoted := h.ClaimPlayer(c); demoted != nil {
				h.SendTo(demoted, RoleMsg{Type: "role", Role: "observer"})
			}
			h.SendTo(c, RoleMsg{Type: "role", Role: "player", PlayerID: 1})
		}
	case "set_mode":
		var sm SetModeMsg
		if err := json.Unmarshal(msg, &sm); err == nil {
			h.DispatchMode(c, sm)
		}
	}
}

// ClaimPlayer makes c the player-1 unconditionally, demoting the current
// player-1 (if any) to observer. Returns the demoted client.
func (h *Hub) ClaimPlayer(c *Client) (demoted *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.player == c {
		return nil
	}
	demoted = h.player
	if demoted != nil {
		demoted.role = "observer"
		demoted.playerID = 0
	}
	h.player = c
	c.role = "player"
	c.playerID = 1
	return demoted
}

// ClaimPlayer2 makes c the player-2 slot, demoting any existing player-2.
// Returns the demoted client.
func (h *Hub) ClaimPlayer2(c *Client) (demoted *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.player2 == c {
		return nil
	}
	demoted = h.player2
	if demoted != nil {
		demoted.role = "observer"
		demoted.playerID = 0
	}
	h.player2 = c
	c.role = "player"
	c.playerID = 2
	return demoted
}
