package farwest

import (
	"embed"
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"strconv"

	"github.com/bramca/Far-West/actors"
	"github.com/bramca/Far-West/helpers"
	"github.com/bramca/Far-West/utils"
	"github.com/bramca/Far-West/world"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type Mode int

const (
	ModeTitle Mode = iota
	ModeGame
	ModeGameOver
	ModePause
	ModeLevelComplete
)

const (
	ScreenWidth  = 1280
	ScreenHeight = 860

	playerHealthBarSize = 9.0
	enemyHealthBarSize  = 7.0

	// world
	worldTilesX   = 90
	worldTilesY   = 60
	worldTileSize = 64.0
	cactusAmount  = 220

	// levels
	initialEnemies   = 6
	maxLevelEnemies  = 12
	enemyCountGrowth = 1
	// every other level the enemies deal one more damage
	levelDamageBonus = 1
	// and they get two more hitpoints every level
	levelHealthBonus = 2
	// the enemies shoot a bit faster every level, never faster than this
	enemyFireRateFloor = 15
	// enemies spawn at least this far away from the player
	enemyMinSpawnDistance = 600.0
	enemyMaxSpawnDistance = 1600.0
	// amount of frames a corpse stays on the field
	corpseDuration = 300

	scorePerKill = 100

	// weapons
	playerMagazineSize   = 6
	playerReloadDuration = 90
	playerShootCooldown  = 12
	enemyMagazineSize    = 6
	enemyReloadDuration  = 150

	enemyHitboxOffset = 16
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

	titleTexts    []string
	gameOverTexts []string
	pauseTexts    []string

	fontSize                int
	titleFontSize           int
	hudFontSize             int
	playerHealthBarFontSize int
	enemyHealthBarFontSize  int
	hitFontSize             int
	titleFontColorScale     ebiten.ColorScale

	titleArcadeFont     font.Face
	arcadeFont          font.Face
	hudFont             font.Face
	playerHealthBarFont font.Face
	enemyHealthBarFont  font.Face
	hitTextFont         font.Face

	titleFace  *text.GoXFace
	arcadeFace *text.GoXFace
	hudFace    *text.GoXFace
	hitFace    *text.GoXFace

	backgroundColor       color.RGBA
	playerHealthbarColors []color.RGBA
	enemyHealthbarColors  []color.RGBA

	camX float64
	camY float64

	// text geo matrices
	titleGeoMatrix    ebiten.GeoM
	gameOverGeoMatrix ebiten.GeoM
	pauseGeoMatrix    ebiten.GeoM

	// text padding
	newlinePadding int

	framesPerSecond int

	// draw options
	titleDrawOptions    *text.DrawOptions
	gameOverDrawOptions *text.DrawOptions
	pauseDrawOptions    *text.DrawOptions

	// actors
	player        *actors.Player
	playerSprites []*ebiten.Image
	enemySprites  []*ebiten.Image
	bulletSprite  *ebiten.Image
	enemies       []*actors.Enemy

	// world
	island         *world.Island
	cactusSprites  []*ebiten.Image
	cactusHitboxes []*actors.HitBox
	cacti          []*world.Cactus

	// gameplay
	frameCount   int
	maxFramCount int

	// level state
	level         int
	levelEnemies  int
	elapsedFrames int
	score         int

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
		titleTexts:              []string{"FAR WEST", "PRESS SPACE KEY OR START BUTTON"},
		gameOverTexts:           []string{"GAME OVER!", "PRESS SPACE KEY OR START BUTTON"},
		pauseTexts:              []string{"PAUSED", "PRESS SPACE KEY OR START BUTTON"},
		fontSize:                24,
		titleFontSize:           36,
		hudFontSize:             16,
		playerHealthBarFontSize: playerHealthBarSize,
		enemyHealthBarFontSize:  enemyHealthBarSize,
		hitFontSize:             8,
		backgroundColor:         color.RGBA{R: 76, G: 70, B: 50, A: 1},
		playerHealthbarColors:   []color.RGBA{{0, 255, 0, 240}, {255, 0, 0, 240}},
		enemyHealthbarColors:    []color.RGBA{{0, 255, 0, 240}, {255, 0, 0, 240}},
		camX:                    0.0,
		camY:                    0.0,
		newlinePadding:          20,
		framesPerSecond:         60,
		assets:                  assets,
		frameCount:              1,
		maxFramCount:            60,
		buttonsPressed:          map[string]bool{},
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
	game.hudFont, _ = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    float64(game.hudFontSize),
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	game.playerHealthBarFont, _ = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    float64(game.playerHealthBarFontSize),
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	game.enemyHealthBarFont, _ = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    float64(game.enemyHealthBarFontSize),
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	game.hitTextFont, _ = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    float64(game.hitFontSize),
		DPI:     dpi,
		Hinting: font.HintingVertical,
	})

	game.titleFontColorScale.ScaleWithColor(color.White)

	game.titleFace = text.NewGoXFace(game.titleArcadeFont)
	game.arcadeFace = text.NewGoXFace(game.arcadeFont)
	game.hudFace = text.NewGoXFace(game.hudFont)
	game.hitFace = text.NewGoXFace(game.hitTextFont)

	game.titleGeoMatrix.Translate(float64(ScreenWidth-len(game.titleTexts[0])*game.titleFontSize)/2, float64(4*game.titleFontSize))
	game.gameOverGeoMatrix.Translate(float64(ScreenWidth-len(game.gameOverTexts[0])*game.fontSize)/2, float64(8*game.fontSize))
	game.pauseGeoMatrix.Translate(float64((ScreenWidth-len(game.pauseTexts[0])*game.fontSize)/2), float64(8*game.fontSize))

	// set text draw options
	game.titleDrawOptions = &text.DrawOptions{
		DrawImageOptions: ebiten.DrawImageOptions{
			GeoM:       game.titleGeoMatrix,
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

	game.playerSprites = helpers.LoadSprites(assets, []string{
		"assets/player-no-gun.png",
		"assets/player-revolver.png",
		"assets/player-dead.png",
	}, 32, 32)

	game.enemySprites = helpers.LoadSprites(assets, []string{
		"assets/enemy-1-no-gun.png",
		"assets/enemy-1-revolver.png",
		"assets/enemy-1-dead.png",
	}, 32, 32)

	game.bulletSprite = helpers.LoadSprites(assets, []string{
		"assets/bullet.png",
	}, 32, 32)[0]

	game.cactusSprites = helpers.LoadSprites(assets, []string{
		"assets/cactus.png",
	}, 32, 32)

	game.cactusHitboxes = helpers.InitializeCactusHitboxes()

	game.Initialize()

	return game
}

// Initialize starts a brand new run from level one. It is called on startup
// and after every game over.
func (g *Game) Initialize() {
	g.island = world.NewIsland(worldTilesX, worldTilesY, worldTileSize)
	g.player = g.newPlayer()
	g.score = 0
	g.elapsedFrames = 0

	g.startLevel(1)
}

// startLevel drops the player on a brand new island for the given level,
// spawns the enemies to beat and restores the player.
func (g *Game) startLevel(level int) {
	g.level = level

	g.island = world.NewIsland(worldTilesX, worldTilesY, worldTileSize)
	g.cacti = helpers.SpawnCacti(g.island, cactusAmount, 4.0, g.cactusSprites, g.cactusHitboxes)

	x, y := g.island.Center()
	g.player.X = x
	g.player.Y = y
	g.player.Bullets = nil
	g.player.Hits = nil
	g.player.Health = g.player.MaxHealth
	g.player.Ammo = g.player.MagazineSize
	g.player.Reloading = false
	g.player.ReloadTimer = 0
	g.player.Dead = false
	g.player.CurrentAction = actors.Action{}
	g.player.UpdateHitbox()
	g.player.UpdateHealhBar()
	g.player.DrawWeapon(g.player.CurrentWeapon)
	g.clearSpawnArea()

	g.enemies = nil
	g.spawnLevel()
}

// spawnLevel spawns every enemy the player has to defeat in the current level.
func (g *Game) spawnLevel() {
	g.levelEnemies = enemyCountForLevel(g.level)
	for range g.levelEnemies {
		g.spawnEnemy()
	}
}

// enemyCountForLevel returns how many enemies the player has to beat in a
// level, it grows with the level up to a cap so the field never gets silly.
func enemyCountForLevel(level int) int {
	count := initialEnemies + (level-1)*enemyCountGrowth
	if count > maxLevelEnemies {
		count = maxLevelEnemies
	}

	return count
}

func (g *Game) newPlayer() *actors.Player {
	x, y := g.island.Center()
	player := &actors.Player{
		X:              x,
		Y:              y,
		W:              float64(g.playerSprites[0].Bounds().Dx()),
		H:              float64(g.playerSprites[0].Bounds().Dy()),
		Sprites:        g.playerSprites,
		Scale:          2,
		Speed:          2.0,
		DodgeSpeed:     1.7,
		DodgeDuration:  20,
		AnimationSpeed: 15,
		DrawOptions:    &ebiten.DrawImageOptions{},
		BulletSprite:   g.bulletSprite,
		Health:         20,
		MaxHealth:      20,
		MagazineSize:   playerMagazineSize,
		Ammo:           playerMagazineSize,
		ReloadDuration: playerReloadDuration,
		ShootCooldown:  playerShootCooldown,
		Damage:         3,
		Hitbox: &actors.HitBox{
			X: float32(x),
			Y: float32(y),
			W: float32(g.playerSprites[0].Bounds().Dx() - 5),
			H: float32(g.playerSprites[0].Bounds().Dy()),
		},
	}

	player.Healthbar = &actors.HealthBar{
		X:               40,
		Y:               ScreenHeight - 40,
		W:               100,
		H:               playerHealthBarSize,
		FixedSize:       true,
		FixedPos:        true,
		Points:          player.Health,
		MaxPoints:       player.MaxHealth,
		HealthBarColor:  g.playerHealthbarColors[0],
		HealthLostColor: g.playerHealthbarColors[1],
		TextFont:        text.NewGoXFace(g.playerHealthBarFont),
		FontColor:       color.RGBA{0, 0, 0, 240},
		FontSize:        g.playerHealthBarFontSize,
	}
	player.Healthbar.SetDrawOptions()
	player.SavePosition()

	return player
}

// enemySpawnPoint looks for a land tile that is far enough from the player so
// enemies never pop up right next to him.
// clearSpawnArea removes the cacti that grew on top of the player spawn.
func (g *Game) clearSpawnArea() {
	spawnArea := &actors.HitBox{
		X: g.player.Hitbox.X - g.player.Hitbox.W,
		Y: g.player.Hitbox.Y - g.player.Hitbox.H,
		W: g.player.Hitbox.W * 3,
		H: g.player.Hitbox.H * 3,
	}

	remaining := g.cacti[:0]
	for _, cactus := range g.cacti {
		if spawnArea.CheckCollision(cactus.Hitbox) {
			continue
		}
		remaining = append(remaining, cactus)
	}
	g.cacti = remaining
}

// enemySpawnPoint looks for a land tile that is far enough from the player so
// enemies never pop up right next to him. It reports whether such a spot was
// found, spawning an enemy on an invalid position would leave it stuck there
// for the rest of the run.
func (g *Game) enemySpawnPoint() (float64, float64, bool) {
	spriteWidth := float32(g.enemySprites[actors.PlayerRevolverLeft].Bounds().Dx() - 5)
	spriteHeight := float32(g.enemySprites[actors.PlayerRevolverLeft].Bounds().Dy())

	// the first pass keeps the enemies at a distance, the second pass drops
	// that requirement for the islands that are too small for it
	for pass := range 2 {
		for range 100 {
			x, y := g.island.RandomLandPoint()
			dist := utils.DistanceBetweenPoints(g.player.X, g.player.Y, x, y)
			if pass == 0 && (dist < enemyMinSpawnDistance || dist > enemyMaxSpawnDistance) {
				continue
			}
			hitbox := &actors.HitBox{
				X: float32(x) + enemyHitboxOffset,
				Y: float32(y) + enemyHitboxOffset,
				W: spriteWidth,
				H: spriteHeight,
			}
			if g.canStandAt(hitbox) {
				return x, y, true
			}
		}
	}

	return 0, 0, false
}

func (g *Game) spawnEnemy() {
	x, y, ok := g.enemySpawnPoint()
	if !ok {
		return
	}

	state := actors.PlayerRevolverLeft
	enemy := &actors.Enemy{
		Player: &actors.Player{
			X:              x,
			Y:              y,
			W:              float64(g.enemySprites[state].Bounds().Dx() - 5),
			H:              float64(g.enemySprites[state].Bounds().Dy()),
			Sprites:        g.enemySprites,
			CurrentState:   state,
			CurrentWeapon:  actors.Revolver,
			Scale:          2,
			Speed:          0.5 + rand.Float64(),
			DodgeSpeed:     0.3 + rand.Float64()*0.4,
			AnimationSpeed: 15,
			DrawOptions:    &ebiten.DrawImageOptions{},
			FireRate:       max(enemyFireRateFloor, 25+rand.Intn(15)-(g.level-1)),
			BulletSprite:   g.bulletSprite,
			// the higher the level the tougher and deadlier the enemies get
			Health:         10 + (g.level-1)*levelHealthBonus,
			MaxHealth:      10 + (g.level-1)*levelHealthBonus,
			Damage:         3 + (g.level-1)/2*levelDamageBonus,
			IsNpc:          true,
			HitboxOffset:   enemyHitboxOffset,
			MagazineSize:   enemyMagazineSize,
			Ammo:           enemyMagazineSize,
			ReloadDuration: enemyReloadDuration,
			Hitbox: &actors.HitBox{
				X: float32(x) + enemyHitboxOffset,
				Y: float32(y) + enemyHitboxOffset,
				W: float32(g.enemySprites[state].Bounds().Dx() - 5),
				H: float32(g.enemySprites[state].Bounds().Dy()),
			},
		},
		VisualDist: rand.Intn(350) + 400,
	}
	enemy.Healthbar = &actors.HealthBar{
		X:               enemy.X,
		Y:               enemy.Y - (enemy.H - enemy.H/3),
		W:               enemy.W + 5,
		H:               enemyHealthBarSize,
		Points:          enemy.Health,
		MaxPoints:       enemy.MaxHealth,
		HealthBarColor:  g.enemyHealthbarColors[0],
		HealthLostColor: g.enemyHealthbarColors[1],
		TextFont:        text.NewGoXFace(g.enemyHealthBarFont),
		FontColor:       color.RGBA{0, 0, 0, 240},
		FontSize:        g.enemyHealthBarFontSize,
	}
	enemy.Healthbar.SetDrawOptions()
	enemy.SavePosition()

	g.enemies = append(g.enemies, enemy)
}

// aliveEnemies counts the enemies that are still a threat to the player.
func (g *Game) aliveEnemies() int {
	alive := 0
	for _, enemy := range g.enemies {
		if !enemy.Dead {
			alive++
		}
	}

	return alive
}

// removeCorpses drops enemies that have been dead for a while so the world
// does not fill up with bodies.
func (g *Game) removeCorpses() {
	for i := len(g.enemies) - 1; i >= 0; i-- {
		enemy := g.enemies[i]
		if !enemy.Dead || len(enemy.Bullets) > 0 {
			continue
		}
		enemy.CorpseTimer += 1
		if enemy.CorpseTimer > corpseDuration {
			g.enemies = append(g.enemies[:i], g.enemies[i+1:]...)
		}
	}
}

// survivalTime returns how long the player has been surviving as mm:ss.
func (g *Game) survivalTime() string {
	seconds := g.elapsedFrames / g.framesPerSecond

	return fmt.Sprintf("%02d:%02d", seconds/60, seconds%60)
}

// CheckCollisions resolves every collision of the current frame: bullets,
// the borders of the island and the actors bumping into each other.
func (g *Game) CheckCollisions() {
	g.checkBulletCollisions()
	g.checkWorldCollisions()
	g.checkActorCollisions()
}

// checkBulletCollisions lets bullets damage what they hit and removes the
// bullets that hit something or that left the world.
func (g *Game) checkBulletCollisions() {
	remaining := g.player.Bullets[:0]
	for _, bullet := range g.player.Bullets {
		if g.bulletBlocked(bullet) {
			continue
		}

		hitEnemy := false
		for _, enemy := range g.enemies {
			if enemy.Dead || !bullet.Hitbox.CheckCollision(enemy.Hitbox) {
				continue
			}
			enemy.Health -= bullet.Damage
			g.addHit(&enemy.Hits, enemy.X, enemy.Y-enemy.H/2, "-"+strconv.Itoa(bullet.Damage), color.RGBA{255, 255, 255, 240})
			enemy.UpdateHealhBar()
			if enemy.Health <= 0 {
				g.killEnemy(enemy)
			}
			hitEnemy = true

			break
		}
		if hitEnemy {
			continue
		}

		remaining = append(remaining, bullet)
	}
	g.player.Bullets = remaining

	for _, enemy := range g.enemies {
		remaining := enemy.Bullets[:0]
		for _, bullet := range enemy.Bullets {
			if g.bulletBlocked(bullet) {
				continue
			}

			if !g.player.Dead && bullet.Hitbox.CheckCollision(g.player.Hitbox) {
				g.damagePlayer(bullet.Damage)

				continue
			}

			remaining = append(remaining, bullet)
		}
		enemy.Bullets = remaining
	}
}

// bulletBlocked reports whether the bullet hit a cactus or left the world.
func (g *Game) bulletBlocked(bullet *actors.Bullet) bool {
	x, y := float64(bullet.Hitbox.X), float64(bullet.Hitbox.Y)
	if x < 0 || y < 0 || x > g.island.PixelWidth() || y > g.island.PixelHeight() {
		return true
	}

	for _, cactus := range g.cacti {
		if bullet.Hitbox.CheckCollision(cactus.Hitbox) {
			return true
		}
	}

	return false
}

// checkWorldCollisions keeps the actors on solid ground, the ocean and the
// cacti are the barriers of the world.
func (g *Game) checkWorldCollisions() {
	if !g.player.Dead && !g.canStandAt(g.player.Hitbox) {
		g.player.RestorePosition()
	}

	for _, enemy := range g.enemies {
		if enemy.Dead {
			continue
		}
		if !g.canStandAt(enemy.Hitbox) {
			enemy.RestorePosition()
		}
	}
}

// canStandAt reports whether the given hitbox is completely on walkable land
// and does not overlap a cactus.
func (g *Game) canStandAt(hitbox *actors.HitBox) bool {
	if !g.island.IsRectWalkable(float64(hitbox.X), float64(hitbox.Y), float64(hitbox.W), float64(hitbox.H)) {
		return false
	}

	for _, cactus := range g.cacti {
		if hitbox.CheckCollision(cactus.Hitbox) {
			return false
		}
	}

	return true
}

// checkActorCollisions stops actors from walking through each other.
func (g *Game) checkActorCollisions() {
	for i, enemy := range g.enemies {
		if enemy.Dead {
			continue
		}

		if !g.player.Dead && g.player.Hitbox.CheckCollision(enemy.Hitbox) {
			g.player.RestorePosition()
			enemy.RestorePosition()
		}

		for _, otherEnemy := range g.enemies[i+1:] {
			if otherEnemy.Dead {
				continue
			}
			if enemy.Hitbox.CheckCollision(otherEnemy.Hitbox) {
				enemy.RestorePosition()
			}
		}
	}
}

// damagePlayer applies bullet damage to the player and ends the run when he dies.
func (g *Game) damagePlayer(damage int) {
	g.player.Health -= damage
	g.addHit(&g.player.Hits, g.player.X, g.player.Y-g.player.H/2, "-"+strconv.Itoa(damage), color.RGBA{255, 255, 255, 240})

	if g.player.Health <= 0 {
		g.player.Health = 0
		g.player.Dead = true
		g.player.UpdateCurrentState(actors.PlayerDead)
		g.mode = ModeGameOver
	}

	g.player.UpdateHealhBar()
}

// killEnemy turns the enemy into a corpse and rewards the player with points.
func (g *Game) killEnemy(enemy *actors.Enemy) {
	enemy.Health = 0
	enemy.Healthbar.Update(enemy.Healthbar.X, enemy.Healthbar.Y, enemy.Health, enemy.MaxHealth)
	enemy.Dead = true
	enemy.UpdateCurrentState(actors.PlayerDead)
	g.score += scorePerKill
	g.addHit(&enemy.Hits, enemy.X, enemy.Y-enemy.H/2, "+"+strconv.Itoa(scorePerKill), color.RGBA{255, 215, 0, 240})
}

func (g *Game) addHit(hits *[]actors.Hit, x, y float64, msg string, hitColor color.RGBA) {
	hit := actors.Hit{
		X:        x,
		Y:        y,
		Color:    hitColor,
		Msg:      msg,
		TextFont: g.hitFace,
		Duration: 2 * g.framesPerSecond / 3,
	}
	hit.SetDrawOptions()
	*hits = append(*hits, hit)
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
		if !ebiten.IsStandardGamepadAxisAvailable(id, ebiten.StandardGamepadAxisLeftStickHorizontal) ||
			!ebiten.IsStandardGamepadAxisAvailable(id, ebiten.StandardGamepadAxisRightStickHorizontal) ||
			!ebiten.IsStandardGamepadAxisAvailable(id, ebiten.StandardGamepadAxisLeftStickVertical) ||
			!ebiten.IsStandardGamepadAxisAvailable(id, ebiten.StandardGamepadAxisRightStickVertical) {
			continue
		}
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
		b := ebiten.GamepadButton(0)
		for b = range maxButton {
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

	// Calculate the position of the screen center based on the player's position
	g.camX = g.player.X + g.player.W/2 - ScreenWidth/2
	g.camY = g.player.Y + g.player.H/2 - ScreenHeight/2

	// controls
	switch g.mode {
	case ModeTitle:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || buttonsJustPressed["FBR"] {
			g.Initialize()
			g.mode = ModeGame
		}
	case ModeGameOver:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || buttonsJustPressed["FBR"] {
			g.Initialize()
			g.mode = ModeGame
		}
	case ModePause:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || buttonsJustPressed["FBR"] {
			g.mode = ModeGame
		}
	case ModeLevelComplete:
		if inpututil.IsKeyJustPressed(ebiten.KeyN) || buttonsJustPressed["RB"] {
			g.startLevel(g.level + 1)
			g.mode = ModeGame
		}
	case ModeGame:
		g.elapsedFrames += 1
		g.removeCorpses()

		g.player.SavePosition()
		g.player.UpdateWeapon()
		g.player.MoveDirs = map[actors.Direction]bool{
			actors.Up:    false,
			actors.Down:  false,
			actors.Right: false,
			actors.Left:  false,
		}

		for _, enemy := range g.enemies {
			enemy.SavePosition()
			enemy.UpdateWeapon()
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
			directionKeyPressed = true
		}

		if ebiten.IsKeyPressed(ebiten.KeyZ) || ebiten.IsKeyPressed(ebiten.KeyW) || g.yLeftAxis < -0.5 {
			g.player.Move(actors.Up)
			directionKeyPressed = true
		}

		if ebiten.IsKeyPressed(ebiten.KeyD) || g.xLeftAxis > 0.5 {
			g.player.Move(actors.Right)
			g.player.UpdateHitbox()
			directionKeyPressed = true
		}

		if ebiten.IsKeyPressed(ebiten.KeyQ) || ebiten.IsKeyPressed(ebiten.KeyA) || g.xLeftAxis < -0.5 {
			g.player.Move(actors.Left)
			g.player.UpdateHitbox()
			directionKeyPressed = true
		}

		// free-angle aim: while the right stick is pushed out of its deadzone
		// it takes over, otherwise the gun follows the mouse cursor
		if math.Hypot(g.xRightAxis, g.yRightAxis) > 0.5 {
			g.player.AimAngle = math.Atan2(g.yRightAxis, g.xRightAxis)
		} else {
			mx, my := ebiten.CursorPosition()
			worldX := float64(mx) + g.camX
			worldY := float64(my) + g.camY
			g.player.AimAngle = math.Atan2(worldY-(g.player.Y+g.player.H/2), worldX-(g.player.X+g.player.W/2))
		}
		g.player.FaceFromAim()

		if inpututil.IsKeyJustPressed(ebiten.KeyShiftLeft) || buttonsJustPressed["FTL"] {
			// TODO: only dodge when stamina is replenished
			g.player.CurrentAction = actors.Action{
				Duration: g.player.DodgeDuration,
				Type:     actors.Dodge,
				Actor:    g.player,
			}
		}

		g.player.Act(g.frameCount)

		if directionKeyPressed && g.frameCount%g.player.AnimationSpeed == 0 {
			g.player.Animate()
		}

		if !directionKeyPressed {
			g.player.StopAnimation()
		}

		if ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || g.buttonsPressed["FTR"] {
			g.player.Shoot()
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyR) || buttonsJustPressed["RL"] {
			g.player.Reload()
		}

		for _, enemy := range g.enemies {
			if enemy.Dead {
				continue
			}
			enemy.ThinkAndAct(g.player, g.player.Bullets, g.frameCount)
		}

		g.CheckCollisions()

		// a level is cleared when every enemy of it is dead, the player
		// presses N or the A button to continue to the next one
		if g.mode == ModeGame && len(g.enemies) > 0 && g.aliveEnemies() == 0 {
			g.mode = ModeLevelComplete
		}

		if g.mode == ModeGame && (ebiten.IsKeyPressed(ebiten.KeyP) || buttonsJustPressed["FBR"]) {
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
	screen.Fill(g.backgroundColor)
	switch g.mode {
	case ModeTitle:
		g.drawWorld(screen)
		g.drawCenteredTexts(screen, g.titleTexts, g.titleFace, g.titleFontSize, g.titleDrawOptions, g.titleGeoMatrix)

	case ModeGameOver:
		g.drawWorld(screen)
		g.drawHud(screen)
		gameOverTexts := append([]string{}, g.gameOverTexts[0])
		gameOverTexts = append(gameOverTexts,
			"LEVEL "+strconv.Itoa(g.level)+"  SURVIVED "+g.survivalTime(),
			"SCORE "+strconv.Itoa(g.score),
			g.gameOverTexts[1],
		)
		g.drawCenteredTexts(screen, gameOverTexts, g.arcadeFace, g.fontSize, g.gameOverDrawOptions, g.gameOverGeoMatrix)

	case ModePause:
		g.drawWorld(screen)
		g.drawHud(screen)
		g.drawCenteredTexts(screen, g.pauseTexts, g.arcadeFace, g.fontSize, g.pauseDrawOptions, g.pauseGeoMatrix)

	case ModeLevelComplete:
		g.drawWorld(screen)
		g.drawHud(screen)
		levelCompleteTexts := []string{
			"LEVEL " + strconv.Itoa(g.level) + " COMPLETE!",
			"SCORE " + strconv.Itoa(g.score),
			"TIME  " + g.survivalTime(),
			"PRESS N OR A TO CONTINUE",
		}
		g.drawCenteredTexts(screen, levelCompleteTexts, g.arcadeFace, g.fontSize, g.pauseDrawOptions, g.pauseGeoMatrix)

	case ModeGame:
		g.drawWorld(screen)
		g.drawHud(screen)
	}
}

// drawWorld renders the island and everything that lives on it.
func (g *Game) drawWorld(screen *ebiten.Image) {
	g.island.Draw(screen, g.camX, g.camY)

	for _, cactus := range g.cacti {
		cactus.Draw(screen, g.camX, g.camY)
	}

	// dead enemies are drawn first so corpses never cover a living actor
	for _, enemy := range g.enemies {
		if enemy.Dead {
			enemy.Draw(screen, g.camX, g.camY)
		}
	}

	for _, enemy := range g.enemies {
		if !enemy.Dead {
			enemy.Draw(screen, g.camX, g.camY)
		}
		enemy.DrawBullets(screen, g.camX, g.camY)
	}

	g.player.Draw(screen, g.camX, g.camY)
	g.player.DrawBullets(screen, g.camX, g.camY)
}

// drawHud renders the survival timer, the score, the ammo and the mini map.
func (g *Game) drawHud(screen *ebiten.Image) {
	lines := []string{
		"LEVEL " + strconv.Itoa(g.level),
		"TIME  " + g.survivalTime(),
		"SCORE " + strconv.Itoa(g.score),
		"FOES  " + strconv.Itoa(g.aliveEnemies()) + "/" + strconv.Itoa(g.levelEnemies),
	}
	for i, line := range lines {
		drawOptions := &text.DrawOptions{}
		drawOptions.GeoM.Translate(20, float64(20+i*(g.hudFontSize+6)))
		text.Draw(screen, line, g.hudFace, drawOptions)
	}

	g.drawAmmo(screen)

	miniMapScale := 2.0
	g.island.DrawMiniMap(
		screen,
		float64(ScreenWidth)-float64(worldTilesX)*miniMapScale-20,
		20,
		miniMapScale,
		g.player.X,
		g.player.Y,
		g.enemyPositions(),
	)
}

// drawAmmo shows the bullets left in the magazine and the reload progress.
func (g *Game) drawAmmo(screen *ebiten.Image) {
	x := 40.0
	y := float64(ScreenHeight) - 70

	msg := "AMMO"
	if g.player.Reloading {
		msg = "RELOADING"
	}
	drawOptions := &text.DrawOptions{}
	drawOptions.GeoM.Translate(x, y-float64(g.hudFontSize)-6)
	text.Draw(screen, msg, g.hudFace, drawOptions)

	if g.player.Reloading {
		progress := 1 - float64(g.player.ReloadTimer)/float64(g.player.ReloadDuration)
		vector.FillRect(screen, float32(x), float32(y), 100, 8, color.RGBA{60, 60, 60, 220}, false)
		vector.FillRect(screen, float32(x), float32(y), float32(100*progress), 8, color.RGBA{230, 190, 60, 240}, false)

		return
	}

	bulletWidth, bulletSpacing := float32(6), float32(6)
	for i := range g.player.MagazineSize {
		bulletColor := color.RGBA{90, 70, 40, 200}
		if i < g.player.Ammo {
			bulletColor = color.RGBA{240, 220, 120, 240}
		}
		vector.FillRect(screen, float32(x)+float32(i)*(bulletWidth+bulletSpacing), float32(y), bulletWidth, 14, bulletColor, false)
	}
}

// enemyPositions returns the position of every living enemy for the mini map.
func (g *Game) enemyPositions() [][2]float64 {
	positions := make([][2]float64, 0, len(g.enemies))
	for _, enemy := range g.enemies {
		if enemy.Dead {
			continue
		}
		positions = append(positions, [2]float64{enemy.X, enemy.Y})
	}

	return positions
}

// drawCenteredTexts draws a block of horizontally centered lines of text.
func (g *Game) drawCenteredTexts(screen *ebiten.Image, texts []string, face *text.GoXFace, fontSize int, drawOptions *text.DrawOptions, geoMatrix ebiten.GeoM) {
	for i, line := range texts {
		tx := 0
		if i > 0 {
			tx = (len(texts[i-1]) - len(line)) * fontSize / 2
		}
		drawOptions.GeoM.Translate(float64(tx), float64(i+fontSize+g.newlinePadding))
		text.Draw(screen, line, face, drawOptions)
	}
	drawOptions.GeoM = geoMatrix
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
