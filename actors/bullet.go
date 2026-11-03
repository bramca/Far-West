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
	Damage       int
	Scale        float64
	Duration     int
	Sprite       *ebiten.Image
	Hitbox       HitBox
	HitboxOffset float64
	// TimeScale slows the bullet down without changing how far it travels.
	// A value of zero means the bullet moves at its normal speed.
	TimeScale float64
	// durationAccumulator keeps the lifetime in sync with the time scale so a
	// slowed bullet also lives longer on screen.
	durationAccumulator float64
}

// scale returns the effective time scale of the bullet.
func (b *Bullet) scale() float64 {
	if b.TimeScale <= 0 {
		return 1
	}

	return b.TimeScale
}

func (b *Bullet) Draw(screen *ebiten.Image, camX float64, camY float64) {
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(-b.W/2, -b.H/2)
	op.GeoM.Rotate(b.R)
	op.GeoM.Translate(b.W/2, b.H/2)
	op.GeoM.Scale(b.Scale, b.Scale)
	op.GeoM.Translate(b.X-camX, b.Y-camY)
	screen.DrawImage(b.Sprite, &op)
}

func (b *Bullet) Update() {
	scale := b.scale()
	b.X += b.Speed * scale * math.Cos(b.R)
	b.Y += b.Speed * scale * math.Sin(b.R)

	b.durationAccumulator += scale
	for b.durationAccumulator >= 1 {
		b.Duration -= 1
		b.durationAccumulator -= 1
	}

	b.UpdateHitbox()
}

func (b *Bullet) DrawHitbox(screen *ebiten.Image, camX, camY float64) {
	b.Hitbox.Draw(screen, camX, camY)
}

func (b *Bullet) UpdateHitbox() {
	b.Hitbox.X = float32(b.X + b.HitboxOffset*b.Scale)
	b.Hitbox.Y = float32(b.Y + b.HitboxOffset*b.Scale)
}
