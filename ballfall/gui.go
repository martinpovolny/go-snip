//go:build !nogui

package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func runGUI(game *Game) {
	ebiten.SetWindowSize(winW, winH)
	ebiten.SetWindowTitle("Ball Fall — match 3 to score!")
	// Game logic drives web clients too; keep ticking at full rate even when
	// the ebiten window is behind another window or minimized.
	ebiten.SetRunnableOnUnfocused(true)
	if err := ebiten.RunGame(&EbitenGame{game}); err != nil {
		log.Fatal(err)
	}
}
