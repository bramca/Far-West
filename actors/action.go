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
		if a.Actor.IsNpc {
			a.Actor.Speed += a.Actor.DodgeSpeed
			for _, bullet := range player.Bullets {
				if utils.DistanceBetweenPoints(bullet.X, bullet.Y, a.Actor.X, a.Actor.Y) < 150 {
					// dodge perpendicularly to the incoming bullet
					perpX := -math.Sin(bullet.R)
					perpY := math.Cos(bullet.R)
					moveDir := Up
					if math.Abs(perpX) > math.Abs(perpY) {
						if perpX > 0 {
							moveDir = Right
						} else {
							moveDir = Left
						}
					} else if perpY > 0 {
						moveDir = Down
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
		} else {
			oldSpeed := a.Actor.Speed
			fasterAnimationSpeed := a.Actor.AnimationSpeed * 2 / 3
			a.Actor.Speed = a.Actor.DodgeSpeed
			for dir, moving := range a.Actor.MoveDirs {
				if !moving {
					continue
				}

				a.Actor.Move(dir)
			}
			a.Actor.UpdateHitbox()
			if frameCount%fasterAnimationSpeed == 0 {
				a.Actor.Animate()
			}
			actionPerformed = true
			a.Actor.Speed = oldSpeed
		}
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

		// the enemy aims straight at the player and faces the shot quadrant
		a.Actor.AimAngle = utils.AngleBetweenPoints(a.Actor.X, a.Actor.Y, player.X, player.Y)
		a.Actor.FaceFromAim()

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
