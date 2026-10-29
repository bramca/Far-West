package farwest

import (
	"embed"
	"fmt"
	"image/color"
	"math"
	"math/rand/v2"
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
)

const (
	ScreenWidth  = 1280
	ScreenHeight = 720

	playerHealthBarSize = 9.0
	enemyHealthBarSize  = 7.0

	// world
	worldTilesX   = 90
	worldTilesY   = 60
	worldTileSize = 64.0
	cactusAmount  = 100

	// levels
	initialEnemies   = 6
	maxLevelEnemies  = 12
	enemyCountGrowth = 1
	// every other level the enemies deal one more damage
	levelDamageBonus = 1
	// and they get two more hitpoints every level
	levelHealthBonus = 2
	// the enemies shoot a bit faster every level, never faster than this
	enemyFireRateFloor = 11
	// enemies spawn at least this far away from the player
	enemyMinSpawnDistance = 600.0
	enemyMaxSpawnDistance = 1600.0
	// chance that an enemy also carries a whisky as a second weapon
	commonWhiskyChance    = 0.08
	rareWhiskyChance      = 0.20
	legendaryWhiskyChance = 0.40
	// amount of frames a corpse stays on the field
	corpseDuration = 240
	// how close the player has to be to a gun on the floor to pick it up or
	// compare it with the matching gun in his gunbelt
	lootPickupRadius = 90.0
	// speed at which a swapped gun is tossed in the aim direction
	lootTossSpeed = 2.4

	scorePerKill = 100

	// weapons
	playerReloadDuration          = 60
	playerShootCooldown           = 16
	enemyReloadDuration           = 110
	defaultShotgunBulletSpread    = 30 // lower number is more spread
	defaultMachineGunBulletSpread = 40 // lower number is more spread
	defaultShotgunBulletCount     = 4

	// dodge cooldown keeps the player from spamming the dodge button
	playerDodgeCooldown = 30

	enemyHitboxOffset = 16
)

var (
	defaultMagazineSize = map[actors.Weapon]int{
		actors.Revolver:   6,
		actors.Shotgun:    2,
		actors.MachineGun: 15,
	}

	buttonsJustPressed = map[string]bool{}
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
	player                *actors.Player
	playerSprites         []*ebiten.Image
	enemyCommonSprites    []*ebiten.Image
	enemyRareSprites      []*ebiten.Image
	enemyLegendarySprites []*ebiten.Image
	gunSprites            []*ebiten.Image
	bulletSprite          *ebiten.Image
	enemies               []*actors.Enemy
	particles             []*actors.Particle
	loot                  []*actors.Loot

	// hit feedback
	// shakeFrames and shakeStrength control the decaying screen shake.
	shakeFrames   int
	shakeStrength float64
	// freezeFrames holds the game still for a few frames on a kill.
	freezeFrames int

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
	levelComplete bool
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
		pauseTexts:              []string{"PAUSED", "PRESS SPACE OR START BUTTON"},
		fontSize:                18,
		titleFontSize:           13,
		hudFontSize:             12,
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
		"assets/player-whisky.png",
		"assets/player-revolver.png",
		"assets/player-shotgun.png",
		"assets/player-machine-gun.png",
		"assets/player-dead.png",
	}, 32, 32)

	game.enemyCommonSprites = helpers.LoadSprites(assets, []string{
		"assets/enemy-common-no-gun.png",
		"assets/enemy-common-whisky.png",
		"assets/enemy-common-revolver.png",
		"assets/enemy-common-shotgun.png",
		"assets/enemy-common-machine-gun.png",
		"assets/enemy-common-dead.png",
	}, 32, 32)

	game.enemyRareSprites = helpers.LoadSprites(assets, []string{
		"assets/enemy-rare-no-gun.png",
		"assets/enemy-rare-whisky.png",
		"assets/enemy-rare-revolver.png",
		"assets/enemy-rare-shotgun.png",
		"assets/enemy-rare-machine-gun.png",
		"assets/enemy-rare-dead.png",
	}, 32, 32)

	game.enemyLegendarySprites = helpers.LoadSprites(assets, []string{
		"assets/enemy-legendary-no-gun.png",
		"assets/enemy-legendary-whisky.png",
		"assets/enemy-legendary-revolver.png",
		"assets/enemy-legendary-shotgun.png",
		"assets/enemy-legendary-machine-gun.png",
		"assets/enemy-legendary-dead.png",
	}, 32, 32)

	game.bulletSprite = helpers.LoadSprites(assets, []string{
		"assets/bullet.png",
	}, 32, 32)[0]

	game.cactusSprites = helpers.LoadSprites(assets, []string{
		"assets/cactus.png",
	}, 32, 32)

	game.gunSprites = helpers.LoadSprites(assets, []string{
		"assets/fist.png",
		"assets/whisky.png",
		"assets/revolver.png",
		"assets/shotgun.png",
		"assets/machine-gun.png",
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
	g.levelComplete = false

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
	clear(g.player.Bullets)
	g.player.Bullets = g.player.Bullets[:0]
	g.player.Hits = nil
	g.player.Reloading = false
	g.player.ReloadTimer = 0
	g.player.Healing = false
	g.player.HealTimer = 0
	g.player.HealDuration = 0
	g.player.HealTotal = 0
	g.player.HealApplied = 0
	g.player.DodgeCooldownTimer = 0
	g.player.Dead = false
	g.player.CurrentAction = actors.Action{}
	g.player.UpdateHitbox()
	g.player.UpdateHealhBar()
	g.player.DrawWeapon(g.player.CurrentWeapon)
	g.clearSpawnArea()

	g.enemies = nil
	g.loot = nil
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
	defaultGunBelt := []*actors.Gun{
		{
			Damage:        2,
			GunType:       actors.Whisky,
			Sprite:        g.gunSprites[actors.Whisky],
			DrawOptions:   &ebiten.DrawImageOptions{},
			ShootCooldown: 32,
			Ammo:          20,
			Scale:         2.0,
			X:             ScreenWidth/2 - 128,
			Y:             ScreenHeight - 64,
			AmmoFace:      g.hitFace,
		},
		{
			Damage:        3,
			ShootCooldown: playerShootCooldown,
			GunType:       actors.Fists,
			Sprite:        g.gunSprites[actors.Fists],
			DrawOptions:   &ebiten.DrawImageOptions{},
			Scale:         2.0,
			X:             ScreenWidth/2 - 96,
			Y:             ScreenHeight - 64,
			AmmoFace:      g.hitFace,
		},
		{
			Damage:              3,
			MagazineSize:        defaultMagazineSize[actors.Revolver],
			CurrentMagazineSize: defaultMagazineSize[actors.Revolver],
			Ammo:                99,
			ShootCooldown:       playerShootCooldown,
			BulletDuration:      120,
			BulletSpeed:         6.0,
			GunType:             actors.Revolver,
			MuzzleFlash:         3,
			Sprite:              g.gunSprites[actors.Revolver],
			DrawOptions:         &ebiten.DrawImageOptions{},
			Scale:               2.0,
			X:                   ScreenWidth/2 - 64,
			Y:                   ScreenHeight - 64,
			AmmoFace:            g.hitFace,
		},
		{
			Damage:              3,
			MagazineSize:        defaultMagazineSize[actors.Shotgun],
			CurrentMagazineSize: defaultMagazineSize[actors.Shotgun],
			Ammo:                20,
			ShootCooldown:       playerShootCooldown,
			GunType:             actors.Shotgun,
			BulletDuration:      50,
			BulletSpread:        defaultShotgunBulletSpread,
			BulletCount:         defaultShotgunBulletCount,
			BulletSpeed:         6.0,
			MuzzleFlash:         3,
			Sprite:              g.gunSprites[actors.Shotgun],
			DrawOptions:         &ebiten.DrawImageOptions{},
			Scale:               2.0,
			X:                   ScreenWidth/2 - 32,
			Y:                   ScreenHeight - 64,
			AmmoFace:            g.hitFace,
		},
		{
			Damage:              2,
			MagazineSize:        defaultMagazineSize[actors.MachineGun],
			CurrentMagazineSize: defaultMagazineSize[actors.MachineGun],
			Ammo:                120,
			ShootCooldown:       playerShootCooldown / 2,
			GunType:             actors.MachineGun,
			BulletDuration:      90,
			BulletSpread:        defaultMachineGunBulletSpread,
			BulletSpeed:         9.0,
			MuzzleFlash:         3,
			Sprite:              g.gunSprites[actors.MachineGun],
			DrawOptions:         &ebiten.DrawImageOptions{},
			Scale:               2.0,
			X:                   ScreenWidth / 2,
			Y:                   ScreenHeight - 64,
			AmmoFace:            g.hitFace,
		},
	}
	player := &actors.Player{
		X:              x,
		Y:              y,
		W:              float64(g.playerSprites[0].Bounds().Dx()),
		H:              float64(g.playerSprites[0].Bounds().Dy()),
		Sprites:        g.playerSprites,
		CurrentState:   actors.PlayerNoGunRight,
		CurrentWeapon:  1,
		Scale:          2,
		Speed:          2.8,
		DodgeSpeed:     4.5,
		DodgeDuration:  15,
		AnimationSpeed: 11,
		DrawOptions:    &ebiten.DrawImageOptions{},
		BulletSprite:   g.bulletSprite,
		GunBelt:        defaultGunBelt,
		GunBeltSize:    len(defaultGunBelt),
		Health:         20,
		MaxHealth:      20,
		ReloadDuration: playerReloadDuration,
		DodgeCooldown:  playerDodgeCooldown,
		MoveDirs: map[actors.Direction]bool{
			actors.Up:    false,
			actors.Down:  false,
			actors.Right: false,
			actors.Left:  false,
		},
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
	spriteWidth := float32(g.enemyCommonSprites[actors.PlayerRevolverLeft].Bounds().Dx() - 5)
	spriteHeight := float32(g.enemyCommonSprites[actors.PlayerRevolverLeft].Bounds().Dy())

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

// enemyTierChances returns how likely a legendary or rare enemy is to spawn
// on the given level, the tougher tiers get more common as the run goes on.
func enemyTierChances(level int) (float64, float64) {
	legendary := min(0.05+float64(level-1)*0.01, 0.25)
	rare := min(0.15+float64(level-1)*0.02, 0.40)

	return legendary, rare
}

// rollEnemyTier rolls the tier of a freshly spawned enemy.
func rollEnemyTier(level int) actors.Tier {
	return enemyTierForRoll(level, rand.Float64())
}

// enemyTierForRoll maps a roll in [0, 1) to an enemy tier.
func enemyTierForRoll(level int, roll float64) actors.Tier {
	legendary, rare := enemyTierChances(level)

	switch {
	case roll < legendary:
		return actors.Legendary
	case roll < legendary+rare:
		return actors.Rare
	default:
		return actors.Common
	}
}

// enemyWhiskyChance returns how likely an enemy of the given tier also
// carries a whisky in his gunbelt.
func enemyWhiskyChance(tier actors.Tier) float64 {
	switch tier {
	case actors.Legendary:
		return legendaryWhiskyChance
	case actors.Rare:
		return rareWhiskyChance
	default:
		return commonWhiskyChance
	}
}

// enemyHasWhisky maps a roll in [0, 1) to whether the enemy carries a whisky.
func enemyHasWhisky(tier actors.Tier, roll float64) bool {
	return roll < enemyWhiskyChance(tier)
}

// rollEnemyWhisky rolls whether a freshly spawned enemy carries a whisky.
func rollEnemyWhisky(tier actors.Tier) bool {
	return enemyHasWhisky(tier, rand.Float64())
}

// enemyTierStats holds the extra stats an enemy gets from its tier on top of
// the level scaling. Tougher tiers are deadlier, faster and shoot more often.
type enemyTierStats struct {
	health   int
	damage   int
	speed    float64
	fireRate int
}

func enemyStatsForTier(tier actors.Tier) enemyTierStats {
	switch tier {
	case actors.Legendary:
		return enemyTierStats{health: 12, damage: 2, speed: 0.5, fireRate: 4}
	case actors.Rare:
		return enemyTierStats{health: 5, damage: 1, speed: 0.25, fireRate: 2}
	default:
		return enemyTierStats{}
	}
}

func (g *Game) enemySpritesForTier(tier actors.Tier) []*ebiten.Image {
	switch tier {
	case actors.Legendary:
		return g.enemyLegendarySprites
	case actors.Rare:
		return g.enemyRareSprites
	default:
		return g.enemyCommonSprites
	}
}

// rollEnemyWeapon picks the kind of gun an enemy carries.
func rollEnemyWeapon() actors.Weapon {
	switch roll := rand.Float64(); {
	case roll > 0.75:
		return actors.Shotgun
	case roll < 0.25:
		return actors.MachineGun
	default:
		return actors.Revolver
	}
}

// rollGunTier rolls the tier of the gun an enemy carries.
func rollGunTier(enemyTier actors.Tier) actors.Tier {
	return gunTierForRoll(enemyTier, rand.Float64())
}

// gunTierForRoll maps a roll in [0, 1) to a gun tier, the tougher the enemy
// the higher the chance of a better gun.
func gunTierForRoll(enemyTier actors.Tier, roll float64) actors.Tier {
	legendaryChance, rareChance := 0.05, 0.20
	switch enemyTier {
	case actors.Rare:
		legendaryChance, rareChance = 0.15, 0.40
	case actors.Legendary:
		legendaryChance, rareChance = 0.40, 0.40
	}

	switch {
	case roll < legendaryChance:
		return actors.Legendary
	case roll < legendaryChance+rareChance:
		return actors.Rare
	default:
		return actors.Common
	}
}

// gunTierStats holds the stat bonuses a gun rolls on top of its base stats.
// Higher tiers roll on wider and higher ranges so their stats vary more.
type gunTierStats struct {
	damage        int
	magazine      int
	ammo          int
	shootCooldown int
	bulletSpeed   float64
	bulletSpread  int
	bulletCount   int
}

func rollGunTierStats(tier actors.Tier) gunTierStats {
	switch tier {
	case actors.Legendary:
		return gunTierStats{
			damage:        2 + rand.IntN(3),
			magazine:      2 + rand.IntN(3),
			ammo:          rand.IntN(20),
			shootCooldown: -(4 + rand.IntN(4)),
			bulletSpeed:   1.0 + rand.Float64()*1.5,
			bulletSpread:  8 + rand.IntN(9),
			bulletCount:   2,
		}
	case actors.Rare:
		return gunTierStats{
			damage:        1 + rand.IntN(2),
			magazine:      1 + rand.IntN(2),
			ammo:          rand.IntN(20),
			shootCooldown: -(2 + rand.IntN(3)),
			bulletSpeed:   0.5 + rand.Float64(),
			bulletSpread:  4 + rand.IntN(5),
			bulletCount:   1,
		}
	default:
		return gunTierStats{
			damage:      rand.IntN(2),
			magazine:    rand.IntN(2),
			ammo:        rand.IntN(20),
			bulletSpeed: rand.Float64() * 0.5,
		}
	}
}

// newEnemyGun builds the gun an enemy carries. The tier of the gun improves
// its base stats and widens the random spread of those stats.
func (g *Game) newEnemyGun(gunType actors.Weapon, tier actors.Tier, damage int) actors.Gun {
	gun := actors.Gun{
		Tier:          tier,
		Damage:        damage,
		GunType:       gunType,
		ShootCooldown: playerShootCooldown * 2,
		MuzzleFlash:   3,
		DrawOptions:   &ebiten.DrawImageOptions{},
		Scale:         2.0,
		AmmoFace:      g.hitFace,
	}

	switch gunType {
	case actors.Shotgun:
		gun.MagazineSize = defaultMagazineSize[actors.Shotgun]
		gun.BulletDuration = 50
		gun.BulletSpread = defaultShotgunBulletSpread
		gun.BulletCount = defaultShotgunBulletCount
		gun.Ammo = 15 + rand.IntN(10)
		gun.BulletSpeed = 6.0
		gun.Sprite = g.gunSprites[actors.Shotgun]
	case actors.MachineGun:
		gun.MagazineSize = defaultMagazineSize[actors.MachineGun]
		gun.BulletDuration = 90
		gun.BulletSpread = defaultMachineGunBulletSpread
		gun.BulletSpeed = 9.0
		gun.Ammo = 100 + rand.IntN(50)
		gun.ShootCooldown = playerShootCooldown / 2
		gun.Sprite = g.gunSprites[actors.MachineGun]
	case actors.Revolver:
		gun.MagazineSize = defaultMagazineSize[actors.Revolver]
		gun.BulletDuration = 120
		gun.Ammo = 50 + rand.IntN(50)
		gun.BulletSpeed = 6.0
		gun.Sprite = g.gunSprites[actors.Revolver]
	case actors.Whisky:
		gun.ShootCooldown = playerShootCooldown * 2
		gun.Ammo = 5 + rand.IntN(10)
		gun.MuzzleFlash = 0
		gun.Sprite = g.gunSprites[actors.Whisky]
	}

	stats := rollGunTierStats(tier)
	gun.Damage += stats.damage
	gun.ShootCooldown = max(1, gun.ShootCooldown+stats.shootCooldown)
	if gunType == actors.Whisky {
		gun.Ammo += min(stats.ammo, 2)

		return gun
	}

	gun.MagazineSize += stats.magazine
	gun.Ammo += stats.ammo
	gun.BulletSpeed += stats.bulletSpeed
	gun.BulletSpread += stats.bulletSpread
	gun.BulletCount += stats.bulletCount
	gun.CurrentMagazineSize = gun.MagazineSize

	return gun
}

func (g *Game) spawnEnemy() {
	x, y, ok := g.enemySpawnPoint()
	if !ok {
		return
	}

	tier := rollEnemyTier(g.level)
	tierStats := enemyStatsForTier(tier)
	sprites := g.enemySpritesForTier(tier)

	// the higher the level the tougher and deadlier the enemies get,
	// the base damage is low because every bullet deals at least 1
	damage := 2 + (g.level-1)/2*levelDamageBonus + tierStats.damage
	gun := g.newEnemyGun(rollEnemyWeapon(), rollGunTier(tier), damage)
	gunBelt := []*actors.Gun{&gun}
	if rollEnemyWhisky(tier) {
		whisky := g.newEnemyGun(actors.Whisky, rollGunTier(tier), damage)
		gunBelt = append(gunBelt, &whisky)
	}

	fireRate := max(enemyFireRateFloor, 18+rand.IntN(11)-(g.level-1)-tierStats.fireRate)

	state := actors.PlayerRevolverLeft
	switch gun.GunType {
	case actors.Shotgun:
		state = actors.PlayerShotgunLeft
	case actors.MachineGun:
		state = actors.PlayerMachineGunLeft
		fireRate = fireRate / 2
	}
	enemy := &actors.Enemy{
		Tier: tier,
		Player: &actors.Player{
			X:              x,
			Y:              y,
			W:              float64(g.enemyCommonSprites[state].Bounds().Dx() - 5),
			H:              float64(g.enemyCommonSprites[state].Bounds().Dy()),
			Sprites:        sprites,
			CurrentState:   state,
			CurrentWeapon:  0,
			Scale:          2,
			Speed:          0.7 + rand.Float64()*1.4 + tierStats.speed,
			DodgeSpeed:     0.4 + rand.Float64()*0.6,
			AnimationSpeed: 11,
			DrawOptions:    &ebiten.DrawImageOptions{},
			FireRate:       fireRate,
			GunBelt:        gunBelt,
			GunBeltSize:    len(gunBelt),
			BulletSprite:   g.bulletSprite,
			// the higher the level the tougher and deadlier the enemies get,
			// the base damage is low because every bullet deals at least 1
			Health:    10 + (g.level-1)*levelHealthBonus + tierStats.health,
			MaxHealth: 10 + (g.level-1)*levelHealthBonus + tierStats.health,
			MoveDirs: map[actors.Direction]bool{
				actors.Up:    false,
				actors.Down:  false,
				actors.Right: false,
				actors.Left:  false,
			},
			IsNpc:          true,
			HitboxOffset:   enemyHitboxOffset,
			ReloadDuration: enemyReloadDuration,
			Hitbox: &actors.HitBox{
				X: float32(x) + enemyHitboxOffset,
				Y: float32(y) + enemyHitboxOffset,
				W: float32(g.enemyCommonSprites[state].Bounds().Dx() - 5),
				H: float32(g.enemyCommonSprites[state].Bounds().Dy()),
			},
		},
		VisualDist: rand.IntN(350) + 400,
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

// addShake triggers a screen shake that decays over the given frames, a
// higher strength shakes the camera harder.
func (g *Game) addShake(frames int, strength float64) {
	g.shakeFrames = frames
	g.shakeStrength = strength
}

// spawnHitSparks scatters a few short lived particles around a point so
// impacts and kills have a bit of visual weight.
func (g *Game) spawnHitSparks(x, y float64, count int, clr color.RGBA) {
	for range count {
		angle := rand.Float64() * 2 * math.Pi
		speed := 0.8 + rand.Float64()*2.2
		g.particles = append(g.particles, actors.NewParticle(
			x, y,
			math.Cos(angle)*speed, math.Sin(angle)*speed,
			clr,
		))
	}
}

// dodgeDirection returns the direction the player is dashing in, it reports
// false when no movement key is held.
func (g *Game) dodgeDirection() (float64, float64, bool) {
	dirX, dirY := 0.0, 0.0
	if g.player.MoveDirs[actors.Right] {
		dirX += 1
	}
	if g.player.MoveDirs[actors.Left] {
		dirX -= 1
	}
	if g.player.MoveDirs[actors.Down] {
		dirY += 1
	}
	if g.player.MoveDirs[actors.Up] {
		dirY -= 1
	}

	length := math.Hypot(dirX, dirY)
	if length == 0 {
		return 0, 0, false
	}

	return dirX / length, dirY / length, true
}

// spawnDodgeDust kicks up a puff of dust behind the dashing player so the
// dodge reads as a quick burst of movement.
func (g *Game) spawnDodgeDust(count int) {
	dirX, dirY, moving := g.dodgeDirection()
	cx := g.player.X + g.player.W/2
	cy := g.player.Y + g.player.H - 6

	for range count {
		x := cx + (rand.Float64()*2-1)*7
		y := cy + (rand.Float64()*2-1)*4
		vx := (rand.Float64()*2 - 1) * 0.7
		vy := -rand.Float64() * 0.8

		if moving {
			spread := (rand.Float64()*2 - 1) * 10
			x += -dirY * spread
			y += dirX * spread

			speed := 0.5 + rand.Float64()*1.3
			vx = -dirX*speed + (rand.Float64()*2-1)*0.6
			vy = -dirY*speed - rand.Float64()*0.8
		}

		clr := color.RGBA{205, 185, 135, 200}
		g.particles = append(g.particles, actors.NewParticle(x, y, vx, vy, clr))
	}
}

// updateParticles advances every spark and drops the ones that burned out.
func (g *Game) updateParticles() {
	remaining := g.particles[:0]
	for _, particle := range g.particles {
		particle.Update()
		if particle.Life > 0 {
			remaining = append(remaining, particle)
		}
	}
	g.particles = remaining
}

// knockbackEnemy pushes an enemy a short distance along the bullet path so
// every hit that lands has some weight behind it. checkWorldCollisions pulls
// the enemy back if the push would have put him on the ocean or a cactus.
func (g *Game) knockbackEnemy(enemy *actors.Enemy, bullet *actors.Bullet) {
	const knockbackForce = 7.0
	enemy.X += math.Cos(bullet.R) * knockbackForce
	enemy.Y += math.Sin(bullet.R) * knockbackForce
	enemy.UpdateHitboxOffset(enemyHitboxOffset)
	enemy.UpdateHealhBar()
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
	old := g.player.Bullets
	remaining := old[:0]
	for i := range old {
		bullet := &old[i]
		if g.bulletBlocked(bullet) {
			continue
		}

		hitEnemy := false
		for _, enemy := range g.enemies {
			if enemy.Dead || !bullet.Hitbox.CheckCollision(enemy.Hitbox) {
				continue
			}
			enemy.Health -= bullet.Damage
			enemy.HitFlash = 6
			g.addHit(&enemy.Hits, enemy.X, enemy.Y-enemy.H/2, "-"+strconv.Itoa(bullet.Damage), color.RGBA{255, 255, 255, 240})
			enemy.UpdateHealhBar()
			g.spawnHitSparks(bullet.Hitbox.CenterX(), bullet.Hitbox.CenterY(), 6, color.RGBA{255, 200, 100, 240})
			if enemy.Health <= 0 {
				g.killEnemy(enemy)
			} else {
				g.knockbackEnemy(enemy, bullet)
			}
			hitEnemy = true

			break
		}
		if hitEnemy {
			continue
		}

		remaining = append(remaining, *bullet)
	}
	clear(old[len(remaining):])
	g.player.Bullets = remaining

	for _, enemy := range g.enemies {
		old := enemy.Bullets
		remaining := old[:0]
		for i := range old {
			bullet := &old[i]
			if g.bulletBlocked(bullet) {
				continue
			}

			// a dodging player is untouchable, the bullet flies through him
			if !g.player.Dead && !g.player.IsDodging() && bullet.Hitbox.CheckCollision(g.player.Hitbox) {
				g.damagePlayer(bullet.Damage)

				continue
			}

			remaining = append(remaining, *bullet)
		}
		clear(old[len(remaining):])
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
			g.spawnHitSparks(bullet.Hitbox.CenterX(), bullet.Hitbox.CenterY(), 4, color.RGBA{160, 130, 70, 220})
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
	g.player.HitFlash = 8
	g.addHit(&g.player.Hits, g.player.X, g.player.Y-g.player.H/2, "-"+strconv.Itoa(damage), color.RGBA{255, 255, 255, 240})
	g.addShake(8, 4)
	g.spawnHitSparks(g.player.X+g.player.W/2, g.player.Y+g.player.H/2, 6, color.RGBA{230, 70, 60, 240})

	if g.player.Health <= 0 {
		g.player.Health = 0
		g.player.Dead = true
		g.player.UpdateCurrentState(actors.PlayerDead)
		g.mode = ModeGameOver
	}

	g.player.UpdateHealhBar()
}

// killEnemy turns the enemy into a corpse, drops his gun and rewards the
// player with points.
func (g *Game) killEnemy(enemy *actors.Enemy) {
	if enemy.Dead {
		return
	}

	enemy.Health = 0
	enemy.Healthbar.Update(enemy.Healthbar.X, enemy.Healthbar.Y, enemy.Health, enemy.MaxHealth)
	enemy.Dead = true
	enemy.UpdateCurrentState(actors.PlayerDead)
	g.score += scorePerKill
	g.addHit(&enemy.Hits, enemy.X, enemy.Y-enemy.H/2, "+"+strconv.Itoa(scorePerKill), color.RGBA{255, 215, 0, 240})
	g.addShake(5, 3)
	g.freezeFrames = 3
	g.spawnHitSparks(enemy.X+enemy.W/2, enemy.Y+enemy.H/2, 12, color.RGBA{220, 60, 50, 240})
	g.dropEnemyGun(enemy)
}

// dropEnemyGun leaves the gun of a dead enemy on the floor so the player can
// pick it up.
func (g *Game) dropEnemyGun(enemy *actors.Enemy) {
	if enemy.CurrentWeapon < 0 || enemy.CurrentWeapon >= len(enemy.GunBelt) || enemy.GunBelt[enemy.CurrentWeapon] == nil {
		return
	}

	for i, gun := range enemy.GunBelt {
		g.loot = append(g.loot, actors.NewLoot(gun, enemy.X+enemy.W/2+float64(i)*enemy.W, enemy.Y+enemy.H/2+float64(i)*enemy.H))
	}
}

// gunBeltIndexFor returns the belt slot of the gun with the given type.
func (g *Game) gunBeltIndexFor(weapon actors.Weapon) (int, bool) {
	for i, gun := range g.player.GunBelt {
		if gun.GunType == weapon {
			return i, true
		}
	}

	return 0, false
}

// nearestLoot returns the closest gun on the floor within pickup range that
// has a matching gun type in the player's gunbelt.
func (g *Game) nearestLoot() *actors.Loot {
	var nearest *actors.Loot
	nearestDist := lootPickupRadius
	px := g.player.X + g.player.W/2
	py := g.player.Y + g.player.H/2

	for _, loot := range g.loot {
		if loot.Gun == nil {
			continue
		}
		if _, ok := g.gunBeltIndexFor(loot.Gun.GunType); !ok {
			continue
		}

		dist := utils.DistanceBetweenPoints(px, py, loot.X, loot.Y)
		if dist <= nearestDist {
			nearest = loot
			nearestDist = dist
		}
	}

	return nearest
}

// comparingLoot returns the nearby gun on the floor while the compare input
// is held.
func (g *Game) comparingLoot() *actors.Loot {
	if !ebiten.IsKeyPressed(ebiten.KeyQ) && !g.buttonsPressed["RL"] {
		return nil
	}

	return g.nearestLoot()
}

// pickUpLoot swaps the gun on the floor with the gun of the same type in the
// gunbelt and drops the old gun where the new one was found.
func (g *Game) pickUpLoot(loot *actors.Loot) {
	if loot.Gun == nil {
		return
	}

	index, ok := g.gunBeltIndexFor(loot.Gun.GunType)
	if !ok {
		return
	}

	oldGun := g.player.GunBelt[index]
	newGun := loot.Gun
	newGun.X, newGun.Y = oldGun.X, oldGun.Y
	g.player.GunBelt[index] = newGun
	if g.player.CurrentWeapon == index {
		g.player.Reloading = false
		g.player.ReloadTimer = 0
		g.player.DrawWeapon(index)
	}

	g.removeLoot(loot)

	if oldGun.GunType == actors.Fists {
		return
	}

	vx := math.Cos(g.player.AimAngle) * lootTossSpeed
	vy := math.Sin(g.player.AimAngle) * lootTossSpeed
	g.loot = append(g.loot, actors.NewTossedLoot(oldGun, loot.X, loot.Y, vx, vy))
}

// removeLoot takes a gun off the floor.
func (g *Game) removeLoot(target *actors.Loot) {
	for i, loot := range g.loot {
		if loot == target {
			g.loot = append(g.loot[:i], g.loot[i+1:]...)

			return
		}
	}
}

func (g *Game) addHit(hits *[]actors.Hit, x, y float64, msg string, hitColor color.RGBA) {
	hit := actors.Hit{
		X:        x,
		Y:        y,
		Color:    hitColor,
		Msg:      msg,
		TextFont: g.hitFace,
		Duration: g.framesPerSecond / 2,
	}
	hit.SetDrawOptions()
	*hits = append(*hits, hit)
}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update() error {
	// gamepad logic
	for k := range buttonsJustPressed {
		delete(buttonsJustPressed, k)
	}

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
	if g.shakeFrames > 0 {
		g.camX += (rand.Float64()*2 - 1) * g.shakeStrength
		g.camY += (rand.Float64()*2 - 1) * g.shakeStrength
		g.shakeFrames--
	}

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
	case ModeGame:
		// a kill freezes the game for a couple of frames, the hit-stop
		// gives the fight a bit more punch
		if g.freezeFrames > 0 {
			g.freezeFrames -= 1
			return nil
		}

		g.elapsedFrames += 1
		g.removeCorpses()

		g.player.SavePosition()
		g.player.UpdateWeapon()
		g.player.UpdateDodge()
		g.player.MoveDirs[actors.Up] = false
		g.player.MoveDirs[actors.Down] = false
		g.player.MoveDirs[actors.Right] = false
		g.player.MoveDirs[actors.Left] = false

		for _, enemy := range g.enemies {
			enemy.SavePosition()
			enemy.UpdateWeapon()
			enemy.MoveDirs[actors.Up] = false
			enemy.MoveDirs[actors.Down] = false
			enemy.MoveDirs[actors.Right] = false
			enemy.MoveDirs[actors.Left] = false
			enemy.UpdateBullets()
		}

		g.frameCount += 1

		g.player.UpdateBullets()

		// weapon switching
		if inpututil.IsKeyJustPressed(ebiten.Key5) && len(g.player.GunBelt) > 4 {
			g.player.DrawWeapon(4)
		}

		if inpututil.IsKeyJustPressed(ebiten.Key4) && len(g.player.GunBelt) > 3 {
			g.player.DrawWeapon(3)
		}

		if inpututil.IsKeyJustPressed(ebiten.Key3) && len(g.player.GunBelt) > 2 {
			g.player.DrawWeapon(2)
		}

		if inpututil.IsKeyJustPressed(ebiten.Key2) && len(g.player.GunBelt) > 1 {
			g.player.DrawWeapon(1)
		}

		if inpututil.IsKeyJustPressed(ebiten.Key1) && len(g.player.GunBelt) > 0 {
			g.player.DrawWeapon(0)
		}

		if buttonsJustPressed["RS"] || inpututil.IsKeyJustPressed(ebiten.KeyTab) {
			g.player.DrawWeapon((g.player.CurrentWeapon + 1) % len(g.player.GunBelt))
		}

		if buttonsJustPressed["LB"] {
			newWeapon := g.player.CurrentWeapon - 1
			if newWeapon < 0 {
				newWeapon = len(g.player.GunBelt) - 1
			}
			g.player.DrawWeapon(newWeapon)
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

		movingLeft := ebiten.IsKeyPressed(ebiten.KeyA) || g.xLeftAxis < -0.5
		// Q moves left like on an AZERTY keyboard, unless it is being used to
		// compare a nearby gun
		if ebiten.IsKeyPressed(ebiten.KeyQ) && g.comparingLoot() == nil {
			movingLeft = true
		}
		if movingLeft {
			g.player.Move(actors.Left)
			g.player.UpdateHitbox()
			directionKeyPressed = true
		}

		lookDirX := 0.0
		lookDirY := 0.0
		if ebiten.IsKeyPressed(ebiten.KeyL) {
			lookDirX = 1
		}

		if ebiten.IsKeyPressed(ebiten.KeyH) {
			lookDirX = -1
		}

		if ebiten.IsKeyPressed(ebiten.KeyK) {
			lookDirY = -1
		}

		if ebiten.IsKeyPressed(ebiten.KeyJ) {
			lookDirY = 1
		}

		// free-angle aim: while the right stick is pushed out of its deadzone
		// it takes over, otherwise the gun follows the mouse cursor
		if math.Hypot(g.xRightAxis, g.yRightAxis) > 0.5 {
			g.player.AimAngle = math.Atan2(g.yRightAxis, g.xRightAxis)
		} else if lookDirX != 0 || lookDirY != 0 {
			g.player.AimAngle = math.Atan2(lookDirY, lookDirX)
		} else {
			mx, my := ebiten.CursorPosition()
			worldX := float64(mx) + g.camX
			worldY := float64(my) + g.camY
			g.player.AimAngle = math.Atan2(worldY-(g.player.Y+g.player.H/2), worldX-(g.player.X+g.player.W/2))
		}
		g.player.FaceFromAim()

		// picking up a gun swaps it with the matching gun in the gunbelt and
		// tosses the old one in the aim direction
		if loot := g.nearestLoot(); loot != nil && (inpututil.IsKeyJustPressed(ebiten.KeyE) || buttonsJustPressed["RT"]) {
			g.pickUpLoot(loot)
		}

		if (inpututil.IsKeyJustPressed(ebiten.KeyShiftLeft) || buttonsJustPressed["FTL"]) && g.player.CanDodge() {
			g.player.StartDodge()
			g.player.CurrentAction = actors.Action{
				Duration: g.player.DodgeDuration,
				Type:     actors.Dodge,
				Actor:    g.player,
			}
			g.spawnDodgeDust(10)
		}

		g.player.Act(g.frameCount)

		if g.player.IsDodging() && g.frameCount%2 == 0 {
			g.spawnDodgeDust(2)
		}

		if directionKeyPressed && g.frameCount%g.player.AnimationSpeed == 0 {
			g.player.Animate()
		}

		if !directionKeyPressed {
			g.player.StopAnimation()
		}

		if ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || g.buttonsPressed["FTR"] {
			ammoBefore := g.player.GunBelt[g.player.CurrentWeapon].CurrentMagazineSize
			g.player.Shoot()
			if g.player.GunBelt[g.player.CurrentWeapon].CurrentMagazineSize != ammoBefore {
				g.addShake(2, 1.5)
			}
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyR) || buttonsJustPressed["RR"] {
			g.player.Reload()
		}

		for _, enemy := range g.enemies {
			if enemy.Dead {
				continue
			}
			enemy.ThinkAndAct(g.player, g.player.Bullets, g.frameCount)
		}

		g.CheckCollisions()
		g.updateParticles()

		for _, loot := range g.loot {
			loot.Update()
		}

		// a level is cleared when every enemy of it is dead, the player can
		// keep walking around until he presses N or the A button to continue
		if g.mode == ModeGame && !g.levelComplete && len(g.enemies) > 0 && g.aliveEnemies() == 0 {
			g.levelComplete = true
		}
		if g.mode == ModeGame && g.levelComplete && (inpututil.IsKeyJustPressed(ebiten.KeyN) || buttonsJustPressed["RB"]) {
			g.levelComplete = false
			g.startLevel(g.level + 1)
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

	case ModeGame:
		g.drawWorld(screen)
		g.drawHud(screen)
		if loot := g.comparingLoot(); loot != nil {
			g.drawLootComparison(screen, loot)
		}
		if g.levelComplete {
			g.drawLevelCompleteOverlay(screen)
		}
	}
}

// drawLevelCompleteOverlay shows the summary of the finished level while the
// player is still free to move around the island.
func (g *Game) drawLevelCompleteOverlay(screen *ebiten.Image) {
	levelCompleteTexts := []string{
		"LEVEL " + strconv.Itoa(g.level) + " COMPLETE!",
		"SCORE " + strconv.Itoa(g.score),
		"TIME  " + g.survivalTime(),
		"PRESS N OR A TO CONTINUE",
	}
	g.drawCenteredTexts(screen, levelCompleteTexts, g.arcadeFace, g.fontSize, g.pauseDrawOptions, g.pauseGeoMatrix)
}

// drawWorld renders the island and everything that lives on it.
func (g *Game) drawWorld(screen *ebiten.Image) {
	g.island.Draw(screen, g.camX, g.camY)

	for _, cactus := range g.cacti {
		cactus.Draw(screen, g.camX, g.camY)
	}

	for _, loot := range g.loot {
		loot.Draw(screen, g.camX, g.camY)
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

	for _, particle := range g.particles {
		particle.Draw(screen, g.camX, g.camY)
	}
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
	g.drawDodgeCooldown(screen)
	g.player.DrawGunBelt(screen, 0, 0)
	g.drawLootPrompt(screen)

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

// drawLootPrompt shows the controls of the gun lying next to the player.
func (g *Game) drawLootPrompt(screen *ebiten.Image) {
	if g.mode != ModeGame || g.player.Dead || g.nearestLoot() == nil {
		return
	}

	msg := "E TAKE   Q COMPARE"
	x := (ScreenWidth - len(msg)*g.hudFontSize) / 2
	g.drawSmallText(screen, msg, float64(x), float64(ScreenHeight)-150, color.RGBA{240, 230, 190, 255})
}

// gunStatRow is one line of the gun comparison overlay.
type gunStatRow struct {
	label          string
	now, loot      float64
	higherIsBetter bool
	decimals       int
}

// gunStatRows returns the stats that matter when comparing two guns.
func gunStatRows(current, loot *actors.Gun) []gunStatRow {
	if loot.GunType == actors.Whisky {
		return []gunStatRow{
			{"HEALING", float64(current.Damage), float64(loot.Damage), true, 0},
			{"AMMO", float64(current.Ammo), float64(loot.Ammo), true, 0},
			{"COOLDOWN", float64(current.ShootCooldown), float64(loot.ShootCooldown), false, 0},
		}
	}

	rows := []gunStatRow{
		{"DAMAGE", float64(current.Damage), float64(loot.Damage), true, 0},
		{"MAGAZINE", float64(current.MagazineSize), float64(loot.MagazineSize), true, 0},
		{"AMMO", float64(current.Ammo), float64(loot.Ammo), true, 0},
		{"COOLDOWN", float64(current.ShootCooldown), float64(loot.ShootCooldown), false, 0},
		{"SPEED", current.BulletSpeed, loot.BulletSpeed, true, 1},
	}

	switch loot.GunType {
	case actors.Shotgun:
		rows = append(rows, gunStatRow{"PELLETS", float64(current.BulletCount), float64(loot.BulletCount), true, 0})
	case actors.MachineGun:
		rows = append(rows, gunStatRow{"ACCURACY", float64(current.BulletSpread), float64(loot.BulletSpread), true, 0})
	}

	return rows
}

// statIsBetter reports whether a stat difference is an improvement.
func statIsBetter(delta float64, higherIsBetter bool) bool {
	if delta == 0 {
		return false
	}
	if higherIsBetter {
		return delta > 0
	}

	return delta < 0
}

func formatStatValue(value float64, decimals int) string {
	if decimals == 0 {
		return strconv.Itoa(int(math.Round(value)))
	}

	return strconv.FormatFloat(value, 'f', decimals, 64)
}

func formatStatDelta(delta float64, decimals int) string {
	if math.Abs(delta) < math.Pow10(-decimals)/2 {
		return "0"
	}

	sign := "+"
	if delta < 0 {
		sign = "-"
		delta = -delta
	}

	return sign + formatStatValue(delta, decimals)
}

// drawSmallText draws a hud sized line of text in the given color.
func (g *Game) drawSmallText(screen *ebiten.Image, msg string, x, y float64, clr color.RGBA) {
	drawOptions := &text.DrawOptions{}
	drawOptions.GeoM.Translate(x, y)
	drawOptions.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, msg, g.hudFace, drawOptions)
}

// drawLootComparison shows the stats of the gun on the floor next to the
// stats of the gun the player is holding.
func (g *Game) drawLootComparison(screen *ebiten.Image, loot *actors.Loot) {
	index, ok := g.gunBeltIndexFor(loot.Gun.GunType)
	if !ok {
		return
	}
	beltGun := g.player.GunBelt[index]
	tierColor := actors.TierColor(loot.Gun.Tier)

	const (
		panelW = 370.0
		panelH = 336.0
	)
	panelX := float64(ScreenWidth) - panelW - 20
	panelY := 160.0

	vector.FillRect(screen, float32(panelX), float32(panelY), panelW, panelH, color.RGBA{18, 16, 12, 238}, false)
	vector.StrokeRect(screen, float32(panelX), float32(panelY), panelW, panelH, 2, tierColor, false)

	floorText := "GUN ON THE FLOOR"
	if loot.Gun.GunType == actors.Whisky {
		floorText = "WHISKY ON THE FLOOR"
	}
	g.drawSmallText(screen, floorText, panelX+20, panelY+16, color.RGBA{235, 220, 170, 255})

	iconX := panelX + panelW/2
	iconY := panelY + 66
	loot.DrawIcon(screen, iconX, iconY, 2.0)

	tierText := loot.Gun.Tier.String()
	g.drawSmallText(screen, tierText, iconX-float64(len(tierText)*g.hudFontSize)/2, iconY+40, tierColor)

	nameText := loot.Gun.GunType.String()
	g.drawSmallText(screen, nameText, iconX-float64(len(nameText)*g.hudFontSize)/2, iconY+56, color.RGBA{230, 225, 210, 255})

	headerY := panelY + 150
	headerColor := color.RGBA{170, 160, 140, 255}
	g.drawSmallText(screen, "STAT", panelX+20, headerY, headerColor)
	g.drawSmallText(screen, "YOURS", panelX+170, headerY, headerColor)
	g.drawSmallText(screen, "LOOT", panelX+240, headerY, headerColor)
	g.drawSmallText(screen, "DIFF", panelX+310, headerY, headerColor)

	for i, row := range gunStatRows(beltGun, loot.Gun) {
		y := headerY + 24 + float64(i)*22
		g.drawSmallText(screen, row.label, panelX+20, y, color.RGBA{220, 215, 200, 255})
		g.drawSmallText(screen, formatStatValue(row.now, row.decimals), panelX+170, y, color.RGBA{180, 175, 160, 255})
		g.drawSmallText(screen, formatStatValue(row.loot, row.decimals), panelX+240, y, color.RGBA{240, 235, 220, 255})

		delta := row.loot - row.now
		diffColor := color.RGBA{170, 165, 150, 255}
		switch {
		case statIsBetter(delta, row.higherIsBetter):
			diffColor = color.RGBA{100, 225, 100, 255}
		case delta != 0:
			diffColor = color.RGBA{235, 85, 75, 255}
		}
		g.drawSmallText(screen, formatStatDelta(delta, row.decimals), panelX+310, y, diffColor)
	}

	g.drawSmallText(screen, "E TAKE", panelX+20, panelY+panelH-28, color.RGBA{235, 220, 170, 255})
}

// drawAmmo shows the bullets left in the magazine and the reload progress.
func (g *Game) drawAmmo(screen *ebiten.Image) {
	x := 40.0
	y := float64(ScreenHeight) - 120

	if g.player.Reloading {
		drawOptions := &text.DrawOptions{}
		drawOptions.GeoM.Translate(x, y-float64(g.hudFontSize)-6)
		text.Draw(screen, "RELOADING", g.hudFace, drawOptions)
	}

	if g.player.Reloading {
		progress := 1 - float64(g.player.ReloadTimer)/float64(g.player.ReloadDuration)
		vector.FillRect(screen, float32(x), float32(y), 100, 8, color.RGBA{60, 60, 60, 220}, false)
		vector.FillRect(screen, float32(x), float32(y), float32(100*progress), 8, color.RGBA{230, 190, 60, 240}, false)

		return
	}

	bulletWidth, bulletSpacing := float32(6), float32(6)
	for i := range g.player.GunBelt[g.player.CurrentWeapon].MagazineSize {
		bulletColor := color.RGBA{90, 70, 40, 200}
		if i < g.player.GunBelt[g.player.CurrentWeapon].CurrentMagazineSize {
			bulletColor = color.RGBA{240, 220, 120, 240}
		}
		vector.FillRect(screen, float32(x)+float32(i)*(bulletWidth+bulletSpacing), float32(y), bulletWidth, 14, bulletColor, false)
	}
}

// drawDodgeCooldown shows the dodge cooldown bar under the ammo, the player
// can only dodge again once it refills.
func (g *Game) drawDodgeCooldown(screen *ebiten.Image) {
	x := 40.0
	y := float64(ScreenHeight) - 76

	drawOptions := &text.DrawOptions{}
	drawOptions.GeoM.Translate(x, y-float64(g.hudFontSize)-6)
	text.Draw(screen, "DODGE", g.hudFace, drawOptions)

	progress := 1.0
	if g.player.DodgeCooldown > 0 {
		progress = 1 - float64(g.player.DodgeCooldownTimer)/float64(g.player.DodgeCooldown)
	}
	vector.FillRect(screen, float32(x), float32(y), 100, 8, color.RGBA{60, 60, 60, 220}, false)
	vector.FillRect(screen, float32(x), float32(y), float32(100*progress), 8, color.RGBA{110, 200, 220, 240}, false)
	drawOptions.GeoM.Translate(0, float64(g.hudFontSize)+22)
	text.Draw(screen, "HEALTH", g.hudFace, drawOptions)
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

// drawCenteredTexts draws a block of horizontally centered lines of text, the
// translation of the geo matrix already centers the first line.
func (g *Game) drawCenteredTexts(screen *ebiten.Image, texts []string, face *text.GoXFace, fontSize int, drawOptions *text.DrawOptions, geoMatrix ebiten.GeoM) {
	for i, line := range texts {
		tx := 0
		if i > 0 {
			tx = (len(texts[0]) - len(line)) * fontSize / 2
		}
		drawOptions.GeoM.Translate(float64(tx), float64((i+1)*(fontSize+g.newlinePadding)))
		text.Draw(screen, line, face, drawOptions)
		drawOptions.GeoM = geoMatrix
	}
	drawOptions.GeoM = geoMatrix
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
