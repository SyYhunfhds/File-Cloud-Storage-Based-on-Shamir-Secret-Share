package main

import (
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

type Game struct{}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Ebitengine is working!")
	ebitenutil.DebugPrintAt(screen, "Testing Ebitengine setup...", 10, 20)
	ebitenutil.DebugPrintAt(screen, "If you see this window, Ebitengine is installed correctly.", 10, 40)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ebitengine Test")

	fmt.Println("Testing Ebitengine...")
	fmt.Println("If a window opens with 'Ebitengine is working!', then Ebitengine is installed correctly.")
	fmt.Println("Press ESC to close the window.")

	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
