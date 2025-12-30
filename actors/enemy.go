package actors

import (
	"math/rand"

	"github.com/bramca/Far-West/utils"
	"github.com/hajimehoshi/ebiten/v2"
)

type Enemy struct {
	*Player

	CurrentAction Action
	VisualDist    int
	MoveSpeed     int
	ShootSpeed    int
}

func (e *Enemy) Draw(screen *ebiten.Image, camX float64, camY float64) {
	e.DrawOptions.GeoM.Reset()
	e.DrawOptions.GeoM.Scale(e.Scale, e.Scale)
	e.DrawOptions.GeoM.Translate(float64(e.X-camX), float64(e.Y-camY))
	screen.DrawImage(e.Sprites[e.CurrentState], e.DrawOptions)
}

func (e *Enemy) UpdateHitboxOffset(offset int) {
	e.Hitbox.X = float32(e.X) + float32(offset)
	e.Hitbox.Y = float32(e.Y) + float32(offset)
}

func (e *Enemy) ThinkAndAct(player *Player, playerBullets []*Bullet, frameCount int) {
	// Detect player
	e.CurrentAction.Duration -= 1
	if utils.DistanceBetweenPoints(player.X, player.Y, e.X, e.Y) <= float64(e.VisualDist) && e.CurrentAction.Type != MoveAndShoot && e.CurrentAction.Duration <= 0 {
		actionType := MoveAndShoot
		if rand.Float64() < 0.5 {
			actionType = Dodge
		}
		e.CurrentAction = Action{
			Duration: 120 + rand.Intn(240),
			Type:     actionType,
			actor:    e,
		}
	}

	if e.CurrentAction.Duration <= 0 {
		e.StopAnimation()
		actionType := Move
		if rand.Float64() < 0.5 {
			actionType = Dodge
		}
		dirs := []Direction{
			Up,
			Down,
			Left,
			Right,
		}
		e.CurrentAction = Action{
			Duration: 120 + rand.Intn(240),
			Type:     actionType,
			MoveDir:  dirs[rand.Intn(len(dirs))],
			LookDir:  dirs[rand.Intn(len(dirs))],
			actor:    e,
		}
	}

	e.CurrentAction.PerformAction(player, frameCount)
}
