package actors

import (
	"math"
	"math/rand/v2"

	"github.com/bramca/Far-West/utils"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// an enemy drinks his whisky when his health drops below this fraction.
	enemyHealHealthDivisor = 3
	// homeLeashFraction is the part of the camp radius an idle enemy wanders
	// before it turns back towards the camp center.
	homeLeashFraction = 0.6
)

// EnemyTierStats holds the extra stats an enemy gets from its tier on top of
// the level scaling. Tougher tiers are deadlier, faster and shoot more often.
type EnemyTierStats struct {
	Health   int
	Damage   int
	Speed    float64
	FireRate int
}

type Enemy struct {
	*Player

	Tier Tier

	VisualDist int
	// HomeX and HomeY are the center of the bandit camp the enemy belongs to
	// and HomeRadius is how far it may wander away from it.
	HomeX      float64
	HomeY      float64
	HomeRadius float64
	// Blocked is set when the enemy could not move at all because an obstacle
	// or another actor was in the way. The game uses it to steer the enemy
	// into a free direction instead of letting it push against the obstacle.
	Blocked bool
	// CorpseTimer counts the frames since the enemy died.
	CorpseTimer int
}

func (e *Enemy) Draw(screen *ebiten.Image, camX float64, camY float64) {
	e.DrawOptions.GeoM.Reset()
	e.DrawOptions.GeoM.Scale(e.Scale, e.Scale)
	e.DrawOptions.GeoM.Translate(float64(e.X-camX), float64(e.Y-camY))
	e.DrawOptions.ColorScale.Reset()
	if e.HitFlash > 0 {
		// a bright flash makes every bullet that lands feel like it lands
		e.DrawOptions.ColorScale.Scale(3, 3, 3, 1)
		e.HitFlash--
	}
	screen.DrawImage(e.Sprites[e.CurrentState], e.DrawOptions)
	e.DrawOptions.ColorScale.Reset()
	weapon := e.GunBelt[e.CurrentWeapon].GunType
	if e.MuzzleFlash > 0 && !e.Dead && weapon != Fists && weapon != Whisky {
		drawMuzzleFlash(screen, e.X+e.W-camX, e.Y+e.H-camY, e.AimAngle)
	}
	if !e.Dead {
		e.Healthbar.Draw(screen, camX-16, camY-16)
	}
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

// whiskyIndex returns the belt slot of a whisky that still has a sip left.
func (e *Enemy) whiskyIndex() (int, bool) {
	for i, gun := range e.GunBelt {
		if gun != nil && gun.GunType == Whisky && gun.Ammo > 0 {
			return i, true
		}
	}

	return 0, false
}

// primaryWeaponIndex returns the belt slot of the enemy's fighting gun.
func (e *Enemy) primaryWeaponIndex() int {
	for i, gun := range e.GunBelt {
		if gun != nil && gun.GunType != Whisky {
			return i
		}
	}

	return 0
}

// shouldHeal reports whether the enemy is hurt enough to reach for his whisky.
func (e *Enemy) shouldHeal() bool {
	if e.Dead || e.Health <= 0 || e.Health*enemyHealHealthDivisor > e.MaxHealth {
		return false
	}

	_, ok := e.whiskyIndex()

	return ok
}

func (e *Enemy) switchWeapon(index int) {
	if index < 0 || index >= len(e.GunBelt) || e.CurrentWeapon == index {
		return
	}

	e.Reloading = false
	e.ReloadTimer = 0
	e.ShootTimer = 0
	e.DrawWeapon(index)
}

func (e *Enemy) ThinkAndAct(player *Player, playerBullets []Bullet, frameCount int) {
	e.CurrentAction.Duration -= 1

	if e.CurrentAction.Type == Heal {
		if e.CurrentAction.Duration > 0 {
			e.CurrentAction.PerformAction(player, frameCount)
			e.UpdateHitboxOffset(16)

			return
		}

		e.switchWeapon(e.primaryWeaponIndex())
		e.CurrentAction = Action{}
	}

	if e.shouldHeal() && rand.Float64() > 0.99 {
		if index, ok := e.whiskyIndex(); ok {
			e.switchWeapon(index)
			e.CurrentAction = Action{
				Duration: e.GunBelt[index].ShootCooldown * 2,
				Type:     Heal,
				Actor:    e.Player,
			}
			e.CurrentAction.PerformAction(player, frameCount)
			e.UpdateHitboxOffset(16)

			return
		}
	}

	// Detect player
	if utils.DistanceBetweenPoints(player.X, player.Y, e.X, e.Y) <= float64(e.VisualDist) && e.CurrentAction.Type != MoveAndShoot && e.CurrentAction.Duration <= 0 {
		actionType := MoveAndShoot
		if rand.Float64() < 0.5 {
			actionType = Dodge
		}
		e.CurrentAction = Action{
			Duration: 85 + rand.IntN(170),
			Type:     actionType,
			Actor:    e.Player,
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
			Duration: 85 + rand.IntN(170),
			Type:     actionType,
			MoveDir:  e.wanderDirection(dirs),
			LookDir:  dirs[rand.IntN(len(dirs))],
			Actor:    e.Player,
		}
	}

	// a bandit that wandered too far heads back to his camp instead of
	// drifting over the whole island
	if e.HomeRadius > 0 && e.CurrentAction.Type == Move && e.distanceFromHome() > e.HomeRadius {
		e.CurrentAction.MoveDir = e.directionTowardsPoint(e.HomeX, e.HomeY)
	}

	e.CurrentAction.PerformAction(player, frameCount)
	e.UpdateHitboxOffset(16)
}

// wanderDirection picks a direction for an idle bandit. Once he is far enough
// from the center of his camp he turns back home.
func (e *Enemy) wanderDirection(dirs []Direction) Direction {
	if e.HomeRadius > 0 && e.distanceFromHome() > e.HomeRadius*homeLeashFraction {
		return e.directionTowardsPoint(e.HomeX, e.HomeY)
	}

	return dirs[rand.IntN(len(dirs))]
}

func (e *Enemy) distanceFromHome() float64 {
	return utils.DistanceBetweenPoints(e.X, e.Y, e.HomeX, e.HomeY)
}

// directionTowardsPoint returns the cardinal direction that brings the enemy
// closer to the given point.
func (e *Enemy) directionTowardsPoint(x, y float64) Direction {
	xDelta := x - e.X
	yDelta := y - e.Y

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
