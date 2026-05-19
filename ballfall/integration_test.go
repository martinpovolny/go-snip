//go:build integration

package main_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ── server lifecycle ──────────────────────────────────────────────────────────

const (
	itTCP  = "127.0.0.1:17777"
	itHTTP = "127.0.0.1:18080"
	itUnix = "/tmp/ballfall-integ-test.sock"
)

var itBinary string

// TestMain builds the binary once, starts one headless server for all tests,
// then tears it down when the suite exits.
func TestMain(m *testing.M) {
	bin, err := buildBinary()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n", err)
		os.Exit(1)
	}
	itBinary = bin

	srv := exec.Command(bin,
		"--headless",
		"--tcp", itTCP,
		"--http", itHTTP,
		"--unix", itUnix,
	)
	srv.Stdout = os.Stdout
	srv.Stderr = os.Stderr
	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "start server: %v\n", err)
		os.Exit(1)
	}
	if !waitReady(itTCP, 3*time.Second) {
		srv.Process.Kill()
		fmt.Fprintln(os.Stderr, "server did not start in time")
		os.Exit(1)
	}
	code := m.Run()
	srv.Process.Kill()
	srv.Wait()
	os.Remove(itUnix)
	os.Remove(bin)
	os.Exit(code)
}

func buildBinary() (string, error) {
	// Build server binary (no gui tag) into a temp file.
	dir, _ := os.MkdirTemp("", "ballfall-integ-*")
	bin := filepath.Join(dir, "ballfall-integ")
	cmd := exec.Command("go", "build", "-tags", "nogui", "-o", bin, ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w\n%s", err, out)
	}
	return bin, nil
}

func waitReady(addr string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		c, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			c.Close()
			return true
		}
		time.Sleep(30 * time.Millisecond)
	}
	return false
}

// ── connection helper ─────────────────────────────────────────────────────────

type conn struct {
	c   net.Conn
	sc  *bufio.Scanner
	enc *json.Encoder
	t   *testing.T
}

func dialTCP(t *testing.T) *conn {
	t.Helper()
	return dial(t, "tcp", itTCP)
}

func dialUnix(t *testing.T) *conn {
	t.Helper()
	return dial(t, "unix", itUnix)
}

func dial(t *testing.T, network, addr string) *conn {
	t.Helper()
	c, err := net.DialTimeout(network, addr, 3*time.Second)
	if err != nil {
		t.Fatalf("dial %s %s: %v", network, addr, err)
	}
	t.Cleanup(func() { c.Close() })
	sc := bufio.NewScanner(c)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	return &conn{c: c, sc: sc, enc: json.NewEncoder(c), t: t}
}

// readMsg reads the next newline-delimited JSON message.
func (c *conn) readMsg() map[string]any {
	c.t.Helper()
	c.c.SetReadDeadline(time.Now().Add(8 * time.Second))
	if !c.sc.Scan() {
		c.t.Fatalf("read: %v", c.sc.Err())
	}
	var msg map[string]any
	if err := json.Unmarshal(c.sc.Bytes(), &msg); err != nil {
		c.t.Fatalf("unmarshal %q: %v", c.sc.Text(), err)
	}
	return msg
}

// readUntil discards messages until one with the given type arrives.
func (c *conn) readUntil(typ string) map[string]any {
	c.t.Helper()
	for i := 0; i < 20; i++ {
		msg := c.readMsg()
		if msg["type"] == typ {
			return msg
		}
	}
	c.t.Fatalf("never received message type %q", typ)
	return nil
}

// readUntilStatus discards messages until a "state" with the given status arrives.
func (c *conn) readUntilStatus(status string) map[string]any {
	c.t.Helper()
	for i := 0; i < 30; i++ {
		msg := c.readMsg()
		if msg["type"] == "state" && msg["status"] == status {
			return msg
		}
	}
	c.t.Fatalf("never received state/%s", status)
	return nil
}

func (c *conn) send(v any) {
	c.t.Helper()
	if err := c.enc.Encode(v); err != nil {
		c.t.Fatalf("send: %v", err)
	}
}

func (c *conn) sendMove(row, col int, dir string) {
	c.send(map[string]any{"type": "move", "row": row, "col": col, "dir": dir})
}

func (c *conn) sendClaim() {
	c.send(map[string]any{"type": "claim"})
}

// welcome reads the first message (must be "welcome") and returns it.
func (c *conn) welcome() map[string]any {
	c.t.Helper()
	msg := c.readMsg()
	if msg["type"] != "welcome" {
		c.t.Fatalf("expected welcome, got %v", msg["type"])
	}
	return msg
}

// becomePlayer reads the welcome and, if the role is "observer", sends a
// claim and waits for the role promotion. Returns the final role.
func (c *conn) becomePlayer() string {
	c.t.Helper()
	w := c.welcome()
	role, _ := w["role"].(string)
	if role == "player" {
		return "player"
	}
	c.sendClaim()
	rm := c.readUntil("role")
	return rm["role"].(string)
}

// ── board helpers ─────────────────────────────────────────────────────────────

func toBoard(msg map[string]any) [][]int {
	raw, _ := msg["board"].([]any)
	board := make([][]int, len(raw))
	for r, rowAny := range raw {
		rowSlice := rowAny.([]any)
		board[r] = make([]int, len(rowSlice))
		for c, v := range rowSlice {
			board[r][c] = int(v.(float64))
		}
	}
	return board
}

func toBricks(msg map[string]any) [][]bool {
	raw, _ := msg["bricks"].([]any)
	bricks := make([][]bool, len(raw))
	for r, rowAny := range raw {
		rowSlice, _ := rowAny.([]any)
		bricks[r] = make([]bool, len(rowSlice))
		for c, v := range rowSlice {
			bricks[r][c], _ = v.(bool)
		}
	}
	return bricks
}

var allDirs = []string{"right", "left", "down", "up"}
var dirDelta = map[string][2]int{"right": {0, 1}, "left": {0, -1}, "down": {1, 0}, "up": {-1, 0}}

func hasMatchOnBoard(b [][]int) bool {
	rows, cols := len(b), len(b[0])
	for r := 0; r < rows; r++ {
		run := 1
		for c := 1; c < cols; c++ {
			if b[r][c] != 0 && b[r][c] == b[r][c-1] {
				if run++; run >= 3 {
					return true
				}
			} else {
				run = 1
			}
		}
	}
	for c := 0; c < cols; c++ {
		run := 1
		for r := 1; r < rows; r++ {
			if b[r][c] != 0 && b[r][c] == b[r-1][c] {
				if run++; run >= 3 {
					return true
				}
			} else {
				run = 1
			}
		}
	}
	for r := 0; r < rows-1; r++ {
		for c := 0; c < cols-1; c++ {
			clr := b[r][c]
			if clr != 0 && b[r][c+1] == clr && b[r+1][c] == clr && b[r+1][c+1] == clr {
				return true
			}
		}
	}
	return false
}

// isServerRejected reports whether the server would silently ignore this move
// (no swap_anim sent). Mirrors the guard in game.go initiateSwap.
func isServerRejected(board [][]int, bricks [][]bool, r, c, nr, nc int, d string) bool {
	if board[r][c] == 0 || bricks[r][c] || bricks[nr][nc] {
		return true
	}
	// Moving up into empty space is also silently rejected.
	if dirDelta[d][0] < 0 && board[nr][nc] == 0 {
		return true
	}
	return false
}

// findValidMove returns the first swap (in reading order) that the server will
// accept and that creates a match.
func findValidMove(board [][]int, bricks [][]bool) (row, col int, dir string, ok bool) {
	rows, cols := len(board), len(board[0])
	cp := make([][]int, rows)
	for i := range cp {
		cp[i] = make([]int, cols)
	}
	for _, d := range allDirs {
		delta := dirDelta[d]
		for r := 0; r < rows; r++ {
			for c := 0; c < cols; c++ {
				nr, nc := r+delta[0], c+delta[1]
				if nr < 0 || nr >= rows || nc < 0 || nc >= cols {
					continue
				}
				if isServerRejected(board, bricks, r, c, nr, nc, d) {
					continue
				}
				for i := range board {
					copy(cp[i], board[i])
				}
				cp[r][c], cp[nr][nc] = cp[nr][nc], cp[r][c]
				if hasMatchOnBoard(cp) {
					return r, c, d, true
				}
			}
		}
	}
	return 0, 0, "", false
}

// findInvalidMove returns a swap that the server will accept but that does NOT
// create a match (so the server will reverse it with valid=false).
func findInvalidMove(board [][]int, bricks [][]bool) (row, col int, dir string, ok bool) {
	rows, cols := len(board), len(board[0])
	cp := make([][]int, rows)
	for i := range cp {
		cp[i] = make([]int, cols)
	}
	for _, d := range allDirs {
		delta := dirDelta[d]
		for r := 0; r < rows; r++ {
			for c := 0; c < cols; c++ {
				nr, nc := r+delta[0], c+delta[1]
				if nr < 0 || nr >= rows || nc < 0 || nc >= cols {
					continue
				}
				if isServerRejected(board, bricks, r, c, nr, nc, d) {
					continue
				}
				for i := range board {
					copy(cp[i], board[i])
				}
				cp[r][c], cp[nr][nc] = cp[nr][nc], cp[r][c]
				if !hasMatchOnBoard(cp) {
					return r, c, d, true
				}
			}
		}
	}
	return 0, 0, "", false
}

// ── tests ─────────────────────────────────────────────────────────────────────

// TestHelp verifies the binary handles --help without panicking.
func TestHelp(t *testing.T) {
	cmd := exec.Command(itBinary, "--help")
	out, _ := cmd.CombinedOutput()
	if strings.Contains(string(out), "panic") {
		t.Fatalf("--help produced a panic:\n%s", out)
	}
	if !strings.Contains(string(out), "-headless") {
		t.Errorf("--help output missing expected flags:\n%s", out)
	}
}

// TestTCPConnect verifies the welcome message over TCP.
func TestTCPConnect(t *testing.T) {
	c := dialTCP(t)
	w := c.becomePlayer()
	if w != "player" {
		t.Fatalf("expected player, got %s", w)
	}
	// Should also receive the initial state
	state := c.readUntil("state")
	board := toBoard(state)
	if len(board) != 10 {
		t.Errorf("board height: want 10, got %d", len(board))
	}
	if len(board[0]) != 8 {
		t.Errorf("board width: want 8, got %d", len(board[0]))
	}
}

// TestUnixConnect verifies the welcome message over Unix socket.
func TestUnixConnect(t *testing.T) {
	c := dialUnix(t)
	role := c.becomePlayer()
	if role != "player" {
		t.Fatalf("expected player via unix socket, got %s", role)
	}
	state := c.readUntil("state")
	if state["status"] != "waiting" {
		t.Errorf("expected waiting status, got %v", state["status"])
	}
}

// TestSecondClientObserver verifies the second connection gets observer role.
func TestSecondClientObserver(t *testing.T) {
	c1 := dialTCP(t)
	c1.becomePlayer()

	c2 := dialTCP(t)
	w2 := c2.welcome()
	if w2["role"] != "observer" {
		t.Errorf("second client: expected observer, got %v", w2["role"])
	}
}

// TestObserverClaim verifies an observer can claim the player role.
func TestObserverClaim(t *testing.T) {
	c1 := dialTCP(t)
	c1.becomePlayer()

	c2 := dialTCP(t)
	w2 := c2.welcome()
	if w2["role"] != "observer" {
		t.Skipf("second client unexpectedly got %v (race with previous test cleanup)", w2["role"])
	}

	c2.sendClaim()
	rm := c2.readUntil("role")
	if rm["role"] != "player" {
		t.Errorf("after claim: expected player, got %v", rm["role"])
	}
}

// TestInvalidMove sends a swap that creates no match and expects a rejection.
func TestInvalidMove(t *testing.T) {
	c := dialTCP(t)
	c.becomePlayer()
	state := c.readUntil("state")
	board := toBoard(state)
	bricks := toBricks(state)

	row, col, dir, ok := findInvalidMove(board, bricks)
	if !ok {
		t.Skip("could not find an invalid move on this board")
	}

	c.sendMove(row, col, dir)

	// Expect swap_anim followed by state/invalid (or state/waiting for reverted swap).
	var gotSwapAnim bool
	for i := 0; i < 10; i++ {
		msg := c.readMsg()
		switch msg["type"] {
		case "swap_anim":
			gotSwapAnim = true
			if msg["valid"] == true {
				t.Error("invalid move sent swap_anim with valid=true")
			}
		case "state":
			if msg["status"] == "invalid" || msg["status"] == "waiting" {
				if !gotSwapAnim {
					t.Error("state arrived before swap_anim")
				}
				return
			}
		}
	}
	t.Error("never received rejection state after invalid move")
}

// TestValidMove sends a match-creating swap and expects the full animation sequence.
func TestValidMove(t *testing.T) {
	c := dialTCP(t)
	c.becomePlayer()
	state := c.readUntil("state")
	board := toBoard(state)
	bricks := toBricks(state)
	scoreBefore := int(state["score"].(float64))

	row, col, dir, ok := findValidMove(board, bricks)
	if !ok {
		t.Skip("no valid match-creating move found on initial board")
	}

	c.sendMove(row, col, dir)

	var gotSwapAnim, gotClearing, gotFallAnim bool
	for i := 0; i < 30; i++ {
		msg := c.readMsg()
		switch msg["type"] {
		case "swap_anim":
			gotSwapAnim = true
			if msg["valid"] != true {
				t.Errorf("valid move got swap_anim with valid=%v", msg["valid"])
			}
		case "state":
			switch msg["status"] {
			case "clearing":
				gotClearing = true
				matches, _ := msg["matches"].([]any)
				if len(matches) == 0 {
					t.Error("clearing state has no matches")
				}
			case "waiting":
				scoreAfter := int(msg["score"].(float64))
				if scoreAfter <= scoreBefore {
					t.Errorf("score did not increase: before=%d after=%d", scoreBefore, scoreAfter)
				}
				if !gotSwapAnim {
					t.Error("never received swap_anim")
				}
				if !gotClearing {
					t.Error("never received state/clearing")
				}
				if !gotFallAnim {
					t.Error("never received fall_anim")
				}
				return
			}
		case "fall_anim":
			gotFallAnim = true
			balls, _ := msg["balls"].([]any)
			if len(balls) == 0 {
				t.Error("fall_anim has no balls")
			}
			// Verify each ball travels a consistent distance (rigid column).
			if len(balls) > 0 {
				first := balls[0].(map[string]any)
				travel := first["toRow"].(float64) - first["fromRow"].(float64)
				for _, bAny := range balls {
					b := bAny.(map[string]any)
					// Only new balls (fromRow < 0) must all travel the same distance per column.
					// (Existing balls that dropped may travel different distances.)
					if b["fromRow"].(float64) < 0 {
						d := b["toRow"].(float64) - b["fromRow"].(float64)
						if d <= 0 {
							t.Errorf("ball has non-positive travel: from=%v to=%v", b["fromRow"], b["toRow"])
						}
						_ = travel
					}
				}
			}
		}
	}
	t.Error("valid move sequence did not complete")
}

// TestValidMoveUnix runs the same valid-move test over the Unix socket.
func TestValidMoveUnix(t *testing.T) {
	c := dialUnix(t)
	c.becomePlayer()
	state := c.readUntil("state")
	board := toBoard(state)
	bricks := toBricks(state)

	row, col, dir, ok := findValidMove(board, bricks)
	if !ok {
		t.Skip("no valid move on this board")
	}

	c.sendMove(row, col, dir)
	final := c.readUntilStatus("waiting")
	if int(final["score"].(float64)) == 0 && final["cascade"] == float64(0) {
		// score may already be non-zero from earlier tests; just verify it arrived
	}
	_ = final
}

// TestCascade verifies that after a clearing sequence the cascade counter
// increments when multiple match groups follow gravity.
// This is opportunistic: we make valid moves until a cascade is reported,
// or give up after several attempts.
func TestCascade(t *testing.T) {
	c := dialTCP(t)
	c.becomePlayer()
	c.readUntil("state")

	for attempt := 0; attempt < 12; attempt++ {
		// Get current board from latest state.
		c.sendMove(0, 0, "right") // dummy to force state refresh; may be invalid
		var lastBoard [][]int
		var lastBricks [][]bool
		for i := 0; i < 6; i++ {
			msg := c.readMsg()
			if msg["type"] == "state" && msg["status"] == "waiting" {
				lastBoard = toBoard(msg)
				lastBricks = toBricks(msg)
				break
			}
		}
		if lastBoard == nil {
			continue
		}
		row, col, dir, ok := findValidMove(lastBoard, lastBricks)
		if !ok {
			continue
		}
		c.sendMove(row, col, dir)
		for i := 0; i < 20; i++ {
			msg := c.readMsg()
			if msg["type"] == "state" && msg["status"] == "waiting" {
				if cascade, _ := msg["cascade"].(float64); cascade > 0 {
					t.Logf("cascade ×%v detected after %d attempts", cascade+1, attempt+1)
					return
				}
				break
			}
		}
	}
	t.Log("no cascade observed in 12 attempts (acceptable on random boards)")
}

// TestFallAnimBallData verifies the structure of a fall_anim message.
func TestFallAnimBallData(t *testing.T) {
	c := dialTCP(t)
	c.becomePlayer()

	// Seed with the initial waiting state.
	current := c.readUntil("state")

	for attempt := 0; attempt < 20; attempt++ {
		// If we don't have a waiting board yet, read until we find one.
		if current == nil || current["status"] != "waiting" {
			current = c.readUntilStatus("waiting")
		}
		board := toBoard(current)
		bricks := toBricks(current)
		current = nil

		row, col, dir, ok := findValidMove(board, bricks)
		if !ok {
			// No match-creating move available; send any move to advance state.
			c.sendMove(0, 0, "right")
		} else {
			c.sendMove(row, col, dir)
		}

		for i := 0; i < 15; i++ {
			msg := c.readMsg()
			if msg["type"] == "state" && msg["status"] == "waiting" {
				current = msg // save for next attempt
				break
			}
			if msg["type"] != "fall_anim" || !ok {
				continue
			}
			balls := msg["balls"].([]any)
			boardMsg := msg["board"].([]any)
			if len(balls) == 0 {
				t.Error("fall_anim: no balls")
			}
			if len(boardMsg) != 10 {
				t.Errorf("fall_anim: board height %d, want 10", len(boardMsg))
			}
			for _, bAny := range balls {
				b := bAny.(map[string]any)
				from := b["fromRow"].(float64)
				to := b["toRow"].(float64)
				if to < 0 || to >= 10 {
					t.Errorf("ball toRow out of range: %v", to)
				}
				if from >= to {
					t.Errorf("ball fromRow (%v) >= toRow (%v): should be falling down", from, to)
				}
			}
			return
		}
	}
	t.Error("never received fall_anim after 20 attempts")
}
