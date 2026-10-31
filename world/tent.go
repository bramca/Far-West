package world

import (
	"github.com/bramca/Far-West/actors"
	"github.com/hajimehoshi/ebiten/v2"
)

type Tent struct {
	X, Y        float64
	W, H        float64
	Sprite      *ebiten.Image
	DrawOptions *ebiten.DrawImageOptions
	Scale       float64
	Hitbox      *actors.HitBox
}

func (t *Tent) Draw(screen *ebiten.Image, camX float64, camY float64) {
	t.DrawOptions.GeoM.Reset()
	t.DrawOptions.GeoM.Scale(t.Scale, t.Scale)
	t.DrawOptions.GeoM.Translate(float64(t.X-camX), float64(t.Y-camY))
	screen.DrawImage(t.Sprite, t.DrawOptions)
}

func (t *Tent) DrawHitbox(screen *ebiten.Image, camX float64, camY float64) {
	t.Hitbox.Draw(screen, camX, camY)
}
