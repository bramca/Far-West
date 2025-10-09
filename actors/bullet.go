package actors

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Bullet struct {
	X, Y         float64
	W, H         float64
	R            float64
	Speed        float64
	Scale        float64
	Duration     int
	DrawOptions  *ebiten.DrawImageOptions
	Sprite       *ebiten.Image
	Hitbox       *HitBox
	HitboxOffset float64
}

func (b *Bullet) Draw(screen *ebiten.Image, camX float64, camY float64) {
	b.DrawOptions.GeoM.Reset()
	b.DrawOptions.GeoM.Translate(-b.W/2, -b.H/2)
	b.DrawOptions.GeoM.Rotate(b.R)
	b.DrawOptions.GeoM.Translate(b.W/2, b.H/2)
	b.DrawOptions.GeoM.Scale(b.Scale, b.Scale)
	b.DrawOptions.GeoM.Translate(float64(b.X-camX), float64(b.Y-camY))
	screen.DrawImage(b.Sprite, b.DrawOptions)
}

func (b *Bullet) Update() {
	b.X += b.Speed * math.Cos(b.R)
	b.Y += b.Speed * math.Sin(b.R)

	b.UpdateHitbox()
	b.Duration -= 1
}

func (b *Bullet) DrawHitbox(screen *ebiten.Image, camX, camY float64) {
	b.Hitbox.Draw(screen, camX, camY)
}

func (b *Bullet) UpdateHitbox() {
	b.Hitbox.X = float32(b.X + b.HitboxOffset*b.Scale)
	b.Hitbox.Y = float32(b.Y + b.HitboxOffset*b.Scale)
}
