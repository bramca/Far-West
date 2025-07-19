package world

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Cactus struct {
	X, Y        float64
	W, H        float64
	Sprite      *ebiten.Image
	DrawOptions *ebiten.DrawImageOptions
	Scale       float64
}

func (c *Cactus) Draw(screen *ebiten.Image, camX float64, camY float64) {
	c.DrawOptions.GeoM.Reset()
	c.DrawOptions.GeoM.Scale(c.Scale, c.Scale)
	c.DrawOptions.GeoM.Translate(float64(c.X-camX), float64(c.Y-camY))
	screen.DrawImage(c.Sprite, c.DrawOptions)
}
