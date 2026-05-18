//go:build !nogui

package main

import "github.com/hajimehoshi/ebiten/v2"

func (g *Game) handleMouse() {
	mx, my := ebiten.CursorPosition()
	px, py := float64(mx), float64(my)

	pressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	// ── Click released: check HUD buttons or game-over overlay ──────────────
	if !pressed && g.mouseDown {
		g.mouseDown = false
		dx := px - g.dragStartPx[0]
		dy := py - g.dragStartPx[1]

		// HUD button click (small movement = click, not drag).
		if g.dragStartPx[1] < hudH && abs64(dx) < 8 && abs64(dy) < 8 {
			demoX, attackX, demoW, attackW := modeButtonBounds()
			fx := float32(g.dragStartPx[0])
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
		if g.state == stateGameOver && g.dragStartPx[1] >= hudH {
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
		case g.hub.MoveIn <- MoveMsg{Row: g.dragStartCell.R, Col: g.dragStartCell.C, Dir: dir}:
		default:
		}
		return
	}

	// ── Press: start tracking ────────────────────────────────────────────────
	if pressed && !g.mouseDown {
		g.mouseDown = true
		g.dragStartPx = [2]float64{px, py}
		// Grid cell for drag origin (account for HUD offset).
		gridY := py - hudH
		g.dragStartCell = Pos{int(gridY) / cellSize, int(px) / cellSize}
	}

	// ── Hover highlight (grid area only) ────────────────────────────────────
	gridY := py - hudH
	hc := Pos{int(gridY) / cellSize, int(px) / cellSize}
	if gridY >= 0 && hc.R < GridH && hc.C >= 0 && hc.C < GridW && !g.grid.Bricks[hc.R][hc.C] {
		g.hoverCell = hc
		g.hoverValid = true
	} else {
		g.hoverValid = false
	}

	// ── Space bar: restart Ball Attack from game-over ────────────────────────
	if ebiten.IsKeyPressed(ebiten.KeySpace) && g.state == stateGameOver {
		g.Reset(modeAttack)
	}
}
