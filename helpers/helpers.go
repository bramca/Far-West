package helpers

import (
	"embed"
	"fmt"
	"image"
	"math"
	"math/rand"

	"github.com/bramca/Far-West/world"
	"github.com/hajimehoshi/ebiten/v2"
)

func LoadImage(assets embed.FS, imagePath string) (*ebiten.Image, error) {
	file, err := assets.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image file: %w", err)
	}

	return ebiten.NewImageFromImage(img), nil
}

func AngleBetweenPoints(x1, y1, x2, y2 float64) float64 {
	return math.Atan2(y2-y1, x2-x1)
}

func DistanceBetweenPoints(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1))
}

func LoadSprites(assets embed.FS, fileNames []string, xFrameSize int, yFrameSize int) []*ebiten.Image {
	result := []*ebiten.Image{}
	for _, fileName := range fileNames {
		img, err := LoadImage(assets, fileName)
		if err != nil {
			panic(fmt.Sprintf("could not load player sprite: %e", err))
		}

		verticalFrames := int(img.Bounds().Dy() / yFrameSize)
		horizontalFrames := int(img.Bounds().Dx() / xFrameSize)

		for i := range verticalFrames {
			for j := range horizontalFrames {
				result = append(result, img.SubImage(image.Rect(j*xFrameSize, i*yFrameSize, (j+1)*xFrameSize, (i+1)*yFrameSize)).(*ebiten.Image))
			}
		}
	}

	return result
}

func SpawnCacti(xBound, yBound int, amount int, spriteScale float64, cactusSprites []*ebiten.Image) []*world.Cactus {
	cacti := []*world.Cactus{}
	for range amount {
		x := float64(rand.Intn(xBound))
		y := float64(rand.Intn(yBound))
		sprite := cactusSprites[rand.Intn(len(cactusSprites))]
		cactus := &world.Cactus{
			X:           x,
			Y:           y,
			W:           float64(sprite.Bounds().Dx()),
			H:           float64(sprite.Bounds().Dy()),
			Sprite:      sprite,
			DrawOptions: &ebiten.DrawImageOptions{},
			Scale:       spriteScale,
		}
		cacti = append(cacti, cactus)
	}

	return cacti
}
