package helpers

import (
	"embed"
	"fmt"
	"image"
	"math"
	"math/rand"

	"github.com/bramca/Far-West/actors"
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

func InitializeCactusHitboxes() []*actors.HitBox {
	result := []*actors.HitBox{
		{
			X: 11,
			Y: 12,
			W: 7,
			H: 13,
		},
		{
			X: 10,
			Y: 12,
			W: 9,
			H: 13,
		},
		{
			X: 11,
			Y: 9,
			W: 8,
			H: 16,
		},
		{
			X: 11,
			Y: 6,
			W: 8,
			H: 19,
		},
		{
			X: 11,
			Y: 16,
			W: 7,
			H: 9,
		},
		{
			X: 11,
			Y: 12,
			W: 8,
			H: 13,
		},
	}

	return result
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
				subImg := img.SubImage(image.Rect(j*xFrameSize, i*yFrameSize, (j+1)*xFrameSize, (i+1)*yFrameSize)).(*ebiten.Image)
				result = append(result, subImg)
			}
		}
	}

	return result
}

func SpawnCacti(xBound, yBound int, amount int, spriteScale float64, cactusSprites []*ebiten.Image, hitboxes []*actors.HitBox) []*world.Cactus {
	cacti := []*world.Cactus{}
	for range amount {
		x := float64(rand.Intn(xBound))
		y := float64(rand.Intn(yBound))
		i := rand.Intn(len(cactusSprites))
		sprite := cactusSprites[i]
		hitbox := hitboxes[i]
		cactus := &world.Cactus{
			X:           x,
			Y:           y,
			W:           float64(sprite.Bounds().Dx()),
			H:           float64(sprite.Bounds().Dy()),
			Sprite:      sprite,
			DrawOptions: &ebiten.DrawImageOptions{},
			Scale:       spriteScale,
			Hitbox: &actors.HitBox{
				X: float32(x + float64(hitbox.X*float32(spriteScale))),
				Y: float32(y + float64(hitbox.Y*float32(spriteScale))),
				W: hitbox.W * float32(spriteScale),
				H: hitbox.H * float32(spriteScale),
			},
		}
		cacti = append(cacti, cactus)
	}

	return cacti
}
