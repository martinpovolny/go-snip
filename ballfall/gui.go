//go:build !nogui

package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func runGUI(game *Game) {
	ebiten.SetWindowSize(winW, winH)
	ebiten.SetWindowTitle("Ball Fall — match 3 to score!")
	if err := ebiten.RunGame(&EbitenGame{game}); err != nil {
		log.Fatal(err)
	}
}
