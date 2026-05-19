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

	// Dragged ball: suppress from grid and draw lifted at cursor instead.
	isDragging := g.mouseDown && g.dragStartPx[1] >= hudH && state != stateGameOver
	dragCell := g.dragStartCell

	for r := 0; r < GridH; r++ {
		for c := 0; c < GridW; c++ {
			x := float32(c * cellSize)
			y := float32(hudH + r*cellSize)
			vector.DrawFilledRect(screen, x+2, y+2, cellSize-4, cellSize-4,
				color.RGBA{38, 40, 52, 255}, false)

			if g.grid.Bricks[r][c] {
				drawBrick(screen, x, y)
				continue
			}
			if swapHide[Pos{r, c}] || fallDest[Pos{r, c}] {
				continue
			}
			if isDragging && (Pos{r, c}) == dragCell {
				// Ghost ring where the ball was lifted from.
				cx := x + cellSize/2
				cy := y + cellSize/2
				vector.StrokeCircle(screen, cx, cy, float32(cellSize/2-ballPad), 2,
					color.RGBA{255, 255, 255, 70}, true)
				continue
			}
			clr := snapshot[r][c]
			if clr == ColorNone {
				continue
			}
			p := Pos{r, c}
			if g.matchMS != nil && g.matchMS[p] {
				if flash {
					drawBallFlash(screen, x, y, palette[clr])
				} else {
					drawBall(screen, x, y, palette[clr])
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

	// ── Drag visual ───────────────────────────────────────────────────────────
	if isDragging && dragCell.R >= 0 && dragCell.R < GridH && dragCell.C >= 0 && dragCell.C < GridW {
		clr := snapshot[dragCell.R][dragCell.C]
		if clr != ColorNone {
			ox := float64(dragCell.C)*cellSize + cellSize/2
			oy := float64(hudH+dragCell.R*cellSize) + cellSize/2
			ddx := g.dragCurPx[0] - ox
			ddy := g.dragCurPx[1] - oy
			adx, ady := abs64(ddx), abs64(ddy)

			// Destination cell highlight once direction is clear.
			if adx > cellSize*0.15 || ady > cellSize*0.15 {
				var dr, dc int
				if adx >= ady {
					if ddx > 0 {
						dc = 1
					} else {
						dc = -1
					}
				} else {
					if ddy > 0 {
						dr = 1
					} else {
						dr = -1
					}
				}
				tr, tc := dragCell.R+dr, dragCell.C+dc
				if tr >= 0 && tr < GridH && tc >= 0 && tc < GridW {
					tx := float32(tc)*cellSize + cellSize/2
					ty := float32(hudH+tr*cellSize) + cellSize/2
					vector.StrokeCircle(screen, tx, ty, float32(cellSize/2-ballPad+5), 3,
						color.RGBA{255, 255, 255, 115}, true)
				}
			}

			// Lifted ball clamped to one cell radius.
			maxOff := float64(cellSize)
			clamp := func(v, lo, hi float64) float64 {
				if v < lo {
					return lo
				}
				if v > hi {
					return hi
				}
				return v
			}
			bx := float32(ox+clamp(ddx, -maxOff, maxOff)) - cellSize/2
			by := float32(oy+clamp(ddy, -maxOff, maxOff)) - cellSize/2
			drawBallLifted(screen, bx, by, palette[clr])
		}
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

// Update handles only input on the ebiten goroutine. Game logic ticks via the
// separate time.Ticker goroutine started in runGUI, so the game runs at a
// consistent 60 Hz regardless of window visibility or display refresh rate.
func (e *EbitenGame) Update() error {
	e.handleMouse()
	return nil
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

func drawBrick(screen *ebiten.Image, x, y float32) {
	const pad = 3
	// Base fill
	vector.DrawFilledRect(screen, x+pad, y+pad, cellSize-2*pad, cellSize-2*pad,
		color.RGBA{95, 70, 45, 255}, false)
	// Upper brick face
	const fpad = 5
	half := float32(cellSize-2*fpad) / 2
	vector.DrawFilledRect(screen, x+fpad, y+fpad, cellSize-2*fpad, half-1,
		color.RGBA{120, 90, 58, 255}, false)
	// Lower brick face
	vector.DrawFilledRect(screen, x+fpad, y+fpad+half+1, cellSize-2*fpad, half-1,
		color.RGBA{120, 90, 58, 255}, false)
	// Horizontal mortar line
	vector.DrawFilledRect(screen, x+pad, y+fpad+half-1, cellSize-2*pad, 2,
		color.RGBA{55, 38, 22, 255}, false)
	// Vertical mortar — offset on each row to look like staggered brickwork
	mid := x + cellSize/2
	vector.DrawFilledRect(screen, mid-1, y+fpad, 2, half-1,
		color.RGBA{55, 38, 22, 255}, false)
	vector.DrawFilledRect(screen, x+cellSize/4-1, y+fpad+half+1, 2, half-1,
		color.RGBA{55, 38, 22, 255}, false)
	vector.DrawFilledRect(screen, x+3*cellSize/4-1, y+fpad+half+1, 2, half-1,
		color.RGBA{55, 38, 22, 255}, false)
	// Top highlight
	vector.DrawFilledRect(screen, x+fpad, y+fpad+1, cellSize-2*fpad, 2,
		color.RGBA{150, 115, 78, 255}, false)
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

// drawBallLifted renders a ball being dragged: larger shadow implies elevation.
func drawBallLifted(screen *ebiten.Image, x, y float32, c color.RGBA) {
	cx := x + cellSize/2
	cy := y + cellSize/2
	r := float32(cellSize/2 - ballPad)
	vector.DrawFilledCircle(screen, cx+4, cy+7, r, color.RGBA{0, 0, 0, 115}, true)
	vector.DrawFilledCircle(screen, cx, cy, r, c, true)
	vector.DrawFilledCircle(screen, cx-r/3, cy-r/3, r/4, color.RGBA{255, 255, 255, 110}, true)
}

// drawBallFlash renders a ball during the clearing flash: the original colour
// shows through a semi-transparent white overlay, and a wide outer glow ring
// makes the flash pop without losing colour identity.
func drawBallFlash(screen *ebiten.Image, x, y float32, c color.RGBA) {
	cx := x + cellSize/2
	cy := y + cellSize/2
	r := float32(cellSize/2 - ballPad)
	// Outer glow ring (drawn first so ball sits on top).
	vector.DrawFilledCircle(screen, cx, cy, r+6, color.RGBA{255, 255, 255, 110}, true)
	// Drop shadow.
	vector.DrawFilledCircle(screen, cx+2, cy+3, r, color.RGBA{0, 0, 0, 60}, true)
	// Original colour base.
	vector.DrawFilledCircle(screen, cx, cy, r, c, true)
	// White overlay — partial so the colour bleeds through.
	vector.DrawFilledCircle(screen, cx, cy, r, color.RGBA{255, 255, 255, 155}, true)
	// Bright specular highlight.
	vector.DrawFilledCircle(screen, cx-r/3, cy-r/3, r/3, color.RGBA{255, 255, 255, 210}, true)
}
