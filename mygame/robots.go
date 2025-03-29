package main

import (
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth        = 320
	screenHeight       = 240
	playerSize         = 20
	frameThickness     = 10
	poleSize           = 15
	numOfElectricPoles = 5
)

type Player struct {
	X, Y     float64
	Speed    float64
	Color    color.Color
	HasMoved bool
}

type ElectricPole struct {
	X, Y  float64
	Color color.Color
}

func (ep *ElectricPole) Draw(screen *ebiten.Image) {
	poleImg := ebiten.NewImage(poleSize, poleSize)
	poleImg.Fill(ep.Color)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(ep.X, ep.Y)
	screen.DrawImage(poleImg, opts)
}

func (p *Player) Move() {
	if ebiten.IsKeyPressed(ebiten.KeyLeft) && p.X-p.Speed > float64(frameThickness) {
		p.X -= p.Speed
		p.HasMoved = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) && p.X+p.Speed+playerSize < float64(screenWidth-frameThickness) {
		p.X += p.Speed
		p.HasMoved = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) && p.Y-p.Speed > float64(frameThickness) {
		p.Y -= p.Speed
		p.HasMoved = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) && p.Y+p.Speed+playerSize < float64(screenHeight-frameThickness) {
		p.Y += p.Speed
		p.HasMoved = true
	}
}

func (p *Player) Draw(screen *ebiten.Image) {
	playerRect := ebiten.NewImage(playerSize, playerSize)
	playerRect.Fill(p.Color)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(p.X, p.Y)
	screen.DrawImage(playerRect, opts)
}

type Robot struct {
	X, Y  float64
	Speed float64
	Color color.Color
}

func (r *Robot) MoveTowards(targetX, targetY float64) {
	if r.X < targetX {
		r.X += r.Speed
	} else if r.X > targetX {
		r.X -= r.Speed
	}

	if r.Y < targetY {
		r.Y += r.Speed
	} else if r.Y > targetY {
		r.Y -= r.Speed
	}
}

func (r *Robot) Draw(screen *ebiten.Image) {
	robotImg := ebiten.NewImage(poleSize, poleSize) // Using `poleSize` for simplicity, or you can define a new constant
	robotImg.Fill(r.Color)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(r.X, r.Y)
	screen.DrawImage(robotImg, opts)
}

type Game struct {
	player        *Player
	electricPoles []*ElectricPole
	robots        []*Robot
	gameOver      bool
	playerWon     bool
}

func (g *Game) checkPlayerPoleCollision() {
	for _, pole := range g.electricPoles {
		if g.player.X+playerSize > pole.X && g.player.X < pole.X+poleSize &&
			g.player.Y+playerSize > pole.Y && g.player.Y < pole.Y+poleSize {
			g.gameOver = true
			return
		}
	}
}

func (g *Game) Update() error {
	if !g.gameOver && !g.playerWon {
		g.player.Move()

		// Move each robot towards the player
		if g.player.HasMoved {
			for _, robot := range g.robots {

				robot.MoveTowards(g.player.X, g.player.Y)
			}
		}

		g.checkPlayerPoleCollision()

		// Check for collision between robots and the player
		for _, robot := range g.robots {
			if robot.X+poleSize > g.player.X && robot.X < g.player.X+playerSize &&
				robot.Y+poleSize > g.player.Y && robot.Y < g.player.Y+playerSize {
				g.gameOver = true
				return nil
			}
		}

		// Check for collision between robots and electric poles
		for i := 0; i < len(g.robots); i++ {
			robot := g.robots[i]
			for _, pole := range g.electricPoles {
				if robot.X+poleSize > pole.X && robot.X < pole.X+poleSize &&
					robot.Y+poleSize > pole.Y && robot.Y < pole.Y+poleSize {
					// Remove the robot from the slice
					g.robots = append(g.robots[:i], g.robots[i+1:]...)
					i-- // Adjust index after removal
					break
				}
			}
		}

		// Check for win condition
		if len(g.robots) == 0 {
			g.playerWon = true
		}
	}

	return nil
}

func (g *Game) drawBoundingFrame(screen *ebiten.Image) {
	frameColor := color.RGBA{255, 255, 255, 255} // White color
	thickness := 5.0                             // Thickness of the frame

	// Top
	topRect := ebiten.NewImage(screenWidth, int(thickness))
	topRect.Fill(frameColor)
	screen.DrawImage(topRect, &ebiten.DrawImageOptions{})

	// Bottom
	bottomOpts := &ebiten.DrawImageOptions{}
	bottomOpts.GeoM.Translate(0, screenHeight-thickness)
	screen.DrawImage(topRect, bottomOpts) // Reusing the top rectangle

	// Left
	leftRect := ebiten.NewImage(int(thickness), screenHeight)
	leftRect.Fill(frameColor)
	screen.DrawImage(leftRect, &ebiten.DrawImageOptions{})

	// Right
	rightOpts := &ebiten.DrawImageOptions{}
	rightOpts.GeoM.Translate(screenWidth-thickness, 0)
	screen.DrawImage(leftRect, rightOpts) // Reusing the left rectangle
}

func (g *Game) Draw(screen *ebiten.Image) {
	rand.Seed(time.Now().UnixNano())

	// Clear the screen
	screen.Fill(color.RGBA{0, 0, 0, 255})

	// Draw the bounding frame
	g.drawBoundingFrame(screen)

	// Draw the player
	g.player.Draw(screen)

	// Draw the electric poles
	for _, pole := range g.electricPoles {
		pole.Draw(screen)
	}

	// Draw the robots
	for _, robot := range g.robots {
		robot.Draw(screen)
	}

	// If the game is over, display the "Game Over" message
	if g.gameOver {
		ebitenutil.DebugPrint(screen, "\n\n\n\n\n            GAME OVER")
	}

	// If the player has won, display the "You Won" message
	if g.playerWon {
		ebitenutil.DebugPrint(screen, "\n\n\n\n\n            YOU WON")
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	rand.Seed(time.Now().UnixNano())
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Avoid the Poles")

	player := &Player{
		X:     screenWidth / 2,
		Y:     screenHeight / 2,
		Speed: 4.0,
		Color: color.RGBA{255, 0, 0, 255},
	}

	electricPoles := make([]*ElectricPole, numOfElectricPoles)
	for i := range electricPoles {
		electricPoles[i] = &ElectricPole{
			X:     float64(frameThickness + rand.Intn(screenWidth-2*frameThickness-poleSize)),
			Y:     float64(frameThickness + rand.Intn(screenHeight-2*frameThickness-poleSize)),
			Color: color.RGBA{0, 0, 255, 255},
		}
	}

	game := &Game{
		player:        player,
		electricPoles: electricPoles,
	}

	// Initialize Robots
	numRobots := 5 // or any other number you prefer
	game.robots = make([]*Robot, numRobots)
	for i := range game.robots {
		game.robots[i] = &Robot{
			X:     rand.Float64() * (screenWidth - poleSize),
			Y:     rand.Float64() * (screenHeight - poleSize),
			Speed: 1, // or any other speed you prefer
			Color: color.RGBA{ // random color
				R: uint8(rand.Intn(256)),
				G: uint8(rand.Intn(256)),
				B: uint8(rand.Intn(256)),
				A: 255,
			},
		}
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
