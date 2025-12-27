package actors

import "math/rand"

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
	Type     ActionType
	actor    *Player
}

func (a Action) PerformAction(playerBullets []*Bullet, animate bool) {
	a.Duration -= 1
	switch a.Type {
	case Dodge:
		for _, bullet := range playerBullets {
			if bullet.X > a.actor.X-20 || bullet.X < a.actor.Y+20 {
				moveDir := Up
				if rand.Float64() > 0.5 {
					moveDir = Down
				}

				a.actor.Move(moveDir)
				if animate {
					a.actor.Animate()
				}
				break
			}
			if bullet.Y > a.actor.Y-20 || bullet.Y < a.actor.Y+20 {
				moveDir := Left
				if rand.Float64() > 0.5 {
					moveDir = Right
				}

				a.actor.Move(moveDir)
				if animate {
					a.actor.Animate()
				}
				break
			}
		}
	}
}
