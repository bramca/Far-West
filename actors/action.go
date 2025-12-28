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
	Dir      Direction
	Type     ActionType
	actor    *Enemy
}

func (a Action) PerformAction(playerBullets []*Bullet, animate bool) {
	actionPerformed := false
	switch a.Type {
	case Dodge:
		for _, bullet := range playerBullets {
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
				a.actor.UpdateHitboxOffset(16)
				if animate {
					a.actor.Animate()
				}
				actionPerformed = true
				break
			}
		}
	}
	if !actionPerformed {
		a.actor.StopAnimation()
	}
}
