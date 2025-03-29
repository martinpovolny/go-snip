package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 320
	screenHeight = 240
	playerSize   = 20
)

type Player struct {
	X, Y  float64
	Speed float64
	Color color.Color
}

func (p *Player) Move() {
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		p.X -= p.Speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		p.X += p.Speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		p.Y -= p.Speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		p.Y += p.Speed
	}
}

func (p *Player) Draw(screen *ebiten.Image) {
	playerRect := ebiten.NewImage(playerSize, playerSize)
	playerRect.Fill(p.Color)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(p.X, p.Y)
	screen.DrawImage(playerRect, opts)
}

type Game struct {
	player *Player
}

func (g *Game) Update() error {
	g.player.Move()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.player.Draw(screen)
	ebitenutil.DebugPrint(screen, "Move the player with arrow keys!")
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Top View Game")

	player := &Player{
		X:     screenWidth / 2,
		Y:     screenHeight / 2,
		Speed: 4.0,
		Color: color.RGBA{255, 0, 0, 255},
	}

	game := &Game{
		player: player,
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

