package main

import (
	"log"

	"github.com/bramca/Far-West"

	"github.com/hajimehoshi/ebiten/v2"
)

// TODO: ideas
// 1. add more temporary rewards
// 2. add more interesting enemies
func main() {
	game := farwest.NewGame()
	// Sepcify the window size as you like. Here, a doulbed size is specified.
	ebiten.SetWindowSize(farwest.ScreenWidth, farwest.ScreenHeight)
	ebiten.SetWindowTitle("Far West")
	ebiten.SetCursorMode(ebiten.CursorModeHidden)


	// Call ebiten.RunGame to start your game loop.
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
