package farwest

import (
	"testing"

	"github.com/bramca/Far-West/actors"
)

// TestGameLoop runs the game loop without any player input, the player is
// expected to stay on the island, to be hunted down by the enemies of the
// first level and to end up in a game over.
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
// enemies, that killing them all shows the level complete overlay and that
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

	if err := game.Update(); err != nil {
		t.Fatal(err)
	}

	if game.mode != ModeLevelComplete {
		t.Fatalf("expected the level complete overlay, got mode %d", game.mode)
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
	if game.player.Health != game.player.MaxHealth {
		t.Fatalf("the player did not regain full health on level 2, got %d", game.player.Health)
	}
	if game.aliveEnemies() != enemyCountForLevel(2) {
		t.Fatalf("expected %d enemies on level 2, got %d", enemyCountForLevel(2), game.aliveEnemies())
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