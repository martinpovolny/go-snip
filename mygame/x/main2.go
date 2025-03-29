package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct{}

func (g *Game) Update() error {
	// Update logic here if any
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Define bounding box (frame) parameters
	frameThickness := 5.0
	frameColor := color.RGBA{255, 0, 0, 255} // Red color

	// Draw top frame
	ebitenutil.DrawRect(screen, 0, 0, 320, frameThickness, frameColor)
	// Draw bottom frame
	ebitenutil.DrawRect(screen, 0, 240-frameThickness, 320, frameThickness, frameColor)
	// Draw left frame
	ebitenutil.DrawRect(screen, 0, 0, frameThickness, 240, frameColor)
	// Draw right frame
	ebitenutil.DrawRect(screen, 320-frameThickness, 0, frameThickness, 240, frameColor)

	// Print Hello, World! as before
	ebitenutil.DebugPrint(screen, "Hello, World!")
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Hello, World with Bounding Box!")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}

