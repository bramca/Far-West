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
