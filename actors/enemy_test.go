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
	Action{Type: MoveAndShoot, Actor: advancing, PreferredRange: 200}.PerformAction(target, 1)
	advancingAfter := math.Hypot(advancing.X-target.X, advancing.Y-target.Y)
	if advancingAfter >= advancingBefore {
		t.Fatalf("the enemy moved from %f to %f away from the player, it should close in", advancingBefore, advancingAfter)
	}

}

func TestMoveAndShootDoesNotRetreat(t *testing.T) {
	target := newCombatTestPlayer(500, 500)
	enemy := newCombatTestPlayer(450, 500)

	Action{Type: MoveAndShoot, Actor: enemy, PreferredRange: 200}.PerformAction(target, 1)

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

	Action{
		Type:    Reposition,
		Actor:   enemy,
		MoveDir: Up,
		LookDir: Right,
	}.PerformAction(target, 0)

	if len(enemy.Bullets) != 0 {
		t.Fatal("a repositioning enemy fired at the player")
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
