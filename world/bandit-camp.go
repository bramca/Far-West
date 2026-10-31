package world

import (
	"math"

	"github.com/bramca/Far-West/actors"
)

// BanditCamp is a logical gathering of bandits, tents and chests. It is the
// home of its enemies: they spawn and wander around it instead of roaming the
// whole island.
type BanditCamp struct {
	X, Y   float64
	Radius float64

	Enemies []*actors.Enemy
	Tents   []*Tent
	Chests  []*Chest
}

func NewBanditCamp(x, y, radius float64) *BanditCamp {
	return &BanditCamp{
		X:      x,
		Y:      y,
		Radius: radius,
	}
}

// Contains reports whether the given world position lies inside the camp.
func (c *BanditCamp) Contains(x, y float64) bool {
	return math.Hypot(x-c.X, y-c.Y) <= c.Radius
}

// AddEnemy makes the enemy part of the camp.
func (c *BanditCamp) AddEnemy(enemy *actors.Enemy) {
	c.Enemies = append(c.Enemies, enemy)
}

// RemoveEnemy drops the enemy from the camp, it is called when a corpse is
// cleaned up.
func (c *BanditCamp) RemoveEnemy(enemy *actors.Enemy) {
	for i, campEnemy := range c.Enemies {
		if campEnemy == enemy {
			c.Enemies = append(c.Enemies[:i], c.Enemies[i+1:]...)

			return
		}
	}
}

// AliveEnemies counts the enemies of the camp that are still a threat.
func (c *BanditCamp) AliveEnemies() int {
	alive := 0
	for _, enemy := range c.Enemies {
		if !enemy.Dead {
			alive++
		}
	}

	return alive
}
