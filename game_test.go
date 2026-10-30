package farwest

import (
	"math"
	"testing"

	"github.com/bramca/Far-West/actors"
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
	game.cacti = nil
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

// TestBlockedEnemyPicksFreeDirection makes sure an enemy that walked into an
// obstacle is steered into a direction it can actually move in.
func TestBlockedEnemyPicksFreeDirection(t *testing.T) {
	game := NewGame()
	enemy := game.enemies[0]

	// isolate the enemy in an open part of the island so the test cannot be
	// foiled by a neighbouring bandit standing in every free direction
	game.enemies = []*actors.Enemy{enemy}
	game.cacti = nil
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
	game.player.DrawWeapon(2)

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

// TestBulletsAlwaysDealDamage makes sure the damage roll never produces a
// zero damage bullet, a shot that does nothing is not fun.
func TestBulletsAlwaysDealDamage(t *testing.T) {
	game := NewGame()
	game.player.DrawWeapon(2)

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

	if common.health >= rare.health || rare.health >= legendary.health {
		t.Fatal("higher tier enemies should have more health")
	}
	if common.damage >= rare.damage || rare.damage >= legendary.damage {
		t.Fatal("higher tier enemies should deal more damage")
	}
	if common.speed >= rare.speed || rare.speed >= legendary.speed {
		t.Fatal("higher tier enemies should move faster")
	}
	if common.fireRate >= rare.fireRate || rare.fireRate >= legendary.fireRate {
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
	game.player.DrawWeapon(1)
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
	if game.player.CurrentWeapon != 1 {
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
	game.player.DrawWeapon(0)
	revolverIndex, ok := game.gunBeltIndexFor(actors.Revolver)
	if !ok {
		t.Fatal("the player has no revolver slot")
	}
	oldRevolver := game.player.GunBelt[revolverIndex]

	lootGun := game.newEnemyGun(actors.Revolver, actors.Common, 2)
	loot := actors.NewLoot(&lootGun, game.player.X, game.player.Y)
	game.loot = append(game.loot, loot)

	game.pickUpLoot(loot)

	if game.player.CurrentWeapon != 0 {
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

// TestSpritesCoverEveryState guards against drawing a state that has no sprite.
func TestSpritesCoverEveryState(t *testing.T) {
	game := NewGame()

	if len(game.playerSprites) <= int(actors.PlayerDead) {
		t.Fatalf("the player sprites (%d) do not cover the dead state", len(game.playerSprites))
	}

	if len(game.enemyCommonSprites) <= int(actors.PlayerDead) {
		t.Fatalf("the enemy sprites (%d) do not cover the dead state", len(game.enemyCommonSprites))
	}
}
