package farwest

import (
	"testing"

	"github.com/bramca/Far-West/actors"
)

// TestGameLoop runs the game loop without any player input, the player is
// expected to stay on the island, to be hunted down by the enemies and to end
// up in a game over.
func TestGameLoop(t *testing.T) {
	game := NewGame()
	game.mode = ModeGame

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

		if game.aliveEnemies() > game.maxEnemies {
			t.Fatalf("%d enemies on the field while only %d are allowed", game.aliveEnemies(), game.maxEnemies)
		}

		if game.mode == ModeGameOver {
			break
		}
	}

	if game.mode != ModeGameOver {
		t.Fatal("an idle player is expected to die eventually")
	}

	if game.maxEnemies <= initialMaxEnemies {
		t.Fatalf("the maximum amount of enemies did not grow over time: %d", game.maxEnemies)
	}
}

// TestRestart makes sure a new run starts from a clean slate.
func TestRestart(t *testing.T) {
	game := NewGame()
	game.score = 500
	game.elapsedFrames = 6000
	game.maxEnemies = 20
	game.player.Dead = true
	game.player.Health = 0

	game.Initialize()

	if game.score != 0 || game.elapsedFrames != 0 || game.maxEnemies != initialMaxEnemies {
		t.Fatal("the run state was not reset")
	}

	if game.player.Dead || game.player.Health != game.player.MaxHealth {
		t.Fatal("the player was not revived")
	}

	if len(game.enemies) == 0 || len(game.enemies) > initialMaxEnemies {
		t.Fatalf("expected at most %d enemies after a restart, got %d", initialMaxEnemies, len(game.enemies))
	}
}

// TestReloadLimitsFireRate makes sure the magazine and the reload timer stop
// the player from spamming bullets.
func TestReloadLimitsFireRate(t *testing.T) {
	game := NewGame()
	game.player.DrawWeapon(actors.Revolver)

	shots := 0
	for range game.framesPerSecond {
		bulletsBefore := len(game.player.Bullets)
		game.player.Shoot()
		if len(game.player.Bullets) > bulletsBefore {
			shots++
		}
		game.player.UpdateWeapon()
	}

	if shots > playerMagazineSize+1 {
		t.Fatalf("fired %d bullets in a second, the reload is not slowing the player down", shots)
	}

	if shots == 0 {
		t.Fatal("the player could not fire a single bullet")
	}
}

// TestSpritesCoverEveryState guards against drawing a state that has no sprite.
func TestSpritesCoverEveryState(t *testing.T) {
	game := NewGame()

	if len(game.playerSprites) <= int(actors.PlayerDead) {
		t.Fatalf("the player sprites (%d) do not cover the dead state", len(game.playerSprites))
	}

	if len(game.enemySprites) <= int(actors.PlayerDead) {
		t.Fatalf("the enemy sprites (%d) do not cover the dead state", len(game.enemySprites))
	}
}
