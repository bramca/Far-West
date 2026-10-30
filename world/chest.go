package world

import (
	"github.com/bramca/Far-West/actors"
	"github.com/hajimehoshi/ebiten/v2"
)

type ChestState int

type ChestType int

const (
	Gold ChestType = iota
	Gun
	Ammo
)

const (
	Closed ChestState = iota
	Open
)

type Chest struct {
	X, Y        float64
	W, H        float64
	State       ChestState
	Sprites     []*ebiten.Image
	DrawOptions *ebiten.DrawImageOptions
	Scale       float64
	Hitbox      *actors.HitBox
	Loot        actors.Loot
	Type        ChestType
}

func (c *Chest) Draw(screen *ebiten.Image, camX float64, camY float64) {
	c.DrawOptions.GeoM.Reset()
	c.DrawOptions.GeoM.Scale(c.Scale, c.Scale)
	c.DrawOptions.GeoM.Translate(float64(c.X-camX), float64(c.Y-camY))
	screen.DrawImage(c.Sprites[c.State], c.DrawOptions)
}

func (c *Chest) DrawHitbox(screen *ebiten.Image, camX float64, camY float64) {
	c.Hitbox.Draw(screen, camX, camY)
}
