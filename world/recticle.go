package world

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Recticle struct {
	Size int
}

func (r *Recticle) Draw(screen *ebiten.Image) {
	mx, my := ebiten.CursorPosition()
	rSize := float32(r.Size)
	x1, y1 := float32(mx), float32(my)-rSize-1
	x2, y2 := float32(mx), float32(my-1)
	vector.StrokeLine(screen, x1, y1, x2, y2, 1, color.White, false)
	x3, y3 := float32(mx)-rSize-1, float32(my)
	x4, y4 := float32(mx-1), float32(my)
	vector.StrokeLine(screen, x3, y3, x4, y4, 1, color.White, false)
	x5, y5 := float32(mx), float32(my)+rSize+1
	x6, y6 := float32(mx), float32(my+1)
	vector.StrokeLine(screen, x5, y5, x6, y6, 1, color.White, false)
	x7, y7 := float32(mx)+rSize+1, float32(my)
	x8, y8 := float32(mx+1), float32(my)
	vector.StrokeLine(screen, x7, y7, x8, y8, 1, color.White, false)
}
