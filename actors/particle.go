package actors

import (
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Particle is a tiny short lived speck of sparks or dust, it gives impacts
// and kills a bit of visual weight.
type Particle struct {
	X, Y    float64
	VX, VY  float64
	Life    int
	MaxLife int
	Color   color.RGBA
}

// NewParticle spawns a particle at the given position that flies along the
// given velocity and fades out over its lifetime.
func NewParticle(x, y, vx, vy float64, clr color.RGBA) *Particle {
	life := 14 + rand.Intn(8)
	return &Particle{
		X:       x,
		Y:       y,
		VX:      vx,
		VY:      vy,
		Life:    life,
		MaxLife: life,
		Color:   clr,
	}
}

func (p *Particle) Update() {
	p.X += p.VX
	p.Y += p.VY
	p.VY += 0.12
	p.Life--
}

func (p *Particle) Draw(screen *ebiten.Image, camX, camY float64) {
	clr := p.Color
	clr.A = uint8(float64(clr.A) * float64(p.Life) / float64(p.MaxLife))
	size := 2 + float32(p.Life)/float32(p.MaxLife)*2
	vector.FillRect(screen, float32(p.X-camX), float32(p.Y-camY), size, size, clr, false)
}
