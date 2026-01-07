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
	Actor    *Player
}

func (a Action) PerformAction(player *Player, frameCount int) {
	actionPerformed := false
	switch a.Type {
	case Dodge:
		a.Actor.Speed += a.Actor.DodgeSpeed
		for _, bullet := range player.Bullets {
			if utils.DistanceBetweenPoints(bullet.X, bullet.Y, a.Actor.X, a.Actor.Y) < 150 {
				moveDir := Up

				angle := utils.AngleBetweenPoints(bullet.X, bullet.Y, a.Actor.X, a.Actor.Y)
				if (bullet.R == 0 || bullet.R == math.Pi) && angle >= 0 && angle <= math.Pi {
					moveDir = Down
				}

				if (bullet.R == math.Pi/2 || bullet.R == 3*math.Pi/2) && angle >= math.Pi/2 && angle <= 3*math.Pi/2 {
					moveDir = Left
				}

				if (bullet.R == math.Pi/2 || bullet.R == 3*math.Pi/2) && angle <= math.Pi/2 && angle >= -math.Pi/2 {
					moveDir = Right
				}
				a.Actor.Move(moveDir)
				if frameCount%a.Actor.AnimationSpeed == 0 {
					a.Actor.Animate()
				}
				actionPerformed = true
				break
			}
		}
		a.Actor.Speed -= a.Actor.DodgeSpeed
	case Move:
		a.Actor.Look(a.LookDir)
		a.Actor.Move(a.MoveDir)
		if frameCount%a.Actor.AnimationSpeed == 0 {
			a.Actor.Animate()
		}
		actionPerformed = true
	case MoveAndShoot:
		angle := utils.AngleBetweenPoints(player.X, player.Y, a.Actor.X, a.Actor.Y)
		if math.Abs(angle-3*math.Pi/2) < 0.1 || math.Abs(angle-math.Pi) < 0.1 || math.Abs(angle) < 0.1 || math.Abs(angle-math.Pi/2) < 0.1 {
			a.Actor.StopAnimation()
		}

		xDelta := math.Abs(a.Actor.X - player.X)
		yDelta := math.Abs(a.Actor.Y - player.Y)
		moveX := false
		moveY := false
		if xDelta <= yDelta && angle >= 0 && angle <= math.Pi {
			a.Actor.Look(Up)
			moveX = true
		}
		if xDelta <= yDelta && angle <= 0 && angle >= -math.Pi {
			a.Actor.Look(Down)
			moveX = true
		}
		if yDelta <= xDelta && angle >= math.Pi/2 && angle <= 3*math.Pi/2 {
			a.Actor.Look(Right)
			moveY = true
		}
		if yDelta <= xDelta && angle <= math.Pi/2 && angle >= -math.Pi/2 {
			a.Actor.Look(Left)
			moveY = true
		}
		if moveX && angle <= math.Pi/2 && angle >= -math.Pi/2 {
			a.Actor.Move(Left)
		}
		if moveX && ((angle >= math.Pi/2 && angle <= 3*math.Pi/2) || (angle <= -math.Pi/2 && angle >= -3*math.Pi/2)) {
			a.Actor.Move(Right)
		}
		if moveY && angle >= 0 && angle <= math.Pi {
			a.Actor.Move(Up)
		}
		if moveY && angle <= 0 && angle >= -math.Pi {
			a.Actor.Move(Down)
		}

		if player.Hitbox.CheckCollision(a.Actor.Hitbox) {
			for dir, moving := range a.Actor.MoveDirs {
				if moving {
					switch dir {
					case Up:
						a.Actor.Y += a.Actor.Speed
					case Down:
						a.Actor.Y -= a.Actor.Speed
					case Right:
						a.Actor.X -= a.Actor.Speed
					case Left:
						a.Actor.X += a.Actor.Speed
					}
				}
			}
		}
		if frameCount%a.Actor.AnimationSpeed == 0 {
			a.Actor.Animate()
		}

		if frameCount%a.Actor.FireRate == 0 {
			// the addition and substraction here
			// are for adding the correct offset to the bullet
			if a.Actor.IsNpc {
				a.Actor.X += 16
				a.Actor.Y += 16
			}

			a.Actor.Shoot()

			if a.Actor.IsNpc {
				a.Actor.X -= 16
				a.Actor.Y -= 16
			}
		}
		actionPerformed = true
	}
	if !actionPerformed {
		a.Actor.StopAnimation()
	}
}
