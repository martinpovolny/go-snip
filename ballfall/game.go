package main

import (
	"fmt"
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	cellSize   = 60
	ballPad    = 5
	winW       = GridW * cellSize
	winH       = GridH * cellSize
	clearTicks = 30 // frames for the flash-clear animation

	swapFrames = 12  // frames for swap animation
	fallGrav = 0.04 // acceleration in grid-rows per frame²
	fallVMax = 0.6  // max velocity in grid-rows per frame
)

var palette = [5]color.RGBA{
	{},
	{220, 60, 60, 255},
	{60, 200, 80, 255},
	{60, 120, 220, 255},
	{230, 200, 50, 255},
}

type gameState int

const (
	stateWaiting  gameState = iota
	stateSwapping           // two balls animate toward each other
	stateSwapBack           // invalid swap – animate back
	stateClearing           // flash animation before removing matched cells
	stateFalling            // gravity/fill animation
)

type activeFall struct {
	FallingBall
	cur float64 // current row (float, grid units)
	vel float64 // current velocity (rows/frame)
}

// Game is the central game object. It implements ebiten.Game and owns the grid.
type Game struct {
	mu      sync.Mutex
	grid    *Grid
	score   int
	cascade int

	state   gameState
	timer   int
	current []Pos
	matchMS map[Pos]bool

	// swap animation
	swapA, swapB       Pos
	swapColorA, swapColorB int
	swapProgress       float64 // 0→1
	swapValid          bool    // true = will commit, false = will revert

	// fall animation
	falling []activeFall

	// mouse input
	mouseDown     bool
	dragStartCell Pos
	dragStartPx   [2]float64
	hoverCell     Pos
	hoverValid    bool // whether hoverCell is inside grid

	// pending match groups found after gravity; consumed when fall animation ends
	_pendingGroups [][]Pos

	hub *Hub
}

func NewGame(h *Hub) *Game {
	return &Game{grid: NewGrid(), hub: h}
}

// snapshot must be called with g.mu held.
func (g *Game) snapshot(status string) StateMsg {
	msg := StateMsg{
		Type:    "state",
		Board:   g.grid.Snapshot(),
		Width:   GridW,
		Height:  GridH,
		Score:   g.score,
		Status:  status,
		Cascade: g.cascade,
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
	g.hub.Broadcast(msg)
}

func (g *Game) startGroup(group []Pos) {
	g.current = group
	g.matchMS = MatchSet(group)
	g.timer = clearTicks
	g.state = stateClearing
	// broadcast so clients can flash the matched cells
	g.mu.Lock()
	msg := g.snapshot("clearing")
	g.mu.Unlock()
	g.hub.Broadcast(msg)
}

// checkCascade clears the current match group, applies gravity, and either
// starts the next group's clear or begins the fall animation.
func (g *Game) checkCascade() {
	g.mu.Lock()
	g.grid.ClearMatches(g.current)
	g.score += len(g.current)
	falls := g.grid.GravityAndFill()
	groups := g.grid.FindMatchGroups()
	g.mu.Unlock()

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
		active[i] = activeFall{FallingBall: f, cur: f.FromRow, vel: 0}
	}
	g.falling = active
	g._pendingGroups = nextGroups
	g.state = stateFalling

	// broadcast fall animation so web clients can animate
	balls := make([]FallBallData, len(falls))
	for i, f := range falls {
		balls[i] = FallBallData{Color: f.Color, Col: f.Col, FromRow: int(f.FromRow), ToRow: int(f.ToRow)}
	}
	g.mu.Lock()
	board := g.grid.Snapshot()
	score := g.score
	cascade := g.cascade
	g.mu.Unlock()
	g.hub.Broadcast(FallAnimMsg{
		Type:    "fall_anim",
		Board:   board,
		Balls:   balls,
		Score:   score,
		Cascade: cascade,
	})
}

// Update implements ebiten.Game. Runs on the main goroutine at 60 TPS.
func (g *Game) Update() error {
	g.handleMouse()

	switch g.state {
	case stateWaiting:
		select {
		case m := <-g.hub.MoveIn:
			g.initiateSwap(Pos{m.Row, m.Col}, m.Dir)
		default:
		}

	case stateSwapping, stateSwapBack:
		g.swapProgress += 1.0 / swapFrames
		if g.swapProgress >= 1.0 {
			g.swapProgress = 1.0
			if g.state == stateSwapping && g.swapValid {
				// commit: board already has the swap; find matches
				g.mu.Lock()
				groups := g.grid.FindMatchGroups()
				g.mu.Unlock()
				if len(groups) > 0 {
					g.cascade = 0
					g.startGroup(groups[0])
				} else {
					// shouldn't happen (we pre-checked), settle anyway
					g.settle()
				}
			} else {
				// revert: swap cells back in grid
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
	return nil
}

// initiateSwap validates a swap from pos in dir and starts the animation.
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
	g.grid.Swap(from, to)
	groups := g.grid.FindMatchGroups()
	if len(groups) == 0 {
		// leave grid swapped for now; swapBack animation will undo it
		g.swapValid = false
	} else {
		g.swapValid = true
	}
	g.mu.Unlock()

	g.swapA = from
	g.swapB = to
	g.swapColorA = colorA
	g.swapColorB = colorB
	g.swapProgress = 0
	if g.swapValid {
		g.state = stateSwapping
	} else {
		g.state = stateSwapBack
	}
	g.hub.Broadcast(SwapAnimMsg{
		Type:   "swap_anim",
		RowA:   from.R, ColA: from.C,
		RowB:   to.R, ColB: to.C,
		ColorA: colorA, ColorB: colorB,
		Valid:  g.swapValid,
	})
}

// handleMouse reads ebiten mouse state and drives drag-to-swap.
func (g *Game) handleMouse() {
	if g.state != stateWaiting {
		return
	}
	mx, my := ebiten.CursorPosition()
	px, py := float64(mx), float64(my)

	// track hover
	hc := Pos{int(py) / cellSize, int(px) / cellSize}
	if hc.R >= 0 && hc.R < GridH && hc.C >= 0 && hc.C < GridW {
		g.hoverCell = hc
		g.hoverValid = true
	} else {
		g.hoverValid = false
	}

	pressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	if pressed && !g.mouseDown {
		g.mouseDown = true
		g.dragStartPx = [2]float64{px, py}
		g.dragStartCell = hc
	}
	if !pressed && g.mouseDown {
		g.mouseDown = false
		dx := px - g.dragStartPx[0]
		dy := py - g.dragStartPx[1]
		const minDrag = cellSize * 0.35
		if abs64(dx) < minDrag && abs64(dy) < minDrag {
			return // too small a movement
		}
		var dir string
		if abs64(dx) >= abs64(dy) {
			if dx > 0 {
				dir = "right"
			} else {
				dir = "left"
			}
		} else {
			if dy > 0 {
				dir = "down"
			} else {
				dir = "up"
			}
		}
		g.initiateSwap(g.dragStartCell, dir)
	}
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

// Draw implements ebiten.Game.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{22, 22, 30, 255})

	flash := g.state == stateClearing && (g.timer/5)%2 == 0

	g.mu.Lock()
	snapshot := g.grid.Snapshot()
	score := g.score
	g.mu.Unlock()

	// Suppress the grid-cell render for cells where a falling ball will land.
	fallDest := map[Pos]bool{}
	for _, f := range g.falling {
		fallDest[Pos{int(f.ToRow + 0.5), f.Col}] = true
	}

	// Cells hidden during swap animation (we draw them manually at lerped positions).
	swapHide := map[Pos]bool{}
	if g.state == stateSwapping || g.state == stateSwapBack {
		swapHide[g.swapA] = true
		swapHide[g.swapB] = true
	}

	for r := 0; r < GridH; r++ {
		for c := 0; c < GridW; c++ {
			x := float32(c * cellSize)
			y := float32(r * cellSize)
			vector.DrawFilledRect(screen, x+2, y+2, cellSize-4, cellSize-4,
				color.RGBA{38, 40, 52, 255}, false)

			if swapHide[Pos{r, c}] {
				continue
			}
			if fallDest[Pos{r, c}] {
				continue // falling ball will be drawn at its current position
			}

			clr := snapshot[r][c]
			if clr == ColorNone {
				continue
			}
			p := Pos{r, c}
			if g.matchMS != nil && g.matchMS[p] {
				if flash {
					drawBall(screen, x, y, color.RGBA{255, 255, 255, 220})
				}
			} else {
				if g.hoverValid && g.hoverCell == p && g.state == stateWaiting {
					drawBallHighlight(screen, x, y, palette[clr])
				} else {
					drawBall(screen, x, y, palette[clr])
				}
			}
		}
	}

	// Draw swap animation.
	if g.state == stateSwapping || g.state == stateSwapBack {
		t := smoothstep(g.swapProgress)
		aR := float32(g.swapA.R) * cellSize
		aC := float32(g.swapA.C) * cellSize
		bR := float32(g.swapB.R) * cellSize
		bC := float32(g.swapB.C) * cellSize

		// ball A moves from its start toward B's position
		axNow := aC + float32(t)*(bC-aC)
		ayNow := aR + float32(t)*(bR-aR)
		// ball B moves from its start toward A's position
		bxNow := bC + float32(t)*(aC-bC)
		byNow := bR + float32(t)*(aR-bR)

		// during stateSwapping the grid already has A↔B swapped, so colors are inverted
		cA := g.swapColorA
		cB := g.swapColorB
		drawBall(screen, axNow, ayNow, palette[cA])
		drawBall(screen, bxNow, byNow, palette[cB])
	}

	// Draw falling balls.
	for _, f := range g.falling {
		x := float32(f.Col * cellSize)
		y := float32(f.cur * cellSize)
		drawBall(screen, x, y, palette[f.Color])
	}

	label := fmt.Sprintf("Score: %d", score)
	if g.cascade > 0 {
		label += fmt.Sprintf("  Cascade ×%d", g.cascade+1)
	}
	ebitenutil.DebugPrint(screen, label)
}

func drawBall(screen *ebiten.Image, x, y float32, c color.RGBA) {
	cx := x + cellSize/2
	cy := y + cellSize/2
	r := float32(cellSize/2 - ballPad)
	vector.DrawFilledCircle(screen, cx+2, cy+3, r, color.RGBA{0, 0, 0, 60}, true)
	vector.DrawFilledCircle(screen, cx, cy, r, c, true)
	vector.DrawFilledCircle(screen, cx-r/3, cy-r/3, r/4, color.RGBA{255, 255, 255, 110}, true)
}

func drawBallHighlight(screen *ebiten.Image, x, y float32, c color.RGBA) {
	cx := x + cellSize/2
	cy := y + cellSize/2
	r := float32(cellSize/2 - ballPad)
	// outer glow ring
	vector.DrawFilledCircle(screen, cx, cy, r+4, color.RGBA{255, 255, 255, 60}, true)
	vector.DrawFilledCircle(screen, cx+2, cy+3, r, color.RGBA{0, 0, 0, 60}, true)
	vector.DrawFilledCircle(screen, cx, cy, r, c, true)
	vector.DrawFilledCircle(screen, cx-r/3, cy-r/3, r/4, color.RGBA{255, 255, 255, 110}, true)
}

// Layout implements ebiten.Game.
func (g *Game) Layout(_, _ int) (int, int) { return winW, winH }
