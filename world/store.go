package world

import (
	"github.com/bramca/Far-West/actors"
	"github.com/hajimehoshi/ebiten/v2"
)

type Store struct {
	X, Y        float64
	W, H        float64
	Sprite      *ebiten.Image
	DrawOptions *ebiten.DrawImageOptions
	Scale       float64
	Hitbox      *actors.HitBox
}

func (s *Store) Draw(screen *ebiten.Image, camX float64, camY float64) {
	s.DrawOptions.GeoM.Reset()
	s.DrawOptions.GeoM.Scale(s.Scale, s.Scale)
	s.DrawOptions.GeoM.Translate(float64(s.X-camX), float64(s.Y-camY))
	screen.DrawImage(s.Sprite, s.DrawOptions)
	s.DrawHitbox(screen, camX, camY)
}

func (s *Store) DrawHitbox(screen *ebiten.Image, camX float64, camY float64) {
	s.Hitbox.Draw(screen, camX, camY)
}
