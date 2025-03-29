package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	ballX, ballY       float64
	ballRadius         float64
	ballVelocityX      float64
	ballVelocityY      float64
	frameThickness     float64
}

func (g *Game) Update() error {
	g.ballX += g.ballVelocityX
	g.ballY += g.ballVelocityY

	// Collision with left or right
	if g.ballX-g.ballRadius < g.frameThickness || g.ballX+g.ballRadius > 320-g.frameThickness {
		g.ballVelocityX = -g.ballVelocityX
	}

	// Collision with top or bottom
	if g.ballY-g.ballRadius < g.frameThickness || g.ballY+g.ballRadius > 240-g.frameThickness {
		g.ballVelocityY = -g.ballVelocityY
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Draw bounding box (frame)
	frameColor := color.RGBA{255, 0, 0, 255} // Red color
	ebitenutil.DrawRect(screen, 0, 0, 320, g.frameThickness, frameColor)
	ebitenutil.DrawRect(screen, 0, 240-g.frameThickness, 320, g.frameThickness, frameColor)
	ebitenutil.DrawRect(screen, 0, 0, g.frameThickness, 240, frameColor)
	ebitenutil.DrawRect(screen, 320-g.frameThickness, 0, g.frameThickness, 240, frameColor)

	// Draw ball
	ballColor := color.RGBA{0, 0, 255, 255} // Blue color
	ebitenutil.DrawRect(screen, g.ballX-g.ballRadius, g.ballY-g.ballRadius, 2*g.ballRadius, 2*g.ballRadius, ballColor)

	// Print Hello, World!
	ebitenutil.DebugPrint(screen, "Hello, World!")
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func main() {
	game := &Game{
		ballX:          160,
		ballY:          120,
		ballRadius:     10,
		ballVelocityX:  2,
		ballVelocityY:  2,
		frameThickness: 5,
	}

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Bouncing Ball in Bounding Box")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

