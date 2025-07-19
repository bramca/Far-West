package farwest

import (
	"embed"
	"image/color"

	"github.com/bramca/Far-West/actors"
	"github.com/bramca/Far-West/helpers"
	"github.com/bramca/Far-West/world"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type Mode int

const (
	ModeTitle Mode = iota
	ModeGame
	ModeGameOver
	ModePause
)

const (
	ScreenWidth  = 1280
	ScreenHeight = 860
)

//go:embed assets/*
var assets embed.FS

// Game implements ebiten.Game interface.
type Game struct {
	mode Mode

	assets embed.FS

	titleTexts      []string
	titleTextsExtra []string
	gameOverTexts   []string
	pauseTexts      []string

	fontSize            int
	titleFontSize       int
	titleFontColorScale ebiten.ColorScale

	titleArcadeFont font.Face
	arcadeFont      font.Face

	backgroundColor color.RGBA

	camX float64
	camY float64

	// text geo matrices
	titleGeoMatrix      ebiten.GeoM
	titleExtraGeoMatrix ebiten.GeoM
	gameOverGeoMatrix   ebiten.GeoM
	pauseGeoMatrix      ebiten.GeoM

	// text padding
	newlinePadding int

	framesPerSecond int

	// draw options
	titleDrawOptions          *text.DrawOptions
	titleTextExtraDrawOptions *text.DrawOptions
	gameOverDrawOptions       *text.DrawOptions
	pauseDrawOptions          *text.DrawOptions

	// actors
	player *actors.Player

	// world
	recticle      *world.Recticle
	cactusSprites []*ebiten.Image
	cacti         []*world.Cactus

	// gameplay
	frameCount   int
	maxFramCount int
}

func NewGame() *Game {
	game := &Game{
		titleTexts:      []string{"FAR WEST"},
		titleTextsExtra: []string{"PRESS SPACE KEY"},
		gameOverTexts:   []string{"GAME OVER!", "PRESS SPACE KEY"},
		pauseTexts:      []string{"PAUSED", "PRESS SPACE KEY"},
		fontSize:        24,
		titleFontSize:   36,
		backgroundColor: color.RGBA{R: 76, G: 70, B: 50, A: 1},
		camX:            0.0,
		camY:            0.0,
		newlinePadding:  20,
		framesPerSecond: 60,
		recticle:        &world.Recticle{Size: 6},
		assets:          assets,
		frameCount:      1,
		maxFramCount:    60,
	}

	dpi := 72.0
	tt, _ := opentype.Parse(fonts.PressStart2P_ttf)
	game.titleArcadeFont, _ = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    float64(game.titleFontSize),
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	game.arcadeFont, _ = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    float64(game.fontSize),
		DPI:     dpi,
		Hinting: font.HintingFull,
	})

	game.titleFontColorScale.ScaleWithColor(color.White)

	game.titleGeoMatrix.Translate(float64(ScreenWidth-len(game.titleTexts[0])*game.titleFontSize)/2, float64(4*game.titleFontSize))
	game.titleExtraGeoMatrix.Translate(float64(ScreenWidth-len(game.titleTextsExtra[0])*game.fontSize)/2, float64(10*game.fontSize))
	game.gameOverGeoMatrix.Translate(float64(ScreenWidth-len(game.gameOverTexts[0])*game.fontSize)/2, float64(8*game.fontSize))
	game.pauseGeoMatrix.Translate(float64((ScreenWidth-len(game.pauseTexts[0])*game.fontSize)/2), float64(8*game.fontSize))

	// set text draw options
	game.titleDrawOptions = &text.DrawOptions{
		DrawImageOptions: ebiten.DrawImageOptions{
			GeoM:       game.titleGeoMatrix,
			ColorScale: game.titleFontColorScale,
		},
	}
	game.titleTextExtraDrawOptions = &text.DrawOptions{
		DrawImageOptions: ebiten.DrawImageOptions{
			GeoM:       game.titleExtraGeoMatrix,
			ColorScale: game.titleFontColorScale,
		},
	}
	game.gameOverDrawOptions = &text.DrawOptions{
		DrawImageOptions: ebiten.DrawImageOptions{
			GeoM:       game.gameOverGeoMatrix,
			ColorScale: game.titleFontColorScale,
		},
	}
	game.pauseDrawOptions = &text.DrawOptions{
		DrawImageOptions: ebiten.DrawImageOptions{
			GeoM:       game.pauseGeoMatrix,
			ColorScale: game.titleFontColorScale,
		},
	}

	playerSprites := helpers.LoadSprites(assets, []string{
		"assets/player-no-gun.png",
		"assets/player-revolver.png",
	}, 32, 32)

	game.player = &actors.Player{
		X:              0.0,
		Y:              0.0,
		W:              float64(playerSprites[0].Bounds().Dx()),
		H:              float64(playerSprites[0].Bounds().Dy()),
		Sprites:        playerSprites,
		Scale:          2,
		Speed:          2.0,
		AnimationSpeed: 15,
		DrawOptions:    &ebiten.DrawImageOptions{},
	}

	game.cactusSprites = helpers.LoadSprites(assets, []string{
		"assets/cactus.png",
	}, 32, 32)

	game.cacti = helpers.SpawnCacti(3*ScreenWidth, 3*ScreenHeight, 60, 4, game.cactusSprites)

	return game
}

func (g *Game) Initialize() {
	// Calculate the position of the screen center based on the player's position
	// camX = player.x + player.w/2 - ScreenWidth/2
	// camY = player.y + player.h/2 - ScreenHeight/2
}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	switch g.mode {
	case ModeTitle:
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			g.mode = ModeGame
		}
	case ModeGameOver:
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			g.Initialize()
			g.mode = ModeGame
		}
	case ModePause:
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			g.mode = ModeGame
		}
	case ModeGame:
		// Calculate the position of the screen center based on the player's position
		g.camX = g.player.X + g.player.W/2 - ScreenWidth/2
		g.camY = g.player.Y + g.player.H/2 - ScreenHeight/2

		g.frameCount += 1

		if inpututil.IsKeyJustPressed(ebiten.Key1) {
			g.player.DrawWeapon(actors.Revolver)
		}

		if inpututil.IsKeyJustPressed(ebiten.Key0) {
			g.player.DrawWeapon(actors.Fists)
		}

		directionKeyPressed := false
		if  ebiten.IsKeyPressed(ebiten.KeyS) {
			g.player.Y += g.player.Speed
			directionKeyPressed = true
		}

		if ebiten.IsKeyPressed(ebiten.KeyDown) {
			switch g.player.VisualDir {
			case actors.Left:
				g.player.ChangeVisualDirection(actors.LeftDown)
			case actors.LeftUp:
				g.player.ChangeVisualDirection(actors.LeftDown)
			case actors.Right:
				g.player.ChangeVisualDirection(actors.RightDown)
			case actors.RightUp:
				g.player.ChangeVisualDirection(actors.RightDown)
			}
		}

		if  ebiten.IsKeyPressed(ebiten.KeyZ) || ebiten.IsKeyPressed(ebiten.KeyW) {
			g.player.Y -= g.player.Speed
			directionKeyPressed = true
		}

		if ebiten.IsKeyPressed(ebiten.KeyUp) {
			switch g.player.VisualDir {
			case actors.Left:
				g.player.ChangeVisualDirection(actors.LeftUp)
			case actors.LeftDown:
				g.player.ChangeVisualDirection(actors.LeftUp)
			case actors.Right:
				g.player.ChangeVisualDirection(actors.RightUp)
			case actors.RightDown:
				g.player.ChangeVisualDirection(actors.RightUp)
			}
		}

		if  ebiten.IsKeyPressed(ebiten.KeyD) {
			g.player.X += g.player.Speed
			directionKeyPressed = true
		}

		if ebiten.IsKeyPressed(ebiten.KeyRight) {
			g.player.ChangeVisualDirection(actors.Right)
		}

		if ebiten.IsKeyPressed(ebiten.KeyQ) || ebiten.IsKeyPressed(ebiten.KeyA) {
			g.player.X -= g.player.Speed
			directionKeyPressed = true
		}

		if ebiten.IsKeyPressed(ebiten.KeyLeft) {
			g.player.ChangeVisualDirection(actors.Left)
		}

		if directionKeyPressed && g.frameCount%g.player.AnimationSpeed == 0 {
			g.player.Animate()
		}

		if !directionKeyPressed {
			g.player.StopAnimation()
		}

		if ebiten.IsKeyPressed(ebiten.KeyP) {
			g.mode = ModePause
		}

		if g.frameCount%g.maxFramCount == 0 {
			g.frameCount = 1
		}
	}
	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
	// Write your game's rendering.
	screen.Fill(g.backgroundColor)
	switch g.mode {
	case ModeTitle:
		for _, cactus := range g.cacti {
			cactus.Draw(screen, g.camX, g.camY)
		}
		for i, l := range g.titleTexts {
			tx := 0
			if i-1 > -1 {
				tx = (len(g.titleTexts[i-1]) - len(l)) * g.titleFontSize
			}
			g.titleDrawOptions.GeoM.Translate(float64(tx), float64(i+g.titleFontSize+g.newlinePadding))
			text.Draw(screen, l, text.NewGoXFace(g.titleArcadeFont), g.titleDrawOptions)
		}
		g.titleDrawOptions.GeoM = g.titleGeoMatrix

		for i, l := range g.titleTextsExtra {
			tx := 0
			if i-1 > -1 {
				tx = ((len(g.titleTexts[i-1]) - len(l)) * g.fontSize) / 2
			}
			g.titleTextExtraDrawOptions.GeoM.Translate(float64(tx), float64(i+g.fontSize+g.newlinePadding))
			text.Draw(screen, l, text.NewGoXFace(g.arcadeFont), g.titleTextExtraDrawOptions)
		}
		g.titleTextExtraDrawOptions.GeoM = g.titleExtraGeoMatrix

		// g.recticle.Draw(screen)

	case ModeGameOver:
		for i, l := range g.gameOverTexts {
			tx := 0
			if i-1 > -1 {
				tx = ((len(g.titleTexts[i-1]) - len(l)) * g.fontSize) / 2
			}
			g.gameOverDrawOptions.GeoM.Translate(float64(tx), float64(i+g.fontSize+g.newlinePadding))
			text.Draw(screen, l, text.NewGoXFace(g.arcadeFont), g.gameOverDrawOptions)
		}
		g.gameOverDrawOptions.GeoM = g.gameOverGeoMatrix

	case ModePause:
		for _, cactus := range g.cacti {
			cactus.Draw(screen, g.camX, g.camY)
		}

		for i, l := range g.pauseTexts {
			tx := 0
			if i-1 > -1 {
				tx = (len(g.titleTexts[i-1]) - len(l)) * g.fontSize
			}
			g.pauseDrawOptions.GeoM.Translate(float64(tx), float64(i+g.fontSize+g.newlinePadding))
			text.Draw(screen, l, text.NewGoXFace(g.arcadeFont), g.pauseDrawOptions)
		}
		g.pauseDrawOptions.GeoM = g.pauseGeoMatrix

	case ModeGame:
		for _, cactus := range g.cacti {
			cactus.Draw(screen, g.camX, g.camY)
		}
		// g.recticle.Draw(screen)
		g.player.Draw(screen, float64(g.player.X-g.camX), float64(g.player.Y-g.camY))
	}
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
