package main

import (
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

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
	Color color.Color
}

func (b *Ball) Update(balls []*Ball) {
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

    // Collision with other balls
    for _, otherBall := range balls {
        if b == otherBall {
            continue
        }
        dx := b.X - otherBall.X
        dy := b.Y - otherBall.Y
        distance := math.Sqrt(dx*dx + dy*dy)

        if distance < b.Radius+otherBall.Radius {
            // The balls are colliding
            b.VelocityX, otherBall.VelocityX = otherBall.VelocityX, b.VelocityX
            b.VelocityY, otherBall.VelocityY = otherBall.VelocityY, b.VelocityY

            // Move the balls out of collision
            overlap := 0.5 * (distance - b.Radius - otherBall.Radius)
            b.X -= overlap * (dx / distance)
            b.Y -= overlap * (dy / distance)
            otherBall.X += overlap * (dx / distance)
            otherBall.Y += overlap * (dy / distance)
        }
    }
}

func (b *Ball) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(b.X-b.Radius, b.Y-b.Radius)
	circle := ebiten.NewImageFromImage(circleImage(int(b.Radius), b.Color))
	screen.DrawImage(circle, op)
}

func circleImage(radius int, col color.Color) *ebiten.Image {
    diameter := 2 * radius
    img := ebiten.NewImage(diameter, diameter)
    img.Fill(color.Transparent)
    for y := -radius; y < radius; y++ {
        for x := -radius; x < radius; x++ {
            if x*x+y*y <= radius*radius {
                img.Set(x+radius, y+radius, col)
            }
        }
    }
    return img
}

func randomColor() color.RGBA {
    return color.RGBA{
        R: uint8(rand.Intn(256)),
        G: uint8(rand.Intn(256)),
        B: uint8(rand.Intn(256)),
        A: 255,
    }
}

const frameThickness = 5

type Game struct {
	balls []*Ball
}

func (g *Game) Update() error {
	for _, ball := range g.balls {
		ball.Update(g.balls)
	}
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
	for _, ball := range g.balls {
		ball.Draw(screen)
	}

	// Print Hello, World!
	//ebitenutil.DebugPrint(screen, "Hello, World!")
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func main() {
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator

	game := &Game{}

	numberOfBalls := rand.Intn(10) + 1 // Random number of balls between 1 and 10

	for i := 0; i < numberOfBalls; i++ {
		radius := float64(rand.Intn(15) + 5) // Random radius between 5 and 20
		game.balls = append(game.balls, &Ball{
			X:              float64(rand.Intn(320)),
			Y:              float64(rand.Intn(240)),
			Radius:         radius,
			VelocityX:      rand.Float64()*4 - 2, // Random velocity between -2 and 2
			VelocityY:      rand.Float64()*4 - 2,
			BoundaryWidth:  320,
			BoundaryHeight: 240,
			Color: randomColor(),
		})
	}

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Bouncing Ball in Bounding Box")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

