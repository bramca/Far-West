package actors

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Loot is a gun lying on the floor, waiting to be picked up.
type Loot struct {
	Gun    *Gun
	X, Y   float64
	VX, VY float64
	timer  int
}

// NewLoot creates a gun pickup at the given world position.
func NewLoot(gun *Gun, x, y float64) *Loot {
	return &Loot{Gun: gun, X: x, Y: y}
}

// NewTossedLoot creates a gun pickup that slides in the given direction and
// slows down until it settles.
func NewTossedLoot(gun *Gun, x, y, vx, vy float64) *Loot {
	loot := NewLoot(gun, x, y)
	loot.VX, loot.VY = vx, vy

	return loot
}

// Update advances the small bob animation and the toss of the loot.
func (l *Loot) Update() {
	l.timer++
	l.X += l.VX
	l.Y += l.VY
	l.VX *= 0.86
	l.VY *= 0.86
	if math.Abs(l.VX) < 0.05 {
		l.VX = 0
	}
	if math.Abs(l.VY) < 0.05 {
		l.VY = 0
	}
}

// Draw draws the loot on the ground, it gently bobs so it stands out from the
// world sprites.
func (l *Loot) Draw(screen *ebiten.Image, camX, camY float64) {
	bob := math.Sin(float64(l.timer)/12) * 3
	l.DrawIcon(screen, l.X-camX, l.Y-camY+bob, l.Gun.Scale)
}

// DrawIcon draws the gun sprite centered on the given screen position with a
// ribbon showing its tier.
func (l *Loot) DrawIcon(screen *ebiten.Image, x, y, scale float64) {
	if l.Gun == nil || l.Gun.Sprite == nil {
		return
	}

	width := float64(l.Gun.Sprite.Bounds().Dx()) * scale
	height := float64(l.Gun.Sprite.Bounds().Dy()) * scale
	drawOptions := &ebiten.DrawImageOptions{}
	drawOptions.GeoM.Scale(scale, scale)
	drawOptions.GeoM.Translate(x-width/2, y-height/2)
	screen.DrawImage(l.Gun.Sprite, drawOptions)

	const ribbonHeight = 4.0
	vector.FillRect(
		screen,
		float32(x-width/2), float32(y+height/2),
		float32(width), ribbonHeight,
		TierColor(l.Gun.Tier),
		false,
	)
}
