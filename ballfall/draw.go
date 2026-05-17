//go:build !nogui

package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	hudH    = 40  // pixels reserved at the top for score + mode buttons
	ballPad = 5
	winW    = GridW * cellSize
	winH    = GridH*cellSize + hudH

	// Mode button layout (right-aligned in the HUD).
	btnH    = 26
	btnY    = 7
	btnPad  = 6  // horizontal padding inside button
)

// modeButtonRects returns the pixel rects for the Demo and Ball Attack buttons.
// Returns (demoX, attackX, btnW_demo, btnW_attack) — all buttons share btnH/btnY.
func modeButtonBounds() (demoX, attackX, demoW, attackW float32) {
	attackW = 100
	demoW   = 58
	attackX = float32(winW) - attackW - 4
	demoX   = attackX - demoW - 6
	return
}

var palette = [5]color.RGBA{
	{},
	{220, 60, 60, 255},
	{60, 200, 80, 255},
	{60, 120, 220, 255},
	{230, 200, 50, 255},
}

// EbitenGame wraps Game and implements the full ebiten.Game interface.
type EbitenGame struct{ *Game }

func (e *EbitenGame) Draw(screen *ebiten.Image) {
	g := e.Game
	screen.Fill(color.RGBA{22, 22, 30, 255})

	g.mu.Lock()
	snapshot := g.grid.Snapshot()
	score    := g.score
	mode     := g.mode
	state    := g.state
	g.mu.Unlock()

	// ── HUD ──────────────────────────────────────────────────────────────────
	label := fmt.Sprintf("Score: %d", score)
	if g.cascade > 0 {
		label += fmt.Sprintf("  Cascade ×%d", g.cascade+1)
	}
	ebitenutil.DebugPrintAt(screen, label, 6, 10)

	demoX, attackX, demoW, attackW := modeButtonBounds()
	drawModeBtn(screen, demoX, float32(btnY), demoW, btnH, "Demo", mode == modeDemo)
	drawModeBtn(screen, attackX, float32(btnY), attackW, btnH, "Ball Attack", mode == modeAttack)

	// ── Grid ─────────────────────────────────────────────────────────────────
	flash := state == stateClearing && (g.timer/5)%2 == 0

	fallDest := map[Pos]bool{}
	for _, f := range g.falling {
		fallDest[Pos{int(f.ToRow + 0.5), f.Col}] = true
	}

	swapHide := map[Pos]bool{}
	if state == stateSwapping || state == stateSwapBack {
		swapHide[g.swapA] = true
		swapHide[g.swapB] = true
	}

	for r := 0; r < GridH; r++ {
		for c := 0; c < GridW; c++ {
			x := float32(c * cellSize)
			y := float32(hudH + r*cellSize)
			vector.DrawFilledRect(screen, x+2, y+2, cellSize-4, cellSize-4,
				color.RGBA{38, 40, 52, 255}, false)

			if swapHide[Pos{r, c}] || fallDest[Pos{r, c}] {
				continue
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
			} else if g.hoverValid && g.hoverCell == p && state == stateWaiting {
				drawBallHighlight(screen, x, y, palette[clr])
			} else {
				drawBall(screen, x, y, palette[clr])
			}
		}
	}

	// Swap animation overlay.
	if state == stateSwapping || state == stateSwapBack {
		t := smoothstep(g.swapProgress)
		aR := float32(hudH) + float32(g.swapA.R)*cellSize
		aC := float32(g.swapA.C) * cellSize
		bR := float32(hudH) + float32(g.swapB.R)*cellSize
		bC := float32(g.swapB.C) * cellSize
		drawBall(screen, aC+float32(t)*(bC-aC), aR+float32(t)*(bR-aR), palette[g.swapColorA])
		drawBall(screen, bC+float32(t)*(aC-bC), bR+float32(t)*(aR-bR), palette[g.swapColorB])
	}

	// Fall animation overlay.
	for _, f := range g.falling {
		drawBall(screen, float32(f.Col*cellSize), float32(hudH)+float32(f.cur*cellSize), palette[f.Color])
	}

	// ── Game-over overlay (Ball Attack only) ──────────────────────────────────
	if state == stateGameOver {
		// Semi-transparent backdrop over the grid.
		vector.DrawFilledRect(screen, 0, float32(hudH), float32(winW), float32(GridH*cellSize),
			color.RGBA{0, 0, 0, 180}, false)

		cx := winW / 2
		cy := hudH + GridH*cellSize/2

		ebitenutil.DebugPrintAt(screen, "GAME  OVER", cx-30, cy-30)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d", score), cx-24, cy-14)
		ebitenutil.DebugPrintAt(screen, "Click or Space to play again", cx-84, cy+4)
	}
}

func (e *EbitenGame) Layout(_, _ int) (int, int) { return winW, winH }

func drawModeBtn(screen *ebiten.Image, x, y, w, h float32, label string, active bool) {
	bg := color.RGBA{40, 44, 60, 255}
	fg := color.RGBA{130, 140, 170, 255}
	if active {
		bg = color.RGBA{50, 100, 180, 255}
		fg = color.RGBA{220, 235, 255, 255}
	}
	vector.DrawFilledRect(screen, x, y, w, h, bg, false)
	ebitenutil.DebugPrintAt(screen, label, int(x)+int(btnPad), int(y)+int((h-13)/2))
	_ = fg
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
	vector.DrawFilledCircle(screen, cx, cy, r+4, color.RGBA{255, 255, 255, 60}, true)
	vector.DrawFilledCircle(screen, cx+2, cy+3, r, color.RGBA{0, 0, 0, 60}, true)
	vector.DrawFilledCircle(screen, cx, cy, r, c, true)
	vector.DrawFilledCircle(screen, cx-r/3, cy-r/3, r/4, color.RGBA{255, 255, 255, 110}, true)
}
