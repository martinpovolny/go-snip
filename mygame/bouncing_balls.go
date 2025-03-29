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

const (
    screenWidth  = 320
    screenHeight = 240
)

type Player struct {
    X, Y float64
    Color color.Color
    Speed float64
}

func (p *Player) Draw(screen *ebiten.Image) {
    arrowSize := 20.0
    x1, y1 := p.X, p.Y
    x2, y2 := p.X-arrowSize/2, p.Y+arrowSize
    x3, y3 := p.X+arrowSize/2, p.Y+arrowSize

    ebitenutil.DrawLine(screen, x1, y1, x2, y2, p.Color)
    ebitenutil.DrawLine(screen, x1, y1, x3, y3, p.Color)
    ebitenutil.DrawLine(screen, x2, y2, x3, y3, p.Color)
}

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
    if b.X-b.Radius < float64(frameThickness) {
        b.VelocityX = math.Abs(b.VelocityX)
        b.X = float64(frameThickness) + b.Radius
    } else if b.X+b.Radius > float64(b.BoundaryWidth-frameThickness) {
        b.VelocityX = -math.Abs(b.VelocityX)
        b.X = float64(b.BoundaryWidth-frameThickness) - b.Radius
    }

    // Collision with top or bottom
    if b.Y-b.Radius < float64(frameThickness) {
        b.VelocityY = math.Abs(b.VelocityY)
        b.Y = float64(frameThickness) + b.Radius
    } else if b.Y+b.Radius > float64(b.BoundaryHeight-frameThickness) {
        b.VelocityY = -math.Abs(b.VelocityY)
        b.Y = float64(b.BoundaryHeight-frameThickness) - b.Radius
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
	player *Player
	gameOver bool
}

func (g *Game) Update() error {
	if g.gameOver {
	    return nil
	}

	for _, ball := range g.balls {
		ball.Update(g.balls)
	}

	// Handle player movement
	if ebiten.IsKeyPressed(ebiten.KeyLeft) && g.player.X-(g.player.Speed+10) > float64(frameThickness) {
	    g.player.X -= g.player.Speed
	}

	if ebiten.IsKeyPressed(ebiten.KeyRight) && g.player.X+(g.player.Speed+10) < float64(screenWidth-frameThickness) {
	    g.player.X += g.player.Speed
	}

	g.checkBallPlayerCollision()

	return nil
}

func (g *Game) checkBallPlayerCollision() {
    for _, ball := range g.balls {
        if ball.X-ball.Radius < g.player.X+10 && ball.X+ball.Radius > g.player.X-10 &&
           ball.Y+ball.Radius > g.player.Y {
            g.gameOver = true
            return
        }
    }
}

func (g *Game) Draw(screen *ebiten.Image) {
	// If the game is over, draw the "GAME OVER" message and return
	if g.gameOver {
	    msg := "GAME OVER"
	    textSize := 20 // adjust as needed
	    x := (screenWidth - textSize*len(msg)/2) / 2
	    y := screenHeight / 2
	    ebitenutil.DebugPrintAt(screen, msg, x, y)
	    return
	}

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

	g.player.Draw(screen)

	// Print Hello, World!
	//ebitenutil.DebugPrint(screen, "Hello, World!")
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
    return screenWidth, screenHeight
}

func main() {
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator

	game := &Game{
		player: &Player{
			X:     320 / 2,                       // Center of the screen width
			Y:     240 - frameThickness - 5,      // Near the bottom frame minus some padding
			Color: color.RGBA{255, 255, 255, 255}, // White color for the player
			Speed: 10,
		},
	}

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

