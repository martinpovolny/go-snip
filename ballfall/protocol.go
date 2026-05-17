package main

// Pos is a (row, col) cell coordinate.
type Pos struct{ R, C int }

// MoveMsg is sent by a player client to make a move.
type MoveMsg struct {
	Type string `json:"type"` // "move"
	Row  int    `json:"row"`
	Col  int    `json:"col"`
	Dir  string `json:"dir"` // "left" | "right" | "up" | "down"
}

// StateMsg is broadcast to all clients after every significant state change.
type StateMsg struct {
	Type    string  `json:"type"`    // "state"
	Board   [][]int `json:"board"`   // [row][col], 0=empty, 1-4=color
	Width   int     `json:"width"`
	Height  int     `json:"height"`
	Score   int     `json:"score"`
	Status  string  `json:"status"`            // "waiting" | "clearing" | "invalid"
	Matches [][2]int `json:"matches,omitempty"` // cells being cleared
	Cascade int     `json:"cascade"`
}

// WelcomeMsg is sent once when a client connects.
type WelcomeMsg struct {
	Type   string `json:"type"` // "welcome"
	Role   string `json:"role"` // "player" | "observer"
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// SwapAnimMsg is broadcast when a swap animation begins.
type SwapAnimMsg struct {
	Type   string `json:"type"` // "swap_anim"
	RowA   int    `json:"rowA"`
	ColA   int    `json:"colA"`
	RowB   int    `json:"rowB"`
	ColB   int    `json:"colB"`
	ColorA int    `json:"colorA"`
	ColorB int    `json:"colorB"`
	Valid  bool   `json:"valid"` // false → will revert
}

// FallBallData describes one ball's fall path.
type FallBallData struct {
	Color   int `json:"color"`
	Col     int `json:"col"`
	FromRow int `json:"fromRow"` // may be negative (spawned above grid)
	ToRow   int `json:"toRow"`
}

// FallAnimMsg is broadcast when gravity+fill animation begins.
// Board is the final board state (after gravity and refill).
type FallAnimMsg struct {
	Type  string         `json:"type"` // "fall_anim"
	Board [][]int        `json:"board"`
	Balls []FallBallData `json:"balls"`
	Score int            `json:"score"`
	Cascade int          `json:"cascade"`
}

// RoleMsg is sent to a specific client when its role changes.
type RoleMsg struct {
	Type string `json:"type"` // "role"
	Role string `json:"role"` // "player" | "observer"
}

// InMsg is any message received from a client.
type InMsg struct {
	Type string `json:"type"`
}

func dirToDelta(dir string) (dr, dc int, ok bool) {
	switch dir {
	case "up":
		return -1, 0, true
	case "down":
		return 1, 0, true
	case "left":
		return 0, -1, true
	case "right":
		return 0, 1, true
	}
	return 0, 0, false
}
