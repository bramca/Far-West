package actors

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type Gun struct {
	Damage              int
	MagazineSize        int
	CurrentMagazineSize int
	Ammo                int
	ShootCooldown       int
	GunType             Weapon
	BulletDuration      int
	BulletSpread        int
	BulletCount         int
	BulletSpeed         float64
	MuzzleFlash         int

	X, Y        float64
	Sprite      *ebiten.Image
	DrawOptions *ebiten.DrawImageOptions
	Scale       float64
	AmmoFace    *text.GoXFace
}

func (g *Gun) Draw(screen *ebiten.Image, camX, camY float64) {
	g.DrawOptions.GeoM.Reset()
	g.DrawOptions.GeoM.Scale(g.Scale, g.Scale)
	g.DrawOptions.GeoM.Translate(float64(g.X-camX), float64(g.Y-camY))
	screen.DrawImage(g.Sprite, g.DrawOptions)
	drawOptions := &text.DrawOptions{}
	drawOptions.GeoM.Translate(float64(g.X+32), g.Y+16)
	text.Draw(screen, fmt.Sprintf("%d", g.Ammo), g.AmmoFace, drawOptions)
}
