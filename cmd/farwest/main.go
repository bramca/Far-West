package main

import (
	"log"

	farwest "github.com/bramca/Far-West"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	game := farwest.NewGame()

	ebiten.SetWindowSize(farwest.ScreenWidth, farwest.ScreenHeight)
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowTitle("Far West")
	// ebiten.SetCursorMode(ebiten.CursorModeHidden)

	// Call ebiten.RunGame to start your game loop.
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
