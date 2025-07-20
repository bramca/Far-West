package world

import (
	"github.com/bramca/Far-West/actors"
	"github.com/hajimehoshi/ebiten/v2"
)

type Cactus struct {
	X, Y        float64
	W, H        float64
	Sprite      *ebiten.Image
	DrawOptions *ebiten.DrawImageOptions
	Scale       float64
	Hitbox      *actors.HitBox
}

func (c *Cactus) Draw(screen *ebiten.Image, camX float64, camY float64) {
	c.DrawOptions.GeoM.Reset()
	c.DrawOptions.GeoM.Scale(c.Scale, c.Scale)
	c.DrawOptions.GeoM.Translate(float64(c.X-camX), float64(c.Y-camY))
	screen.DrawImage(c.Sprite, c.DrawOptions)
}

func (c *Cactus) DrawHitbox(screen *ebiten.Image, camX float64, camY float64) {
	c.Hitbox.Draw(screen, camX, camY)
}
