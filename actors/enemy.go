package actors

import (
	"math"
	"math/rand/v2"

	"github.com/bramca/Far-West/utils"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// huntChance is the chance that an enemy that lost track of the player
	// roams in his direction instead of picking a random direction.
	huntChance = 0.75
	// an enemy drinks his whisky when his health drops below this fraction.
	enemyHealHealthDivisor = 3
)

type Enemy struct {
	*Player

	Tier Tier

	VisualDist int
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
		moveDir := dirs[rand.IntN(len(dirs))]
		// most of the time the enemy roams towards the player, that way the
		// player is always hunted instead of being left alone on the island
		if rand.Float64() < huntChance {
			moveDir = e.directionTowards(player)
		}
		e.CurrentAction = Action{
			Duration: 85 + rand.IntN(170),
			Type:     actionType,
			MoveDir:  moveDir,
			LookDir:  dirs[rand.IntN(len(dirs))],
			Actor:    e.Player,
		}
	}

	e.CurrentAction.PerformAction(player, frameCount)
	e.UpdateHitboxOffset(16)
}

// directionTowards returns the direction that brings the enemy closer to the player.
func (e *Enemy) directionTowards(player *Player) Direction {
	xDelta := player.X - e.X
	yDelta := player.Y - e.Y

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
