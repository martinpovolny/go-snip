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
	btnH   = 26
	btnY   = 7
	btnPad = 6 // horizontal padding inside button

	// Versus layout: two boards side by side with a gap.
	versusGap  = 8
	versusWinW = winW*2 + versusGap
	versusWinH = winH
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
// ms holds GUI-only input state that has no place in the game logic.
type EbitenGame struct {
	*Game
	ms mouseState
}

// boardDrawData holds everything required to draw one board, collected under lock.
type boardDrawData struct {
	snapshot       [][]int
	bricks         [GridH][GridW]bool
	score          int
	mode           gameMode
	state          gameState
	timer          int
	cascade        int
	matchMS        map[Pos]bool
	falling        []activeFall
	swapA, swapB   Pos
	swapColorA     int
	swapColorB     int
	swapProgress   float64
	versusWinner   bool // true = this board won (opponent's column filled)
}

func collectBoardData(g *Game) boardDrawData {
	g.mu.Lock()
	d := boardDrawData{
		snapshot:     g.grid.Snapshot(),
		bricks:       g.grid.Bricks,
		score:        g.score,
		mode:         g.mode,
		state:        g.state,
		timer:        g.timer,
		cascade:      g.cascade,
		matchMS:      g.matchMS,
		falling:      append([]activeFall(nil), g.falling...),
		swapA:        g.swapA,
		swapB:        g.swapB,
		swapColorA:   g.swapColorA,
		swapColorB:   g.swapColorB,
		swapProgress: g.swapProgress,
		versusWinner: g.versusWinner,
	}
	g.mu.Unlock()
	return d
}

// drawSoloHUD draws the score label and Demo / Ball Attack mode buttons.
func drawSoloHUD(screen *ebiten.Image, d boardDrawData) {
	label := fmt.Sprintf("Score: %d", d.score)
	if d.cascade > 0 {
		label += fmt.Sprintf("  Cascade ×%d", d.cascade+1)
	}
	ebitenutil.DebugPrintAt(screen, label, 6, 10)

	demoX, attackX, demoW, attackW := modeButtonBounds()
	drawModeBtn(screen, demoX, float32(btnY), demoW, btnH, "Demo", d.mode == modeDemo)
	drawModeBtn(screen, attackX, float32(btnY), attackW, btnH, "Ball Attack", d.mode == modeAttack)
}

// drawVersusHUD draws a minimal HUD for one versus board at the given x offset.
func drawVersusHUD(screen *ebiten.Image, d boardDrawData, x0 float32, label string) {
	scoreStr := fmt.Sprintf("%s  Score: %d", label, d.score)
	if d.cascade > 0 {
		scoreStr += fmt.Sprintf("  ×%d", d.cascade+1)
	}
	ebitenutil.DebugPrintAt(screen, scoreStr, int(x0)+6, 10)
}

// drawBoardGrid draws the grid cells, swap overlay, fall overlay, drag visual,
// and game-over overlay for one board at the given x offset.
func drawBoardGrid(screen *ebiten.Image, d boardDrawData, ms *mouseState, x0 float32) {
	flash := d.state == stateClearing && (d.timer/5)%2 == 0

	fallDest := map[Pos]bool{}
	for _, f := range d.falling {
		fallDest[Pos{int(f.ToRow + 0.5), f.Col}] = true
	}

	swapHide := map[Pos]bool{}
	if d.state == stateSwapping || d.state == stateSwapBack {
		swapHide[d.swapA] = true
		swapHide[d.swapB] = true
	}

	isDragging := ms != nil && ms.mouseDown && ms.dragStartPx[1] >= hudH && d.state != stateGameOver
	var dragCell Pos
	if ms != nil {
		dragCell = ms.dragStartCell
	}

	for r := 0; r < GridH; r++ {
		for c := 0; c < GridW; c++ {
			x := x0 + float32(c*cellSize)
			y := float32(hudH + r*cellSize)
			vector.DrawFilledRect(screen, x+2, y+2, cellSize-4, cellSize-4,
				color.RGBA{38, 40, 52, 255}, false)

			if d.bricks[r][c] {
				drawBrick(screen, x, y)
				continue
			}
			if swapHide[Pos{r, c}] || fallDest[Pos{r, c}] {
				continue
			}
			if isDragging && (Pos{r, c}) == dragCell {
				cx := x + cellSize/2
				cy := y + cellSize/2
				vector.StrokeCircle(screen, cx, cy, float32(cellSize/2-ballPad), 2,
					color.RGBA{255, 255, 255, 70}, true)
				continue
			}
			clr := d.snapshot[r][c]
			if clr == ColorNone {
				continue
			}
			p := Pos{r, c}
			if d.matchMS != nil && d.matchMS[p] {
				if flash {
					drawBallFlash(screen, x, y, palette[clr])
				} else {
					drawBall(screen, x, y, palette[clr])
				}
			} else if ms != nil && ms.hoverValid && ms.hoverCell == p && d.state == stateWaiting {
				drawBallHighlight(screen, x, y, palette[clr])
			} else {
				drawBall(screen, x, y, palette[clr])
			}
		}
	}

	// Swap animation overlay.
	if d.state == stateSwapping || d.state == stateSwapBack {
		t := smoothstep(d.swapProgress)
		aR := float32(hudH) + float32(d.swapA.R)*cellSize
		aC := x0 + float32(d.swapA.C)*cellSize
		bR := float32(hudH) + float32(d.swapB.R)*cellSize
		bC := x0 + float32(d.swapB.C)*cellSize
		drawBall(screen, aC+float32(t)*(bC-aC), aR+float32(t)*(bR-aR), palette[d.swapColorA])
		drawBall(screen, bC+float32(t)*(aC-bC), bR+float32(t)*(aR-bR), palette[d.swapColorB])
	}

	// Fall animation overlay.
	for _, f := range d.falling {
		drawBall(screen, x0+float32(f.Col*cellSize), float32(hudH)+float32(f.cur*cellSize), palette[f.Color])
	}

	// Drag visual.
	if isDragging && dragCell.R >= 0 && dragCell.R < GridH && dragCell.C >= 0 && dragCell.C < GridW {
		clr := d.snapshot[dragCell.R][dragCell.C]
		if clr != ColorNone {
			ox := float64(dragCell.C)*cellSize + cellSize/2
			oy := float64(hudH+dragCell.R*cellSize) + cellSize/2
			ddx := ms.dragCurPx[0] - float64(x0) - ox
			ddy := ms.dragCurPx[1] - oy
			adx, ady := abs64(ddx), abs64(ddy)

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
					tx := x0 + float32(tc)*cellSize + cellSize/2
					ty := float32(hudH+tr*cellSize) + cellSize/2
					vector.StrokeCircle(screen, tx, ty, float32(cellSize/2-ballPad+5), 3,
						color.RGBA{255, 255, 255, 115}, true)
				}
			}

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
			bx := float32(float64(x0)+ox+clamp(ddx, -maxOff, maxOff)) - cellSize/2
			by := float32(oy+clamp(ddy, -maxOff, maxOff)) - cellSize/2
			drawBallLifted(screen, bx, by, palette[clr])
		}
	}

	// Game-over overlay.
	if d.state == stateGameOver {
		vector.DrawFilledRect(screen, x0, float32(hudH), float32(winW), float32(GridH*cellSize),
			color.RGBA{0, 0, 0, 180}, false)

		cx := int(x0) + winW/2
		cy := hudH + GridH*cellSize/2

		if d.mode == modeVersus {
			if d.versusWinner {
				ebitenutil.DebugPrintAt(screen, "WINNER!", cx-21, cy-14)
			} else {
				ebitenutil.DebugPrintAt(screen, "DEFEATED", cx-24, cy-14)
			}
		} else {
			ebitenutil.DebugPrintAt(screen, "GAME  OVER", cx-30, cy-30)
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d", d.score), cx-24, cy-14)
			ebitenutil.DebugPrintAt(screen, "Click or Space to play again", cx-84, cy+4)
		}
	}
}

func (e *EbitenGame) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{22, 22, 30, 255})
	d := collectBoardData(e.Game)
	drawSoloHUD(screen, d)
	drawBoardGrid(screen, d, &e.ms, 0)
}

func (e *EbitenGame) Layout(_, _ int) (int, int) { return winW, winH }

func drawModeBtn(screen *ebiten.Image, x, y, w, h float32, label string, active bool) {
	bg := color.RGBA{40, 44, 60, 255}
	if active {
		bg = color.RGBA{50, 100, 180, 255}
	}
	vector.DrawFilledRect(screen, x, y, w, h, bg, false)
	ebitenutil.DebugPrintAt(screen, label, int(x)+int(btnPad), int(y)+int((h-13)/2))
}

func drawBrick(screen *ebiten.Image, x, y float32) {
	const pad = 3
	vector.DrawFilledRect(screen, x+pad, y+pad, cellSize-2*pad, cellSize-2*pad,
		color.RGBA{95, 70, 45, 255}, false)
	const fpad = 5
	half := float32(cellSize-2*fpad) / 2
	vector.DrawFilledRect(screen, x+fpad, y+fpad, cellSize-2*fpad, half-1,
		color.RGBA{120, 90, 58, 255}, false)
	vector.DrawFilledRect(screen, x+fpad, y+fpad+half+1, cellSize-2*fpad, half-1,
		color.RGBA{120, 90, 58, 255}, false)
	vector.DrawFilledRect(screen, x+pad, y+fpad+half-1, cellSize-2*pad, 2,
		color.RGBA{55, 38, 22, 255}, false)
	mid := x + cellSize/2
	vector.DrawFilledRect(screen, mid-1, y+fpad, 2, half-1,
		color.RGBA{55, 38, 22, 255}, false)
	vector.DrawFilledRect(screen, x+cellSize/4-1, y+fpad+half+1, 2, half-1,
		color.RGBA{55, 38, 22, 255}, false)
	vector.DrawFilledRect(screen, x+3*cellSize/4-1, y+fpad+half+1, 2, half-1,
		color.RGBA{55, 38, 22, 255}, false)
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

func drawBallLifted(screen *ebiten.Image, x, y float32, c color.RGBA) {
	cx := x + cellSize/2
	cy := y + cellSize/2
	r := float32(cellSize/2 - ballPad)
	vector.DrawFilledCircle(screen, cx+4, cy+7, r, color.RGBA{0, 0, 0, 115}, true)
	vector.DrawFilledCircle(screen, cx, cy, r, c, true)
	vector.DrawFilledCircle(screen, cx-r/3, cy-r/3, r/4, color.RGBA{255, 255, 255, 110}, true)
}

func drawBallFlash(screen *ebiten.Image, x, y float32, c color.RGBA) {
	cx := x + cellSize/2
	cy := y + cellSize/2
	r := float32(cellSize/2 - ballPad)
	vector.DrawFilledCircle(screen, cx, cy, r+6, color.RGBA{255, 255, 255, 110}, true)
	vector.DrawFilledCircle(screen, cx+2, cy+3, r, color.RGBA{0, 0, 0, 60}, true)
	vector.DrawFilledCircle(screen, cx, cy, r, c, true)
	vector.DrawFilledCircle(screen, cx, cy, r, color.RGBA{255, 255, 255, 155}, true)
	vector.DrawFilledCircle(screen, cx-r/3, cy-r/3, r/3, color.RGBA{255, 255, 255, 210}, true)
}
