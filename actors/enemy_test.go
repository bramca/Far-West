package actors

import (
	"math"
	"testing"
)

func TestEnemyAlertExpires(t *testing.T) {
	enemy := &Enemy{Player: &Player{}}
	enemy.Alert(10, 20)

	if !enemy.Alerted {
		t.Fatal("the enemy did not get alerted")
	}
	if enemy.LastSeenX != 10 || enemy.LastSeenY != 20 {
		t.Fatalf("the enemy remembered the player at %f,%f", enemy.LastSeenX, enemy.LastSeenY)
	}

	for range enemyAlertDuration - 1 {
		enemy.TickAlert()
	}
	if !enemy.Alerted {
		t.Fatal("the enemy calmed down before the alert timer ran out")
	}

	enemy.TickAlert()
	if enemy.Alerted {
		t.Fatal("the enemy stayed alerted after the alert timer ran out")
	}
}

func TestEnemyEngageRangeFollowsWeapon(t *testing.T) {
	tests := []struct {
		weapon Weapon
		want   float64
	}{
		{Fists, fistsEngageRange},
		{Revolver, revolverEngageRange},
		{Shotgun, shotgunEngageRange},
		{MachineGun, machineGunEngageRange},
	}

	for _, test := range tests {
		enemy := &Enemy{Player: &Player{
			CurrentWeapon: 0,
			GunBelt:       []*Gun{{GunType: test.weapon}},
		}}
		if got := enemy.engageRange(); got != test.want {
			t.Fatalf("an enemy with a %s keeps %f distance, want %f", test.weapon, got, test.want)
		}
	}
}

func TestEnemyEngageRangeFallsBackToPrimaryWeapon(t *testing.T) {
	enemy := &Enemy{Player: &Player{
		CurrentWeapon: 0,
		GunBelt: []*Gun{
			{GunType: Whisky},
			{GunType: MachineGun},
		},
	}}

	if got := enemy.engageRange(); got != machineGunEngageRange {
		t.Fatalf("an enemy holding a whisky keeps %f distance, want its machine gun range %f", got, machineGunEngageRange)
	}
}

func TestMoveAndShootKeepsPreferredRange(t *testing.T) {
	target := newCombatTestPlayer(500, 500)

	advancing := newCombatTestPlayer(100, 500)
	advancingBefore := math.Hypot(advancing.X-target.X, advancing.Y-target.Y)
	action := Action{Type: MoveAndShoot, Actor: advancing, PreferredRange: 200}
	action.PerformAction(target, 1)
	advancingAfter := math.Hypot(advancing.X-target.X, advancing.Y-target.Y)
	if advancingAfter >= advancingBefore {
		t.Fatalf("the enemy moved from %f to %f away from the player, it should close in", advancingBefore, advancingAfter)
	}

}

func TestMoveAndShootDoesNotRetreat(t *testing.T) {
	target := newCombatTestPlayer(500, 500)
	enemy := newCombatTestPlayer(450, 500)

	action := Action{Type: MoveAndShoot, Actor: enemy, PreferredRange: 200}
	action.PerformAction(target, 1)

	if enemy.X < 450 {
		t.Fatalf("the enemy backed off to %f, it should hold its ground", enemy.X)
	}
}

func TestEngageRepositionsAfterAttacking(t *testing.T) {
	target := newCombatTestPlayer(500, 500)
	enemy := &Enemy{Player: newCombatTestPlayer(300, 500)}
	enemy.CurrentAction = Action{Type: MoveAndShoot, Actor: enemy.Player}

	enemy.engage(target)

	if enemy.CurrentAction.Type != Reposition {
		t.Fatalf("the enemy did not reposition after attacking, action type %d", enemy.CurrentAction.Type)
	}
	if enemy.CurrentAction.Duration <= 0 {
		t.Fatal("the reposition had no duration")
	}
}

func TestEngageKeepsActiveReposition(t *testing.T) {
	target := newCombatTestPlayer(500, 500)
	enemy := &Enemy{Player: newCombatTestPlayer(300, 500)}
	enemy.CurrentAction = Action{
		Duration: 10,
		Type:     Reposition,
		MoveDir:  Up,
		LookDir:  Right,
		Actor:    enemy.Player,
	}

	enemy.engage(target)

	if enemy.CurrentAction.Type != Reposition || enemy.CurrentAction.MoveDir != Up {
		t.Fatal("the active reposition was replaced by the attack logic")
	}
}

func TestRepositionDoesNotShoot(t *testing.T) {
	target := newCombatTestPlayer(500, 500)
	enemy := newCombatTestPlayer(300, 500)
	enemy.FireRate = 1

	action := Action{
		Type:    Reposition,
		Actor:   enemy,
		MoveDir: Up,
		LookDir: Right,
	}
	action.PerformAction(target, 0)

	if len(enemy.Bullets) != 0 {
		t.Fatal("a repositioning enemy fired at the player")
	}
}

func TestEnemyRolesTargetDifferentPoints(t *testing.T) {
	player := newCombatTestPlayer(500, 500)
	flanker := &Enemy{Player: newCombatTestPlayer(500, 300), Role: RoleFlank, FlankAngle: math.Pi / 2}
	pin := &Enemy{Player: newCombatTestPlayer(500, 300), Role: RolePin}

	offsetX, offsetY := flanker.engageOffset()
	if math.Abs(offsetX) > 0.001 || math.Abs(offsetY-flankOffsetDistance) > 0.001 {
		t.Fatalf("the flanker targets offset %f,%f, want 0,%f", offsetX, offsetY, flankOffsetDistance)
	}

	focusX, focusY := flanker.focusPoint(player)
	if math.Hypot(focusX-player.X, focusY-player.Y) < flankOffsetDistance-0.001 {
		t.Fatal("the flanker focused on the player instead of a point beside him")
	}

	offsetX, offsetY = pin.engageOffset()
	if offsetX != 0 || offsetY != 0 {
		t.Fatalf("the pin targets offset %f,%f, want the player himself", offsetX, offsetY)
	}
}

func TestRusherClosesInMoreThanPin(t *testing.T) {
	gunBelt := []*Gun{{GunType: Revolver}}
	pin := &Enemy{Player: &Player{CurrentWeapon: 0, GunBelt: gunBelt}, Role: RolePin}
	rusher := &Enemy{Player: &Player{CurrentWeapon: 0, GunBelt: gunBelt}, Role: RoleRusher}

	if rusher.engageRange() >= pin.engageRange() {
		t.Fatalf("a rusher keeps %f distance while a pin keeps %f, the rusher should close in", rusher.engageRange(), pin.engageRange())
	}
}

func TestFlankerMovesTowardsOffsetPoint(t *testing.T) {
	target := newCombatTestPlayer(500, 500)
	enemy := newCombatTestPlayer(500, 300)

	action := Action{
		Type:           MoveAndShoot,
		Actor:          enemy,
		PreferredRange: 100,
		TargetOffsetX:  300,
	}
	action.PerformAction(target, 1)

	if enemy.X <= 500 {
		t.Fatalf("the flanker moved to %f, it should head for the offset point to the right", enemy.X)
	}
}

func TestFleeingBountyKeepsFighting(t *testing.T) {
	player := newCombatTestPlayer(310, 500)
	enemy := &Enemy{
		Player:     newCombatTestPlayer(300, 500),
		Fleeing:    true,
		FleeX:      0,
		FleeY:      500,
		VisualDist: 500,
	}

	enemy.ThinkAndAct(player, nil, 1)

	if enemy.CurrentAction.Type != Flee {
		t.Fatalf("the fleeing bounty picked action %d instead of fleeing", enemy.CurrentAction.Type)
	}
	if !enemy.CurrentAction.Fire {
		t.Fatal("the fleeing bounty stopped fighting the player")
	}
	if enemy.X >= 300 {
		t.Fatalf("the fleeing bounty did not run toward the boat, x is %f", enemy.X)
	}
}

func TestEnemyAttackProfileBurstsFollowWeapon(t *testing.T) {
	tests := []struct {
		weapon    Weapon
		wantShots int
		wantCone  bool
	}{
		{Shotgun, shotgunBurstShots, true},
		{Revolver, revolverBurstShots, false},
		{MachineGun, machineGunBurstShots, false},
	}

	for _, test := range tests {
		enemy := &Enemy{Player: &Player{
			CurrentWeapon: 0,
			GunBelt:       []*Gun{{GunType: test.weapon}},
		}}

		profile := enemy.attackProfile()
		if profile.windup <= 0 {
			t.Fatalf("the %s attack has no windup", test.weapon)
		}
		if profile.shots != test.wantShots {
			t.Fatalf("the %s bursts %d shots, want %d", test.weapon, profile.shots, test.wantShots)
		}
		if profile.interval <= 0 {
			t.Fatalf("the %s burst has no interval between shots", test.weapon)
		}

		shape, halfAngle := telegraphForGun(enemy.GunBelt[0])
		if test.wantCone && (shape != TelegraphCone || halfAngle <= 0) {
			t.Fatalf("the %s should telegraph a cone", test.weapon)
		}
		if !test.wantCone && shape != TelegraphLine {
			t.Fatalf("the %s should telegraph a straight line", test.weapon)
		}
	}
}

func TestEngageStartsWithATelegraphedAttack(t *testing.T) {
	target := newCombatTestPlayer(500, 500)
	enemy := &Enemy{Player: newCombatTestPlayer(300, 500)}
	enemy.FireRate = 10

	// the engage logic sometimes picks a dodge instead of an attack, try a
	// few times until it picks the attack so the test stays deterministic
	for range 100 {
		enemy.CurrentAction = Action{}
		enemy.engage(target)
		if enemy.CurrentAction.Type == MoveAndShoot {
			break
		}
	}

	if enemy.CurrentAction.Type != MoveAndShoot {
		t.Fatal("the enemy never picked an attack")
	}
	if enemy.CurrentAction.Windup <= 0 || enemy.CurrentAction.WindupFrames != enemy.CurrentAction.Windup {
		t.Fatal("the attack started without a windup telegraph")
	}
	if enemy.CurrentAction.BurstShots <= 0 {
		t.Fatal("the attack started without a burst")
	}
}

func TestEnemyWaitsForWindupThenFiresBurst(t *testing.T) {
	target := newCombatTestPlayer(500, 500)
	enemy := newCombatTestPlayer(300, 500)
	enemy.GunBelt[0] = &Gun{GunType: Fists, Damage: 1, ShootCooldown: 1}
	shots := 0
	enemy.OnMelee = func(int) { shots++ }

	const windup = 20
	action := Action{
		Type:          MoveAndShoot,
		Actor:         enemy,
		Duration:      100,
		Windup:        windup,
		WindupFrames:  windup,
		BurstShots:    3,
		BurstInterval: 8,
	}

	for frame := range windup - 1 {
		action.PerformAction(target, frame)
		enemy.UpdateWeapon()
	}
	if shots != 0 {
		t.Fatalf("the enemy fired %d shots during the windup", shots)
	}

	for frame := range 100 {
		action.PerformAction(target, frame)
		enemy.UpdateWeapon()
	}
	if shots != 3 {
		t.Fatalf("the enemy fired %d shots, want a burst of 3", shots)
	}
}

func newCombatTestPlayer(x, y float64) *Player {
	return &Player{
		X:              x,
		Y:              y,
		Speed:          2,
		AnimationSpeed: 11,
		FireRate:       1000,
		IsNpc:          true,
		GunBelt:        []*Gun{{GunType: Revolver, BulletSpeed: 6}},
		MoveDirs:       map[Direction]bool{},
		Hitbox:         &HitBox{X: float32(x), Y: float32(y), W: 10, H: 10},
		Healthbar:      &HealthBar{},
	}
}
