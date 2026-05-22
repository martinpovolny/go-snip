package main

import (
	"math/rand"
	"sync"
)

const (
	cellSize   = 60
	clearTicks = 30 // frames for flash-clear animation
	swapFrames = 12 // frames for swap animation
	fallGrav   = 0.04
	fallVMax   = 0.6

	attackDropInterval = 60 // ticks between drops in Ball Attack (1 s at 60 fps)
)

type gameState int

const (
	stateWaiting  gameState = iota
	stateSwapping
	stateSwapBack
	stateClearing
	stateFalling
	stateGameOver
)

type gameMode int

const (
	modeDemo   gameMode = iota
	modeAttack
	modeVersus
)

type activeFall struct {
	FallingBall
	cur float64
	vel float64
}

// Game is the central game object holding all logic state. GUI input state
// (mouse, hover) lives in EbitenGame.ms in mouse_gui.go instead.
type Game struct {
	mu      sync.Mutex
	grid    *Grid
	score   int
	cascade int

	mode  gameMode
	state gameState
	timer int

	current []Pos
	matchMS map[Pos]bool

	swapA, swapB           Pos
	swapColorA, swapColorB int
	swapProgress           float64
	swapValid              bool

	falling        []activeFall
	_pendingGroups [][]Pos

	dropTicks int // Ball Attack / Versus: ticks until next periodic ball drop

	hub    *Hub
	moveIn chan MoveMsg // this game's move input channel (hub.MoveIn or hub.MoveIn2)

	// Versus mode fields.
	playerID     int   // 1 or 2 (0 = solo)
	opponent     *Game // nil in solo modes
	penaltyQueue int   // penalty balls waiting to drop; guarded by mu
}

func NewGame(h *Hub) *Game {
	return &Game{
		grid:      NewGrid(),
		mode:      modeDemo,
		dropTicks: attackDropInterval,
		hub:       h,
		moveIn:    h.MoveIn,
	}
}

// broadcast sends msg to all connected clients, storing it in the per-player
// last-state cache when in versus mode.
func (g *Game) broadcast(msg any) {
	g.hub.BroadcastTagged(msg, g.playerID)
}

// AddPenalty queues n penalty balls to drop on this game's next waiting tick.
// Safe to call from another goroutine (opponent's game loop).
func (g *Game) AddPenalty(n int) {
	if n <= 0 {
		return
	}
	g.mu.Lock()
	g.penaltyQueue += n
	g.mu.Unlock()
}

// BroadcastInitial sends the current board to all connected clients.
// Called once at startup after all servers are listening.
func (g *Game) BroadcastInitial() {
	g.mu.Lock()
	msg := g.snapshot("waiting")
	g.mu.Unlock()
	g.hub.Broadcast(msg)
}

func (g *Game) modeString() string {
	switch g.mode {
	case modeAttack:
		return "attack"
	case modeVersus:
		return "versus"
	default:
		return "demo"
	}
}

// Reset reinitializes the game for the given mode, broadcasting a fresh state.
// Called directly by the GUI mouse handler or via the ModeIn channel.
// Do not call directly in versus mode; use VersusGame.Reset() instead.
func (g *Game) Reset(mode gameMode) {
	g.mu.Lock()
	g.mode = mode
	g.score = 0
	g.cascade = 0
	g.state = stateWaiting
	g.timer = 0
	g.current = nil
	g.matchMS = nil
	g.falling = nil
	g._pendingGroups = nil
	g.swapProgress = 0
	g.dropTicks = attackDropInterval
	g.penaltyQueue = 0

	switch mode {
	case modeAttack, modeVersus:
		g.grid = NewAttackGrid()
	default:
		g.grid = NewGrid()
	}
	msg := g.snapshot("waiting")
	g.mu.Unlock()

	g.hub.SetMode(g.modeString())
	g.broadcast(msg)
}

func (g *Game) snapshot(status string) StateMsg {
	msg := StateMsg{
		Type:    "state",
		Board:   g.grid.Snapshot(),
		Bricks:  g.grid.BrickSnapshot(),
		Width:   GridW,
		Height:  GridH,
		Score:   g.score,
		Status:  status,
		Cascade: g.cascade,
		Mode:    g.modeString(),
		Player:  g.playerID,
	}
	if g.current != nil {
		msg.Matches = MatchesAsArray(g.current)
	}
	return msg
}

func (g *Game) settle() {
	g.cascade = 0
	g.state = stateWaiting
	g.current = nil
	g.matchMS = nil
	g.mu.Lock()
	msg := g.snapshot("waiting")
	g.mu.Unlock()
	g.broadcast(msg)
}

func (g *Game) startGroup(group []Pos) {
	g.current = group
	g.matchMS = MatchSet(group)
	g.timer = clearTicks
	g.state = stateClearing
	g.mu.Lock()
	msg := g.snapshot("clearing")
	g.mu.Unlock()
	g.broadcast(msg)
}

func (g *Game) checkCascade() {
	g.mu.Lock()
	cleared := len(g.current)
	g.grid.ClearMatches(g.current)
	g.score += cleared
	g.grid.DestroyAdjacentBricks(g.current)
	var falls []FallingBall
	if g.mode == modeAttack || g.mode == modeVersus {
		falls = g.grid.GravityOnly()
	} else {
		falls = g.grid.GravityAndFill()
	}
	groups := g.grid.FindMatchGroups()
	g.mu.Unlock()

	// Send penalty balls to the opponent: one ball per 3 cleared.
	if g.opponent != nil && cleared >= 3 {
		g.opponent.AddPenalty(cleared / 3)
	}

	if len(falls) > 0 {
		g.startFalling(falls, groups)
	} else if len(groups) > 0 {
		g.cascade++
		g.startGroup(groups[0])
	} else {
		g.settle()
	}
}

func (g *Game) startFalling(falls []FallingBall, nextGroups [][]Pos) {
	active := make([]activeFall, len(falls))
	for i, f := range falls {
		active[i] = activeFall{FallingBall: f, cur: f.FromRow}
	}
	g.falling = active
	g._pendingGroups = nextGroups
	g.state = stateFalling

	balls := make([]FallBallData, len(falls))
	for i, f := range falls {
		balls[i] = FallBallData{Color: f.Color, Col: f.Col, FromRow: int(f.FromRow), ToRow: int(f.ToRow)}
	}
	g.mu.Lock()
	board := g.grid.Snapshot()
	bricks := g.grid.BrickSnapshot()
	score := g.score
	cascade := g.cascade
	g.mu.Unlock()
	g.broadcast(FallAnimMsg{
		Type:    "fall_anim",
		Board:   board,
		Bricks:  bricks,
		Balls:   balls,
		Score:   score,
		Cascade: cascade,
		Player:  g.playerID,
	})
}

// dropBall is the periodic drop: resets the drop timer then drops at a random column.
func (g *Game) dropBall() {
	g.dropTicks = attackDropInterval
	g.dropBallAt(rand.Intn(GridW))
}

// dropPenaltyBall drops a penalty ball (sent by the opponent) without resetting the drop timer.
func (g *Game) dropPenaltyBall() {
	g.dropBallAt(rand.Intn(GridW))
}

// dropBallAt places a ball in the given column, applies the slide rule if isolated,
// then runs match detection. Used by both periodic and penalty drops.
func (g *Game) dropBallAt(col int) {
	g.mu.Lock()
	// Find landing row: first occupied row from top minus 1.
	landRow := GridH - 1
	for r := 0; r < GridH; r++ {
		if g.grid.Cells[r][col] != ColorNone {
			landRow = r - 1
			break
		}
	}

	if landRow < 0 {
		// Column is full — game over.
		g.state = stateGameOver
		score := g.score
		if g.opponent != nil {
			g.opponent.mu.Lock()
			g.opponent.state = stateGameOver
			g.opponent.mu.Unlock()
		}
		g.mu.Unlock()
		if g.opponent != nil {
			g.broadcast(GameOverMsg{
				Type:   "game_over",
				Score:  score,
				Mode:   "versus",
				Player: g.playerID,
				Winner: g.opponent.playerID,
			})
		} else {
			g.broadcast(GameOverMsg{Type: "game_over", Score: score, Mode: "attack"})
		}
		return
	}

	clr := rand.Intn(numColors) + 1
	g.grid.Cells[landRow][col] = clr

	// Slide rule: if the ball landed on top of another ball with no
	// horizontal neighbors, let it slide into an adjacent column.
	hasBelow := landRow+1 < GridH && g.grid.Cells[landRow+1][col] != ColorNone
	noLeft   := col == 0 || g.grid.Cells[landRow][col-1] == ColorNone
	noRight  := col == GridW-1 || g.grid.Cells[landRow][col+1] == ColorNone

	if hasBelow && noLeft && noRight {
		tryDirs := [2]int{-1, 1}
		if rand.Intn(2) == 0 {
			tryDirs = [2]int{1, -1}
		}
		for _, dc := range tryDirs {
			nc := col + dc
			if nc < 0 || nc >= GridW {
				continue
			}
			// Find lowest empty row in the adjacent column.
			newLand := -1
			for r := GridH - 1; r >= 0; r-- {
				if g.grid.Cells[r][nc] == ColorNone {
					newLand = r
					break
				}
			}
			if newLand >= 0 {
				g.grid.Cells[landRow][col] = ColorNone
				g.grid.Cells[newLand][nc] = clr
				col = nc
				landRow = newLand
				break
			}
		}
	}
	groups := g.grid.FindMatchGroups()
	g.mu.Unlock()

	falls := []FallingBall{{Color: clr, Col: col, FromRow: -1, ToRow: float64(landRow)}}
	g.startFalling(falls, groups)
}

// LogicTick advances the game state machine by one tick. It is called by a
// fixed-rate time.Ticker goroutine in GUI mode and directly from the headless
// ticker loop, keeping game speed independent of the display refresh rate.
func (g *Game) LogicTick() {
	// Mode changes from remote clients take effect immediately (not in versus mode).
	if g.mode != modeVersus {
		select {
		case sm := <-g.hub.ModeIn:
			mode := modeDemo
			if sm.Mode == "attack" {
				mode = modeAttack
			}
			g.Reset(mode)
			return
		default:
		}
	}

	if g.state == stateGameOver {
		return
	}

	switch g.state {
	case stateWaiting:
		// Drain one penalty ball before the regular periodic drop.
		if g.mode == modeVersus {
			g.mu.Lock()
			pq := g.penaltyQueue
			if pq > 0 {
				g.penaltyQueue--
			}
			g.mu.Unlock()
			if pq > 0 {
				g.dropPenaltyBall()
				return
			}
		}

		// Ball Attack / Versus: advance drop timer only while waiting.
		if g.mode == modeAttack || g.mode == modeVersus {
			g.dropTicks--
			if g.dropTicks <= 0 {
				g.dropBall()
				return
			}
		}

		select {
		case m := <-g.moveIn:
			g.initiateSwap(Pos{m.Row, m.Col}, m.Dir)
		default:
		}

	case stateSwapping, stateSwapBack:
		g.swapProgress += 1.0 / swapFrames
		if g.swapProgress >= 1.0 {
			g.swapProgress = 1.0
			if g.state == stateSwapping && g.swapValid {
				g.mu.Lock()
				groups := g.grid.FindMatchGroups()
				g.mu.Unlock()
				if len(groups) > 0 {
					g.cascade = 0
					g.startGroup(groups[0])
				} else {
					g.settle()
				}
			} else {
				g.mu.Lock()
				g.grid.Swap(g.swapA, g.swapB)
				g.mu.Unlock()
				g.settle()
			}
		}

	case stateClearing:
		g.timer--
		if g.timer <= 0 {
			g.checkCascade()
		}

	case stateFalling:
		allDone := true
		for i := range g.falling {
			f := &g.falling[i]
			if f.cur >= f.ToRow {
				continue
			}
			allDone = false
			f.vel += fallGrav
			if f.vel > fallVMax {
				f.vel = fallVMax
			}
			f.cur += f.vel
			if f.cur > f.ToRow {
				f.cur = f.ToRow
			}
		}
		if allDone {
			if len(g._pendingGroups) > 0 {
				g.cascade++
				g.startGroup(g._pendingGroups[0])
			} else {
				g.settle()
			}
			g.falling = nil
			g._pendingGroups = nil
		}
	}
}

// Update is the headless entry point: called by the time.Ticker in main.go.
// In GUI mode, EbitenGame.Update() in mouse_gui.go handles input; game logic
// runs in a separate goroutine via LogicTick.
func (g *Game) Update() error {
	g.LogicTick()
	return nil
}

func (g *Game) initiateSwap(from Pos, dir string) {
	dr, dc, ok := dirToDelta(dir)
	if !ok {
		return
	}
	to := Pos{from.R + dr, from.C + dc}
	if to.R < 0 || to.R >= GridH || to.C < 0 || to.C >= GridW {
		return
	}

	g.mu.Lock()
	colorA := g.grid.Cells[from.R][from.C]
	colorB := g.grid.Cells[to.R][to.C]
	// Can't drag from an empty cell, and moving a ball upward into empty
	// space makes no sense in a gravity game.
	if colorA == ColorNone || (colorB == ColorNone && dr < 0) {
		g.mu.Unlock()
		return
	}
	// Bricks are immovable obstacles.
	if g.grid.Bricks[from.R][from.C] || g.grid.Bricks[to.R][to.C] {
		g.mu.Unlock()
		return
	}
	g.grid.Swap(from, to)
	groups := g.grid.FindMatchGroups()
	g.swapValid = len(groups) > 0
	g.mu.Unlock()

	g.swapA, g.swapB = from, to
	g.swapColorA, g.swapColorB = colorA, colorB
	g.swapProgress = 0
	if g.swapValid {
		g.state = stateSwapping
	} else {
		g.state = stateSwapBack
	}
	g.broadcast(SwapAnimMsg{
		Type:   "swap_anim",
		RowA:   from.R, ColA: from.C,
		RowB:   to.R, ColB: to.C,
		ColorA: colorA, ColorB: colorB,
		Valid:  g.swapValid,
		Player: g.playerID,
	})
}

func abs64(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func smoothstep(t float64) float64 {
	return t * t * (3 - 2*t)
}
