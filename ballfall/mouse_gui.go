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
	handleBoardMouse(e.Game, &e.ms, 0, float64(winW))
	return nil
}

// handleBoardMouse processes mouse input for one board.
//   - g: the game whose move channel receives drag-to-swap moves
//   - ms: per-board mouse state
//   - boardX: left pixel edge of this board (0 for P1, winW+gap for P2)
//   - maxX: right pixel edge of this board
func handleBoardMouse(g *Game, ms *mouseState, boardX, maxX float64) {
	mx, my := ebiten.CursorPosition()
	px, py := float64(mx), float64(my)

	pressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	// Ignore events outside this board's column.
	if px < boardX || px >= maxX {
		if ms.mouseDown && !pressed {
			ms.mouseDown = false
		}
		ms.hoverValid = false
		return
	}

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

		// HUD button click (small movement = click, not drag) — solo mode only.
		if g.mode != modeVersus && ms.dragStartPx[1] < hudH && abs64(dx) < 8 && abs64(dy) < 8 {
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

		// Game-over overlay click → restart Ball Attack (solo mode only).
		if g.mode != modeVersus && g.state == stateGameOver && ms.dragStartPx[1] >= hudH {
			g.Reset(modeAttack)
			return
		}

		if g.state == stateGameOver {
			return
		}

		// Grid drag-to-swap.
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
		case g.moveIn <- MoveMsg{Row: ms.dragStartCell.R, Col: ms.dragStartCell.C, Dir: dir}:
		default:
		}
		return
	}

	// ── Press: start tracking ────────────────────────────────────────────────
	if pressed && !ms.mouseDown {
		ms.mouseDown = true
		ms.dragStartPx = [2]float64{px, py}
		ms.dragCurPx   = [2]float64{px, py}
		gridY := py - hudH
		gridX := px - boardX
		ms.dragStartCell = Pos{int(gridY) / cellSize, int(gridX) / cellSize}
	}
	if pressed {
		ms.dragCurPx = [2]float64{px, py}
	}

	// ── Hover highlight (grid area only, not while dragging) ─────────────────
	gridY := py - hudH
	gridX := px - boardX
	hc := Pos{int(gridY) / cellSize, int(gridX) / cellSize}
	if !pressed && gridY >= 0 && hc.R < GridH && hc.C >= 0 && hc.C < GridW && !g.grid.Bricks[hc.R][hc.C] {
		ms.hoverCell  = hc
		ms.hoverValid = true
	} else {
		ms.hoverValid = false
	}

	// ── Space bar: restart Ball Attack from game-over (solo mode only) ────────
	if g.mode != modeVersus && ebiten.IsKeyPressed(ebiten.KeySpace) && g.state == stateGameOver {
		g.Reset(modeAttack)
	}
}
