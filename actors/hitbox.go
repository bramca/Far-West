package actors

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type HitBox struct {
	X, Y float32
	W, H float32
}

func (h *HitBox) Draw(screen *ebiten.Image, camX float64, camY float64) {
	vector.StrokeRect(screen, h.X-float32(camX), h.Y-float32(camY), h.W, h.H, 2.0, color.Black, false)
}

func (h *HitBox) CheckCollision(hitbox *HitBox) bool {
	return h.X+h.W >= hitbox.X && // h right edge past hitbox left
		h.X <= hitbox.X+hitbox.W && // h left edge past hitbox right
		h.Y+h.H >= hitbox.Y && // h top edge past hitbox bottom
		h.Y <= hitbox.Y+hitbox.H
}

// CenterX returns the horizontal center of the hitbox.
func (h *HitBox) CenterX() float64 {
	return float64(h.X + h.W/2)
}

// CenterY returns the vertical center of the hitbox.
func (h *HitBox) CenterY() float64 {
	return float64(h.Y + h.H/2)
}
