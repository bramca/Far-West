package world

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type TileType int

const (
	DeepWater TileType = iota
	ShallowWater
	Beach
	Land
)

// oceanMargin is the amount of tiles around the island that is always water,
// it guarantees the island never touches the edge of the world.
const oceanMargin = 3

var tileColors = map[TileType]color.RGBA{
	DeepWater:    {R: 22, G: 66, B: 110, A: 255},
	ShallowWater: {R: 44, G: 118, B: 168, A: 255},
	Beach:        {R: 200, G: 180, B: 120, A: 255},
	Land:         {R: 122, G: 112, B: 78, A: 255},
}

// Island is a randomly shaped chunk of land surrounded by water.
// The water acts as the barrier of the world.
type Island struct {
	TileSize float64
	// Width and Height are expressed in tiles.
	Width  int
	Height int

	tiles [][]TileType
	// shade holds a small per tile brightness offset so the terrain does not look flat.
	shade [][]int

	miniMap *ebiten.Image
}

// NewIsland generates a random island of width x height tiles.
func NewIsland(width, height int, tileSize float64) *Island {
	island := &Island{
		TileSize: tileSize,
		Width:    width,
		Height:   height,
	}

	land := island.generateLandMask()
	land = smoothLandMask(land, width, height, 4)
	land = keepLargestLandMass(land, width, height)

	island.tiles = classifyTiles(land, width, height)
	island.shade = make([][]int, width)
	for x := range island.shade {
		island.shade[x] = make([]int, height)
		for y := range island.shade[x] {
			island.shade[x][y] = rand.Intn(15) - 7
		}
	}

	island.miniMap = island.renderMiniMap()

	return island
}

// generateLandMask builds the raw shape of the island using overlapping
// metaballs that get distorted by a coarse noise field. Every blob pulls the
// coastline outwards while the noise carves out bays and peninsulas, which
// results in an organic, non circular shape.
func (i *Island) generateLandMask() [][]bool {
	type blob struct {
		x, y, r float64
	}

	centerX := float64(i.Width) / 2
	centerY := float64(i.Height) / 2
	maxRadius := math.Min(centerX, centerY) - oceanMargin

	blobs := []blob{{x: centerX, y: centerY, r: maxRadius * 0.45}}
	nBlobs := 5 + rand.Intn(4)
	for range nBlobs {
		angle := rand.Float64() * 2 * math.Pi
		dist := (0.3 + rand.Float64()*0.6) * maxRadius
		blobs = append(blobs, blob{
			x: centerX + math.Cos(angle)*dist,
			y: centerY + math.Sin(angle)*dist,
			r: maxRadius * (0.15 + rand.Float64()*0.2),
		})
	}

	noise := newValueNoise(i.Width, i.Height, 7)

	land := make([][]bool, i.Width)
	for x := range land {
		land[x] = make([]bool, i.Height)
		for y := range land[x] {
			value := 0.0
			for _, b := range blobs {
				dx := float64(x) - b.x
				dy := float64(y) - b.y
				value += (b.r * b.r) / (dx*dx + dy*dy + 1)
			}
			// the noise pushes the coastline in and out
			value *= 0.6 + 1.1*noise(float64(x), float64(y))
			land[x][y] = value > 1
		}
	}

	return land
}

// newValueNoise returns a smooth noise function in the [0, 1] range, it
// interpolates between random values placed on a coarse grid.
func newValueNoise(width, height, cellSize int) func(x, y float64) float64 {
	cols := width/cellSize + 2
	rows := height/cellSize + 2
	values := make([][]float64, cols)
	for c := range values {
		values[c] = make([]float64, rows)
		for r := range values[c] {
			values[c][r] = rand.Float64()
		}
	}

	// smoothstep removes the linear, blocky look of the interpolation
	smoothstep := func(t float64) float64 {
		return t * t * (3 - 2*t)
	}

	return func(x, y float64) float64 {
		gx := x / float64(cellSize)
		gy := y / float64(cellSize)
		x0, y0 := int(gx), int(gy)
		tx := smoothstep(gx - float64(x0))
		ty := smoothstep(gy - float64(y0))

		top := values[x0][y0]*(1-tx) + values[x0+1][y0]*tx
		bottom := values[x0][y0+1]*(1-tx) + values[x0+1][y0+1]*tx

		return top*(1-ty) + bottom*ty
	}
}

// smoothLandMask runs a cellular automaton that turns the noisy mask into
// smooth, connected land.
func smoothLandMask(land [][]bool, width, height, iterations int) [][]bool {
	for range iterations {
		next := make([][]bool, width)
		for x := range next {
			next[x] = make([]bool, height)
			for y := range next[x] {
				if isBorder(x, y, width, height) {
					continue
				}
				neighbours := 0
				for dx := -1; dx <= 1; dx++ {
					for dy := -1; dy <= 1; dy++ {
						if dx == 0 && dy == 0 {
							continue
						}
						if land[x+dx][y+dy] {
							neighbours++
						}
					}
				}
				next[x][y] = neighbours > 4 || (land[x][y] && neighbours == 4)
			}
		}
		land = next
	}

	return land
}

// keepLargestLandMass removes every small island that got disconnected from
// the main land mass, so the player can always reach the whole map.
func keepLargestLandMass(land [][]bool, width, height int) [][]bool {
	visited := make([][]bool, width)
	for x := range visited {
		visited[x] = make([]bool, height)
	}

	best := []image.Point{}
	for x := range width {
		for y := range height {
			if !land[x][y] || visited[x][y] {
				continue
			}
			region := []image.Point{}
			queue := []image.Point{{X: x, Y: y}}
			visited[x][y] = true
			for len(queue) > 0 {
				p := queue[0]
				queue = queue[1:]
				region = append(region, p)
				for _, n := range []image.Point{{X: p.X - 1, Y: p.Y}, {X: p.X + 1, Y: p.Y}, {X: p.X, Y: p.Y - 1}, {X: p.X, Y: p.Y + 1}} {
					if n.X < 0 || n.Y < 0 || n.X >= width || n.Y >= height {
						continue
					}
					if !land[n.X][n.Y] || visited[n.X][n.Y] {
						continue
					}
					visited[n.X][n.Y] = true
					queue = append(queue, n)
				}
			}
			if len(region) > len(best) {
				best = region
			}
		}
	}

	result := make([][]bool, width)
	for x := range result {
		result[x] = make([]bool, height)
	}
	for _, p := range best {
		result[p.X][p.Y] = true
	}

	return result
}

// classifyTiles turns the boolean land mask into renderable tile types.
func classifyTiles(land [][]bool, width, height int) [][]TileType {
	tiles := make([][]TileType, width)
	for x := range tiles {
		tiles[x] = make([]TileType, height)
		for y := range tiles[x] {
			switch {
			case land[x][y] && hasNeighbour(land, width, height, x, y, false, 1):
				tiles[x][y] = Beach
			case land[x][y]:
				tiles[x][y] = Land
			case hasNeighbour(land, width, height, x, y, true, 2):
				tiles[x][y] = ShallowWater
			default:
				tiles[x][y] = DeepWater
			}
		}
	}

	return tiles
}

func hasNeighbour(land [][]bool, width, height, x, y int, wanted bool, radius int) bool {
	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			nx, ny := x+dx, y+dy
			if nx < 0 || ny < 0 || nx >= width || ny >= height {
				continue
			}
			if land[nx][ny] == wanted {
				return true
			}
		}
	}

	return false
}

func isBorder(x, y, width, height int) bool {
	return x < oceanMargin || y < oceanMargin || x >= width-oceanMargin || y >= height-oceanMargin
}

// PixelWidth returns the width of the world in pixels.
func (i *Island) PixelWidth() float64 {
	return float64(i.Width) * i.TileSize
}

// PixelHeight returns the height of the world in pixels.
func (i *Island) PixelHeight() float64 {
	return float64(i.Height) * i.TileSize
}

// TileAt returns the tile type at the given world coordinates.
func (i *Island) TileAt(x, y float64) TileType {
	tx := int(x / i.TileSize)
	ty := int(y / i.TileSize)
	if tx < 0 || ty < 0 || tx >= i.Width || ty >= i.Height {
		return DeepWater
	}

	return i.tiles[tx][ty]
}

// IsWalkable reports whether the given world coordinates are on solid ground.
func (i *Island) IsWalkable(x, y float64) bool {
	tile := i.TileAt(x, y)

	return tile == Land || tile == Beach
}

// IsRectWalkable reports whether every corner of the given rectangle is on solid ground.
func (i *Island) IsRectWalkable(x, y, w, h float64) bool {
	return i.IsWalkable(x, y) &&
		i.IsWalkable(x+w, y) &&
		i.IsWalkable(x, y+h) &&
		i.IsWalkable(x+w, y+h)
}

// RandomLandPoint returns the center of a random land tile.
func (i *Island) RandomLandPoint() (float64, float64) {
	for range 10000 {
		x := rand.Intn(i.Width)
		y := rand.Intn(i.Height)
		if i.tiles[x][y] == Land {
			return (float64(x) + 0.5) * i.TileSize, (float64(y) + 0.5) * i.TileSize
		}
	}

	return i.Center()
}

// Center returns the center of the land mass, it is used as the player spawn.
func (i *Island) Center() (float64, float64) {
	sumX, sumY, count := 0.0, 0.0, 0.0
	for x := range i.Width {
		for y := range i.Height {
			if i.tiles[x][y] == Land {
				sumX += float64(x)
				sumY += float64(y)
				count++
			}
		}
	}
	if count == 0 {
		return i.PixelWidth() / 2, i.PixelHeight() / 2
	}

	centerX, centerY := sumX/count, sumY/count
	// the average position can land on water for crescent shaped islands,
	// so snap to the closest land tile
	bestX, bestY, bestDist := centerX, centerY, math.MaxFloat64
	for x := range i.Width {
		for y := range i.Height {
			if i.tiles[x][y] != Land {
				continue
			}
			dx, dy := float64(x)-centerX, float64(y)-centerY
			if dist := dx*dx + dy*dy; dist < bestDist {
				bestX, bestY, bestDist = float64(x), float64(y), dist
			}
		}
	}

	return (bestX + 0.5) * i.TileSize, (bestY + 0.5) * i.TileSize
}

// Draw renders every tile that is visible on screen.
func (i *Island) Draw(screen *ebiten.Image, camX, camY float64) {
	screen.Fill(tileColors[DeepWater])

	startX := max(int(camX/i.TileSize), 0)
	startY := max(int(camY/i.TileSize), 0)
	endX := min(int((camX+float64(screen.Bounds().Dx()))/i.TileSize)+1, i.Width)
	endY := min(int((camY+float64(screen.Bounds().Dy()))/i.TileSize)+1, i.Height)

	for x := startX; x < endX; x++ {
		for y := startY; y < endY; y++ {
			tile := i.tiles[x][y]
			if tile == DeepWater {
				continue
			}
			vector.FillRect(
				screen,
				float32(float64(x)*i.TileSize-camX),
				float32(float64(y)*i.TileSize-camY),
				float32(i.TileSize)+1,
				float32(i.TileSize)+1,
				shadeColor(tileColors[tile], i.shade[x][y]),
				false,
			)
		}
	}
}

func shadeColor(c color.RGBA, offset int) color.RGBA {
	clamp := func(v int) uint8 {
		return uint8(min(max(v, 0), 255))
	}

	return color.RGBA{
		R: clamp(int(c.R) + offset),
		G: clamp(int(c.G) + offset),
		B: clamp(int(c.B) + offset),
		A: c.A,
	}
}

func (i *Island) renderMiniMap() *ebiten.Image {
	img := ebiten.NewImage(i.Width, i.Height)
	for x := range i.Width {
		for y := range i.Height {
			img.Set(x, y, tileColors[i.tiles[x][y]])
		}
	}

	return img
}

// DrawMiniMap draws a scaled down version of the world with the position of
// the player and of every enemy on it.
func (i *Island) DrawMiniMap(screen *ebiten.Image, x, y, scale float64, playerX, playerY float64, enemies [][2]float64) {
	drawOptions := &ebiten.DrawImageOptions{}
	drawOptions.GeoM.Scale(scale, scale)
	drawOptions.GeoM.Translate(x, y)
	drawOptions.ColorScale.ScaleAlpha(0.85)
	screen.DrawImage(i.miniMap, drawOptions)

	vector.StrokeRect(screen, float32(x), float32(y), float32(float64(i.Width)*scale), float32(float64(i.Height)*scale), 2, color.RGBA{0, 0, 0, 200}, false)

	toMiniMap := func(worldX, worldY float64) (float32, float32) {
		return float32(x + worldX/i.TileSize*scale), float32(y + worldY/i.TileSize*scale)
	}

	for _, enemy := range enemies {
		ex, ey := toMiniMap(enemy[0], enemy[1])
		vector.FillRect(screen, ex, ey, 2, 2, color.RGBA{200, 40, 40, 255}, false)
	}

	px, py := toMiniMap(playerX, playerY)
	vector.FillRect(screen, px-1, py-1, 4, 4, color.RGBA{255, 255, 255, 255}, false)
}
