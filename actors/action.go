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
	Heal
	// Reposition is a short strafe between attack bursts. The enemy moves
	// without firing, which gives the player a chance to shoot back.
	Reposition
	// Flee runs toward an escape point while optionally shooting over the
	// actor's shoulder, used by a bounty that is running for the boat.
	Flee
)

const (
	// enemyLeadFactor scales how far the enemies lead a moving target, a
	// value of 1 produces exact intercept courses.
	enemyLeadFactor = 0.75
	// enemyMaxLead caps how far ahead of the player the enemies may aim so
	// the prediction never carries the shots out of the play field.
	enemyMaxLead = 140.0
	// engageRangeTolerance is how far an enemy may drift from the range its
	// weapon prefers before it closes in or backs off.
	engageRangeTolerance = 45.0
	// aimLockThreshold is the threshold on how many frames of the final steps
	// of the windup the aim angle is locked
	aimLockThreshold = 2
)

type Action struct {
	Duration int
	MoveDir  Direction
	LookDir  Direction
	Type     ActionType
	Actor    *Player
	// PreferredRange is the distance from the player the actor tries to keep
	// while moving and shooting. Zero means the actor only strafes.
	PreferredRange float64
	// TargetOffsetX and TargetOffsetY shift the point the actor moves toward
	// away from the player, so camp members can flank instead of stacking up.
	TargetOffsetX float64
	TargetOffsetY float64
	// FleeX and FleeY are the escape point a fleeing actor runs toward.
	FleeX float64
	FleeY float64
	// Fire makes a fleeing actor shoot at the player while he runs.
	Fire bool

	// Windup counts down the frames an actor spends aiming before it opens
	// fire. While it runs the shot is telegraphed so the player can dodge.
	Windup int
	// WindupFrames is the full length of the windup, used to draw how far it
	// has progressed.
	WindupFrames int
	// BurstShots is the number of shots left in the current burst.
	BurstShots int
	// BurstInterval is the number of frames between the shots of a burst.
	BurstInterval int
	// ShotTimer counts down to the next shot of a burst.
	ShotTimer int
	// attackTimerAccumulator keeps the burst timers in sync with TimeScale.
	attackTimerAccumulator float64
}

// isTelegraphing reports whether the action is still winding up an attack.
func (a *Action) isTelegraphing() bool {
	if a.WindupFrames <= 0 || a.Windup <= 0 {
		return false
	}

	// a fleeing bounty only telegraphs while he can actually see the player
	return a.Type != Flee || a.Fire
}

// telegraphProgress returns how much of the windup has played, from 0 to 1.
func (a *Action) telegraphProgress() float64 {
	if a.WindupFrames <= 0 {
		return 1
	}

	progress := 1 - float64(a.Windup)/float64(a.WindupFrames)

	return min(1, max(0, progress))
}

// advanceAttack ticks the windup and fires the shots of the current burst. It
// only advances once per simulation step so dead eye actually slows the
// enemies down instead of letting them fire every frame.
func (a *Action) advanceAttack() {
	if a.Actor == nil {
		return
	}

	if a.Windup > 0 {
		a.Windup -= 1

		return
	}

	if a.BurstShots <= 0 {
		return
	}

	if a.ShotTimer > 0 {
		a.ShotTimer -= 1
	}

	if a.ShotTimer > 0 {
		return
	}

	// the offset makes the bullet leave from the barrel instead of the
	// middle of the enemy sprite
	if a.Actor.IsNpc {
		a.Actor.X += 16
		a.Actor.Y += 16
	}

	a.Actor.Shoot()

	if a.Actor.IsNpc {
		a.Actor.X -= 16
		a.Actor.Y -= 16
	}

	a.BurstShots -= 1
	a.ShotTimer = a.BurstInterval
}

// directionTowardsPoint returns the cardinal direction that brings an actor
// closer to the given point.
func directionTowardsPoint(fromX, fromY, toX, toY float64) Direction {
	xDelta := toX - fromX
	yDelta := toY - fromY

	if math.Abs(xDelta) > math.Abs(yDelta) {
		if xDelta > 0 {
			return Right
		}

		return Left
	}

	if yDelta > 0 {
		return Down
	}

	return Up
}

func (a *Action) PerformAction(player *Player, frameCount int) {
	actionPerformed := false
	if a.Actor == nil {
		return
	}
	switch a.Type {
	case Dodge:
		if a.Actor.IsNpc {
			a.Actor.Speed += a.Actor.DodgeSpeed
			for i := range player.Bullets {
				bullet := &player.Bullets[i]
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
	case Move, Reposition:
		a.Actor.Look(a.LookDir)
		a.Actor.Move(a.MoveDir)
		if frameCount%a.Actor.AnimationSpeed == 0 {
			a.Actor.Animate()
		}
		actionPerformed = true
	case Flee:
		dir := directionTowardsPoint(a.Actor.X, a.Actor.Y, a.FleeX, a.FleeY)
		a.Actor.Move(dir)

		// a fleeing bounty does not stop fighting: while he can see the
		// player he keeps aiming and firing over his shoulder
		if a.Fire && a.Actor.IsNpc && !player.Dead {
			if a.Windup > aimLockThreshold {
				a.Actor.AimAngle = utils.LeadAngle(
					a.Actor.X, a.Actor.Y,
					player.X, player.Y,
					player.X-player.PrevX, player.Y-player.PrevY,
					a.Actor.GunBelt[0].BulletSpeed, enemyLeadFactor, enemyMaxLead,
				)
				a.Actor.FaceFromAim()
			}
			for range a.Actor.consumeTime(&a.attackTimerAccumulator) {
				a.advanceAttack()
			}
		} else {
			a.Actor.Look(dir)
		}

		if frameCount%a.Actor.AnimationSpeed == 0 {
			a.Actor.Animate()
		}
		actionPerformed = true
	case MoveAndShoot:
		focusX := player.X + a.TargetOffsetX
		focusY := player.Y + a.TargetOffsetY
		angle := utils.AngleBetweenPoints(focusX, focusY, a.Actor.X, a.Actor.Y)
		if math.Abs(angle-3*math.Pi/2) < 0.1 || math.Abs(angle-math.Pi) < 0.1 || math.Abs(angle) < 0.1 || math.Abs(angle-math.Pi/2) < 0.1 {
			a.Actor.StopAnimation()
		}

		distance := utils.DistanceBetweenPoints(a.Actor.X, a.Actor.Y, focusX, focusY)
		advance := a.PreferredRange > 0 && distance > a.PreferredRange+engageRangeTolerance

		if advance {
			dir := directionTowardsPoint(a.Actor.X, a.Actor.Y, focusX, focusY)
			a.Actor.Look(dir)
			a.Actor.Move(dir)
		} else {
			xDelta := math.Abs(a.Actor.X - focusX)
			yDelta := math.Abs(a.Actor.Y - focusY)
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
		}

		// the enemy leads the running player so the bullets land where he
		// will be when they arrive, the player keeps aiming at the cursor
		if a.Actor.IsNpc && !player.Dead {
			if a.Windup > aimLockThreshold {
				a.Actor.AimAngle = utils.LeadAngle(
					a.Actor.X, a.Actor.Y,
					player.X, player.Y,
					player.X-player.PrevX, player.Y-player.PrevY,
					a.Actor.GunBelt[0].BulletSpeed, enemyLeadFactor, enemyMaxLead,
				)
			}
			for range a.Actor.consumeTime(&a.attackTimerAccumulator) {
				a.advanceAttack()
			}
		} else {
			a.Actor.AimAngle = utils.AngleBetweenPoints(a.Actor.X, a.Actor.Y, player.X, player.Y)
		}
		a.Actor.FaceFromAim()

		if player.Hitbox.CheckCollision(a.Actor.Hitbox) {
			step := a.Actor.Speed * a.Actor.timeScale()
			for dir, moving := range a.Actor.MoveDirs {
				if moving {
					switch dir {
					case Up:
						a.Actor.Y += step
					case Down:
						a.Actor.Y -= step
					case Right:
						a.Actor.X -= step
					case Left:
						a.Actor.X += step
					}
				}
			}
		}
		if frameCount%a.Actor.AnimationSpeed == 0 {
			a.Actor.Animate()
		}

		actionPerformed = true
	case Heal:
		a.Actor.Shoot()

		xDelta := a.Actor.X - player.X
		yDelta := a.Actor.Y - player.Y
		if xDelta != 0 {
			if xDelta > 0 {
				a.Actor.Move(Right)
			} else {
				a.Actor.Move(Left)
			}
		}
		if yDelta != 0 {
			if yDelta > 0 {
				a.Actor.Move(Down)
			} else {
				a.Actor.Move(Up)
			}
		}

		switch {
		case math.Abs(xDelta) >= math.Abs(yDelta):
			if xDelta >= 0 {
				a.Actor.Look(Right)
			} else {
				a.Actor.Look(Left)
			}
		case yDelta >= 0:
			a.Actor.Look(Down)
		default:
			a.Actor.Look(Up)
		}

		if frameCount%a.Actor.AnimationSpeed == 0 {
			a.Actor.Animate()
		}
		actionPerformed = true
	}
	if !actionPerformed {
		a.Actor.StopAnimation()
	}
}
