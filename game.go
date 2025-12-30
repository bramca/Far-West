package farwest

import (
	"embed"
	"image/color"
	"math/rand"

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

// gamepad mappings
var standardButtonToString = map[ebiten.StandardGamepadButton]string{
	ebiten.StandardGamepadButtonRightBottom:      "RB",
	ebiten.StandardGamepadButtonRightRight:       "RR",
	ebiten.StandardGamepadButtonRightLeft:        "RL",
	ebiten.StandardGamepadButtonRightTop:         "RT",
	ebiten.StandardGamepadButtonFrontTopLeft:     "FTL",
	ebiten.StandardGamepadButtonFrontTopRight:    "FTR",
	ebiten.StandardGamepadButtonFrontBottomLeft:  "FBL",
	ebiten.StandardGamepadButtonFrontBottomRight: "FBR",
	ebiten.StandardGamepadButtonCenterLeft:       "CL",
	ebiten.StandardGamepadButtonCenterRight:      "CR",
	ebiten.StandardGamepadButtonLeftStick:        "LS",
	ebiten.StandardGamepadButtonRightStick:       "RS",
	ebiten.StandardGamepadButtonLeftBottom:       "LB",
	ebiten.StandardGamepadButtonLeftRight:        "LR",
	ebiten.StandardGamepadButtonLeftLeft:         "LL",
	ebiten.StandardGamepadButtonLeftTop:          "LT",
	ebiten.StandardGamepadButtonCenterCenter:     "CC",
}

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
	player       *actors.Player
	bulletSprite *ebiten.Image
	enemies      []*actors.Enemy

	// world
	cactusSprites  []*ebiten.Image
	cactusHitboxes []*actors.HitBox
	cacti          []*world.Cactus

	// gameplay
	frameCount   int
	maxFramCount int

	// gamepad
	gamepadIDsBuf  []ebiten.GamepadID
	gamepadIDs     map[ebiten.GamepadID]struct{}
	xLeftAxis      float64
	yLeftAxis      float64
	xRightAxis     float64
	yRightAxis     float64
	buttonsPressed map[string]bool
}

func NewGame() *Game {
	game := &Game{
		titleTexts:      []string{"FAR WEST"},
		titleTextsExtra: []string{"PRESS SPACE KEY OR START BUTTON"},
		gameOverTexts:   []string{"GAME OVER!", "PRESS SPACE KEY OR START BUTTON"},
		pauseTexts:      []string{"PAUSED", "PRESS SPACE KEY OR START BUTTON"},
		fontSize:        24,
		titleFontSize:   36,
		backgroundColor: color.RGBA{R: 76, G: 70, B: 50, A: 1},
		camX:            0.0,
		camY:            0.0,
		newlinePadding:  20,
		framesPerSecond: 60,
		assets:          assets,
		frameCount:      1,
		maxFramCount:    60,
		buttonsPressed:  map[string]bool{},
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

	enemySprites := helpers.LoadSprites(assets, []string{
		"assets/enemy-1-no-gun.png",
		"assets/enemy-1-revolver.png",
	}, 32, 32)

	game.bulletSprite = helpers.LoadSprites(assets, []string{
		"assets/bullet.png",
	}, 32, 32)[0]

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
		BulletSprite:   game.bulletSprite,
		Hitbox: &actors.HitBox{
			X: 0.0,
			Y: 0.0,
			W: float32(playerSprites[0].Bounds().Dx() - 5),
			H: float32(playerSprites[0].Bounds().Dy()),
		},
	}

	nEnemies := 5
	for range nEnemies {
		x := rand.Float64()*ScreenWidth + 20
		y := rand.Float64()*ScreenHeight + 20
		state := actors.PlayerRevolverLeft
		game.enemies = append(game.enemies, &actors.Enemy{
			Player: &actors.Player{
				X:              x,
				Y:              y,
				W:              float64(enemySprites[state].Bounds().Dx() - 5),
				H:              float64(enemySprites[state].Bounds().Dy()),
				Sprites:        enemySprites,
				CurrentState:   state,
				CurrentWeapon:  actors.Revolver,
				Scale:          2,
				Speed:          2.0,
				AnimationSpeed: 15,
				DrawOptions:    &ebiten.DrawImageOptions{},
				BulletSprite:   game.bulletSprite,
				Hitbox: &actors.HitBox{
					X: float32(x) + 16,
					Y: float32(y) + 16,
					W: float32(enemySprites[state].Bounds().Dx() - 5),
					H: float32(enemySprites[state].Bounds().Dy()),
				},
			},
			VisualDist: rand.Intn(200) + 250,
			MoveSpeed:  2,
			ShootSpeed: 25 + rand.Intn(15),
		})
	}

	game.cactusSprites = helpers.LoadSprites(assets, []string{
		"assets/cactus.png",
	}, 32, 32)

	game.cactusHitboxes = helpers.InitializeCactusHitboxes()

	cactusAmount := 60
	cactusSpawnBoundY := 3 * ScreenHeight
	cactusSpawnBoundX := 3 * ScreenWidth
	cactusSpriteScale := 4.0
	game.cacti = helpers.SpawnCacti(cactusSpawnBoundX, cactusSpawnBoundY, cactusAmount, cactusSpriteScale, game.cactusSprites, game.cactusHitboxes)

	return game
}

func (g *Game) Initialize() {
	// TODO: What happens after game over?
	// Calculate the position of the screen center based on the player's position
	// camX = player.x + player.w/2 - ScreenWidth/2
	// camY = player.y + player.h/2 - ScreenHeight/2
}

func (g *Game) CheckCollisions() {
	// TODO: check bullet collision and damage the environment
	for _, cactus := range g.cacti {
		removeIndices := []int{}
		for i, bullet := range g.player.Bullets {
			if bullet.Hitbox.CheckCollision(cactus.Hitbox) {
				removeIndices = append(removeIndices, i)
			}
		}
		for _, index := range removeIndices {
			g.player.Bullets = append(g.player.Bullets[:index], g.player.Bullets[index+1:]...)
		}

		for i, enemy := range g.enemies {
			if enemy.Hitbox.CheckCollision(cactus.Hitbox) {
				for dir, moving := range enemy.MoveDirs {
					if moving {
						switch dir {
						case actors.Up:
							enemy.Y += enemy.Speed
						case actors.Down:
							enemy.Y -= enemy.Speed
						case actors.Right:
							enemy.X -= enemy.Speed
						case actors.Left:
							enemy.X += enemy.Speed
						}
						enemy.UpdateHitboxOffset(16)
					}
				}
			}
			for j, otherEnemy := range g.enemies {
				if i == j {
					continue
				}

				if enemy.Hitbox.CheckCollision(otherEnemy.Hitbox) {
					for dir, moving := range enemy.MoveDirs {
						if moving {
							switch dir {
							case actors.Up:
								enemy.Y += enemy.Speed
							case actors.Down:
								enemy.Y -= enemy.Speed
							case actors.Right:
								enemy.X -= enemy.Speed
							case actors.Left:
								enemy.X += enemy.Speed
							}
							enemy.UpdateHitboxOffset(16)
						}
					}
				}

			}

			removeIndices := []int{}
			for i, bullet := range enemy.Bullets {
				if bullet.Hitbox.CheckCollision(cactus.Hitbox) {
					removeIndices = append(removeIndices, i)
				}
				// TODO: damage player
				if bullet.Hitbox.CheckCollision(g.player.Hitbox) {
					removeIndices = append(removeIndices, i)
				}
			}
			for _, index := range removeIndices {
				enemy.Bullets = append(enemy.Bullets[:index], enemy.Bullets[index+1:]...)
			}
		}

		if g.player.Hitbox.CheckCollision(cactus.Hitbox) {
			for dir, moving := range g.player.MoveDirs {
				if moving {
					switch dir {
					case actors.Up:
						g.player.Y += g.player.Speed
					case actors.Down:
						g.player.Y -= g.player.Speed
					case actors.Right:
						g.player.X -= g.player.Speed
					case actors.Left:
						g.player.X += g.player.Speed
					}
					g.player.UpdateHitbox()
				}
			}
		}
	}

	for _, enemy := range g.enemies {
		removeIndices := []int{}
		for i, bullet := range g.player.Bullets {
			// TODO: damage enemy
			if bullet.Hitbox.CheckCollision(enemy.Hitbox) {
				removeIndices = append(removeIndices, i)
			}
		}
		for _, index := range removeIndices {
			g.player.Bullets = append(g.player.Bullets[:index], g.player.Bullets[index+1:]...)
		}

		if g.player.Hitbox.CheckCollision(enemy.Hitbox) {
			for dir, moving := range g.player.MoveDirs {
				if moving {
					switch dir {
					case actors.Up:
						g.player.Y += g.player.Speed
					case actors.Down:
						g.player.Y -= g.player.Speed
					case actors.Right:
						g.player.X -= g.player.Speed
					case actors.Left:
						g.player.X += g.player.Speed
					}
					g.player.UpdateHitbox()
				}
			}
		}
	}
}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	// gamepad logic
	buttonsJustPressed := map[string]bool{}
	if g.gamepadIDs == nil {
		g.gamepadIDs = map[ebiten.GamepadID]struct{}{}
	}

	// log gamepad connection events
	g.gamepadIDsBuf = inpututil.AppendJustConnectedGamepadIDs(g.gamepadIDsBuf[:0])
	for _, id := range g.gamepadIDsBuf {
		// log.Printf("gamepad connected: id: %d, SDL id: $s", id, ebiten.GamepadSDLID(id))
		g.gamepadIDs[id] = struct{}{}
	}
	for id := range g.gamepadIDs {
		if inpututil.IsGamepadJustDisconnected(id) {
			// log.Printf("gamepad disconnected: id: %d", id)
			delete(g.gamepadIDs, id)
		}
	}

	for id := range g.gamepadIDs {
		// log axis events
		xLeftAxisPressed := ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal)
		if xLeftAxisPressed != g.xLeftAxis {
			g.xLeftAxis = xLeftAxisPressed
			// log.Printf("Left Stick X: %+0.2f", g.xLeftAxis)
		}
		yLeftAyisPressed := ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical)
		if yLeftAyisPressed != g.yLeftAxis {
			g.yLeftAxis = yLeftAyisPressed
			// log.Printf("Left Stick Y: %+0.2f", g.yLeftAxis)
		}
		xRightAxisPressed := ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisRightStickHorizontal)
		if xRightAxisPressed != g.xRightAxis {
			g.xRightAxis = xRightAxisPressed
			// log.Printf("Right Stick X: %+0.2f", g.xRightAxis)
		}
		yRightAxisPressed := ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisRightStickVertical)
		if yRightAxisPressed != g.yRightAxis {
			g.yRightAxis = yRightAxisPressed
			// log.Printf("Right Stick Y: %+0.2f", g.yRightAxis)
		}

		// log button events
		maxButton := ebiten.GamepadButton(ebiten.GamepadButtonCount(id))
		for b := ebiten.GamepadButton(0); b < maxButton; b++ {
			if inpututil.IsGamepadButtonJustPressed(id, b) {
				// log.Printf("button pressed: id: %d, button: %d - %s", id, b, standardButtonToString[ebiten.StandardGamepadButton(b)])
				g.buttonsPressed[standardButtonToString[ebiten.StandardGamepadButton(b)]] = true
				buttonsJustPressed[standardButtonToString[ebiten.StandardGamepadButton(b)]] = true
			}

			if inpututil.IsGamepadButtonJustReleased(id, b) {
				// log.Printf("button released: id %d, button: %d - %s", id, b, standardButtonToString[ebiten.StandardGamepadButton(b)])
				g.buttonsPressed[standardButtonToString[ebiten.StandardGamepadButton(b)]] = false
			}
		}

	}

	// controls
	switch g.mode {
	case ModeTitle:
		if ebiten.IsKeyPressed(ebiten.KeySpace) || g.buttonsPressed["FBR"] {
			g.mode = ModeGame
		}
	case ModeGameOver:
		if ebiten.IsKeyPressed(ebiten.KeySpace) || g.buttonsPressed["FBR"] {
			g.Initialize()
			g.mode = ModeGame
		}
	case ModePause:
		if ebiten.IsKeyPressed(ebiten.KeySpace) || g.buttonsPressed["FBR"] {
			g.mode = ModeGame
		}
	case ModeGame:
		// Calculate the position of the screen center based on the player's position
		g.camX = g.player.X + g.player.W/2 - ScreenWidth/2
		g.camY = g.player.Y + g.player.H/2 - ScreenHeight/2

		g.player.MoveDirs = map[actors.Direction]bool{
			actors.Up:    false,
			actors.Down:  false,
			actors.Right: false,
			actors.Left:  false,
		}

		for _, enemy := range g.enemies {
			enemy.MoveDirs = map[actors.Direction]bool{
				actors.Up:    false,
				actors.Down:  false,
				actors.Right: false,
				actors.Left:  false,
			}

			enemy.UpdateBullets()
		}

		g.frameCount += 1

		g.player.UpdateBullets()

		// weapon switching
		if inpututil.IsKeyJustPressed(ebiten.Key1) {
			g.player.DrawWeapon(actors.Revolver)
		}

		if inpututil.IsKeyJustPressed(ebiten.Key0) {
			g.player.DrawWeapon(actors.Fists)
		}

		if buttonsJustPressed["RT"] {
			g.player.DrawWeapon((g.player.CurrentWeapon + 1) % (actors.Revolver + 1))
		}

		directionKeyPressed := false
		if ebiten.IsKeyPressed(ebiten.KeyS) || g.yLeftAxis > 0.5 {
			g.player.Move(actors.Down)
			g.player.UpdateHitbox()
			directionKeyPressed = true
		}

		if ebiten.IsKeyPressed(ebiten.KeyDown) || g.yRightAxis > 0.5 {
			g.player.Look(actors.Down)
		}

		if ebiten.IsKeyPressed(ebiten.KeyZ) || ebiten.IsKeyPressed(ebiten.KeyW) || g.yLeftAxis < -0.5 {
			g.player.Move(actors.Up)
			g.player.UpdateHitbox()
			directionKeyPressed = true
		}

		if ebiten.IsKeyPressed(ebiten.KeyUp) || g.yRightAxis < -0.5 {
			g.player.Look(actors.Up)
		}

		if ebiten.IsKeyPressed(ebiten.KeyD) || g.xLeftAxis > 0.5 {
			g.player.Move(actors.Right)
			g.player.UpdateHitbox()
			directionKeyPressed = true
		}

		if ebiten.IsKeyPressed(ebiten.KeyRight) || g.xRightAxis > 0.5 {
			g.player.Look(actors.Right)
		}

		if ebiten.IsKeyPressed(ebiten.KeyQ) || ebiten.IsKeyPressed(ebiten.KeyA) || g.xLeftAxis < -0.5 {
			g.player.Move(actors.Left)
			g.player.UpdateHitbox()
			directionKeyPressed = true
		}

		if ebiten.IsKeyPressed(ebiten.KeyLeft) || g.xRightAxis < -0.5 {
			g.player.Look(actors.Left)
		}

		if directionKeyPressed && g.frameCount%g.player.AnimationSpeed == 0 {
			g.player.Animate()
		}

		if !directionKeyPressed {
			g.player.StopAnimation()
		}

		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || buttonsJustPressed["FTR"] {
			g.player.Shoot()
		}

		for _, enemy := range g.enemies {
			enemy.ThinkAndAct(g.player, g.player.Bullets, g.frameCount)
		}

		g.CheckCollisions()

		if ebiten.IsKeyPressed(ebiten.KeyP) || buttonsJustPressed["FBR"] {
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
		for _, enemy := range g.enemies {
			enemy.Draw(screen, g.camX, g.camY)
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
		for _, enemy := range g.enemies {
			enemy.Draw(screen, g.camX, g.camY)
			enemy.DrawBullets(screen, g.camX, g.camY)
		}

		g.player.Draw(screen, g.camX, g.camY)
		g.player.DrawBullets(screen, g.camX, g.camY)

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
			// cactus.DrawHitbox(screen, g.camX, g.camY)
		}
		for _, enemy := range g.enemies {
			enemy.Draw(screen, g.camX, g.camY)
			enemy.DrawBullets(screen, g.camX, g.camY)
			// enemy.DrawHitbox(screen, g.camX, g.camY)
		}
		g.player.Draw(screen, g.camX, g.camY)
		// g.player.DrawHitbox(screen, g.camX, g.camY)
		g.player.DrawBullets(screen, g.camX, g.camY)
	}
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
