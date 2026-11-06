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
	// enemyAlertDuration is how long a bandit keeps hunting after the player
	// was last seen before he calms down and returns to his camp.
	enemyAlertDuration = 3 * 60
	// enemyLastSeenRadius is how close an alerted bandit has to get to the
	// last place he saw the player before he gives up the chase.
	enemyLastSeenRadius = 200.0
	// enemyCombatDodgeChance is how often a fighting enemy picks a defensive
	// dodge over a strafing attack.
	enemyCombatDodgeChance = 0.12
	// enemyIdleDodgeChance is how often a wandering enemy stops to dodge.
	enemyIdleDodgeChance = 0.25
	// enemyCombatDodgeDuration is how long a fighting enemy dodges before it
	// attacks again. Dodges are kept short so they do not stall the fight.
	enemyCombatDodgeMinDuration = 2 * 60
	enemyCombatDodgeMaxDuration = 4 * 60
	// enemyRepositionMinDuration and enemyRepositionMaxDuration bound the
	// short strafe an enemy takes between attack bursts. It does not fire
	// during this window.
	enemyRepositionMinDuration = 1 * 60
	enemyRepositionMaxDuration = 3 * 60
	// engage ranges keep the different guns in their own niche: shotguns
	// close in, revolvers hold the middle and machine guns keep their
	// distance.
	fistsEngageRange      = 40.0
	shotgunEngageRange    = 110.0
	revolverEngageRange   = 220.0
	machineGunEngageRange = 320.0
	// flankOffsetDistance is how far to the side of the player a flanking
	// bandit aims to stand, so the camp spreads out instead of stacking up.
	flankOffsetDistance = 120.0
	// rusherEngageFraction scales the distance a rusher keeps, so it charges
	// in while the pins hold the player at their weapon's range.
	rusherEngageFraction = 0.5
	// bountyFleeAggroRange widens the detection range of the bandits that
	// chase the player while the bounty runs for the boat.
	bountyFleeAggroRange = 1.75
)

// EnemyRole decides how a bandit approaches a fight.
type EnemyRole int

const (
	// RolePin holds the player's attention at the weapon's usual range.
	RolePin EnemyRole = iota
	// RoleFlank circles around to the side of the player.
	RoleFlank
	// RoleRusher charges in close.
	RoleRusher
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
	// IsBounty marks the one bounty target of a level. Bounties are optional
	// side targets and have to be knocked out with fists to be captured.
	IsBounty bool
	// KnockedOut marks a bounty that was defeated with fists and can be
	// collected at the jail.
	KnockedOut bool
	// Fleeing is true while a bounty is running for the boat. A fleeing
	// bounty keeps shooting at the player instead of only running.
	Fleeing bool
	// FleeX and FleeY are the point a fleeing bounty runs toward.
	FleeX float64
	FleeY float64
	// Escaped is set when a fleeing bounty reached the boat, his reward is
	// lost and he is no longer part of the fight.
	Escaped bool
	// Aggressive widens the detection range of the bandits that hunt the
	// player while the bounty is running for the boat.
	Aggressive bool

	VisualDist int
	// Alerted is true while the enemy is hunting the player. The whole camp
	// is alerted when one of its bandits spots the player or gets shot.
	Alerted bool
	// AlertTimer counts the frames since the enemy last saw the player, once
	// it runs out the enemy calms down again.
	AlertTimer int
	// LastSeenX and LastSeenY remember where the player was when the camp
	// was alerted, so bandits can hunt him after he breaks line of sight.
	LastSeenX float64
	LastSeenY float64
	// CampID identifies the bandit camp the enemy belongs to, bandits of the
	// same camp raise the alarm for each other.
	CampID int
	// Role decides how the bandit approaches the player during a fight.
	Role EnemyRole
	// FlankAngle is the world angle around the player a flanker targets.
	FlankAngle float64
	// HomeX and HomeY are the center of the bandit camp the enemy belongs to
	// and HomeRadius is how far it may wander away from it.
	HomeX      float64
	HomeY      float64
	HomeRadius float64
	// Blocked is set when the enemy could not move at all because an obstacle
	// or another actor was in the way. The game uses it to steer the enemy
	// into a free direction instead of letting it push against the obstacle.
	Blocked bool
	// SteerTimer keeps an unstick move from being replaced by the AI for a
	// short while, so the enemy can walk around the obstacle it ran into.
	SteerTimer int
	// CorpseTimer counts the frames since the enemy died.
	CorpseTimer int
	// StateSpriteIndex returns the index in the enemy sprite array for a certain state
	StateSpriteIndex map[PlayerState]int
	// actionTimerAccumulator keeps the action duration in sync with TimeScale.
	actionTimerAccumulator float64
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
	screen.DrawImage(e.Sprites[e.StateSpriteIndex[e.CurrentState]], e.DrawOptions)
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
	step := e.Speed * e.timeScale()
	switch d {
	case Right:
		e.X += step
	case Left:
		e.X -= step
	case Up:
		e.Y -= step
	case Down:
		e.Y += step
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
	e.CurrentAction.Duration -= e.consumeTime(&e.actionTimerAccumulator)

	if e.CurrentAction.Type == Heal {
		if e.CurrentAction.Duration > 0 {
			e.CurrentAction.PerformAction(player, frameCount)
			e.UpdateHitboxOffset(16)

			return
		}

		e.switchWeapon(e.primaryWeaponIndex())
		e.CurrentAction = Action{}
	}

	if !e.Fleeing && e.shouldHeal() && rand.Float64() > 0.99 {
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

	// an unstick move has to play out before the hunt logic can take over,
	// otherwise the enemy walks straight back into the obstacle
	if e.SteerTimer > 0 {
		e.SteerTimer--
		e.CurrentAction.PerformAction(player, frameCount)
		e.UpdateHitboxOffset(16)

		return
	}

	seesPlayer := e.CanSee(player)
	if seesPlayer && !e.Alerted {
		e.Alert(player.X, player.Y)
	}

	if e.Fleeing {
		e.flee(seesPlayer)
		e.CurrentAction.PerformAction(player, frameCount)
		e.UpdateHitboxOffset(16)

		return
	}

	if e.Alerted {
		if seesPlayer {
			e.engage(player)
		} else {
			e.huntLastSeen()
		}
	}

	if e.CurrentAction.Duration <= 0 {
		e.wander()
	}

	// a calm bandit that wandered too far heads back to his camp instead of
	// drifting over the whole island
	if !e.Alerted && e.HomeRadius > 0 && e.CurrentAction.Type == Move && e.distanceFromHome() > e.HomeRadius {
		e.CurrentAction.MoveDir = e.directionTowardsPoint(e.HomeX, e.HomeY)
	}

	e.CurrentAction.PerformAction(player, frameCount)
	e.UpdateHitboxOffset(16)
}

// CanSee reports whether the enemy is close enough to spot the player. While
// the bounty runs for the boat the other bandits are on high alert and spot
// the player from further away.
func (e *Enemy) CanSee(player *Player) bool {
	visualDist := float64(e.VisualDist)
	if e.Aggressive {
		visualDist *= bountyFleeAggroRange
	}

	return !player.Dead && utils.DistanceBetweenPoints(player.X, player.Y, e.X, e.Y) <= visualDist
}

// Alert makes the enemy hunt the player at the given position. The whole camp
// shares the alarm, so one bandit spotting the player warns his friends.
func (e *Enemy) Alert(x, y float64) {
	e.Alerted = true
	e.AlertTimer = enemyAlertDuration
	e.LastSeenX, e.LastSeenY = x, y
}

// TickAlert counts the hunt down, once it runs out the enemy calms down and
// returns to his camp.
func (e *Enemy) TickAlert() {
	if !e.Alerted {
		return
	}

	e.AlertTimer--
	if e.AlertTimer <= 0 {
		e.CalmDown()
	}
}

// CalmDown ends the hunt and cancels the current fight so the enemy goes back
// to wandering around his camp.
func (e *Enemy) CalmDown() {
	e.Alerted = false
	e.AlertTimer = 0
	if e.CurrentAction.Type == MoveAndShoot || e.CurrentAction.Type == Dodge || e.CurrentAction.Type == Reposition {
		e.CurrentAction = Action{}
	}
}

// flee makes the bounty run for his escape point. The action is refreshed
// regularly so he keeps aiming and shooting at the player while he runs.
func (e *Enemy) flee(seesPlayer bool) {
	// a bounty that was reaching for his whisky has to put it away before
	// he can shoot over his shoulder again
	if e.CurrentWeapon >= 0 && e.CurrentWeapon < len(e.GunBelt) &&
		e.GunBelt[e.CurrentWeapon] != nil && e.GunBelt[e.CurrentWeapon].GunType == Whisky {
		e.switchWeapon(e.primaryWeaponIndex())
	}

	if e.CurrentAction.Type == Flee && e.CurrentAction.Duration > 0 {
		e.CurrentAction.FleeX = e.FleeX
		e.CurrentAction.FleeY = e.FleeY
		e.CurrentAction.Fire = seesPlayer

		return
	}

	e.CurrentAction = Action{
		Duration: 30 + rand.IntN(15),
		Type:     Flee,
		FleeX:    e.FleeX,
		FleeY:    e.FleeY,
		Fire:     seesPlayer,
		Actor:    e.Player,
	}
}

// engage picks a fight action that keeps the enemy at the range its weapon
// likes. After an attack burst the enemy repositions without firing, so the
// player gets a clear window to shoot back.
func (e *Enemy) engage(player *Player) {
	if e.CurrentAction.Duration > 0 {
		switch e.CurrentAction.Type {
		case MoveAndShoot:
			e.CurrentAction.PreferredRange = e.engageRange()
			e.CurrentAction.TargetOffsetX, e.CurrentAction.TargetOffsetY = e.engageOffset()

			return
		case Dodge, Reposition:
			return
		}
	}

	if e.CurrentAction.Type == MoveAndShoot {
		focusX, focusY := e.focusPoint(player)
		moveDir := e.repositionDirection(focusX, focusY)
		e.CurrentAction = Action{
			Duration: enemyRepositionMinDuration + rand.IntN(enemyRepositionMaxDuration-enemyRepositionMinDuration+1),
			Type:     Reposition,
			MoveDir:  moveDir,
			LookDir:  e.directionTowardsPoint(player.X, player.Y),
			Actor:    e.Player,
		}

		return
	}

	actionType := MoveAndShoot
	duration := 85 + rand.IntN(170)
	if rand.Float64() < enemyCombatDodgeChance {
		actionType = Dodge
		duration = enemyCombatDodgeMinDuration + rand.IntN(enemyCombatDodgeMaxDuration-enemyCombatDodgeMinDuration+1)
	}

	offsetX, offsetY := e.engageOffset()
	e.CurrentAction = Action{
		Duration:       duration,
		Type:           actionType,
		Actor:          e.Player,
		PreferredRange: e.engageRange(),
		TargetOffsetX:  offsetX,
		TargetOffsetY:  offsetY,
	}
}

// focusPoint returns the point around the player the enemy moves toward. Pins
// and rushers aim straight at the player, flankers aim at a spot beside him.
func (e *Enemy) focusPoint(player *Player) (float64, float64) {
	offsetX, offsetY := e.engageOffset()

	return player.X + offsetX, player.Y + offsetY
}

// engageOffset returns the offset from the player the enemy's role targets.
func (e *Enemy) engageOffset() (float64, float64) {
	if e.Role != RoleFlank {
		return 0, 0
	}

	return math.Cos(e.FlankAngle) * flankOffsetDistance, math.Sin(e.FlankAngle) * flankOffsetDistance
}

// repositionDirection strafes around the focus point instead of closing in or
// backing off, so the reposition after an attack never walks the enemy into
// or away from the player.
func (e *Enemy) repositionDirection(x, y float64) Direction {
	if math.Abs(x-e.X) > math.Abs(y-e.Y) {
		if y < e.Y {
			return Down
		}

		return Up
	}

	if x < e.X {
		return Right
	}

	return Left
}

// huntLastSeen walks toward the place the player was last seen, so a camp
// keeps searching after the player breaks line of sight.
func (e *Enemy) huntLastSeen() {
	if utils.DistanceBetweenPoints(e.X, e.Y, e.LastSeenX, e.LastSeenY) <= enemyLastSeenRadius {
		e.CalmDown()

		return
	}

	dir := e.directionTowardsPoint(e.LastSeenX, e.LastSeenY)
	e.CurrentAction = Action{
		Duration: 45 + rand.IntN(30),
		Type:     Move,
		MoveDir:  dir,
		LookDir:  dir,
		Actor:    e.Player,
	}
}

// wander picks a new random direction for an idle bandit.
func (e *Enemy) wander() {
	e.StopAnimation()
	actionType := Move
	if rand.Float64() < enemyIdleDodgeChance {
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

// engageRange returns the distance the enemy tries to keep from the player,
// based on the gun in his hands. Shotguns close in, machine guns keep their
// distance.
func (e *Enemy) engageRange() float64 {
	index := e.CurrentWeapon
	if index < 0 || index >= len(e.GunBelt) || e.GunBelt[index] == nil || e.GunBelt[index].GunType == Whisky {
		index = e.primaryWeaponIndex()
	}
	if index < 0 || index >= len(e.GunBelt) || e.GunBelt[index] == nil {
		return revolverEngageRange
	}

	var engageRange float64
	switch e.GunBelt[index].GunType {
	case Shotgun:
		engageRange = shotgunEngageRange
	case MachineGun:
		engageRange = machineGunEngageRange
	case Fists:
		engageRange = fistsEngageRange
	default:
		engageRange = revolverEngageRange
	}

	// a rusher charges closer than his weapon would normally like
	if e.Role == RoleRusher {
		return max(fistsEngageRange, engageRange*rusherEngageFraction)
	}

	return engageRange
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
	return directionTowardsPoint(e.X, e.Y, x, y)
}
