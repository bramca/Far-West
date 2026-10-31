package world

import (
	"github.com/bramca/Far-West/actors"
	"github.com/hajimehoshi/ebiten/v2"
)

type Plant struct {
	X, Y        float64
	W, H        float64
	Sprite      *ebiten.Image
	DrawOptions *ebiten.DrawImageOptions
	Scale       float64
	Hitbox      *actors.HitBox
}

func (p *Plant) Draw(screen *ebiten.Image, camX float64, camY float64) {
	p.DrawOptions.GeoM.Reset()
	p.DrawOptions.GeoM.Scale(p.Scale, p.Scale)
	p.DrawOptions.GeoM.Translate(float64(p.X-camX), float64(p.Y-camY))
	screen.DrawImage(p.Sprite, p.DrawOptions)
}

func (p *Plant) DrawHitbox(screen *ebiten.Image, camX float64, camY float64) {
	p.Hitbox.Draw(screen, camX, camY)
}
