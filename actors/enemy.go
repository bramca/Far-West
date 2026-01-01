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
}

func (e *Enemy) Draw(screen *ebiten.Image, camX float64, camY float64) {
	e.DrawOptions.GeoM.Reset()
	e.DrawOptions.GeoM.Scale(e.Scale, e.Scale)
	e.DrawOptions.GeoM.Translate(float64(e.X-camX), float64(e.Y-camY))
	screen.DrawImage(e.Sprites[e.CurrentState], e.DrawOptions)
	e.Healthbar.Draw(screen, camX-16, camY-16)
	for i := len(e.Hits) - 1; i >= 0; i-- {
		if e.Hits[i].Duration > 0 {
			e.Hits[i].Update()
			e.Hits[i].Draw(screen, camX, camY)
		} else {
			e.Hits[i] = e.Hits[len(e.Hits)-1]
			e.Hits = e.Hits[:len(e.Hits)-1]
		}
	}
}

func (e *Enemy) UpdateHitboxOffset(offset int) {
	e.Hitbox.X = float32(e.X) + float32(offset)
	e.Hitbox.Y = float32(e.Y) + float32(offset)
}

func (e *Enemy) Move(d Direction) {
	e.MoveDirs[d] = true
	switch d {
	case Right:
		e.X += e.Speed
	case Left:
		e.X -= e.Speed
	case Up:
		e.Y -= e.Speed
	case Down:
		e.Y += e.Speed
	}
	e.UpdateHitboxOffset(16)
	e.Healthbar.Update(e.X-e.W/2, e.Y-(e.H-e.H/3), e.Health, e.MaxHealth)
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
			actor:    e.Player,
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
			actor:    e.Player,
		}
	}

	e.CurrentAction.PerformAction(player, frameCount)
	e.UpdateHitboxOffset(16)
}
