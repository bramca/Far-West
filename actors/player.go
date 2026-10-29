package actors

import (
	"image/color"
	"math"
	"math/rand/v2"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type (
	PlayerState int
	Direction   int
	Weapon      int
)

const (
	Right Direction = iota
	RightUp
	RightDown
	Left
	LeftUp
	LeftDown
	Up
	Down
)

const (
	Fists Weapon = iota
	Whisky
	Revolver
	Shotgun
	MachineGun
)

const (
	PlayerNoGunRight PlayerState = iota
	PlayerNoGunLeft
	PlayerNoGunRunRight
	PlayerNoGunRunLeft
	PlayerWhiskyRight
	PlayerWhiskyRightUp
	PlayerWhiskyLeft
	PlayerWhiskyLeftUp
	PlayerWhiskyRunRight
	PlayerWhiskyRunRightUp
	PlayerWhiskyRunLeft
	PlayerWhiskyRunLeftUp
	PlayerRevolverRight
	PlayerRevolverRightUp
	PlayerRevolverRightDown
	PlayerRevolverLeft
	PlayerRevolverLeftUp
	PlayerRevolverLeftDown
	PlayerRevolverRunRight
	PlayerRevolverRunRightUp
	PlayerRevolverRunRightDown
	PlayerRevolverRunLeft
	PlayerRevolverRunLeftUp
	PlayerRevolverRunLeftDown
	PlayerShotgunRight
	PlayerShotgunRightUp
	PlayerShotgunRightDown
	PlayerShotgunLeft
	PlayerShotgunLeftUp
	PlayerShotgunLeftDown
	PlayerShotgunRunRight
	PlayerShotgunRunRightUp
	PlayerShotgunRunRightDown
	PlayerShotgunRunLeft
	PlayerShotgunRunLeftUp
	PlayerShotgunRunLeftDown
	PlayerMachineGunRight
	PlayerMachineGunRightUp
	PlayerMachineGunRightDown
	PlayerMachineGunLeft
	PlayerMachineGunLeftUp
	PlayerMachineGunLeftDown
	PlayerMachineGunRunRight
	PlayerMachineGunRunRightUp
	PlayerMachineGunRunRightDown
	PlayerMachineGunRunLeft
	PlayerMachineGunRunLeftUp
	PlayerMachineGunRunLeftDown
	PlayerDead
)

type Player struct {
	X, Y           float64
	PrevX, PrevY   float64
	W, H           float64
	Sprites        []*ebiten.Image
	CurrentState   PlayerState
	Scale          float64
	Speed          float64
	DodgeDuration  int
	DodgeSpeed     float64
	AnimationSpeed int
	DrawOptions    *ebiten.DrawImageOptions
	CurrentAction  Action
	VisualDir      Direction
	// AimAngle is the free-angle in radians the gun is pointed at (0 is
	// right, grows clockwise because the screen has a flipped Y axis). The
	// bullets fly along this angle while VisualDir only selects the sprite.
	AimAngle      float64
	MoveDirs      map[Direction]bool
	CurrentWeapon int
	Hitbox        *HitBox
	HitboxOffset  float64
	Bullets       []Bullet
	BulletSprite  *ebiten.Image
	FireRate      int
	Healthbar     *HealthBar
	Health        int
	MaxHealth     int
	IsNpc         bool
	Running       bool
	Hits          []Hit
	Dead          bool

	// hit feedback
	// HitFlash counts the frames the sprite is drawn brightened after a hit.
	HitFlash int
	// MuzzleFlash counts the frames a muzzle flash is drawn after a shot,
	// it also telegraphs when the enemies pull the trigger.
	MuzzleFlash int

	// weapon handling
	GunBelt        []*Gun
	GunBeltSize    int
	ReloadDuration int
	ReloadTimer    int
	Reloading      bool
	ShootTimer     int

	// healing handling
	Healing      bool
	HealTimer    int
	HealDuration int
	HealTotal    int
	HealApplied  int

	// dodge handling
	DodgeCooldown      int
	DodgeCooldownTimer int
}

func (p *Player) Draw(screen *ebiten.Image, camX, camY float64) {
	// Draw the player
	p.DrawOptions.GeoM.Reset()
	p.DrawOptions.GeoM.Scale(p.Scale, p.Scale)
	p.DrawOptions.GeoM.Translate(-float64(p.W/2), -float64(p.H/2))
	p.DrawOptions.GeoM.Translate(p.X-camX, p.Y-camY)
	if !p.Dead {
		// a small reticle hovering at a fixed distance in front of the gun
		drawAimCursor(screen, p.X+p.W/2-camX, p.Y+p.H/2-camY, p.AimAngle, color.RGBA{255, 255, 255, 140})
	}
	p.DrawOptions.ColorScale.Reset()
	if p.HitFlash > 0 {
		// a red tint tells the player he got hit
		p.DrawOptions.ColorScale.Scale(3, 0.6, 0.6, 1)
		p.HitFlash--
	}
	if p.IsDodging() && p.CurrentAction.Duration%8 < 4 {
		// blink while the dodge makes the player untouchable
		p.DrawOptions.ColorScale.SetA(0.45)
	}
	screen.DrawImage(p.Sprites[p.CurrentState], p.DrawOptions)
	p.DrawOptions.ColorScale.Reset()
	weapon := p.GunBelt[p.CurrentWeapon].GunType
	if p.MuzzleFlash > 0 && !p.Dead && weapon != Fists && weapon != Whisky {
		drawMuzzleFlash(screen, p.X+p.W/2-camX, p.Y+p.H/2-camY, p.AimAngle)
	}
	p.Healthbar.Draw(screen, camX, camY)
	for i := len(p.Hits) - 1; i >= 0; i-- {
		if p.Hits[i].Duration > 0 {
			p.Hits[i].Update()
			p.Hits[i].Draw(screen, camX, camY)
		} else {
			p.Hits[i] = p.Hits[len(p.Hits)-1]
			p.Hits = p.Hits[:len(p.Hits)-1]
		}
	}
}

func (p *Player) DrawHitbox(screen *ebiten.Image, camX, camY float64) {
	p.Hitbox.Draw(screen, camX, camY)
}

func (p *Player) DrawGunBelt(screen *ebiten.Image, camX, camY float64) {
	for _, gun := range p.GunBelt {
		gun.Draw(screen, camX, camY)
	}

	if p.CurrentWeapon < 0 || p.CurrentWeapon >= len(p.GunBelt) {
		return
	}

	// a golden bar under the selected gun marks which one is in the player's
	// hands
	gun := p.GunBelt[p.CurrentWeapon]
	const (
		markerWidth  = 38.0
		markerHeight = 3.0
	)
	spriteWidth := float64(gun.Sprite.Bounds().Dx()) * gun.Scale
	spriteHeight := float64(gun.Sprite.Bounds().Dy()) * gun.Scale
	markerX := gun.X + spriteWidth/2 - markerWidth/2
	markerY := gun.Y + spriteHeight - 14
	vector.FillRect(
		screen,
		float32(markerX-camX), float32(markerY-camY),
		markerWidth, markerHeight,
		TierColor(gun.Tier),
		false,
	)
}

func (p *Player) UpdateHitbox() {
	p.Hitbox.X = float32(p.X + p.HitboxOffset)
	p.Hitbox.Y = float32(p.Y + p.HitboxOffset)
}

// SavePosition stores the current position so it can be restored when the
// actor ends up somewhere it is not allowed to be.
func (p *Player) SavePosition() {
	p.PrevX, p.PrevY = p.X, p.Y
}

// RestorePosition moves the actor back to the last saved position.
func (p *Player) RestorePosition() {
	p.X, p.Y = p.PrevX, p.PrevY
	p.UpdateHitbox()
	p.UpdateHealhBar()
}

func (p *Player) UpdateHealhBar() {
	if p.IsNpc {
		p.Healthbar.Update(p.X, p.Y-(p.H-p.H/3), p.Health, p.MaxHealth)
	}
	if !p.IsNpc {
		p.Healthbar.Update(p.Healthbar.X, p.Healthbar.Y, p.Health, p.MaxHealth)
	}
}

func (p *Player) Shoot() {
	gun := p.GunBelt[p.CurrentWeapon]
	if gun.GunType == Fists || p.Dead {
		return
	}

	if gun.GunType == Whisky {
		p.drinkWhisky(gun)

		return
	}

	if !p.CanShoot() {
		// running dry automatically starts a reload
		if gun.CurrentMagazineSize <= 0 {
			p.Reload()
		}

		return
	}

	switch gun.GunType {
	case Revolver:
		// the lifetime is scaled down with the speed to keep the range the same
		p.addBullet(p.BulletSprite, gun.BulletSpeed, p.AimAngle, gun.BulletDuration, gun.Damage)
		p.MuzzleFlash = gun.MuzzleFlash
	case Shotgun:
		for i := range gun.BulletCount {
			p.addBullet(p.BulletSprite, gun.BulletSpeed+rand.Float64(), p.AimAngle-(float64(i-gun.BulletCount/2)*math.Pi)/float64(gun.BulletSpread), gun.BulletDuration, int(float64(gun.Damage)*rand.Float64())+1)
		}
		p.MuzzleFlash = gun.MuzzleFlash
	case MachineGun:
		p.addBullet(p.BulletSprite, gun.BulletSpeed, p.AimAngle-((0.5+rand.Float64())*math.Pi/float64(gun.BulletSpread)), gun.BulletDuration, gun.Damage)
		p.MuzzleFlash = gun.MuzzleFlash
	}

	p.GunBelt[p.CurrentWeapon].Ammo -= 1
	p.GunBelt[p.CurrentWeapon].CurrentMagazineSize -= 1
	p.ShootTimer = gun.ShootCooldown
	if p.GunBelt[p.CurrentWeapon].CurrentMagazineSize <= 0 {
		p.Reload()
	}
}

// CanShoot reports whether the weapon is loaded and ready to fire again.
func (p *Player) CanShoot() bool {
	gun := p.GunBelt[p.CurrentWeapon]
	if gun.GunType == Whisky {
		return !p.Reloading && !p.Healing && p.ShootTimer <= 0 && gun.Ammo > 0 && p.Health < p.MaxHealth
	}

	return !p.Reloading && gun.CurrentMagazineSize > 0 && p.ShootTimer <= 0 && gun.Ammo > 0
}

// CanReload reports whether the weapon can be reloaded according to it's ammo status.
func (p *Player) CanReload() bool {
	return !p.Reloading && p.GunBelt[p.CurrentWeapon].Ammo > p.GunBelt[p.CurrentWeapon].CurrentMagazineSize
}

// CanDodge reports whether the player is ready to dodge again.
func (p *Player) CanDodge() bool {
	return !p.Dead && p.DodgeCooldownTimer <= 0
}

// IsDodging reports whether the actor is in the middle of a dodge roll,
// while it lasts the player takes no damage from bullets.
func (p *Player) IsDodging() bool {
	return !p.Dead && p.CurrentAction.Type == Dodge && p.CurrentAction.Duration > 0
}

// StartDodge begins the dodge cooldown, it has to be called once per dodge
// to keep the player from spamming the dodge button.
func (p *Player) StartDodge() {
	p.DodgeCooldownTimer = p.DodgeCooldown
}

// UpdateDodge ticks the dodge cooldown timer, it has to be called once per frame.
func (p *Player) UpdateDodge() {
	if p.DodgeCooldownTimer > 0 {
		p.DodgeCooldownTimer -= 1
	}
}

// Reload starts the reload of the weapon, during the reload no bullet can be fired.
func (p *Player) Reload() {
	if !p.CanReload() || p.Reloading || p.GunBelt[p.CurrentWeapon].CurrentMagazineSize >= p.GunBelt[p.CurrentWeapon].MagazineSize {
		return
	}
	p.Reloading = true
	p.ReloadTimer = p.ReloadDuration
}

// UpdateWeapon ticks the reload and fire rate timers, it has to be called once per frame.
func (p *Player) UpdateWeapon() {
	if p.ShootTimer > 0 {
		p.ShootTimer -= 1
	}
	if p.MuzzleFlash > 0 {
		p.MuzzleFlash -= 1
	}
	if p.Healing {
		p.updateHealing()
	}

	if !p.Reloading {
		return
	}

	p.ReloadTimer -= 1
	if p.ReloadTimer <= 0 {
		p.Reloading = false
		p.GunBelt[p.CurrentWeapon].CurrentMagazineSize = min(p.GunBelt[p.CurrentWeapon].MagazineSize, p.GunBelt[p.CurrentWeapon].Ammo)
	}
}

func (p *Player) Move(d Direction) {
	p.MoveDirs[d] = true
	switch d {
	case Right:
		p.X += p.Speed
	case Left:
		p.X -= p.Speed
	case Up:
		p.Y -= p.Speed
	case Down:
		p.Y += p.Speed
	}
	p.UpdateHitbox()
	p.UpdateHealhBar()
}

func (p *Player) Act(frameCount int) {
	if p.CurrentAction.Duration > 0 {
		p.CurrentAction.Duration -= 1
		p.CurrentAction.PerformAction(p, frameCount)
	}
}

func (p *Player) Look(d Direction) {
	switch d {
	case Up:
		switch p.VisualDir {
		case Left:
			p.ChangeVisualDirection(LeftUp)
		case LeftDown:
			p.ChangeVisualDirection(LeftUp)
		case Right:
			p.ChangeVisualDirection(RightUp)
		case RightDown:
			p.ChangeVisualDirection(RightUp)
		}
	case Down:
		switch p.VisualDir {
		case Left:
			p.ChangeVisualDirection(LeftDown)
		case LeftUp:
			p.ChangeVisualDirection(LeftDown)
		case Right:
			p.ChangeVisualDirection(RightDown)
		case RightUp:
			p.ChangeVisualDirection(RightDown)
		}
	case Right:
		p.ChangeVisualDirection(Right)
	case Left:
		p.ChangeVisualDirection(Left)
	}
}

// FaceFromAim turns the free-angle aim into one of the eight visual facing
// directions, so the sprite follows the quadrant the gun is pointed at.
func (p *Player) FaceFromAim() {
	if p.Dead {
		return
	}

	a := math.Mod(p.AimAngle, 2*math.Pi)
	if a < 0 {
		a += 2 * math.Pi
	}

	var dir Direction
	switch {
	case a < math.Pi/4:
		dir = Right
	case a < math.Pi/2:
		dir = RightDown
	case a < 3*math.Pi/4:
		dir = LeftDown
	case a < 5*math.Pi/4:
		dir = Left
	case a < 3*math.Pi/2:
		dir = LeftUp
	case a < 7*math.Pi/4:
		dir = RightUp
	default:
		dir = Right
	}
	p.ChangeVisualDirection(dir)
}

func (p *Player) UpdateBullets() {
	old := p.Bullets
	alive := old[:0]
	for i := range old {
		old[i].Update()
		if old[i].Duration >= 1 {
			alive = append(alive, old[i])
		}
	}
	clear(old[len(alive):])
	p.Bullets = alive
}

func (p *Player) DrawBullets(screen *ebiten.Image, camX, camY float64) {
	for i := range p.Bullets {
		p.Bullets[i].Draw(screen, camX, camY)
	}
}

func (p *Player) UpdateCurrentState(newState PlayerState) {
	p.CurrentState = newState
}

func (p *Player) ChangeVisualDirection(newDir Direction) {
	p.VisualDir = newDir
	switch p.GunBelt[p.CurrentWeapon].GunType {
	case Revolver:
		switch p.VisualDir {
		case Left:
			if p.Running {
				p.UpdateCurrentState(PlayerRevolverRunLeft)
			} else {
				p.UpdateCurrentState(PlayerRevolverLeft)
			}
		case LeftUp:
			if p.Running {
				p.UpdateCurrentState(PlayerRevolverRunLeftUp)
			} else {
				p.UpdateCurrentState(PlayerRevolverLeftUp)
			}
		case LeftDown:
			if p.Running {
				p.UpdateCurrentState(PlayerRevolverRunLeftDown)
			} else {
				p.UpdateCurrentState(PlayerRevolverLeftDown)
			}
		case Right:
			if p.Running {
				p.UpdateCurrentState(PlayerRevolverRunRight)
			} else {
				p.UpdateCurrentState(PlayerRevolverRight)
			}
		case RightUp:
			if p.Running {
				p.UpdateCurrentState(PlayerRevolverRunRightUp)
			} else {
				p.UpdateCurrentState(PlayerRevolverRightUp)
			}
		case RightDown:
			if p.Running {
				p.UpdateCurrentState(PlayerRevolverRunRightDown)
			} else {
				p.UpdateCurrentState(PlayerRevolverRightDown)
			}
		}
	case Shotgun:
		switch p.VisualDir {
		case Left:
			if p.Running {
				p.UpdateCurrentState(PlayerShotgunRunLeft)
			} else {
				p.UpdateCurrentState(PlayerShotgunLeft)
			}
		case LeftUp:
			if p.Running {
				p.UpdateCurrentState(PlayerShotgunRunLeftUp)
			} else {
				p.UpdateCurrentState(PlayerShotgunLeftUp)
			}
		case LeftDown:
			if p.Running {
				p.UpdateCurrentState(PlayerShotgunRunLeftDown)
			} else {
				p.UpdateCurrentState(PlayerShotgunLeftDown)
			}
		case Right:
			if p.Running {
				p.UpdateCurrentState(PlayerShotgunRunRight)
			} else {
				p.UpdateCurrentState(PlayerShotgunRight)
			}
		case RightUp:
			if p.Running {
				p.UpdateCurrentState(PlayerShotgunRunRightUp)
			} else {
				p.UpdateCurrentState(PlayerShotgunRightUp)
			}
		case RightDown:
			if p.Running {
				p.UpdateCurrentState(PlayerShotgunRunRightDown)
			} else {
				p.UpdateCurrentState(PlayerShotgunRightDown)
			}
		}
	case MachineGun:
		switch p.VisualDir {
		case Left:
			if p.Running {
				p.UpdateCurrentState(PlayerMachineGunRunLeft)
			} else {
				p.UpdateCurrentState(PlayerMachineGunLeft)
			}
		case LeftUp:
			if p.Running {
				p.UpdateCurrentState(PlayerMachineGunRunLeftUp)
			} else {
				p.UpdateCurrentState(PlayerMachineGunLeftUp)
			}
		case LeftDown:
			if p.Running {
				p.UpdateCurrentState(PlayerMachineGunRunLeftDown)
			} else {
				p.UpdateCurrentState(PlayerMachineGunLeftDown)
			}
		case Right:
			if p.Running {
				p.UpdateCurrentState(PlayerMachineGunRunRight)
			} else {
				p.UpdateCurrentState(PlayerMachineGunRight)
			}
		case RightUp:
			if p.Running {
				p.UpdateCurrentState(PlayerMachineGunRunRightUp)
			} else {
				p.UpdateCurrentState(PlayerMachineGunRightUp)
			}
		case RightDown:
			if p.Running {
				p.UpdateCurrentState(PlayerMachineGunRunRightDown)
			} else {
				p.UpdateCurrentState(PlayerMachineGunRightDown)
			}
		}
	case Fists:
		switch p.VisualDir {
		case Left, LeftUp, LeftDown:
			if p.Running {
				p.UpdateCurrentState(PlayerNoGunRunLeft)
			} else {
				p.UpdateCurrentState(PlayerNoGunLeft)
			}
		case Right, RightUp, RightDown:
			if p.Running {
				p.UpdateCurrentState(PlayerNoGunRunRight)
			} else {
				p.UpdateCurrentState(PlayerNoGunRight)
			}
		}

	case Whisky:
		switch p.VisualDir {
		case Left, LeftUp, LeftDown:
			if p.Running {
				p.UpdateCurrentState(PlayerWhiskyRunLeft)
			} else {
				p.UpdateCurrentState(PlayerWhiskyLeft)
			}
		case Right, RightUp, RightDown:
			if p.Running {
				p.UpdateCurrentState(PlayerWhiskyRunRight)
			} else {
				p.UpdateCurrentState(PlayerWhiskyRight)
			}
		}
		p.UpdateDrinkingState()
	}
}

// UpdateDrinkingState shows the drinking pose while the whisky is being
// consumed, the up-facing frames hold the bottle to the player's mouth.
func (p *Player) UpdateDrinkingState() {
	if !p.Healing {
		return
	}

	left := p.VisualDir == Left || p.VisualDir == LeftUp || p.VisualDir == LeftDown
	switch {
	case left && p.Running:
		p.UpdateCurrentState(PlayerWhiskyRunLeftUp)
	case left:
		p.UpdateCurrentState(PlayerWhiskyLeftUp)
	case p.Running:
		p.UpdateCurrentState(PlayerWhiskyRunRightUp)
	default:
		p.UpdateCurrentState(PlayerWhiskyRightUp)
	}
}

func (p *Player) DrawWeapon(beltIndex int) {
	p.CurrentWeapon = beltIndex
	weapon := p.GunBelt[beltIndex].GunType
	switch weapon {
	case Revolver:
		switch p.VisualDir {
		case Left:
			p.UpdateCurrentState(PlayerRevolverLeft)
		case LeftUp:
			p.UpdateCurrentState(PlayerRevolverLeftUp)
		case LeftDown:
			p.UpdateCurrentState(PlayerRevolverLeftDown)
		case Right:
			p.UpdateCurrentState(PlayerRevolverRight)
		case RightUp:
			p.UpdateCurrentState(PlayerRevolverRightUp)
		case RightDown:
			p.UpdateCurrentState(PlayerRevolverRightDown)
		}
	case Shotgun:
		switch p.VisualDir {
		case Left:
			p.UpdateCurrentState(PlayerShotgunLeft)
		case LeftUp:
			p.UpdateCurrentState(PlayerShotgunLeftUp)
		case LeftDown:
			p.UpdateCurrentState(PlayerShotgunLeftDown)
		case Right:
			p.UpdateCurrentState(PlayerShotgunRight)
		case RightUp:
			p.UpdateCurrentState(PlayerShotgunRightUp)
		case RightDown:
			p.UpdateCurrentState(PlayerShotgunRightDown)
		}
	case MachineGun:
		switch p.VisualDir {
		case Left:
			p.UpdateCurrentState(PlayerMachineGunLeft)
		case LeftUp:
			p.UpdateCurrentState(PlayerMachineGunLeftUp)
		case LeftDown:
			p.UpdateCurrentState(PlayerMachineGunLeftDown)
		case Right:
			p.UpdateCurrentState(PlayerMachineGunRight)
		case RightUp:
			p.UpdateCurrentState(PlayerMachineGunRightUp)
		case RightDown:
			p.UpdateCurrentState(PlayerMachineGunRightDown)
		}
	case Fists:
		switch p.VisualDir {
		case Left:
			p.UpdateCurrentState(PlayerNoGunLeft)
		case LeftUp:
			p.UpdateCurrentState(PlayerNoGunLeft)
		case LeftDown:
			p.UpdateCurrentState(PlayerNoGunLeft)
		case Right:
			p.UpdateCurrentState(PlayerNoGunRight)
		case RightUp:
			p.UpdateCurrentState(PlayerNoGunRight)
		case RightDown:
			p.UpdateCurrentState(PlayerNoGunRight)
		}
	case Whisky:
		switch p.VisualDir {
		case Left:
			p.UpdateCurrentState(PlayerWhiskyLeft)
		case LeftUp:
			p.UpdateCurrentState(PlayerWhiskyLeft)
		case LeftDown:
			p.UpdateCurrentState(PlayerWhiskyLeft)
		case Right:
			p.UpdateCurrentState(PlayerWhiskyRight)
		case RightUp:
			p.UpdateCurrentState(PlayerWhiskyRight)
		case RightDown:
			p.UpdateCurrentState(PlayerWhiskyRight)
		}
		p.UpdateDrinkingState()
	}
}

func (p *Player) Animate() {
	switch p.CurrentState {
	case PlayerNoGunRight:
		p.UpdateCurrentState(PlayerNoGunRunRight)
		p.Running = true
	case PlayerNoGunLeft:
		p.UpdateCurrentState(PlayerNoGunRunLeft)
		p.Running = true
	case PlayerNoGunRunRight:
		p.UpdateCurrentState(PlayerNoGunRight)
		p.Running = false
	case PlayerNoGunRunLeft:
		p.UpdateCurrentState(PlayerNoGunLeft)
		p.Running = false
	case PlayerWhiskyRight:
		p.UpdateCurrentState(PlayerWhiskyRunRight)
		p.Running = true
	case PlayerWhiskyLeft:
		p.UpdateCurrentState(PlayerWhiskyRunLeft)
		p.Running = true
	case PlayerWhiskyRunRight:
		p.UpdateCurrentState(PlayerWhiskyRight)
		p.Running = false
	case PlayerWhiskyRunLeft:
		p.UpdateCurrentState(PlayerWhiskyLeft)
		p.Running = false
	case PlayerWhiskyRightUp:
		p.UpdateCurrentState(PlayerWhiskyRunRightUp)
		p.Running = true
	case PlayerWhiskyLeftUp:
		p.UpdateCurrentState(PlayerWhiskyRunLeftUp)
		p.Running = true
	case PlayerWhiskyRunRightUp:
		p.UpdateCurrentState(PlayerWhiskyRightUp)
		p.Running = false
	case PlayerWhiskyRunLeftUp:
		p.UpdateCurrentState(PlayerWhiskyLeftUp)
		p.Running = false
	case PlayerRevolverRight:
		p.UpdateCurrentState(PlayerRevolverRunRight)
		p.Running = true
	case PlayerRevolverRightUp:
		p.UpdateCurrentState(PlayerRevolverRunRightUp)
		p.Running = true
	case PlayerRevolverRightDown:
		p.UpdateCurrentState(PlayerRevolverRunRightDown)
		p.Running = true
	case PlayerRevolverLeft:
		p.UpdateCurrentState(PlayerRevolverRunLeft)
		p.Running = true
	case PlayerRevolverLeftUp:
		p.UpdateCurrentState(PlayerRevolverRunLeftUp)
		p.Running = true
	case PlayerRevolverLeftDown:
		p.UpdateCurrentState(PlayerRevolverRunLeftDown)
		p.Running = true
	case PlayerRevolverRunRight:
		p.UpdateCurrentState(PlayerRevolverRight)
		p.Running = false
	case PlayerRevolverRunRightUp:
		p.UpdateCurrentState(PlayerRevolverRightUp)
		p.Running = false
	case PlayerRevolverRunRightDown:
		p.UpdateCurrentState(PlayerRevolverRightDown)
		p.Running = false
	case PlayerRevolverRunLeft:
		p.UpdateCurrentState(PlayerRevolverLeft)
		p.Running = false
	case PlayerRevolverRunLeftUp:
		p.UpdateCurrentState(PlayerRevolverLeftUp)
		p.Running = false
	case PlayerRevolverRunLeftDown:
		p.UpdateCurrentState(PlayerRevolverLeftDown)
		p.Running = false
	case PlayerShotgunRight:
		p.UpdateCurrentState(PlayerShotgunRunRight)
		p.Running = true
	case PlayerShotgunRightUp:
		p.UpdateCurrentState(PlayerShotgunRunRightUp)
		p.Running = true
	case PlayerShotgunRightDown:
		p.UpdateCurrentState(PlayerShotgunRunRightDown)
		p.Running = true
	case PlayerShotgunLeft:
		p.UpdateCurrentState(PlayerShotgunRunLeft)
		p.Running = true
	case PlayerShotgunLeftUp:
		p.UpdateCurrentState(PlayerShotgunRunLeftUp)
		p.Running = true
	case PlayerShotgunLeftDown:
		p.UpdateCurrentState(PlayerShotgunRunLeftDown)
		p.Running = true
	case PlayerShotgunRunRight:
		p.UpdateCurrentState(PlayerShotgunRight)
		p.Running = false
	case PlayerShotgunRunRightUp:
		p.UpdateCurrentState(PlayerShotgunRightUp)
		p.Running = false
	case PlayerShotgunRunRightDown:
		p.UpdateCurrentState(PlayerShotgunRightDown)
		p.Running = false
	case PlayerShotgunRunLeft:
		p.UpdateCurrentState(PlayerShotgunLeft)
		p.Running = false
	case PlayerShotgunRunLeftUp:
		p.UpdateCurrentState(PlayerShotgunLeftUp)
		p.Running = false
	case PlayerShotgunRunLeftDown:
		p.UpdateCurrentState(PlayerShotgunLeftDown)
		p.Running = false
	case PlayerMachineGunRight:
		p.UpdateCurrentState(PlayerMachineGunRunRight)
		p.Running = true
	case PlayerMachineGunRightUp:
		p.UpdateCurrentState(PlayerMachineGunRunRightUp)
		p.Running = true
	case PlayerMachineGunRightDown:
		p.UpdateCurrentState(PlayerMachineGunRunRightDown)
		p.Running = true
	case PlayerMachineGunLeft:
		p.UpdateCurrentState(PlayerMachineGunRunLeft)
		p.Running = true
	case PlayerMachineGunLeftUp:
		p.UpdateCurrentState(PlayerMachineGunRunLeftUp)
		p.Running = true
	case PlayerMachineGunLeftDown:
		p.UpdateCurrentState(PlayerMachineGunRunLeftDown)
		p.Running = true
	case PlayerMachineGunRunRight:
		p.UpdateCurrentState(PlayerMachineGunRight)
		p.Running = false
	case PlayerMachineGunRunRightUp:
		p.UpdateCurrentState(PlayerMachineGunRightUp)
		p.Running = false
	case PlayerMachineGunRunRightDown:
		p.UpdateCurrentState(PlayerMachineGunRightDown)
		p.Running = false
	case PlayerMachineGunRunLeft:
		p.UpdateCurrentState(PlayerMachineGunLeft)
		p.Running = false
	case PlayerMachineGunRunLeftUp:
		p.UpdateCurrentState(PlayerMachineGunLeftUp)
		p.Running = false
	case PlayerMachineGunRunLeftDown:
		p.UpdateCurrentState(PlayerMachineGunLeftDown)
		p.Running = false
	}
}

func (p *Player) StopAnimation() {
	p.Running = false
	switch p.CurrentState {
	case PlayerMachineGunRunRight:
		p.UpdateCurrentState(PlayerMachineGunRight)
	case PlayerMachineGunRunLeft:
		p.UpdateCurrentState(PlayerMachineGunLeft)
	case PlayerShotgunRunRight:
		p.UpdateCurrentState(PlayerShotgunRight)
	case PlayerShotgunRunLeft:
		p.UpdateCurrentState(PlayerShotgunLeft)
	case PlayerRevolverRunRight:
		p.UpdateCurrentState(PlayerRevolverRight)
	case PlayerRevolverRunLeft:
		p.UpdateCurrentState(PlayerRevolverLeft)
	case PlayerNoGunRunRight:
		p.UpdateCurrentState(PlayerNoGunRight)
	case PlayerNoGunRunLeft:
		p.UpdateCurrentState(PlayerNoGunLeft)
	case PlayerWhiskyRunRight:
		p.UpdateCurrentState(PlayerWhiskyRight)
	case PlayerWhiskyRunLeft:
		p.UpdateCurrentState(PlayerWhiskyLeft)
	case PlayerWhiskyRunRightUp:
		p.UpdateCurrentState(PlayerWhiskyRightUp)
	case PlayerWhiskyRunLeftUp:
		p.UpdateCurrentState(PlayerWhiskyLeftUp)
	}
}

func (p *Player) addBullet(bulletSprite *ebiten.Image, bulletSpeed float64, bulletRotation float64, duration int, damage int) {
	scale := float64(4)
	bounds := bulletSprite.Bounds()
	bw := float64(bounds.Dx())
	bh := float64(bounds.Dy())
	// The middle of the bullet starts at the middle of the player
	x := p.X - bw*scale/2 + p.W/2
	y := p.Y - bh*scale/2 + p.H/2
	bulletPixelWidth := 3.0
	bulletPixelHeight := 3.0
	offset := float64(14)
	bullet := Bullet{
		X:        x,
		Y:        y,
		W:        bw,
		H:        bh,
		R:        bulletRotation,
		Scale:    scale,
		Speed:    bulletSpeed,
		Sprite:   bulletSprite,
		Damage:   1 + rand.IntN(damage),
		Duration: duration,
		Hitbox: HitBox{
			X: float32(x + offset*scale),
			Y: float32(y + offset*scale),
			W: float32(bulletPixelWidth * scale),
			H: float32(bulletPixelHeight * scale),
		},
		HitboxOffset: offset,
	}

	p.Bullets = append(p.Bullets, bullet)
}

// drawAimCursor draws a small cross reticle at a fixed distance from the actor
// along the aim angle, it shows where the bullets go without cluttering the
// middle of the screen.
func drawAimCursor(screen *ebiten.Image, cx, cy, angle float64, clr color.RGBA) {
	const (
		distance = 52.0
		size     = 7.0
		width    = 2.0
	)
	ax := cx + math.Cos(angle)*distance
	ay := cy + math.Sin(angle)*distance
	vector.StrokeLine(screen, float32(ax-size), float32(ay), float32(ax+size), float32(ay), width, clr, true)
	vector.StrokeLine(screen, float32(ax), float32(ay-size), float32(ax), float32(ay+size), width, clr, true)
}

// drawMuzzleFlash draws a brief flash at the tip of the gun, it gives the
// shots some punch and shows the player when an enemy pulls the trigger.
func drawMuzzleFlash(screen *ebiten.Image, cx, cy, angle float64) {
	const distance = 26.0

	angle = math.Mod(angle, 2*math.Pi)
	if angle < 0 {
		angle += 2 * math.Pi
	}
	switch {
	case angle < math.Pi/4:
		// dir = Right
		angle = 2 * math.Pi
	case angle < math.Pi/2:
		// dir = RightDown
		angle = 3 * math.Pi / 8
	case angle < 3*math.Pi/4:
		// dir = LeftDown
		angle = 5 * math.Pi / 8
	case angle < 5*math.Pi/4:
		// dir = Left
		angle = math.Pi
	case angle < 3*math.Pi/2:
		// dir = LeftUp
		angle = 23 * math.Pi / 16
	case angle < 7*math.Pi/4:
		// dir = RightUp
		angle = 25 * math.Pi / 16
	default:
		// dir = Right
		angle = 2 * math.Pi
	}

	x := cx + math.Cos(angle)*distance
	y := cy + math.Sin(angle)*distance
	vector.FillCircle(screen, float32(x), float32(y), 6, color.RGBA{255, 220, 110, 200}, true)
	vector.FillCircle(screen, float32(x), float32(y), 3, color.RGBA{255, 255, 230, 255}, true)
}

// drinkWhisky starts healing the player over the shoot cooldown, the damage of
// the whisky is the amount of health it restores.
func (p *Player) drinkWhisky(gun *Gun) {
	if !p.CanShoot() {
		return
	}

	heal := min(gun.Damage, p.MaxHealth-p.Health)
	if heal <= 0 {
		return
	}

	gun.Ammo -= 1
	p.ShootTimer = gun.ShootCooldown
	p.Healing = true
	p.HealTimer = gun.ShootCooldown
	p.HealDuration = gun.ShootCooldown
	p.HealTotal = heal
	p.HealApplied = 0
	p.UpdateDrinkingState()
}

// updateHealing spreads the health gain evenly over the duration of the drink
// and shows the restored amount once the cooldown ran out.
func (p *Player) updateHealing() {
	if p.HealDuration <= 0 {
		p.finishHealing()

		return
	}

	if p.HealTimer > 0 {
		p.HealTimer -= 1
	}
	elapsed := p.HealDuration - p.HealTimer

	if elapsed >= p.HealDuration {
		p.finishHealing()
	}
}

func (p *Player) finishHealing() {
	heal := p.HealTotal
	if heal > 0 {
		p.Health += heal
		p.HealApplied += heal
		p.UpdateHealhBar()
	}
	healed := p.HealApplied
	p.Healing = false
	p.HealTimer = 0
	p.HealDuration = 0
	p.HealTotal = 0
	p.HealApplied = 0

	if healed <= 0 {
		return
	}

	gun := p.GunBelt[p.CurrentWeapon]
	if gun.AmmoFace == nil {
		return
	}
	hit := Hit{
		X:        p.X,
		Y:        p.Y - p.H/2,
		Color:    color.RGBA{90, 220, 100, 240},
		Msg:      "+" + strconv.Itoa(healed),
		TextFont: gun.AmmoFace,
		Duration: 30,
	}
	hit.SetDrawOptions()
	p.Hits = append(p.Hits, hit)
}
