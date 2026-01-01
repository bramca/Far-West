package actors

import (
	"math"

	"github.com/bramca/Far-West/utils"
)

type ActionType int

const (
	Dodge ActionType = iota
	FindCover
	Shoot
	Move
	MoveAndShoot
)

type Action struct {
	Duration int
	MoveDir  Direction
	LookDir  Direction
	Type     ActionType
	actor    *Player
}

func (a Action) PerformAction(player *Player, frameCount int) {
	actionPerformed := false
	switch a.Type {
	case Dodge:
		a.actor.Speed += 0.5
		for _, bullet := range player.Bullets {
			if utils.DistanceBetweenPoints(bullet.X, bullet.Y, a.actor.X, a.actor.Y) < 150 {
				moveDir := Up

				angle := utils.AngleBetweenPoints(bullet.X, bullet.Y, a.actor.X, a.actor.Y)
				if (bullet.R == 0 || bullet.R == math.Pi) && angle >= 0 && angle <= math.Pi {
					moveDir = Down
				}

				if (bullet.R == math.Pi/2 || bullet.R == 3*math.Pi/2) && angle >= math.Pi/2 && angle <= 3*math.Pi/2 {
					moveDir = Left
				}

				if (bullet.R == math.Pi/2 || bullet.R == 3*math.Pi/2) && angle <= math.Pi/2 && angle >= -math.Pi/2 {
					moveDir = Right
				}
				a.actor.Move(moveDir)
				if frameCount%a.actor.AnimationSpeed == 0 {
					a.actor.Animate()
				}
				actionPerformed = true
				break
			}
		}
		a.actor.Speed -= 0.5
	case Move:
		a.actor.Look(a.LookDir)
		a.actor.Move(a.MoveDir)
		if frameCount%a.actor.AnimationSpeed == 0 {
			a.actor.Animate()
		}
		actionPerformed = true
	case MoveAndShoot:
		angle := utils.AngleBetweenPoints(player.X, player.Y, a.actor.X, a.actor.Y)
		if math.Abs(angle-3*math.Pi/2) < 0.1 || math.Abs(angle-math.Pi) < 0.1 || math.Abs(angle) < 0.1 || math.Abs(angle-math.Pi/2) < 0.1 {
			a.actor.StopAnimation()
		}

		xDelta := math.Abs(a.actor.X - player.X)
		yDelta := math.Abs(a.actor.Y - player.Y)
		moveX := false
		moveY := false
		if xDelta <= yDelta && angle >= 0 && angle <= math.Pi {
			a.actor.Look(Up)
			moveX = true
		}
		if xDelta <= yDelta && angle <= 0 && angle >= -math.Pi {
			a.actor.Look(Down)
			moveX = true
		}
		if yDelta <= xDelta && angle >= math.Pi/2 && angle <= 3*math.Pi/2 {
			a.actor.Look(Right)
			moveY = true
		}
		if yDelta <= xDelta && angle <= math.Pi/2 && angle >= -math.Pi/2 {
			a.actor.Look(Left)
			moveY = true
		}
		if moveX && angle <= math.Pi/2 && angle >= -math.Pi/2 {
			a.actor.Move(Left)
		}
		if moveX && ((angle >= math.Pi/2 && angle <= 3*math.Pi/2) || (angle <= -math.Pi/2 && angle >= -3*math.Pi/2)) {
			a.actor.Move(Right)
		}
		if moveY && angle >= 0 && angle <= math.Pi {
			a.actor.Move(Up)
		}
		if moveY && angle <= 0 && angle >= -math.Pi {
			a.actor.Move(Down)
		}

		if player.Hitbox.CheckCollision(a.actor.Hitbox) {
			for dir, moving := range a.actor.MoveDirs {
				if moving {
					switch dir {
					case Up:
						a.actor.Y += a.actor.Speed
					case Down:
						a.actor.Y -= a.actor.Speed
					case Right:
						a.actor.X -= a.actor.Speed
					case Left:
						a.actor.X += a.actor.Speed
					}
				}
			}
		}
		if frameCount%a.actor.AnimationSpeed == 0 {
			a.actor.Animate()
		}

		if frameCount%a.actor.FireRate == 0 {
			// the addition and substraction here
			// are for adding the correct offset to the bullet
			if a.actor.IsNpc {
				a.actor.X += 16
				a.actor.Y += 16
			}

			a.actor.Shoot()

			if a.actor.IsNpc {
				a.actor.X -= 16
				a.actor.Y -= 16
			}
		}
		actionPerformed = true
	}
	if !actionPerformed {
		a.actor.StopAnimation()
	}
}
