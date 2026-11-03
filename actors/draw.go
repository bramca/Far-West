package actors

import (
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	whitePixelOnce   sync.Once
	whitePixel       *ebiten.Image
	muzzleFlashOnce  sync.Once
	muzzleFlashImage *ebiten.Image
)

func whitePixelImage() *ebiten.Image {
	whitePixelOnce.Do(func() {
		whitePixel = ebiten.NewImage(1, 1)
		whitePixel.Fill(color.White)
	})
	return whitePixel
}

func muzzleFlashSprite() *ebiten.Image {
	muzzleFlashOnce.Do(func() {
		const size = 14
		img := ebiten.NewImage(size, size)
		vector.FillCircle(img, size/2, size/2, 6, color.RGBA{255, 220, 110, 200}, true)
		vector.FillCircle(img, size/2, size/2, 3, color.RGBA{255, 255, 230, 255}, true)
		muzzleFlashImage = img
	})
	return muzzleFlashImage
}

func fillRect(screen *ebiten.Image, x, y, w, h float32, clr color.RGBA) {
	if w <= 0 || h <= 0 || clr.A == 0 {
		return
	}

	alpha := float32(clr.A) / 255
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(float64(w), float64(h))
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.Scale(
		float32(clr.R)/255*alpha,
		float32(clr.G)/255*alpha,
		float32(clr.B)/255*alpha,
		alpha,
	)
	screen.DrawImage(whitePixelImage(), &op)
}
