// player connects to a running ballfall game via TCP and plays automatically.
// It tries every possible swap in reading order and picks the first one that
// creates a match, or falls back to a random swap.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"
)

const (
	gridW = 8
	gridH = 10
)

type StateMsg struct {
	Type   string  `json:"type"`
	Board  [][]int `json:"board"`
	Score  int     `json:"score"`
	Status string  `json:"status"`
}

type WelcomeMsg struct {
	Type string `json:"type"`
	Role string `json:"role"`
}

type MoveMsg struct {
	Type string `json:"type"`
	Row  int    `json:"row"`
	Col  int    `json:"col"`
	Dir  string `json:"dir"`
}

var dirs = []string{"left", "right", "up", "down"}

var deltas = map[string][2]int{
	"left":  {0, -1},
	"right": {0, 1},
	"up":    {-1, 0},
	"down":  {1, 0},
}

func colorAt(board [][]int, r, c int) int {
	if r < 0 || r >= gridH || c < 0 || c >= gridW {
		return -1
	}
	return board[r][c]
}

// findMatch returns the first swap (in reading order) that creates a match,
// or a random swap if no match-creating move is found.
func findMove(board [][]int) MoveMsg {
	type candidate struct {
		r, c int
		dir  string
	}

	// Simulate a swap and check for runs of 3+.
	createsMatch := func(r, c int, dir string) bool {
		dr, dc := deltas[dir][0], deltas[dir][1]
		r2, c2 := r+dr, c+dc
		if r2 < 0 || r2 >= gridH || c2 < 0 || c2 >= gridW {
			return false
		}
		// swap
		b := make([][]int, gridH)
		for i := range b {
			b[i] = make([]int, gridW)
			copy(b[i], board[i])
		}
		b[r][c], b[r2][c2] = b[r2][c2], b[r][c]
		return hasMatch(b)
	}

	for r := 0; r < gridH; r++ {
		for c := 0; c < gridW; c++ {
			for _, dir := range dirs {
				if createsMatch(r, c, dir) {
					return MoveMsg{Type: "move", Row: r, Col: c, Dir: dir}
				}
			}
		}
	}

	// No match-creating move: pick a random valid swap.
	var candidates []candidate
	for r := 0; r < gridH; r++ {
		for c := 0; c < gridW; c++ {
			for _, dir := range dirs {
				dr, dc := deltas[dir][0], deltas[dir][1]
				r2, c2 := r+dr, c+dc
				if r2 >= 0 && r2 < gridH && c2 >= 0 && c2 < gridW {
					candidates = append(candidates, candidate{r, c, dir})
				}
			}
		}
	}
	pick := candidates[rand.Intn(len(candidates))]
	return MoveMsg{Type: "move", Row: pick.r, Col: pick.c, Dir: pick.dir}
}

func hasMatch(board [][]int) bool {
	// horizontal runs
	for r := 0; r < gridH; r++ {
		run := 1
		for c := 1; c < gridW; c++ {
			if board[r][c] != 0 && board[r][c] == board[r][c-1] {
				run++
				if run >= 3 {
					return true
				}
			} else {
				run = 1
			}
		}
	}
	// vertical runs
	for c := 0; c < gridW; c++ {
		run := 1
		for r := 1; r < gridH; r++ {
			if board[r][c] != 0 && board[r][c] == board[r-1][c] {
				run++
				if run >= 3 {
					return true
				}
			} else {
				run = 1
			}
		}
	}
	// 2×2 blocks
	for r := 0; r < gridH-1; r++ {
		for c := 0; c < gridW-1; c++ {
			clr := board[r][c]
			if clr != 0 && board[r][c+1] == clr && board[r+1][c] == clr && board[r+1][c+1] == clr {
				return true
			}
		}
	}
	return false
}

func main() {
	addr := flag.String("addr", "localhost:7777", "game server address")
	delay := flag.Duration("delay", 500*time.Millisecond, "pause between moves")
	flag.Parse()

	conn, err := net.Dial("tcp", *addr)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close()
	log.Printf("connected to %s", *addr)

	scanner := bufio.NewScanner(conn)
	enc := json.NewEncoder(conn)

	// Read welcome.
	if !scanner.Scan() {
		log.Fatal("no welcome message")
	}
	var welcome WelcomeMsg
	if err := json.Unmarshal(scanner.Bytes(), &welcome); err != nil {
		log.Fatalf("parse welcome: %v", err)
	}
	fmt.Printf("role: %s\n", welcome.Role)

	if welcome.Role != "player" {
		log.Println("connected as observer — watching only")
		for scanner.Scan() {
			var msg StateMsg
			if err := json.Unmarshal(scanner.Bytes(), &msg); err == nil && msg.Type == "state" {
				fmt.Printf("score=%d status=%s\n", msg.Score, msg.Status)
			}
		}
		return
	}

	for scanner.Scan() {
		var msg StateMsg
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}
		if msg.Type != "state" {
			continue
		}
		fmt.Printf("score=%d status=%s\n", msg.Score, msg.Status)
		if msg.Status != "waiting" {
			continue // game is busy clearing/cascading, wait for next state
		}
		time.Sleep(*delay)
		move := findMove(msg.Board)
		fmt.Printf("move → row=%d col=%d dir=%s\n", move.Row, move.Col, move.Dir)
		if err := enc.Encode(move); err != nil {
			log.Printf("send: %v", err)
			return
		}
	}
}
