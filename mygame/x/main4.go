package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Ball struct {
	X, Y          float64
	Radius        float64
	VelocityX     float64
	VelocityY     float64
	BoundaryWidth int
	BoundaryHeight int
}

func (b *Ball) Update() {
	b.X += b.VelocityX
	b.Y += b.VelocityY

	// Collision with left or right
	if b.X-b.Radius < float64(frameThickness) || b.X+b.Radius > float64(b.BoundaryWidth-frameThickness) {
		b.VelocityX = -b.VelocityX
	}

	// Collision with top or bottom
	if b.Y-b.Radius < float64(frameThickness) || b.Y+b.Radius > float64(b.BoundaryHeight-frameThickness) {
		b.VelocityY = -b.VelocityY
	}
}

func (b *Ball) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(b.X-b.Radius, b.Y-b.Radius)
	circle := ebiten.NewImageFromImage(circleImage(int(b.Radius)))
	screen.DrawImage(circle, op)
}

// Helper function to create a circle image.
func circleImage(radius int) *ebiten.Image {
	diameter := 2 * radius
	img := ebiten.NewImage(diameter, diameter)
	img.Fill(color.RGBA{0, 0, 255, 255})
	mask := ebiten.NewImage(diameter, diameter)
	mask.Fill(color.Transparent)
	op := &ebiten.DrawImageOptions{}
	op.CompositeMode = ebiten.CompositeModeSourceOut
	img.DrawImage(mask, op)
	return img
}

const frameThickness = 5

type Game struct {
	ball *Ball
}

func (g *Game) Update() error {
	g.ball.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Draw bounding box (frame)
	frameColor := color.RGBA{255, 0, 0, 255} // Red color
	ebitenutil.DrawRect(screen, 0, 0, 320, frameThickness, frameColor)
	ebitenutil.DrawRect(screen, 0, 240-frameThickness, 320, frameThickness, frameColor)
	ebitenutil.DrawRect(screen, 0, 0, frameThickness, 240, frameColor)
	ebitenutil.DrawRect(screen, 320-frameThickness, 0, frameThickness, 240, frameColor)

	// Draw ball
	g.ball.Draw(screen)

	// Print Hello, World!
	ebitenutil.DebugPrint(screen, "Hello, World!")
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func main() {
	game := &Game{
		ball: &Ball{
			X:              160,
			Y:              120,
			Radius:         10,
			VelocityX:      2,
			VelocityY:      2,
			BoundaryWidth:  320,
			BoundaryHeight: 240,
		},
	}

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Bouncing Ball in Bounding Box")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

