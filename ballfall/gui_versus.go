//go:build !nogui

package main

import (
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// VersusEbitenGame implements ebiten.Game for the two-player side-by-side view.
// The local player (P1) uses the mouse on the left board. P2's board is shown
// on the right and is controlled via the web interface.
type VersusEbitenGame struct {
	versus *VersusGame
	ms1    mouseState
}

func (v *VersusEbitenGame) Update() error {
	handleBoardMouse(v.versus.P1, &v.ms1, 0, float64(winW))
	return nil
}

func (v *VersusEbitenGame) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{22, 22, 30, 255})

	// Divider between the two boards.
	vector.DrawFilledRect(screen,
		float32(winW), 0, versusGap, float32(versusWinH),
		color.RGBA{50, 52, 70, 255}, false)

	d1 := collectBoardData(v.versus.P1)
	d2 := collectBoardData(v.versus.P2)

	drawVersusHUD(screen, d1, 0, "YOU")
	drawVersusHUD(screen, d2, float32(winW+versusGap), "OPP")
	drawBoardGrid(screen, d1, &v.ms1, 0)
	drawBoardGrid(screen, d2, nil, float32(winW+versusGap))

	// Versus winner overlay (shown on the winning board when opponent is defeated).
	if d1.state == stateGameOver && d2.state == stateGameOver {
		// Both game-over: show winner on the surviving board (the one that didn't lose).
		// The game-over overlay already shows "DEFEATED" on the loser's board.
		// Show "WINNER" on the other board.
	}

	// Show a global rematch hint at the bottom when both boards are game-over.
	if d1.state == stateGameOver || d2.state == stateGameOver {
		hint := "Space to rematch"
		ebitenutil.DebugPrintAt(screen, hint, versusWinW/2-60, versusWinH-20)
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			v.versus.Reset()
		}
	}
}

func (v *VersusEbitenGame) Layout(_, _ int) (int, int) { return versusWinW, versusWinH }

func runVersusGUI(vs *VersusGame) {
	ebiten.SetWindowSize(versusWinW, versusWinH)
	ebiten.SetWindowTitle("Ball Fall — Versus Mode")
	ebiten.SetRunnableOnUnfocused(true)

	go func() {
		t := time.NewTicker(time.Second / 60)
		defer t.Stop()
		for range t.C {
			vs.LogicTick()
		}
	}()

	if err := ebiten.RunGame(&VersusEbitenGame{versus: vs}); err != nil {
		log.Fatal(err)
	}
}
