package actors

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
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
	Revolver
)

const (
	PlayerNoGunRight PlayerState = iota
	PlayerNoGunLeft
	PlayerNoGunRunRight
	PlayerNoGunRunLeft
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
	PlayerDead
)

type Player struct {
	X, Y           float64
	W, H           float64
	Sprites        []*ebiten.Image
	CurrentState   PlayerState
	Scale          float64
	Speed          float64
	DodgeSpeed     float64
	AnimationSpeed int
	DrawOptions    *ebiten.DrawImageOptions
	VisualDir      Direction
	MoveDirs       map[Direction]bool
	CurrentWeapon  Weapon
	Hitbox         *HitBox
	Bullets        []*Bullet
	BulletSprite   *ebiten.Image
	FireRate       int
	Healthbar      *HealthBar
	Health         int
	MaxHealth      int
	IsNpc          bool
	Running        bool
	Hits           []Hit
	Dead           bool
}

func (p *Player) Draw(screen *ebiten.Image, camX, camY float64) {
	// Draw the player
	p.DrawOptions.GeoM.Reset()
	p.DrawOptions.GeoM.Scale(p.Scale, p.Scale)
	p.DrawOptions.GeoM.Translate(-float64(p.W/2), -float64(p.H/2))
	p.DrawOptions.GeoM.Translate(p.X-camX, p.Y-camY)
	screen.DrawImage(p.Sprites[p.CurrentState], p.DrawOptions)
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

func (p *Player) UpdateHitbox() {
	p.Hitbox.X = float32(p.X)
	p.Hitbox.Y = float32(p.Y)
}

func (p *Player) Shoot() {
	if p.CurrentWeapon == Fists {
		return
	}
	switch p.CurrentWeapon {
	case Revolver:
		bulletSpeed := 4.0
		bulletDuration := 500
		bulletDamage := 3
		bulletDirection := map[Direction]float64{
			Right:     0,
			LeftUp:    3 * math.Pi / 2,
			RightUp:   3 * math.Pi / 2,
			Left:      math.Pi,
			LeftDown:  math.Pi / 2,
			RightDown: math.Pi / 2,
		}
		p.addBullet(p.BulletSprite, bulletSpeed, bulletDirection[p.VisualDir], bulletDuration, bulletDamage)
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
	if p.IsNpc {
		p.Healthbar.Update(p.X, p.Y-(p.H-p.H/3), p.Health, p.MaxHealth)
	}
	if !p.IsNpc {
		p.Healthbar.Update(p.Healthbar.X, p.Healthbar.Y, p.Health, p.MaxHealth)
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

func (p *Player) UpdateBullets() {
	toRemove := []int{}
	for i, bullet := range p.Bullets {
		bullet.Update()
		if bullet.Duration < 1 {
			toRemove = append(toRemove, i)
		}
	}

	for _, index := range toRemove {
		p.Bullets = removeFromBullets(p.Bullets, index)
	}
}

func (p *Player) DrawBullets(screen *ebiten.Image, camX, camY float64) {
	for _, bullet := range p.Bullets {
		bullet.Draw(screen, camX, camY)
		// bullet.DrawHitbox(screen, camX, camY)
	}
}

func (p *Player) UpdateCurrentState(newState PlayerState) {
	p.CurrentState = newState
}

func (p *Player) ChangeVisualDirection(newDir Direction) {
	p.VisualDir = newDir
	switch p.CurrentWeapon {
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
	case Fists:
		switch p.VisualDir {
		case Left:
			p.UpdateCurrentState(PlayerNoGunLeft)
		case Right:
			p.UpdateCurrentState(PlayerNoGunRight)
		}
	}
}

func (p *Player) DrawWeapon(weapon Weapon) {
	switch weapon {
	case Revolver:
		p.CurrentWeapon = Revolver
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
	case Fists:
		p.CurrentWeapon = Fists
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
	}
}

func (p *Player) StopAnimation() {
	switch p.CurrentState {
	case PlayerRevolverRunRight:
		p.UpdateCurrentState(PlayerRevolverRight)
	case PlayerRevolverRunLeft:
		p.UpdateCurrentState(PlayerRevolverLeft)
	case PlayerNoGunRunRight:
		p.UpdateCurrentState(PlayerNoGunRight)
	case PlayerNoGunRunLeft:
		p.UpdateCurrentState(PlayerNoGunLeft)
	}
}

func removeFromBullets(bullets []*Bullet, index int) []*Bullet {
	return append(bullets[:index], bullets[index+1:]...)
}

func (p *Player) addBullet(bulletSprite *ebiten.Image, bulletSpeed float64, bulletRotation float64, duration int, damage int) {
	scale := float64(4)
	// The middle of the bullet starts at the middle of the player
	x := p.X - float64(bulletSprite.Bounds().Dx())*scale/2 + p.W/2
	y := p.Y - float64(bulletSprite.Bounds().Dy())*scale/2 + p.H/2
	bulletPixelWidth := 3.0
	bulletPixelHeight := 3.0
	offset := float64(14)
	bullet := &Bullet{
		X:           x,
		Y:           y,
		W:           float64(bulletSprite.Bounds().Dx()),
		H:           float64(bulletSprite.Bounds().Dy()),
		R:           bulletRotation,
		DrawOptions: &ebiten.DrawImageOptions{},
		Scale:       scale,
		Speed:       bulletSpeed,
		Sprite:      bulletSprite,
		Damage:      rand.Intn(damage + 1),
		Duration:    duration,
		Hitbox: &HitBox{
			X: float32(x + offset*scale),
			Y: float32(y + offset*scale),
			W: float32(bulletPixelWidth * scale),
			H: float32(bulletPixelHeight * scale),
		},
		HitboxOffset: offset,
	}

	p.Bullets = append(p.Bullets, bullet)
}
