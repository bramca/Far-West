package helpers

import (
	"embed"
	"fmt"
	"image"
	"math/rand/v2"

	"github.com/bramca/Far-West/actors"
	"github.com/bramca/Far-West/world"
	"github.com/hajimehoshi/ebiten/v2"
)

func LoadImage(assets embed.FS, imagePath string) (*ebiten.Image, error) {
	file, err := assets.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			panic(fmt.Sprintf("Failed to close image file: %e", err))
		}
	}()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image file: %w", err)
	}

	return ebiten.NewImageFromImage(img), nil
}

func InitializeTentHitbox() *actors.HitBox {
	return &actors.HitBox{
		X: 9,
		Y: 11,
		W: 13,
		H: 12,
	}
}

func InitializeChestHitbox() *actors.HitBox {
	return &actors.HitBox{
		X: 8,
		Y: 10,
		W: 16,
		H: 13,
	}
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

// SpawnCacti spreads cacti over the land tiles of the given island.
func SpawnCacti(island *world.Island, amount int, spriteScale float64, cactusSprites []*ebiten.Image, hitboxes []*actors.HitBox) []*world.Cactus {
	cacti := []*world.Cactus{}
	for range amount {
		x, y := island.RandomLandPoint()
		i := rand.IntN(len(cactusSprites))
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

// SpawnTents spreads tents over the land tiles of the given island.
func SpawnTents(island *world.Island, amount int, spriteScale float64, sprite *ebiten.Image, hitbox *actors.HitBox) []*world.Tent {
	tents := []*world.Tent{}
	for range amount {
		x, y := island.RandomLandPoint()
		tent := &world.Tent{
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
		tents = append(tents, tent)
	}

	return tents
}

// SpawnChests spreads chests over the land tiles of the given island.
func SpawnChests(island *world.Island, amount int, spriteScale float64, chestSprites []*ebiten.Image, hitbox *actors.HitBox) []*world.Chest {
	chests := []*world.Chest{}
	for range amount {
		x, y := island.RandomLandPoint()
		i := rand.IntN(len(chestSprites)/2) * 2
		sprites := chestSprites[i : i+1]
		chest := &world.Chest{
			X:           x,
			Y:           y,
			Type:        world.ChestType(i / 2),
			W:           float64(sprites[0].Bounds().Dx()),
			H:           float64(sprites[0].Bounds().Dy()),
			Sprites:     sprites,
			DrawOptions: &ebiten.DrawImageOptions{},
			Scale:       spriteScale,
			Hitbox: &actors.HitBox{
				X: float32(x + float64(hitbox.X*float32(spriteScale))),
				Y: float32(y + float64(hitbox.Y*float32(spriteScale))),
				W: hitbox.W * float32(spriteScale),
				H: hitbox.H * float32(spriteScale),
			},
		}
		chests = append(chests, chest)
	}

	return chests
}
