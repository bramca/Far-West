package actors

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type Tier int

const (
	Common Tier = iota
	Rare
	Legendary
)

// String returns the display name of the tier.
func (t Tier) String() string {
	switch t {
	case Rare:
		return "RARE"
	case Legendary:
		return "LEGENDARY"
	default:
		return "COMMON"
	}
}

// TierColor returns the color that represents the tier, common guns use the
// default color while rare and legendary guns stand out.
func TierColor(tier Tier) color.RGBA {
	switch tier {
	case Rare:
		return color.RGBA{70, 140, 235, 240}
	case Legendary:
		return color.RGBA{245, 205, 55, 240}
	default:
		return color.RGBA{205, 195, 170, 230}
	}
}

// String returns the display name of the weapon.
func (w Weapon) String() string {
	switch w {
	case Fists:
		return "FISTS"
	case Whisky:
		return "WHISKY"
	case Shotgun:
		return "SHOTGUN"
	case MachineGun:
		return "MACHINE GUN"
	default:
		return "REVOLVER"
	}
}

type Gun struct {
	Tier                Tier
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
	if g == nil || g.Sprite == nil {
		return
	}
	if g.DrawOptions == nil {
		g.DrawOptions = &ebiten.DrawImageOptions{}
	}
	g.DrawOptions.GeoM.Reset()
	g.DrawOptions.GeoM.Scale(g.Scale, g.Scale)
	g.DrawOptions.GeoM.Translate(float64(g.X-camX), float64(g.Y-camY))
	screen.DrawImage(g.Sprite, g.DrawOptions)
	if g.AmmoFace != nil {
		drawOptions := &text.DrawOptions{}
		drawOptions.GeoM.Translate(float64(g.X+32), g.Y+16)
		text.Draw(screen, fmt.Sprintf("%d", g.Ammo), g.AmmoFace, drawOptions)
	}
}
