//go:build !nogui

package main

import "github.com/hajimehoshi/ebiten/v2"

// mouseState holds all GUI-side input state. It lives on EbitenGame, not on
// Game, so the game logic struct stays free of display concerns.
type mouseState struct {
	mouseDown     bool
	dragStartCell Pos
	dragStartPx   [2]float64
	dragCurPx     [2]float64
	hoverCell     Pos
	hoverValid    bool
}

// Update handles input on the ebiten goroutine. Game logic ticks independently
// via the time.Ticker goroutine in gui.go.
func (e *EbitenGame) Update() error {
	e.handleMouse()
	return nil
}

func (e *EbitenGame) handleMouse() {
	g := e.Game
	ms := &e.ms

	mx, my := ebiten.CursorPosition()
	px, py := float64(mx), float64(my)

	pressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	// If game just transitioned to game-over while we were dragging, discard
	// the held drag so the next release doesn't accidentally restart.
	if g.state == stateGameOver && ms.mouseDown {
		ms.mouseDown = false
	}

	// ── Click released: check HUD buttons or game-over overlay ──────────────
	if !pressed && ms.mouseDown {
		ms.mouseDown = false
		dx := px - ms.dragStartPx[0]
		dy := py - ms.dragStartPx[1]

		// HUD button click (small movement = click, not drag).
		if ms.dragStartPx[1] < hudH && abs64(dx) < 8 && abs64(dy) < 8 {
			demoX, attackX, demoW, attackW := modeButtonBounds()
			fx := float32(ms.dragStartPx[0])
			if fx >= demoX && fx < demoX+demoW {
				g.Reset(modeDemo)
				return
			}
			if fx >= attackX && fx < attackX+attackW {
				g.Reset(modeAttack)
				return
			}
			return
		}

		// Game-over overlay click → restart Ball Attack.
		if g.state == stateGameOver && ms.dragStartPx[1] >= hudH {
			g.Reset(modeAttack)
			return
		}

		// Grid drag-to-swap. Queue through hub.MoveIn so the move is held
		// if an animation is in progress and applied when stateWaiting resumes.
		if g.state == stateGameOver {
			return
		}
		const minDrag = cellSize * 0.35
		if abs64(dx) < minDrag && abs64(dy) < minDrag {
			return
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
		select {
		case g.hub.MoveIn <- MoveMsg{Row: ms.dragStartCell.R, Col: ms.dragStartCell.C, Dir: dir}:
		default:
		}
		return
	}

	// ── Press: start tracking ────────────────────────────────────────────────
	if pressed && !ms.mouseDown {
		ms.mouseDown = true
		ms.dragStartPx = [2]float64{px, py}
		ms.dragCurPx = [2]float64{px, py}
		// Grid cell for drag origin (account for HUD offset).
		gridY := py - hudH
		ms.dragStartCell = Pos{int(gridY) / cellSize, int(px) / cellSize}
	}
	if pressed {
		ms.dragCurPx = [2]float64{px, py}
	}

	// ── Hover highlight (grid area only, not while dragging) ─────────────────
	gridY := py - hudH
	hc := Pos{int(gridY) / cellSize, int(px) / cellSize}
	if !pressed && gridY >= 0 && hc.R < GridH && hc.C >= 0 && hc.C < GridW && !g.grid.Bricks[hc.R][hc.C] {
		ms.hoverCell = hc
		ms.hoverValid = true
	} else {
		ms.hoverValid = false
	}

	// ── Space bar: restart Ball Attack from game-over ────────────────────────
	if ebiten.IsKeyPressed(ebiten.KeySpace) && g.state == stateGameOver {
		g.Reset(modeAttack)
	}
}
