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
	stateWaiting  gameState = iota // awaiting a player move
	stateClearing                  // animating + clearing current match group
)

// Game is the central game object. It implements ebiten.Game and owns the grid.
type Game struct {
	mu      sync.Mutex // guards grid, score, state visible to socket handlers
	grid    *Grid
	score   int
	cascade int

	state   gameState
	timer   int
	current []Pos      // match group being cleared right now
	matchMS map[Pos]bool // fast lookup of current cleared set (Draw uses this)

	hub *Hub
}

func NewGame(h *Hub) *Game {
	g := &Game{
		grid: NewGrid(),
		hub:  h,
	}
	return g
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

// settle transitions to stateWaiting and broadcasts the stable board.
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

// startGroup begins the clear animation for the next match group.
func (g *Game) startGroup(group []Pos) {
	g.current = group
	g.matchMS = MatchSet(group)
	g.timer = clearTicks
	g.state = stateClearing
}

// Update implements ebiten.Game. Runs on the main goroutine at 60 TPS.
func (g *Game) Update() error {
	switch g.state {
	case stateWaiting:
		select {
		case m := <-g.hub.MoveIn:
			g.handleMove(m)
		default:
		}

	case stateClearing:
		g.timer--
		if g.timer > 0 {
			return nil
		}
		// Timer expired: commit the clear, apply gravity, refill.
		g.mu.Lock()
		g.grid.ClearMatches(g.current)
		g.score += len(g.current)
		g.grid.ApplyGravity()
		g.grid.FillEmpty()
		groups := g.grid.FindMatchGroups()
		g.mu.Unlock()

		if len(groups) > 0 {
			g.cascade++
			g.startGroup(groups[0])
		} else {
			g.settle()
		}
	}
	return nil
}

func (g *Game) handleMove(m MoveMsg) {
	dr, dc, ok := dirToDelta(m.Dir)
	if !ok {
		return
	}
	a := Pos{m.Row, m.Col}
	b := Pos{m.Row + dr, m.Col + dc}
	if b.R < 0 || b.R >= GridH || b.C < 0 || b.C >= GridW {
		return
	}

	g.mu.Lock()
	g.grid.Swap(a, b)
	groups := g.grid.FindMatchGroups()
	if len(groups) == 0 {
		g.grid.Swap(a, b) // revert
		msg := g.snapshot("invalid")
		g.mu.Unlock()
		g.hub.Broadcast(msg)
		return
	}
	g.mu.Unlock()

	g.cascade = 0
	g.startGroup(groups[0])
}

// Draw implements ebiten.Game.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{22, 22, 30, 255})

	flash := g.state == stateClearing && (g.timer/5)%2 == 0

	g.mu.Lock()
	snapshot := g.grid.Snapshot()
	score := g.score
	g.mu.Unlock()

	for r := 0; r < GridH; r++ {
		for c := 0; c < GridW; c++ {
			x := float32(c * cellSize)
			y := float32(r * cellSize)
			// cell background
			vector.DrawFilledRect(screen, x+2, y+2, cellSize-4, cellSize-4,
				color.RGBA{38, 40, 52, 255}, false)

			clr := snapshot[r][c]
			if clr == ColorNone {
				continue
			}
			p := Pos{r, c}
			if g.matchMS != nil && g.matchMS[p] {
				if flash {
					drawBall(screen, x, y, color.RGBA{255, 255, 255, 220})
				}
				// else invisible during off-phase of flash
			} else {
				drawBall(screen, x, y, palette[clr])
			}
		}
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
	// specular highlight
	vector.DrawFilledCircle(screen, cx-r/3, cy-r/3, r/4, color.RGBA{255, 255, 255, 110}, true)
}

// Layout implements ebiten.Game.
func (g *Game) Layout(_, _ int) (int, int) { return winW, winH }
