//go:build !nogui

package main

import (
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func runGUI(game *Game) {
	ebiten.SetWindowSize(winW, winH)
	ebiten.SetWindowTitle("Ball Fall — match 3 to score!")
	ebiten.SetRunnableOnUnfocused(true)

	// Game logic runs on its own fixed-rate ticker, completely independent of
	// the display refresh rate. This guarantees consistent game speed even when
	// the ebiten window is hidden behind other windows (macOS throttles VSync
	// for occluded windows, which would otherwise slow the game down).
	go func() {
		t := time.NewTicker(time.Second / 60)
		defer t.Stop()
		for range t.C {
			game.LogicTick()
		}
	}()

	if err := ebiten.RunGame(&EbitenGame{game}); err != nil {
		log.Fatal(err)
	}
}
