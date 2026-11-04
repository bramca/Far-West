package world

import (
	"github.com/bramca/Far-West/actors"
	"github.com/hajimehoshi/ebiten/v2"
)

// Jail is the town building where the player can read the wanted poster and
// collect the reward for a bounty that was brought in alive.
type Jail struct {
	X, Y        float64
	W, H        float64
	Sprite      *ebiten.Image
	DrawOptions *ebiten.DrawImageOptions
	Scale       float64
	Hitbox      *actors.HitBox
}

func (j *Jail) Draw(screen *ebiten.Image, camX float64, camY float64) {
	j.DrawOptions.GeoM.Reset()
	j.DrawOptions.GeoM.Scale(j.Scale, j.Scale)
	j.DrawOptions.GeoM.Translate(j.X-camX, j.Y-camY)
	screen.DrawImage(j.Sprite, j.DrawOptions)
}

func (j *Jail) DrawHitbox(screen *ebiten.Image, camX float64, camY float64) {
	j.Hitbox.Draw(screen, camX, camY)
}
