package actors

import "github.com/hajimehoshi/ebiten/v2"

type PlayerState int

const (
	PlayerNoGun PlayerState = iota
	PlayerNoGunRun
	PlayerRevolver
	PlayerRevolverRun
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
}

func (p *Player) Draw(screen *ebiten.Image, x float64, y float64) {
	// Draw the player
	p.DrawOptions.GeoM.Reset()
	p.DrawOptions.GeoM.Scale(p.Scale, p.Scale)
	p.DrawOptions.GeoM.Translate(-float64(p.W/2), -float64(p.H/2))
	p.DrawOptions.GeoM.Translate(x, y)
	screen.DrawImage(p.Sprites[p.CurrentState], p.DrawOptions)
}

func (p *Player) UpdateCurrentState(newState PlayerState) {
	p.CurrentState = newState
}

func (p *Player) Animate() {
	switch p.CurrentState {
	case PlayerNoGun:
		p.UpdateCurrentState(PlayerNoGunRun)
	case PlayerNoGunRun:
		p.UpdateCurrentState(PlayerNoGun)
	case PlayerRevolver:
		p.UpdateCurrentState(PlayerRevolverRun)
	case PlayerRevolverRun:
		p.UpdateCurrentState(PlayerRevolver)
	}
}

func (p *Player) StopAnimation() {
	switch p.CurrentState {
	case PlayerRevolverRun:
		p.UpdateCurrentState(PlayerRevolver)
	case PlayerNoGunRun:
		p.UpdateCurrentState(PlayerNoGun)
	}
}
