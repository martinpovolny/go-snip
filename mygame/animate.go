package main

import (
    "bytes"
    _ "embed"
    "image"
    "log"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

//go:embed assets/Run.png
var spriteSheetData []byte

var spriteSheet *ebiten.Image

const (
    screenWidth  = 800
    screenHeight = 600
    frameWidth   = 128
    frameHeight  = 128
)

func loadSpriteSheet() {
    img, _, err := ebitenutil.NewImageFromReader(bytes.NewReader(spriteSheetData))
    if err != nil {
        log.Fatalf("failed to load embedded sprite sheet: %v", err)
    }
    spriteSheet = img
}

type Animation struct {
    Frames       []*ebiten.Image
    currentFrame int
    tickCount    int
    ticksPerFrame int
}

func NewAnimation(sheet *ebiten.Image, frameCount, frameWidth, frameHeight, ticksPerFrame int) *Animation {
    frames := make([]*ebiten.Image, frameCount)
    for i := 0; i < frameCount; i++ {
        x := i * frameWidth
        frames[i] = sheet.SubImage(image.Rect(x, 0, x+frameWidth, frameHeight)).(*ebiten.Image)
    }
    return &Animation{Frames: frames, ticksPerFrame: ticksPerFrame}
}

func (a *Animation) Update() {
    a.tickCount++
    if a.tickCount >= a.ticksPerFrame {
        a.currentFrame = (a.currentFrame + 1) % len(a.Frames)
        a.tickCount = 0
    }
}

func (a *Animation) CurrentFrame() *ebiten.Image {
    return a.Frames[a.currentFrame]
}

type Game struct {
    playerAnim *Animation
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) Update() error {
    g.playerAnim.Update()
    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    currentFrame := g.playerAnim.CurrentFrame()
    op := &ebiten.DrawImageOptions{}
    op.GeoM.Translate((screenWidth-frameWidth)/2, (screenHeight-frameHeight)/2)
    screen.DrawImage(currentFrame, op)
}

func main() {
    ebiten.SetWindowSize(screenWidth, screenHeight)
    ebiten.SetWindowTitle("Sprite Animation")
    loadSpriteSheet()
    game := &Game{
        playerAnim: NewAnimation(spriteSheet, 8, frameWidth, frameHeight, 5),
    }

    if err := ebiten.RunGame(game); err != nil {
        log.Fatal(err)
    }
}
