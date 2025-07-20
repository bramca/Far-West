package actors

import "github.com/hajimehoshi/ebiten/v2"

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
)

type Player struct {
	X, Y           float64
	W, H           float64
	Sprites        []*ebiten.Image
	CurrentState   PlayerState
	Scale          float64
	Speed          float64
	AnimationSpeed int
	DrawOptions    *ebiten.DrawImageOptions
	VisualDir      Direction
	CurrentWeapon  Weapon
	Hitbox         *HitBox
}

func (p *Player) Draw(screen *ebiten.Image, camX float64, camY float64) {
	// Draw the player
	p.DrawOptions.GeoM.Reset()
	p.DrawOptions.GeoM.Scale(p.Scale, p.Scale)
	p.DrawOptions.GeoM.Translate(-float64(p.W/2), -float64(p.H/2))
	p.DrawOptions.GeoM.Translate(p.X-camX, p.Y-camY)
	screen.DrawImage(p.Sprites[p.CurrentState], p.DrawOptions)
}

func (p *Player) DrawHitbox(screen *ebiten.Image, camX, camY float64) {
	p.Hitbox.Draw(screen, camX, camY)
}

func (p *Player) UpdateHitbox() {
	p.Hitbox.X = float32(p.X)
	p.Hitbox.Y = float32(p.Y)
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
	case PlayerNoGunLeft:
		p.UpdateCurrentState(PlayerNoGunRunLeft)
	case PlayerNoGunRunRight:
		p.UpdateCurrentState(PlayerNoGunRight)
	case PlayerNoGunRunLeft:
		p.UpdateCurrentState(PlayerNoGunLeft)
	case PlayerRevolverRight:
		p.UpdateCurrentState(PlayerRevolverRunRight)
	case PlayerRevolverRightUp:
		p.UpdateCurrentState(PlayerRevolverRunRightUp)
	case PlayerRevolverRightDown:
		p.UpdateCurrentState(PlayerRevolverRunRightDown)
	case PlayerRevolverLeft:
		p.UpdateCurrentState(PlayerRevolverRunLeft)
	case PlayerRevolverLeftUp:
		p.UpdateCurrentState(PlayerRevolverRunLeftUp)
	case PlayerRevolverLeftDown:
		p.UpdateCurrentState(PlayerRevolverRunLeftDown)
	case PlayerRevolverRunRight:
		p.UpdateCurrentState(PlayerRevolverRight)
	case PlayerRevolverRunRightUp:
		p.UpdateCurrentState(PlayerRevolverRightUp)
	case PlayerRevolverRunRightDown:
		p.UpdateCurrentState(PlayerRevolverRightDown)
	case PlayerRevolverRunLeft:
		p.UpdateCurrentState(PlayerRevolverLeft)
	case PlayerRevolverRunLeftUp:
		p.UpdateCurrentState(PlayerRevolverLeftUp)
	case PlayerRevolverRunLeftDown:
		p.UpdateCurrentState(PlayerRevolverLeftDown)
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
