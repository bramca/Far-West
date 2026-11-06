package farwest

import (
	"math"
	"testing"

	"github.com/bramca/Far-West/actors"
	"github.com/bramca/Far-West/world"
)

// TestGameLoop runs the game loop without any player input, the player is
// expected to stay on the island, to be hunted down by the bandits of the
// first camp and to end up in a game over.
func TestGameLoop(t *testing.T) {
	game := NewGame()
	game.mode = ModeGame

	// bandits stay near their camp, so put the idle player in the middle of
	// the first camp to let them find him. Clear the cover of the camp so this
	// test is about combat and not about the player hiding behind a tent.
	camp := game.camps[0]
	game.plants = nil
	game.tents = nil
	game.chests = nil
	game.player.X = camp.X
	game.player.Y = camp.Y
	game.player.UpdateHitbox()
	game.player.SavePosition()

	for range game.framesPerSecond * 300 {
		if err := game.Update(); err != nil {
			t.Fatal(err)
		}

		if !game.player.Dead && !game.island.IsWalkable(game.player.X, game.player.Y) {
			t.Fatalf("the player walked into the ocean at %f,%f", game.player.X, game.player.Y)
		}

		for _, enemy := range game.enemies {
			if enemy.Dead {
				continue
			}
			if !game.island.IsWalkable(enemy.Hitbox.CenterX(), enemy.Hitbox.CenterY()) {
				t.Fatalf("an enemy walked into the ocean at %f,%f", enemy.X, enemy.Y)
			}
		}

		if game.aliveEnemies() > game.levelEnemies {
			t.Fatalf("%d enemies on the field while the level only has %d", game.aliveEnemies(), game.levelEnemies)
		}

		if game.player.Dead {
			break
		}
	}

	if !game.player.Dead {
		t.Fatal("an idle player is expected to die eventually")
	}
}

// TestRestart makes sure a new run starts from a clean slate on level one.
func TestRestart(t *testing.T) {
	game := NewGame()
	game.score = 500
	game.elapsedFrames = 6000
	game.level = 5
	game.player.Dead = true
	game.player.Health = 0

	game.Initialize()

	if game.score != 0 || game.elapsedFrames != 0 || game.level != 1 {
		t.Fatal("the run state was not reset")
	}

	if game.player.Dead || game.player.Health != game.player.MaxHealth {
		t.Fatal("the player was not revived")
	}

	if n := game.aliveEnemies(); n < 1 || n > initialEnemies {
		t.Fatalf("expected at most %d enemies on a fresh run, got %d", initialEnemies, n)
	}
}

// TestLevelClearsAndAdvances makes sure a level starts with a set number of
// enemies, that killing them all flags the level complete overlay and that
// continuing starts the next level on a fresh island with full health.
func TestLevelClearsAndAdvances(t *testing.T) {
	game := NewGame()
	game.mode = ModeGame

	firstIsland := game.island
	if game.level != 1 {
		t.Fatalf("expected the game to start at level 1, got level %d", game.level)
	}
	if game.aliveEnemies() != enemyCountForLevel(1) {
		t.Fatalf("expected %d enemies on level 1, got %d", enemyCountForLevel(1), game.aliveEnemies())
	}

	game.player.Health -= game.player.MaxHealth / 2

	for _, enemy := range game.enemies {
		game.killEnemy(enemy)
	}

	// the kill hit-stop holds the game still for a few frames before the
	// level complete overlay is flagged
	for range 5 {
		if err := game.Update(); err != nil {
			t.Fatal(err)
		}
	}

	if game.mode != ModeGame {
		t.Fatalf("the game should keep running behind the overlay, got mode %d", game.mode)
	}
	if !game.levelComplete {
		t.Fatal("the level complete overlay was not shown")
	}
	if game.level != 1 {
		t.Fatalf("the level advanced without a continue input, still at %d", game.level)
	}

	game.startLevel(game.level + 1)
	if err := game.Update(); err != nil {
		t.Fatal(err)
	}

	if game.level != 2 {
		t.Fatalf("the level did not advance after continuing, still at %d", game.level)
	}
	if game.island == firstIsland {
		t.Fatal("the island was not regenerated for the next level")
	}
	if game.aliveEnemies() != enemyCountForLevel(2) {
		t.Fatalf("expected %d enemies on level 2, got %d", enemyCountForLevel(2), game.aliveEnemies())
	}
}

// TestBanditCamps makes sure every level has a small number of camps and that
// each camp is a real place on the island with enemies, tents and chests.
func TestBanditCamps(t *testing.T) {
	game := NewGame()

	if count := len(game.camps); count < 1 || count > maxCamps {
		t.Fatalf("expected 1 to %d bandit camps, got %d", maxCamps, count)
	}

	totalEnemies := 0
	for _, camp := range game.camps {
		if !game.island.IsWalkable(camp.X, camp.Y) {
			t.Fatalf("the camp at %f,%f is not on land", camp.X, camp.Y)
		}
		if len(camp.Tents) < campTentMin || len(camp.Tents) > campTentMax {
			t.Fatalf("the camp has %d tents, expected %d to %d", len(camp.Tents), campTentMin, campTentMax)
		}
		if len(camp.Chests) < campChestMin || len(camp.Chests) > campChestMax {
			t.Fatalf("the camp has %d chests, expected %d to %d", len(camp.Chests), campChestMin, campChestMax)
		}
		if len(camp.Enemies) == 0 {
			t.Fatal("a bandit camp spawned without enemies")
		}

		for _, enemy := range camp.Enemies {
			if distance := math.Hypot(enemy.X-camp.X, enemy.Y-camp.Y); distance > camp.Radius {
				t.Fatalf("an enemy spawned %f away from its camp with radius %f", distance, camp.Radius)
			}
		}

		type bounds struct {
			x, y, w, h float64
		}
		objects := make([]bounds, 0, len(camp.Tents)+len(camp.Chests))
		for _, tent := range camp.Tents {
			objects = append(objects, bounds{tent.X, tent.Y, tent.W * tent.Scale, tent.H * tent.Scale})
		}
		for _, chest := range camp.Chests {
			objects = append(objects, bounds{chest.X, chest.Y, chest.W * chest.Scale, chest.H * chest.Scale})
		}
		for i := range objects {
			for j := i + 1; j < len(objects); j++ {
				a, b := objects[i], objects[j]
				if a.x < b.x+b.w && a.x+a.w > b.x && a.y < b.y+b.h && a.y+a.h > b.y {
					t.Fatalf("camp objects %d and %d overlap", i, j)
				}
			}
		}

		totalEnemies += len(camp.Enemies)
	}

	if totalEnemies != game.levelEnemies {
		t.Fatalf("the camps hold %d enemies while the level has %d", totalEnemies, game.levelEnemies)
	}
}

// TestEnemiesWanderAroundTheirCamp makes sure an enemy that is not hunting the
// player stays close to the center of its camp instead of roaming the island.
func TestEnemiesWanderAroundTheirCamp(t *testing.T) {
	game := NewGame()
	farPlayer := &actors.Player{X: -100000, Y: -100000}

	for _, enemy := range game.enemies {
		for frame := range 1200 {
			enemy.ThinkAndAct(farPlayer, nil, frame)

			distance := math.Hypot(enemy.X-enemy.HomeX, enemy.Y-enemy.HomeY)
			if distance > enemy.HomeRadius+10 {
				t.Fatalf("an idle enemy wandered %f away from its camp with radius %f", distance, enemy.HomeRadius)
			}
		}
	}
}

// TestEnemiesKnowTheirCamp makes sure every spawned bandit can tell which
// camp it belongs to, the camp alert relies on it.
func TestEnemiesKnowTheirCamp(t *testing.T) {
	game := NewGame()

	for i, camp := range game.camps {
		for _, enemy := range camp.Enemies {
			if enemy.CampID != i {
				t.Fatalf("an enemy of camp %d has camp id %d", i, enemy.CampID)
			}
		}
	}
}

// TestCampAlertRalliesTheWholeCamp makes sure one bandit raising the alarm
// alerts every bandit of the same camp and leaves other camps alone.
func TestCampAlertRalliesTheWholeCamp(t *testing.T) {
	game := NewGame()
	first := &actors.Enemy{Player: &actors.Player{}, CampID: 0}
	second := &actors.Enemy{Player: &actors.Player{}, CampID: 0}
	other := &actors.Enemy{Player: &actors.Player{}, CampID: 1}
	game.enemies = []*actors.Enemy{first, second, other}

	game.alertCamp(first, 120, 340)

	if !first.Alerted || !second.Alerted {
		t.Fatal("the alarm did not spread to every bandit of the camp")
	}
	if other.Alerted {
		t.Fatal("the alarm spread to a bandit of another camp")
	}
	if second.LastSeenX != 120 || second.LastSeenY != 340 {
		t.Fatalf("the camp remembered the player at %f,%f, want 120,340", second.LastSeenX, second.LastSeenY)
	}
}

// TestBanditsSpotThePlayerAndAlertTheirCamp makes sure a camp raises the
// alarm when one of its bandits spots the player.
func TestBanditsSpotThePlayerAndAlertTheirCamp(t *testing.T) {
	game := NewGame()
	camp := game.camps[0]
	game.player.X = camp.X
	game.player.Y = camp.Y

	game.updateEnemyAlerts()

	for _, enemy := range camp.Enemies {
		if !enemy.Alerted {
			t.Fatal("a camp member was not alerted when the player was spotted")
		}
	}
}

// TestFreshSightingRalliesTheWholeCamp makes sure one bandit spotting the
// player still warns camp members that cannot see him yet.
func TestFreshSightingRalliesTheWholeCamp(t *testing.T) {
	game := NewGame()
	seer := &actors.Enemy{
		Player:     &actors.Player{X: 100, Y: 100},
		CampID:     0,
		VisualDist: 1000,
	}
	sleeper := &actors.Enemy{
		Player:     &actors.Player{X: -10000, Y: -10000},
		CampID:     0,
		VisualDist: 1,
	}
	game.enemies = []*actors.Enemy{seer, sleeper}
	game.player.X = 100
	game.player.Y = 100

	game.updateEnemyAlerts()

	if !seer.Alerted || !sleeper.Alerted {
		t.Fatal("a fresh sighting did not rally the whole camp")
	}
}

// TestCampAlertDoesNotRefreshBanditsThatLostSight makes sure a camp-wide hunt
// does not last forever just because one bandit can still see the player.
func TestCampAlertDoesNotRefreshBanditsThatLostSight(t *testing.T) {
	game := NewGame()
	seer := &actors.Enemy{
		Player:     &actors.Player{X: 100, Y: 100},
		CampID:     0,
		VisualDist: 1000,
	}
	seer.Alert(100, 100)
	seer.AlertTimer = 1
	sleeper := &actors.Enemy{
		Player:     &actors.Player{X: -10000, Y: -10000},
		CampID:     0,
		VisualDist: 1,
	}
	sleeper.Alert(100, 100)
	sleeper.AlertTimer = 100
	game.enemies = []*actors.Enemy{seer, sleeper}
	game.player.X = 100
	game.player.Y = 100

	game.updateEnemyAlerts()

	if seer.AlertTimer <= 1 {
		t.Fatal("the bandit that still sees the player stopped hunting")
	}
	if sleeper.AlertTimer != 99 {
		t.Fatalf("the bandit that lost sight kept a %d frame alert, want 99", sleeper.AlertTimer)
	}
}

// TestBlockedEnemyPicksFreeDirection makes sure an enemy that walked into an
// obstacle is steered into a direction it can actually move in.
func TestBlockedEnemyPicksFreeDirection(t *testing.T) {
	game := NewGame()
	enemy := game.enemies[0]

	// isolate the enemy in an open part of the island so the test cannot be
	// foiled by a neighbouring bandit standing in every free direction
	game.enemies = []*actors.Enemy{enemy}
	game.plants = nil
	game.tents = nil
	game.chests = nil
	game.player.X = -10000
	game.player.Y = -10000
	game.player.UpdateHitbox()
	enemy.X, enemy.Y = game.island.Center()
	enemy.HomeX, enemy.HomeY = enemy.X, enemy.Y
	enemy.UpdateHitboxOffset(enemyHitboxOffset)
	enemy.CurrentAction = actors.Action{}
	enemy.Blocked = true
	game.unstickEnemy(enemy)

	if enemy.Blocked {
		t.Fatal("the blocked enemy was not unstuck")
	}
	if enemy.CurrentAction.Type != actors.Move {
		t.Fatalf("the blocked enemy did not start moving, action type %d", enemy.CurrentAction.Type)
	}
	if !game.canEnemyStep(enemy, enemy.CurrentAction.MoveDir) {
		t.Fatal("the blocked enemy was steered into another obstacle")
	}
}

// TestUnstickMoveSurvivesTheAlertLogic makes sure a bandit that was steered
// around an obstacle keeps walking that way instead of turning straight back
// into it while hunting the player.
func TestUnstickMoveSurvivesTheAlertLogic(t *testing.T) {
	game := NewGame()
	enemy := game.enemies[0]
	enemy.Alert(0, 0)
	enemy.LastSeenX = enemy.X - 1000
	enemy.LastSeenY = enemy.Y
	enemy.VisualDist = 0
	enemy.CurrentAction = actors.Action{
		Duration: 50,
		Type:     actors.Move,
		MoveDir:  actors.Right,
		LookDir:  actors.Right,
		Actor:    enemy.Player,
	}
	enemy.SteerTimer = 50

	enemy.ThinkAndAct(game.player, nil, 1)

	if enemy.CurrentAction.Type != actors.Move || enemy.CurrentAction.MoveDir != actors.Right {
		t.Fatal("the unstick move was replaced by the hunt logic")
	}
}

// TestEnemyCountForLevel guards the enemy scaling.
func TestEnemyCountForLevel(t *testing.T) {
	if got := enemyCountForLevel(1); got != initialEnemies {
		t.Fatalf("level 1 has %d enemies, expected %d", got, initialEnemies)
	}

	if killed := enemyCountForLevel(maxLevelEnemies - initialEnemies + 1); killed != maxLevelEnemies {
		t.Fatalf("the enemy count should be capped at %d, got %d", maxLevelEnemies, killed)
	}
}

// TestReloadLimitsFireRate makes sure the magazine and the reload timer stop
// the player from spamming bullets.
func TestReloadLimitsFireRate(t *testing.T) {
	game := NewGame()
	game.player.DrawWeapon(int(actors.Revolver))

	shots := 0
	for range game.framesPerSecond {
		bulletsBefore := len(game.player.Bullets)
		game.player.Shoot()
		if len(game.player.Bullets) > bulletsBefore {
			shots++
		}
		game.player.UpdateWeapon()
	}

	if shots > game.player.GunBelt[game.player.CurrentWeapon].MagazineSize+1 {
		t.Fatalf("fired %d bullets in a second, the reload is not slowing the player down", shots)
	}

	if shots == 0 {
		t.Fatal("the player could not fire a single bullet")
	}
}

// TestDodgeCooldownLimitsSpamming makes sure the cooldown stops the player
// from dodging again before it ran out.
func TestDodgeCooldownLimitsSpamming(t *testing.T) {
	game := NewGame()
	if game.player.DodgeCooldown <= 0 {
		t.Fatal("the player has no dodge cooldown")
	}

	game.player.StartDodge()
	game.player.UpdateDodge()
	if game.player.CanDodge() {
		t.Fatal("the player can dodge again immediately after dodging")
	}

	for range game.player.DodgeCooldown {
		game.player.UpdateDodge()
	}
	if !game.player.CanDodge() {
		t.Fatal("the player can never dodge again after the cooldown")
	}
}

// TestDodgeGivesIFrames makes sure the player takes no damage while the
// dodge roll lasts, that the bullet flies through him and that he can be
// hit again once the roll is over.
func TestDodgeGivesIFrames(t *testing.T) {
	game := NewGame()

	enemy := game.enemies[0]
	enemy.Bullets = []actors.Bullet{
		{
			X:        game.player.X,
			Y:        game.player.Y,
			W:        12,
			H:        12,
			Speed:    0,
			Damage:   3,
			Duration: 100,
			Hitbox: actors.HitBox{
				X: game.player.Hitbox.X,
				Y: game.player.Hitbox.Y,
				W: game.player.Hitbox.W,
				H: game.player.Hitbox.H,
			},
		},
	}
	healthBefore := game.player.Health

	game.player.CurrentAction = actors.Action{
		Duration: game.player.DodgeDuration,
		Type:     actors.Dodge,
		Actor:    game.player,
	}

	game.CheckCollisions()
	if game.player.Health != healthBefore {
		t.Fatal("the player took damage while dodging")
	}
	if len(enemy.Bullets) != 1 {
		t.Fatal("the bullet should fly through the dodging player")
	}

	game.player.CurrentAction.Duration = 0
	game.CheckCollisions()
	if game.player.Health >= healthBefore {
		t.Fatal("the player took no damage after the dodge ended")
	}
}

// TestDeadEyeDrainsAndRegenerates makes sure the dead eye gauge runs dry
// while the ability is active and only trickles back afterwards.
func TestDeadEyeDrainsAndRegenerates(t *testing.T) {
	game := NewGame()
	game.player.DeadEyeCharge = game.player.DeadEyeMaxCharge

	game.toggleDeadEye()
	if !game.deadEyeActive {
		t.Fatal("dead eye did not activate with a full gauge")
	}

	chargeBefore := game.player.DeadEyeCharge
	game.updateDeadEye()
	if game.player.DeadEyeCharge >= chargeBefore {
		t.Fatal("dead eye did not drain while active")
	}

	game.player.DeadEyeCharge = deadEyeDrainPerFrame / 2
	game.updateDeadEye()
	if game.deadEyeActive {
		t.Fatal("dead eye stayed active after the gauge ran dry")
	}
	if game.player.DeadEyeCharge != 0 {
		t.Fatalf("dead eye charge is %f after running dry, want 0", game.player.DeadEyeCharge)
	}

	game.updateDeadEye()
	if game.player.DeadEyeCharge <= 0 {
		t.Fatal("dead eye did not regenerate after it was switched off")
	}
}

// TestDeadEyeNeedsCharge makes sure an empty gauge cannot be activated.
func TestDeadEyeNeedsCharge(t *testing.T) {
	game := NewGame()
	game.player.DeadEyeCharge = 0

	game.toggleDeadEye()
	if game.deadEyeActive {
		t.Fatal("dead eye activated without any charge")
	}
}

// TestDeadEyeChargeUpgrade makes sure a fresh player starts with a full gauge
// and that the store upgrade raises the maximum and refills some charge.
func TestDeadEyeChargeUpgrade(t *testing.T) {
	game := NewGame()
	if game.player.DeadEyeMaxCharge != actors.DefaultDeadEyeMaxCharge {
		t.Fatalf("fresh dead eye max charge is %f, want %f", game.player.DeadEyeMaxCharge, actors.DefaultDeadEyeMaxCharge)
	}
	if game.player.DeadEyeCharge != game.player.DeadEyeMaxCharge {
		t.Fatal("a fresh player does not start with a full dead eye gauge")
	}

	game.player.DeadEyeCharge = 40
	if !game.player.ApplyUpgrade(actors.UpgradeDeadEye, defaultUpgradeLimits) {
		t.Fatal("the dead eye upgrade could not be applied")
	}

	if game.player.DeadEyeMaxCharge != actors.DefaultDeadEyeMaxCharge+actors.DeadEyeChargeUpgrade {
		t.Fatalf("dead eye max charge is %f after the upgrade", game.player.DeadEyeMaxCharge)
	}
	if game.player.DeadEyeCharge != 40+actors.DeadEyeChargeUpgrade {
		t.Fatalf("dead eye charge is %f after the upgrade", game.player.DeadEyeCharge)
	}
}

// TestDeadEyeSlowsWorldButNotPlayer makes sure enemies and their bullets slow
// down while the player keeps moving and firing at full speed.
func TestDeadEyeSlowsWorldButNotPlayer(t *testing.T) {
	game := NewGame()
	enemy := game.enemies[0]

	enemyX := enemy.X
	enemy.TimeScale = 0
	enemy.Move(actors.Right)
	normalStep := enemy.X - enemyX

	enemy.X = enemyX
	enemy.TimeScale = deadEyeTimeScale
	enemy.Move(actors.Right)
	slowStep := enemy.X - enemyX

	if slowStep >= normalStep {
		t.Fatalf("the enemy moved %f under dead eye and %f normally, it was not slowed down", slowStep, normalStep)
	}
	if math.Abs(slowStep-normalStep*deadEyeTimeScale) > 1e-9 {
		t.Fatalf("the enemy moved %f, want %f", slowStep, normalStep*deadEyeTimeScale)
	}

	playerX := game.player.X
	game.deadEyeActive = true
	game.player.Move(actors.Right)
	playerStep := game.player.X - playerX
	if math.Abs(playerStep-game.player.Speed) > 1e-9 {
		t.Fatalf("the player moved %f under dead eye, want %f", playerStep, game.player.Speed)
	}

	normalBullet := &actors.Bullet{Speed: 10, Duration: 10}
	normalBullet.Update()
	slowBullet := &actors.Bullet{Speed: 10, Duration: 10, TimeScale: deadEyeTimeScale}
	slowBullet.Update()

	if slowBullet.X >= normalBullet.X {
		t.Fatalf("the slowed bullet travelled %f while the normal bullet travelled %f", slowBullet.X, normalBullet.X)
	}
}

// TestBulletsAlwaysDealDamage makes sure the damage roll never produces a
// zero damage bullet, a shot that does nothing is not fun.
func TestBulletsAlwaysDealDamage(t *testing.T) {
	game := NewGame()
	game.player.DrawWeapon(int(actors.Revolver))
	game.player.GunBelt[game.player.CurrentWeapon].Ammo = 1000

	for range 50 {
		game.player.ShootTimer = 0
		game.player.GunBelt[game.player.CurrentWeapon].CurrentMagazineSize = game.player.GunBelt[game.player.CurrentWeapon].MagazineSize
		game.player.Reloading = false
		game.player.ReloadTimer = 0
		game.player.Shoot()
	}

	if len(game.player.Bullets) != 50 {
		t.Fatalf("expected 50 bullets, got %d", len(game.player.Bullets))
	}
	for _, bullet := range game.player.Bullets {
		if bullet.Damage < 1 {
			t.Fatalf("a bullet dealt %d damage, every shot should hurt", bullet.Damage)
		}
	}
}

func TestFistsShootWithoutAmmo(t *testing.T) {
	game := NewGame()
	game.player.DrawWeapon(int(actors.Fists))
	fists := game.player.GunBelt[game.player.CurrentWeapon]
	if fists.GunType != actors.Fists {
		t.Fatal("the second belt slot does not hold the fists")
	}
	fists.Ammo = 0
	fists.CurrentMagazineSize = 0
	game.player.ShootTimer = 0
	game.player.Reloading = false

	if !game.player.CanShoot() {
		t.Fatal("the fists cannot be used without ammo")
	}
	if fists.AmmoFace != nil {
		t.Fatal("the fists still show an ammo counter")
	}

	game.player.Shoot()

	if fists.Ammo != 0 {
		t.Fatal("punching used ammo")
	}
	if game.player.Reloading {
		t.Fatal("punching started a reload")
	}
	if game.player.MeleeTimer <= 0 {
		t.Fatal("punching did not start the sweep animation")
	}
	if game.player.ShootTimer <= 0 {
		t.Fatal("punching did not start the cooldown")
	}
}

func TestFistSweepHitsEnemiesInFront(t *testing.T) {
	game := NewGame()
	player := game.player
	player.DrawWeapon(int(actors.Fists))
	player.AimAngle = 0
	player.ShootTimer = 0
	player.Reloading = false

	target := game.enemies[0]
	target.Health = 100
	target.MaxHealth = 100
	target.X = player.X + player.W/2 + 40 - target.W/2
	target.Y = player.Y + player.H/2 - target.H/2
	target.UpdateHitboxOffset(enemyHitboxOffset)

	behind := game.enemies[1]
	behind.Health = 100
	behind.MaxHealth = 100
	behind.X = player.X + player.W/2 - 40 - behind.W/2
	behind.Y = player.Y + player.H/2 - behind.H/2
	behind.UpdateHitboxOffset(enemyHitboxOffset)

	healthBefore := target.Health
	behindBefore := behind.Health
	damage := player.GunBelt[player.CurrentWeapon].Damage + player.FistDamageBonus

	player.Shoot()

	if target.Health != healthBefore-damage {
		t.Fatalf("the enemy in front took %d damage, expected %d", healthBefore-target.Health, damage)
	}
	if behind.Health != behindBefore {
		t.Fatal("the enemy behind the player was hit by the sweep")
	}
}

func TestFistUpgradeIncreasesSweepDamage(t *testing.T) {
	game := NewGame()
	player := game.player
	player.DrawWeapon(int(actors.Fists))
	player.AimAngle = 0
	player.ShootTimer = 0
	player.Reloading = false

	target := game.enemies[0]
	target.Health = 100
	target.MaxHealth = 100
	target.X = player.X + player.W/2 + 40 - target.W/2
	target.Y = player.Y + player.H/2 - target.H/2
	target.UpdateHitboxOffset(enemyHitboxOffset)

	baseDamage := player.GunBelt[player.CurrentWeapon].Damage
	player.Shoot()
	baseDealt := 100 - target.Health
	if baseDealt != baseDamage {
		t.Fatalf("the base punch dealt %d damage, expected %d", baseDealt, baseDamage)
	}

	player.ApplyUpgrade(actors.UpgradeFists, defaultUpgradeLimits)
	target.Health = 100
	player.ShootTimer = 0
	player.Shoot()
	upgradedDealt := 100 - target.Health
	if upgradedDealt != baseDamage+player.FistDamageBonus {
		t.Fatalf("the upgraded punch dealt %d damage, expected %d", upgradedDealt, baseDamage+player.FistDamageBonus)
	}
	if upgradedDealt <= baseDealt {
		t.Fatal("the fist upgrade did not increase the punch damage")
	}
}

// TestKillTriggersHitStop makes sure a kill freezes the game for a couple of
// frames before it keeps running.
func TestKillTriggersHitStop(t *testing.T) {
	game := NewGame()
	game.mode = ModeGame

	game.killEnemy(game.enemies[0])
	if game.freezeFrames <= 0 {
		t.Fatal("a kill did not start the hit-stop")
	}

	framesBefore := game.elapsedFrames
	if err := game.Update(); err != nil {
		t.Fatal(err)
	}
	if game.elapsedFrames != framesBefore {
		t.Fatal("the game kept running during the kill hit-stop")
	}

	for range 3 {
		if err := game.Update(); err != nil {
			t.Fatal(err)
		}
	}
	if game.elapsedFrames <= framesBefore {
		t.Fatal("the game never resumed after the kill hit-stop")
	}
}

// TestEnemyTierRolls makes sure the tier roll is ordered and that the rare and
// legendary bandits can actually spawn.
func TestEnemyTierRolls(t *testing.T) {
	legendary, rare := enemyTierChances(1)
	if legendary <= 0 || rare <= 0 {
		t.Fatal("rare and legendary enemies can never spawn")
	}
	if legendary+rare >= 1 {
		t.Fatal("the enemy tier chances overlap")
	}

	if tier := enemyTierForRoll(1, 0); tier != actors.Legendary {
		t.Fatalf("the lowest roll should produce a legendary enemy, got %d", tier)
	}
	if tier := enemyTierForRoll(1, 0.99); tier != actors.Common {
		t.Fatalf("the highest roll should produce a common enemy, got %d", tier)
	}

	for _, roll := range []float64{0, 0.1, 0.3, 0.5, 0.99} {
		tier := enemyTierForRoll(1, roll)
		if tier < actors.Common || tier > actors.Legendary {
			t.Fatalf("roll %f produced the invalid tier %d", roll, tier)
		}
	}
}

// TestEnemyTierStats makes sure a higher tier is tougher in every stat.
func TestEnemyTierStats(t *testing.T) {
	common := enemyStatsForTier(actors.Common)
	rare := enemyStatsForTier(actors.Rare)
	legendary := enemyStatsForTier(actors.Legendary)

	if common.Health >= rare.Health || rare.Health >= legendary.Health {
		t.Fatal("higher tier enemies should have more health")
	}
	if common.Damage >= rare.Damage || rare.Damage >= legendary.Damage {
		t.Fatal("higher tier enemies should deal more damage")
	}
	if common.Speed >= rare.Speed || rare.Speed >= legendary.Speed {
		t.Fatal("higher tier enemies should move faster")
	}
	if common.FireRate >= rare.FireRate || rare.FireRate >= legendary.FireRate {
		t.Fatal("higher tier enemies should shoot faster")
	}
}

// TestGunTierRolls makes sure tougher enemies are more likely to carry
// tougher guns and that the highest gun tier is reachable.
func TestGunTierRolls(t *testing.T) {
	if tier := gunTierForRoll(actors.Legendary, 0); tier != actors.Legendary {
		t.Fatalf("a legendary enemy should be able to carry a legendary gun, got %d", tier)
	}
	if tier := gunTierForRoll(actors.Common, 0.99); tier != actors.Common {
		t.Fatalf("a common enemy should not always carry a rare gun, got %d", tier)
	}

	// the better the enemy the bigger the chance of a better gun
	for _, roll := range []float64{0.1, 0.3, 0.6, 0.9} {
		common := gunTierForRoll(actors.Common, roll)
		rare := gunTierForRoll(actors.Rare, roll)
		legendary := gunTierForRoll(actors.Legendary, roll)
		if common > rare || rare > legendary {
			t.Fatalf("gun tiers do not grow with the enemy tier at roll %f", roll)
		}
	}
}

// TestNewEnemyGunTiers makes sure a gun keeps its tier and that higher tier
// guns roll better minimum stats.
func TestNewEnemyGunTiers(t *testing.T) {
	game := NewGame()

	common := game.newEnemyGun(actors.Revolver, actors.Common, 2)
	if common.Tier != actors.Common {
		t.Fatal("the gun did not keep its common tier")
	}
	if common.Damage < 2 || common.MagazineSize < defaultMagazineSize[actors.Revolver] {
		t.Fatal("a common gun rolled worse stats than its base")
	}
	if common.CurrentMagazineSize != common.MagazineSize {
		t.Fatal("the common gun was not loaded")
	}
	if common.Ammo <= 0 || common.ShootCooldown <= 0 || common.BulletSpeed <= 0 {
		t.Fatal("the common gun has invalid stats")
	}

	legendary := game.newEnemyGun(actors.Revolver, actors.Legendary, 2)
	if legendary.Tier != actors.Legendary {
		t.Fatal("the gun did not keep its legendary tier")
	}
	if legendary.Damage < 4 || legendary.MagazineSize < defaultMagazineSize[actors.Revolver]+2 {
		t.Fatal("a legendary gun should roll better stats than its base")
	}
	if legendary.CurrentMagazineSize != legendary.MagazineSize {
		t.Fatal("the legendary gun was not loaded")
	}
}

// TestKillEnemyDropsGun makes sure an enemy leaves his gun on the floor.
func TestKillEnemyDropsGun(t *testing.T) {
	game := NewGame()
	enemy := game.enemies[0]
	enemyGun := enemy.GunBelt[0]

	game.killEnemy(enemy)

	if len(game.loot) > 2 {
		t.Fatalf("expected one gun or one gun and whisky on the floor, got %d", len(game.loot))
	}
	dropped := game.loot[0]
	if dropped.Gun.GunType != enemyGun.GunType || dropped.Gun.Tier != enemyGun.Tier {
		t.Fatal("the dropped gun does not match the enemy gun")
	}
	if dropped.X != enemy.X+enemy.W/2 || dropped.Y != enemy.Y+enemy.H/2 {
		t.Fatal("the gun was not dropped at the enemy position")
	}
}

// TestNearestLoot makes sure only nearby guns are considered for pick up.
func TestNearestLoot(t *testing.T) {
	game := NewGame()
	if game.nearestLoot() != nil {
		t.Fatal("there is loot on a fresh run")
	}

	px := game.player.X + game.player.W/2
	py := game.player.Y + game.player.H/2
	near := actors.NewLoot(&actors.Gun{}, px+10, py)
	far := actors.NewLoot(&actors.Gun{}, px+lootPickupRadius+10, py)

	game.loot = []*actors.Loot{far, near}
	if got := game.nearestLoot(); got != near {
		t.Fatal("the nearest gun on the floor was not selected")
	}

	game.loot = []*actors.Loot{far}
	if game.nearestLoot() != nil {
		t.Fatal("a gun out of reach was selected")
	}
}

// TestPickUpLootSwapsSameTypeGun makes sure picking up a gun replaces the gun
// of the same type in the belt, not the one the player is holding.
func TestPickUpLootSwapsSameTypeGun(t *testing.T) {
	game := NewGame()
	game.player.DrawWeapon(int(actors.Fists))
	shotgunIndex, ok := game.gunBeltIndexFor(actors.Shotgun)
	if !ok {
		t.Fatal("the player has no shotgun slot")
	}
	oldGun := game.player.GunBelt[shotgunIndex]

	lootGun := game.newEnemyGun(actors.Shotgun, actors.Rare, 2)
	loot := actors.NewLoot(&lootGun, game.player.X, game.player.Y)
	game.loot = append(game.loot, loot)

	game.pickUpLoot(loot)

	if game.player.GunBelt[shotgunIndex] != &lootGun {
		t.Fatal("the player did not equip the looted gun in the shotgun slot")
	}
	if game.player.CurrentWeapon != int(actors.Fists) {
		t.Fatal("picking up a gun of another type switched the held weapon")
	}
	if game.player.GunBelt[shotgunIndex].X != oldGun.X || game.player.GunBelt[shotgunIndex].Y != oldGun.Y {
		t.Fatal("the looted gun did not take over the HUD slot of the old gun")
	}
	if len(game.loot) != 1 || game.loot[0].Gun != oldGun {
		t.Fatal("the old gun was not dropped on the floor")
	}
	if game.loot[0].X != loot.X || game.loot[0].Y != loot.Y {
		t.Fatal("the old gun was not dropped where the new one was found")
	}
}

// TestPickUpLootTossesOldGun makes sure the swapped-out gun slides in the
// direction the player is aiming.
func TestPickUpLootTossesOldGun(t *testing.T) {
	game := NewGame()
	game.player.AimAngle = math.Pi / 2
	revolverIndex, ok := game.gunBeltIndexFor(actors.Revolver)
	if !ok {
		t.Fatal("the player has no revolver slot")
	}
	oldGun := game.player.GunBelt[revolverIndex]

	lootGun := game.newEnemyGun(actors.Revolver, actors.Rare, 2)
	loot := actors.NewLoot(&lootGun, game.player.X, game.player.Y)
	game.loot = append(game.loot, loot)

	game.pickUpLoot(loot)

	if len(game.loot) != 1 || game.loot[0].Gun != oldGun {
		t.Fatal("the old gun was not dropped")
	}

	dropped := game.loot[0]
	startX, startY := dropped.X, dropped.Y
	if math.Abs(dropped.VX) > 0.01 || dropped.VY <= 0 {
		t.Fatal("the dropped gun was not tossed in the aim direction")
	}

	dropped.Update()
	if dropped.Y <= startY {
		t.Fatal("the dropped gun did not move forward")
	}
	if math.Abs(dropped.X-startX) > 0.01 {
		t.Fatal("the dropped gun moved sideways")
	}
}

// TestPickUpLootWithFistsKeepsFists makes sure picking up a gun only touches
// its matching slot while the player is holding his fists.
func TestPickUpLootWithFistsKeepsFists(t *testing.T) {
	game := NewGame()
	game.player.DrawWeapon(int(actors.Fists))
	revolverIndex, ok := game.gunBeltIndexFor(actors.Revolver)
	if !ok {
		t.Fatal("the player has no revolver slot")
	}
	oldRevolver := game.player.GunBelt[revolverIndex]

	lootGun := game.newEnemyGun(actors.Revolver, actors.Common, 2)
	loot := actors.NewLoot(&lootGun, game.player.X, game.player.Y)
	game.loot = append(game.loot, loot)

	game.pickUpLoot(loot)

	if game.player.CurrentWeapon != int(actors.Fists) {
		t.Fatal("picking up a gun switched the held weapon")
	}
	if game.player.GunBelt[0].GunType != actors.Whisky {
		t.Fatal("the whisky was replaced")
	}
	if game.player.GunBelt[revolverIndex] != &lootGun {
		t.Fatal("the revolver slot was not replaced")
	}
	if len(game.loot) != 1 || game.loot[0].Gun != oldRevolver {
		t.Fatal("the old revolver was not dropped")
	}
}

// TestPickUpLootRequiresMatchingType makes sure loot without a matching gun
// type in the belt can neither be selected nor picked up.
func TestPickUpLootRequiresMatchingType(t *testing.T) {
	game := NewGame()
	for _, gun := range game.player.GunBelt {
		if gun.GunType == actors.Shotgun {
			gun.GunType = actors.Fists
		}
	}

	px := game.player.X + game.player.W/2
	py := game.player.Y + game.player.H/2
	lootGun := game.newEnemyGun(actors.Shotgun, actors.Common, 2)
	loot := actors.NewLoot(&lootGun, px, py)
	game.loot = []*actors.Loot{loot}

	if game.nearestLoot() != nil {
		t.Fatal("loot without a matching gun type was selected")
	}

	game.pickUpLoot(loot)
	if len(game.loot) != 1 || game.loot[0] != loot {
		t.Fatal("loot without a matching gun type was picked up")
	}
}

// TestGunBeltIndexFor makes sure every gun type can be found in the belt.
func TestGunBeltIndexFor(t *testing.T) {
	game := NewGame()
	for _, weapon := range []actors.Weapon{actors.Fists, actors.Revolver, actors.Shotgun, actors.MachineGun} {
		index, ok := game.gunBeltIndexFor(weapon)
		if !ok {
			t.Fatalf("no belt slot found for weapon %d", weapon)
		}
		if game.player.GunBelt[index].GunType != weapon {
			t.Fatalf("the belt slot of weapon %d holds the wrong gun", weapon)
		}
	}
}

// TestStatIsBetter guards the color logic of the comparison overlay.
func TestStatIsBetter(t *testing.T) {
	if !statIsBetter(2, true) || statIsBetter(-2, true) || statIsBetter(0, true) {
		t.Fatal("higher is better stats are compared incorrectly")
	}
	if !statIsBetter(-3, false) || statIsBetter(3, false) || statIsBetter(0, false) {
		t.Fatal("lower is better stats are compared incorrectly")
	}
}

// TestGunStatRows makes sure the comparison only shows the stats that matter.
func TestGunStatRows(t *testing.T) {
	current := &actors.Gun{Damage: 3, MagazineSize: 6, Ammo: 99, ShootCooldown: 16, BulletSpeed: 6}
	loot := &actors.Gun{
		GunType:       actors.MachineGun,
		Damage:        5,
		MagazineSize:  8,
		Ammo:          120,
		ShootCooldown: 12,
		BulletSpeed:   7,
		BulletSpread:  42,
	}

	rows := gunStatRows(current, loot)
	if len(rows) != 6 {
		t.Fatalf("expected six comparison rows, got %d", len(rows))
	}
	if rows[0].now != 3 || rows[0].loot != 5 {
		t.Fatal("the damage row does not show both guns")
	}
	if rows[len(rows)-1].label != "ACCURACY" {
		t.Fatal("the machine gun accuracy is missing from the comparison")
	}

	shotgun := &actors.Gun{GunType: actors.Shotgun}
	rows = gunStatRows(current, shotgun)
	if rows[len(rows)-1].label != "PELLETS" {
		t.Fatal("the shotgun pellet count is missing from the comparison")
	}
}

// TestEnemyWhiskyRolls makes sure tougher enemies are more likely to carry a
// whisky as a secondary weapon.
func TestEnemyWhiskyRolls(t *testing.T) {
	common := enemyWhiskyChance(actors.Common)
	rare := enemyWhiskyChance(actors.Rare)
	legendary := enemyWhiskyChance(actors.Legendary)

	if common >= rare || rare >= legendary {
		t.Fatal("higher tier enemies should be more likely to carry whisky")
	}
	if !enemyHasWhisky(actors.Common, 0) || enemyHasWhisky(actors.Common, 0.99) {
		t.Fatal("the common whisky roll is out of bounds")
	}
	if !enemyHasWhisky(actors.Legendary, 0.3) {
		t.Fatal("a legendary enemy should be able to carry whisky on a moderate roll")
	}
}

// TestNewEnemyWhisky makes sure enemy whisky has usable healing stats.
func TestNewEnemyWhisky(t *testing.T) {
	game := NewGame()
	whisky := game.newEnemyGun(actors.Whisky, actors.Legendary, 2)

	if whisky.GunType != actors.Whisky || whisky.Tier != actors.Legendary {
		t.Fatal("the whisky did not keep its type and tier")
	}
	if whisky.Damage < 2 || whisky.Ammo <= 0 || whisky.ShootCooldown <= 0 {
		t.Fatal("the whisky has invalid stats")
	}
	if whisky.Sprite == nil {
		t.Fatal("the whisky has no sprite")
	}
}

// TestGunStatRowsWhisky makes sure the comparison card only shows the stats
// that matter for a bottle of whisky.
func TestGunStatRowsWhisky(t *testing.T) {
	current := &actors.Gun{GunType: actors.Whisky, Damage: 2, Ammo: 3, ShootCooldown: 32}
	loot := &actors.Gun{GunType: actors.Whisky, Damage: 4, Ammo: 5, ShootCooldown: 24}

	rows := gunStatRows(current, loot)
	if len(rows) != 3 {
		t.Fatalf("expected three whisky comparison rows, got %d", len(rows))
	}
	if rows[0].label != "HEALING" || rows[0].now != 2 || rows[0].loot != 4 {
		t.Fatal("the healing row does not compare both bottles")
	}
}

// TestNewEnemyCigar makes sure a cigar drop has usable dead eye refill stats.
func TestNewEnemyCigar(t *testing.T) {
	game := NewGame()
	cigar := game.newEnemyGun(actors.Cigar, actors.Legendary, 4)

	if cigar.GunType != actors.Cigar || cigar.Tier != actors.Legendary {
		t.Fatal("the cigar did not keep its type and tier")
	}
	if cigar.Damage < 4 || cigar.Ammo <= 0 || cigar.ShootCooldown <= 0 {
		t.Fatal("the cigar has invalid stats")
	}
	if cigar.Sprite == nil {
		t.Fatal("the cigar has no sprite")
	}
}

// TestGunStatRowsCigar makes sure the comparison card shows the dead eye
// refill of a cigar instead of combat stats.
func TestGunStatRowsCigar(t *testing.T) {
	current := &actors.Gun{GunType: actors.Cigar, Damage: 4, Ammo: 3, ShootCooldown: 32}
	loot := &actors.Gun{GunType: actors.Cigar, Damage: 6, Ammo: 5, ShootCooldown: 24}

	rows := gunStatRows(current, loot)
	if len(rows) != 3 {
		t.Fatalf("expected three cigar comparison rows, got %d", len(rows))
	}
	if rows[0].label != "DEAD EYE" || rows[0].now != 4 || rows[0].loot != 6 {
		t.Fatal("the dead eye row does not compare both cigars")
	}
}

// TestOpenCigarChestDropsCigar makes sure a cigar chest leaves a cigar on the
// floor instead of gold or a gun.
func TestOpenCigarChestDropsCigar(t *testing.T) {
	game := NewGame()
	lootBefore := len(game.loot)
	chest := &world.Chest{
		X:     100,
		Y:     100,
		W:     32,
		H:     32,
		Scale: 2,
		State: world.Closed,
		Type:  world.Cigar,
	}

	game.openChest(chest)

	if chest.State != world.Open {
		t.Fatal("the cigar chest did not open")
	}
	if len(game.loot) != lootBefore+1 {
		t.Fatalf("expected one cigar drop, got %d new loot items", len(game.loot)-lootBefore)
	}
	if game.loot[len(game.loot)-1].Gun.GunType != actors.Cigar {
		t.Fatal("the cigar chest did not drop a cigar")
	}
}

// TestNewStoreCigar makes sure the store can offer cigars with random stats.
func TestNewStoreCigar(t *testing.T) {
	game := NewGame()
	item := game.newStoreCigarItem()

	if item.Kind != world.StoreCigar || item.Gun == nil {
		t.Fatal("the store cigar item is incomplete")
	}
	if item.Gun.GunType != actors.Cigar || item.Gun.Sprite == nil {
		t.Fatal("the store cigar has no usable gun")
	}
	if item.Amount <= 0 || item.Gun.Ammo != item.Amount || item.Price <= 0 {
		t.Fatal("the store cigar has invalid stats or price")
	}
}

func firstStoreItem(game *Game, kind world.StoreItemKind) *world.StoreItem {
	for i := range game.store.Stock {
		if game.store.Stock[i].Kind == kind {
			return &game.store.Stock[i]
		}
	}

	return nil
}

func addStoreItem(game *Game, item world.StoreItem) *world.StoreItem {
	game.store.Stock = append(game.store.Stock, item)

	return &game.store.Stock[len(game.store.Stock)-1]
}

func TestStoreStockGeneration(t *testing.T) {
	game := NewGame()
	if game.store == nil {
		t.Fatal("a level was spawned without a store")
	}

	counts := map[world.StoreItemKind]int{}
	upgrades := map[actors.Upgrade]bool{}
	for _, item := range game.store.Stock {
		counts[item.Kind]++
		if item.Price <= 0 {
			t.Fatalf("store item %q has no price", item.Name)
		}

		switch item.Kind {
		case world.StoreAmmo:
			if item.Amount <= 0 {
				t.Fatalf("ammo item %q has no ammo", item.Name)
			}
		case world.StoreWeapon, world.StoreWhisky, world.StoreCigar:
			if item.Gun == nil {
				t.Fatalf("gun item %q has no gun", item.Name)
			}
		case world.StoreUpgrade:
			if upgrades[item.Upgrade] {
				t.Fatalf("upgrade %q is offered twice", item.Name)
			}
			upgrades[item.Upgrade] = true
		}
	}

	if counts[world.StoreAmmo] < storeAmmoOffersMin || counts[world.StoreAmmo] > storeAmmoOffersMax {
		t.Fatalf("expected %d to %d ammo offers, got %d", storeAmmoOffersMin, storeAmmoOffersMax, counts[world.StoreAmmo])
	}
	if counts[world.StoreWeapon]+counts[world.StoreWhisky]+counts[world.StoreCigar] != storeWeaponOffers {
		t.Fatalf("expected %d weapon or consumable offers, got %d", storeWeaponOffers, counts[world.StoreWeapon]+counts[world.StoreWhisky]+counts[world.StoreCigar])
	}
	if counts[world.StoreUpgrade] < storeUpgradeOffersMin || counts[world.StoreUpgrade] > storeUpgradeOffersMax {
		t.Fatalf("expected %d to %d upgrade offers, got %d", storeUpgradeOffersMin, storeUpgradeOffersMax, counts[world.StoreUpgrade])
	}
}

func TestBuyStoreAmmo(t *testing.T) {
	game := NewGame()
	item := firstStoreItem(game, world.StoreAmmo)
	if item == nil {
		t.Fatal("the store has no ammo")
	}

	index, ok := game.gunBeltIndexFor(item.Weapon)
	if !ok {
		t.Fatal("the player has no gun for the ammo on sale")
	}
	ammoBefore := game.player.GunBelt[index].Ammo
	game.player.Gold = item.Price + 7

	game.buyStoreItem(item)

	if game.player.Gold != 7 {
		t.Fatalf("expected 7 gold left, got %d", game.player.Gold)
	}
	if game.player.GunBelt[index].Ammo != ammoBefore+item.Amount {
		t.Fatal("the bought ammo was not added to the matching gun")
	}
	if !item.Sold {
		t.Fatal("the ammo offer was not marked as sold")
	}
}

func TestBuyStoreWeaponSwapsAndDropsOldGun(t *testing.T) {
	game := NewGame()
	index, ok := game.gunBeltIndexFor(actors.Shotgun)
	if !ok {
		t.Fatal("the player has no shotgun slot")
	}
	oldGun := game.player.GunBelt[index]
	gun := game.newEnemyGun(actors.Shotgun, actors.Rare, 3)
	item := addStoreItem(game, world.StoreItem{
		Kind:   world.StoreWeapon,
		Name:   "SHOTGUN",
		Price:  120,
		Weapon: actors.Shotgun,
		Gun:    &gun,
	})
	game.player.Gold = 200
	game.loot = nil

	game.buyStoreItem(item)

	if game.player.GunBelt[index] != item.Gun {
		t.Fatal("the bought gun was not equipped in the matching slot")
	}
	if game.player.GunBelt[index].X != oldGun.X || game.player.GunBelt[index].Y != oldGun.Y {
		t.Fatal("the bought gun did not take over the HUD position of the old gun")
	}
	if len(game.loot) != 1 || game.loot[0].Gun != oldGun {
		t.Fatal("the old gun was not dropped")
	}
	if game.player.Gold != 80 {
		t.Fatalf("expected 80 gold left, got %d", game.player.Gold)
	}
	if !item.Sold {
		t.Fatal("the weapon offer was not marked as sold")
	}
}

func TestBuyStoreUpgrade(t *testing.T) {
	game := NewGame()
	upgrade := actors.UpgradeDamage
	upgradeLimit := defaultUpgradeLimits[upgrade]
	items := []*world.StoreItem{}
	for range upgradeLimit {
		items = append(items,
			addStoreItem(game, world.StoreItem{
				Kind:    world.StoreUpgrade,
				Name:    upgrade.String(),
				Price:   100,
				Upgrade: upgrade,
			}))
	}
	game.player.Gold = upgradeLimit * 100

	for _, item := range items {
		game.buyStoreItem(item)
	}

	if game.player.DamageBonus != 2*upgradeLimit {
		t.Fatalf("expected %d bonus damage, got %d", 2*upgradeLimit, game.player.DamageBonus)
	}
	if game.player.GetUpgradeCount(actors.UpgradeDamage) == 0 {
		t.Fatal("the upgrade was not recorded as owned")
	}
	if game.player.Gold != 0 {
		t.Fatalf("expected no gold left, got %d", game.player.Gold)
	}
	for _, item := range items {
		if !item.Sold {
			t.Fatal("the upgrade offer was not marked as sold")
		}
	}

	for _, upgrade := range game.availableUpgrades() {
		if upgrade == actors.UpgradeDamage {
			t.Fatal("an owned upgrade is still offered")
		}
	}
}

func TestBuyStoreMaxHealthUpgrade(t *testing.T) {
	game := NewGame()
	maxHealth := 20
	health := 10
	game.player.Health = health
	game.player.MaxHealth = maxHealth
	game.player.UpdateHealhBar()
	upgrade := actors.UpgradeMaxHealth
	upgradeLimit := defaultUpgradeLimits[upgrade]
	items := []*world.StoreItem{}
	for range upgradeLimit {
		items = append(items,
			addStoreItem(game, world.StoreItem{
				Kind:    world.StoreUpgrade,
				Name:    upgrade.String(),
				Price:   100,
				Upgrade: upgrade,
			}))
	}
	game.player.Gold = upgradeLimit * 100

	for _, item := range items {
		game.buyStoreItem(item)
	}

	if game.player.MaxHealth != upgradeLimit*10+maxHealth {
		t.Fatalf("expected %d max health, got %d", upgradeLimit*10+maxHealth, game.player.MaxHealth)
	}
	if game.player.Health != upgradeLimit*10+health {
		t.Fatalf("expected %d health, got %d", upgradeLimit*5+health, game.player.Health)
	}
	if game.player.Healthbar.MaxPoints != game.player.MaxHealth {
		t.Fatal("the healthbar was not updated for the new max health")
	}
}

func TestBuyStoreItemWithoutEnoughGold(t *testing.T) {
	game := NewGame()
	gun := game.newEnemyGun(actors.Revolver, actors.Rare, 3)
	item := addStoreItem(game, world.StoreItem{
		Kind:   world.StoreWeapon,
		Name:   "REVOLVER",
		Price:  999,
		Weapon: actors.Revolver,
		Gun:    &gun,
	})
	index, ok := game.gunBeltIndexFor(actors.Revolver)
	if !ok {
		t.Fatal("the player has no revolver slot")
	}
	oldGun := game.player.GunBelt[index]
	game.player.Gold = 10

	game.buyStoreItem(item)

	if game.player.Gold != 10 {
		t.Fatalf("gold changed without a purchase, got %d", game.player.Gold)
	}
	if item.Sold {
		t.Fatal("an unaffordable item was marked as sold")
	}
	if game.player.GunBelt[index] != oldGun {
		t.Fatal("an unaffordable weapon was equipped")
	}
}

func TestOwnedUpgradesAreNotStocked(t *testing.T) {
	game := NewGame()
	for _, upgrade := range actors.AllUpgrades {
		for range defaultUpgradeLimits[upgrade] {
			game.player.ApplyUpgrade(upgrade, defaultUpgradeLimits)
		}
	}

	game.spawnStoreStock()

	for _, item := range game.store.Stock {
		if item.Kind == world.StoreUpgrade {
			t.Fatalf("owned upgrade %q was stocked again", item.Name)
		}
	}
}

func TestKillEnemyGrantsGold(t *testing.T) {
	game := NewGame()
	tiers := []struct {
		tier actors.Tier
		gold int
	}{
		{actors.Common, commonKillGold},
		{actors.Rare, rareKillGold},
		{actors.Legendary, legendaryKillGold},
	}

	for i, tc := range tiers {
		enemy := game.enemies[i]
		enemy.Tier = tc.tier
		goldBefore := game.player.Gold
		game.killEnemy(enemy)
		if game.player.Gold != goldBefore+tc.gold {
			t.Fatalf("expected %d gold for a %s kill, got %d", tc.gold, tc.tier, game.player.Gold-goldBefore)
		}
	}
}

func TestStoreDoesNotFreezeWorld(t *testing.T) {
	game := NewGame()
	game.mode = ModeGame
	game.storeOpen = true
	framesBefore := game.frameCount
	elapsedBefore := game.elapsedFrames
	xBefore, yBefore := game.player.X, game.player.Y

	if err := game.Update(); err != nil {
		t.Fatal(err)
	}

	if game.elapsedFrames != elapsedBefore+1 {
		t.Fatal("time did not advance while the store was open")
	}
	if game.frameCount == framesBefore {
		t.Fatal("the animation frame counter did not advance while the store was open")
	}
	if game.player.X != xBefore || game.player.Y != yBefore {
		t.Fatal("the player moved while shopping")
	}
	if !game.storeOpen {
		t.Fatal("the store closed without input")
	}
}

func TestNearestStore(t *testing.T) {
	game := NewGame()
	game.player.X = game.store.Hitbox.CenterX() - game.player.W/2
	game.player.Y = game.store.Hitbox.CenterY() - game.player.H/2
	game.player.UpdateHitbox()
	if game.nearestStore() != game.store {
		t.Fatal("the store in reach was not found")
	}

	game.player.X += storeInteractRadius * 3
	game.player.UpdateHitbox()
	if game.nearestStore() != nil {
		t.Fatal("a store out of reach was found")
	}
}

func TestBountySpawnsOnEveryLevel(t *testing.T) {
	game := NewGame()

	if game.bounty == nil {
		t.Fatal("the level did not spawn a bounty target")
	}
	if !game.bounty.IsBounty {
		t.Fatal("the bounty target is not marked as a bounty")
	}
	if game.bounty.Dead {
		t.Fatal("the bounty spawned dead")
	}
	if game.bountyReward < bountyBaseReward(game.bounty.Tier) {
		t.Fatalf("the bounty reward %d is below its base price", game.bountyReward)
	}
	if game.aliveEnemies() != game.levelEnemies {
		t.Fatalf("the bounty was counted as a level enemy: %d foes for %d enemies", game.aliveEnemies(), game.levelEnemies)
	}
}

func TestBountyRewardScalesWithTierAndLevel(t *testing.T) {
	common := bountyRewardFor(actors.Common, 1, 0)
	rare := bountyRewardFor(actors.Rare, 1, 0)
	legendary := bountyRewardFor(actors.Legendary, 1, 0)
	if common >= rare || rare >= legendary {
		t.Fatal("higher rarity bounties should pay more")
	}
	if bountyRewardFor(actors.Common, 5, 0) <= common {
		t.Fatal("the bounty reward did not grow with the level")
	}
	if bountyRewardFor(actors.Common, 1, 0.99) <= common {
		t.Fatal("the random addition did not increase the reward")
	}
}

func TestFistsKnockOutBounty(t *testing.T) {
	game := NewGame()
	bounty := game.bounty
	player := game.player

	bounty.Health = 1
	bounty.MaxHealth = 1
	bounty.X = player.X + player.W/2 + 30 - bounty.W/2
	bounty.Y = player.Y + player.H/2 - bounty.H/2
	bounty.UpdateHitboxOffset(enemyHitboxOffset)
	player.MeleeAngle = 0

	game.performMeleeSweep(50)

	if !bounty.KnockedOut {
		t.Fatal("the bounty was not knocked out by the fist sweep")
	}
	if game.bountyClaimed {
		t.Fatal("the bounty was paid out before it reached the jail")
	}
}

func TestGunsKillBountyWithoutCapture(t *testing.T) {
	game := NewGame()
	corpseReward := game.bountyReward / bountyKillRewardDivisor

	game.killEnemy(game.bounty)

	goldBefore := game.player.Gold

	if !game.bounty.Dead {
		t.Fatal("the bounty did not die")
	}
	if game.bounty.KnockedOut {
		t.Fatal("a bounty killed by a gun was marked as captured")
	}

	game.claimBountyReward()
	if game.player.Gold != goldBefore+corpseReward {
		t.Fatal("a killed bounty could still be turned in for the full reward")
	}
}

func TestBountyFleesWhenLowHealth(t *testing.T) {
	game := NewGame()
	bounty := game.bounty
	bounty.Health = bounty.MaxHealth / bountyFleeHealthDivisor

	game.updateBountyFlee()

	if !bounty.Fleeing {
		t.Fatal("a badly wounded bounty did not flee")
	}
	if bounty.FleeX != game.boat.ShoreX || bounty.FleeY != game.boat.ShoreY {
		t.Fatalf("the bounty flees to %f,%f instead of the boat at %f,%f", bounty.FleeX, bounty.FleeY, game.boat.ShoreX, game.boat.ShoreY)
	}
}

func TestBountyFleesAfterSustainedAlert(t *testing.T) {
	game := NewGame()
	bounty := game.bounty
	bounty.Alerted = true
	bounty.Health = bounty.MaxHealth

	for range bountyFleeAlertFrames - 1 {
		game.updateBountyFlee()
	}
	if bounty.Fleeing {
		t.Fatal("the bounty fled before the alert timer ran out")
	}

	game.updateBountyFlee()
	if !bounty.Fleeing {
		t.Fatal("the bounty did not run for the boat after being hunted")
	}
}

func TestBountyFleeRalliesNearbyBandits(t *testing.T) {
	game := NewGame()
	bounty := game.bounty
	bounty.Health = bounty.MaxHealth / bountyFleeHealthDivisor

	game.updateBountyFlee()

	for _, enemy := range game.enemies {
		if enemy.IsBounty {
			continue
		}
		if !enemy.Aggressive {
			t.Fatal("a bandit stayed calm while the bounty was fleeing")
		}
	}
}

func TestBountyEscapesAtBoat(t *testing.T) {
	game := NewGame()
	bounty := game.bounty
	bounty.Health = bounty.MaxHealth / bountyFleeHealthDivisor
	game.updateBountyFlee()

	bounty.X = game.boat.ShoreX
	bounty.Y = game.boat.ShoreY
	game.updateBountyFlee()

	if !bounty.Escaped {
		t.Fatal("the bounty did not escape after reaching the boat")
	}
	for _, enemy := range game.enemies {
		if enemy == bounty {
			t.Fatal("the escaped bounty is still on the field")
		}
	}

	goldBefore := game.player.Gold
	game.claimBountyReward()
	if game.player.Gold != goldBefore {
		t.Fatal("an escaped bounty has no reward")
	}
}

func TestClaimBountyRewardAtJail(t *testing.T) {
	game := NewGame()
	game.bounty.Dead = true
	game.bounty.KnockedOut = true

	game.player.X = game.jail.Hitbox.CenterX() - game.player.W/2
	game.player.Y = game.jail.Hitbox.CenterY() - game.player.H/2
	game.player.UpdateHitbox()

	if game.nearestJail() == nil {
		t.Fatal("the jail in reach was not found")
	}

	goldBefore := game.player.Gold
	game.claimBountyReward()

	if !game.bountyClaimed {
		t.Fatal("the bounty reward was not marked as claimed")
	}
	if game.player.Gold != goldBefore+game.bountyReward {
		t.Fatalf("expected %d gold, got %d", game.bountyReward, game.player.Gold-goldBefore)
	}

	game.claimBountyReward()
	if game.player.Gold != goldBefore+game.bountyReward {
		t.Fatal("the bounty reward was paid out twice")
	}
}
