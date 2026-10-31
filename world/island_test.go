package world

import "testing"

func TestNewIslandIsPlayable(t *testing.T) {
	for range 20 {
		island := NewIsland(90, 60, 64)

		landTiles := 0
		for x := range island.Width {
			for y := range island.Height {
				tile := island.tiles[x][y]
				if tile == Land || tile == Beach {
					landTiles++
				}
				if isBorder(x, y, island.Width, island.Height) && tile != DeepWater && tile != ShallowWater {
					t.Fatalf("tile %d,%d on the map border is not water", x, y)
				}
			}
		}

		if landTiles < island.Width*island.Height/10 {
			t.Fatalf("island is too small: %d land tiles", landTiles)
		}

		centerX, centerY := island.Center()
		if !island.IsWalkable(centerX, centerY) {
			t.Fatalf("the player spawn %f,%f is not walkable", centerX, centerY)
		}

		for range 50 {
			x, y := island.RandomLandPoint()
			if !island.IsWalkable(x, y) {
				t.Fatalf("random spawn point %f,%f is not walkable", x, y)
			}
		}
	}
}

func TestIsRectWalkableStopsAtTheOcean(t *testing.T) {
	island := NewIsland(90, 60, 64)

	if island.IsWalkable(-1, -1) {
		t.Fatal("a position outside of the world should not be walkable")
	}

	if island.IsRectWalkable(0, 0, island.TileSize, island.TileSize) {
		t.Fatal("the corner of the world is ocean and should not be walkable")
	}
}
