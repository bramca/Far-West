package actors

import "testing"

func newHurtEnemyWithWhisky() (*Enemy, *Player) {
	player := &Player{
		X:              120,
		Y:              120,
		W:              16,
		H:              32,
		Speed:          1,
		AnimationSpeed: 1,
		Health:         2,
		MaxHealth:      10,
		IsNpc:          true,
		MoveDirs:       map[Direction]bool{},
		Hitbox:         &HitBox{},
		Healthbar:      &HealthBar{},
		GunBelt: []*Gun{
			{GunType: Revolver, Damage: 2, Ammo: 10, MagazineSize: 6, CurrentMagazineSize: 6, ShootCooldown: 1},
			{GunType: Whisky, Damage: 3, Ammo: 1, ShootCooldown: 3},
		},
	}

	return &Enemy{Player: player}, &Player{X: 0, Y: 0}
}

func TestEnemyDrinksWhiskyWhenHurt(t *testing.T) {
	enemy, target := newHurtEnemyWithWhisky()

	enemy.ThinkAndAct(target, nil, 0)

	if enemy.CurrentWeapon != 1 {
		t.Fatal("the hurt enemy did not draw his whisky")
	}
	if !enemy.Healing || enemy.CurrentAction.Type != Heal {
		t.Fatal("the hurt enemy did not start drinking")
	}
	if enemy.X <= 120 || enemy.Y <= 120 {
		t.Fatal("the drinking enemy did not flee from the player")
	}

	for frame := 1; frame <= 3; frame++ {
		enemy.UpdateWeapon()
		enemy.ThinkAndAct(target, nil, frame)
	}

	if enemy.Health != 5 {
		t.Fatalf("expected the whisky to heal three health, got %d", enemy.Health)
	}
	if enemy.Healing {
		t.Fatal("the enemy is still healing after the drink finished")
	}
	if enemy.CurrentWeapon != 0 {
		t.Fatal("the enemy did not switch back to his fighting gun")
	}
	if enemy.GunBelt[1].Ammo != 0 {
		t.Fatal("the whisky did not consume its sip")
	}
}
