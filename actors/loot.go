package actors

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Loot is a gun lying on the floor, waiting to be picked up.
type Loot struct {
	Gun   *Gun
	X, Y  float64
	timer int
}

// NewLoot creates a gun pickup at the given world position.
func NewLoot(gun *Gun, x, y float64) *Loot {
	return &Loot{Gun: gun, X: x, Y: y}
}

// Update advances the small bob animation of the loot.
func (l *Loot) Update() {
	l.timer++
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
