package helpers

import (
	"embed"
	"fmt"
	"image"
	"math"
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

const (
	// tents are placed on an outer ring and chests on an inner ring so a camp
	// looks like a small organised settlement instead of a pile of sprites.
	tentRingFraction  = 0.68
	chestRingFraction = 0.22
	ringRadiusJitter  = 0.04
	tentAngleJitter   = 0.12
	chestAngleJitter  = 0.08
)

// campObject is the visual rectangle of a tent or chest, it is used to keep
// the objects of a camp from overlapping.
type campObject struct {
	x, y, w, h float64
}

func (o campObject) overlaps(other campObject) bool {
	return o.x < other.x+other.w &&
		o.x+o.w > other.x &&
		o.y < other.y+other.h &&
		o.y+o.h > other.y
}

func (o campObject) overlapsAny(objects []campObject) bool {
	for _, other := range objects {
		if o.overlaps(other) {
			return true
		}
	}

	return false
}

// placeCampObject looks for a free spot on a ring around the camp center. The
// first attempts stay close to the ring so the layout stays organised, later
// attempts may pick any angle to find room when the ring is crowded.
func placeCampObject(island *world.Island, camp *world.BanditCamp, width, height, baseAngle, ringFraction, angleJitter float64, occupied []campObject) (campObject, bool) {
	for attempt := range 100 {
		angle := baseAngle + (rand.Float64()*2-1)*angleJitter
		if attempt >= 50 {
			angle = rand.Float64() * 2 * math.Pi
		}
		radius := camp.Radius*ringFraction + (rand.Float64()*2-1)*camp.Radius*ringRadiusJitter
		x := camp.X + math.Cos(angle)*radius - width/2
		y := camp.Y + math.Sin(angle)*radius - height/2
		object := campObject{x: x, y: y, w: width, h: height}
		if !island.IsRectWalkable(object.x, object.y, object.w, object.h) {
			continue
		}
		if object.overlapsAny(occupied) {
			continue
		}

		return object, true
	}

	return campObject{}, false
}

// fallbackCampObject returns a random free spot in the camp, it is only used
// when the ring layout could not fit an object.
func fallbackCampObject(island *world.Island, camp *world.BanditCamp, width, height float64, occupied []campObject) campObject {
	for range 500 {
		x, y, ok := island.RandomLandPointWithin(camp.X, camp.Y, camp.Radius*0.8)
		if !ok {
			break
		}
		object := campObject{x: x - width/2, y: y - height/2, w: width, h: height}
		if !island.IsRectWalkable(object.x, object.y, object.w, object.h) || object.overlapsAny(occupied) {
			continue
		}

		return object
	}

	return campObject{x: camp.X - width/2, y: camp.Y - height/2, w: width, h: height}
}

// SpawnTents lays the tents of a bandit camp out on a ring around its center.
func SpawnTents(island *world.Island, camp *world.BanditCamp, amount int, spriteScale float64, sprite *ebiten.Image, hitbox *actors.HitBox) []*world.Tent {
	tents := []*world.Tent{}
	width := float64(sprite.Bounds().Dx()) * spriteScale
	height := float64(sprite.Bounds().Dy()) * spriteScale
	occupied := make([]campObject, 0, amount)
	angleOffset := rand.Float64() * 2 * math.Pi
	for i := range amount {
		baseAngle := angleOffset + float64(i)/float64(amount)*2*math.Pi
		object, ok := placeCampObject(island, camp, width, height, baseAngle, tentRingFraction, tentAngleJitter, occupied)
		if !ok {
			object = fallbackCampObject(island, camp, width, height, occupied)
		}
		occupied = append(occupied, object)

		x, y := object.x, object.y
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

// SpawnChests places the chests on a smaller inner ring, keeping them clear of
// the tents that are already on the camp.
func SpawnChests(island *world.Island, camp *world.BanditCamp, amount int, spriteScale float64, chestSprites []*ebiten.Image, hitbox *actors.HitBox) []*world.Chest {
	chests := []*world.Chest{}
	occupied := make([]campObject, 0, len(camp.Tents)+amount)
	for _, tent := range camp.Tents {
		occupied = append(occupied, campObject{
			x: tent.X,
			y: tent.Y,
			w: tent.W * tent.Scale,
			h: tent.H * tent.Scale,
		})
	}

	width := float64(chestSprites[0].Bounds().Dx()) * spriteScale
	height := float64(chestSprites[0].Bounds().Dy()) * spriteScale
	angleOffset := rand.Float64() * 2 * math.Pi
	for i := range amount {
		baseAngle := angleOffset + float64(i)/float64(amount)*2*math.Pi
		object, ok := placeCampObject(island, camp, width, height, baseAngle, chestRingFraction, chestAngleJitter, occupied)
		if !ok {
			object = fallbackCampObject(island, camp, width, height, occupied)
		}
		occupied = append(occupied, object)

		x, y := object.x, object.y
		spriteIndex := rand.IntN(len(chestSprites)/2) * 2
		sprites := chestSprites[spriteIndex : spriteIndex+1]
		chest := &world.Chest{
			X:           x,
			Y:           y,
			Type:        world.ChestType(spriteIndex / 2),
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
