package main

import (
	"log"

	"simulator/internal/algorithm"
	"simulator/internal/ui"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 1280
	screenHeight = 720
	screenTitle  = "Shamir Secret Sharing Simulator"
)

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle(screenTitle)

	config := &algorithm.Config{
		Threshold: 3,
		NumShares: 5,
		Prime:     257,
	}

	game, err := ui.NewGame(config)
	if err != nil {
		log.Fatalf("Failed to create game: %v", err)
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
