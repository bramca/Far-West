package world

import (
	"testing"

	"github.com/bramca/Far-West/actors"
)

func TestBanditCampContains(t *testing.T) {
	camp := NewBanditCamp(100, 100, 50)

	if !camp.Contains(100, 100) || !camp.Contains(140, 100) {
		t.Fatal("a position inside the camp was not reported as inside")
	}
	if camp.Contains(160, 100) {
		t.Fatal("a position outside the camp was reported as inside")
	}
}

func TestBanditCampAssignsFightRoles(t *testing.T) {
	camp := NewBanditCamp(100, 100, 50)
	enemies := make([]*actors.Enemy, 6)
	for i := range enemies {
		enemies[i] = &actors.Enemy{Player: &actors.Player{}}
		camp.AddEnemy(enemies[i])
	}

	roles := map[actors.EnemyRole]int{}
	for _, enemy := range enemies {
		roles[enemy.Role]++
	}
	if roles[actors.RolePin] != 2 || roles[actors.RoleFlank] != 2 || roles[actors.RoleRusher] != 2 {
		t.Fatalf("the camp handed out roles %v, want two of each", roles)
	}
	if enemies[1].FlankAngle == enemies[4].FlankAngle {
		t.Fatal("the two flankers target the same side of the player")
	}
}

func TestBanditCampEnemies(t *testing.T) {
	camp := NewBanditCamp(100, 100, 50)
	first := &actors.Enemy{Player: &actors.Player{}}
	second := &actors.Enemy{Player: &actors.Player{}}

	camp.AddEnemy(first)
	camp.AddEnemy(second)
	if camp.AliveEnemies() != 2 {
		t.Fatalf("the camp counted %d living enemies, expected 2", camp.AliveEnemies())
	}

	first.Dead = true
	if camp.AliveEnemies() != 1 {
		t.Fatalf("the camp counted %d living enemies, expected 1", camp.AliveEnemies())
	}

	camp.RemoveEnemy(first)
	if len(camp.Enemies) != 1 || camp.Enemies[0] != second {
		t.Fatal("the dead enemy was not removed from the camp")
	}
}
