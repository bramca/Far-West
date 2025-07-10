package helpers

import (
	"embed"
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

func loadImage(assets embed.FS, imagePath string) (*ebiten.Image, error) {
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

func angleBetweenPoints(x1, y1, x2, y2 float64) float64 {
	return math.Atan2(y2-y1, x2-x1)
}

func distanceBetweenPoints(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1))
}
