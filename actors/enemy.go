package actors

import (
	"math"
	"math/rand"

	"github.com/bramca/Far-West/utils"
	"github.com/hajimehoshi/ebiten/v2"
)

type Enemy struct {
	*Player

	CurrentAction Action
	VisualDist    int
	MoveSpeed     int
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
	performAction := true
	e.CurrentAction.Duration -= 1
	if utils.DistanceBetweenPoints(player.X, player.Y, e.X, e.Y) <= float64(e.VisualDist) {
		move := false
		angle := utils.AngleBetweenPoints(player.X, player.Y, e.X, e.Y)
		if e.CurrentAction.Type == Move || e.CurrentAction.Type == MoveAndShoot {
			performAction = false
			move = frameCount%e.MoveSpeed == 0
			if math.Abs(angle-3*math.Pi/2) < 0.1 || math.Abs(angle-math.Pi) < 0.1 || math.Abs(angle) < 0.1 || math.Abs(angle-math.Pi/2) < 0.1 {
				move = false
				e.StopAnimation()
			}
		}

		xDelta := math.Abs(e.X - player.X)
		yDelta := math.Abs(e.Y - player.Y)
		moveX := false
		moveY := false
		if xDelta <= yDelta && angle >= 0 && angle <= math.Pi {
			e.Look(Up)
			moveX = true
		}
		if xDelta <= yDelta && angle <= 0 && angle >= -math.Pi {
			e.Look(Down)
			moveX = true
		}
		if yDelta <= xDelta && angle >= math.Pi/2 && angle <= 3*math.Pi/2 {
			e.Look(Right)
			moveY = true
		}
		if yDelta <= xDelta && angle <= math.Pi/2 && angle >= -math.Pi/2 {
			e.Look(Left)
			moveY = true
		}
		if move && moveX && angle <= math.Pi/2 && angle >= -math.Pi/2 {
			e.Move(Left)
			e.UpdateHitboxOffset(16)
		}
		if move && moveX && ((angle >= math.Pi/2 && angle <= 3*math.Pi/2) || (angle <= -math.Pi/2 && angle >= -3*math.Pi/2)) {
			e.Move(Right)
			e.UpdateHitboxOffset(16)
		}
		if move && moveY && angle >= 0 && angle <= math.Pi {
			e.Move(Up)
			e.UpdateHitboxOffset(16)
		}
		if move && moveY && angle <= 0 && angle >= -math.Pi {
			e.Move(Down)
			e.UpdateHitboxOffset(16)
		}

		if player.Hitbox.CheckCollision(e.Hitbox) {
			for dir, moving := range e.MoveDirs {
				if moving {
					switch dir {
					case Up:
						e.Y += e.Speed
					case Down:
						e.Y -= e.Speed
					case Right:
						e.X -= e.Speed
					case Left:
						e.X += e.Speed
					}
					e.UpdateHitboxOffset(16)
				}
			}
		}
		if move && frameCount%e.AnimationSpeed == 0 {
			e.Animate()
		}
	}

	if e.CurrentAction.Duration <= 0 {
		e.StopAnimation()
		actionType := Move
		if rand.Float64() < 0.5 {
			actionType = Dodge
		}
		e.CurrentAction = Action{
			Duration: 300 + rand.Intn(240),
			Type:     actionType,
			actor:    e,
		}
	}

	if performAction {
		e.CurrentAction.PerformAction(playerBullets, frameCount%e.AnimationSpeed == 0)
	}
}
