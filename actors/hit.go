package actors

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type Hit struct {
	X, Y      float64
	Color     color.RGBA
	Msg       string
	TextFont  *text.GoXFace
	TextImage *ebiten.Image
	Duration  int
}

func (h *Hit) PrepareText() {
	clr := h.Color
	clr.A = h.Color.G
	if clr.A == 0 {
		h.TextImage = nil
		return
	}
	h.TextImage = cachedTextImage(h.Msg, h.TextFont, clr)
}

func (h *Hit) Update() {
	h.Y -= 1.4
	h.Duration -= 1
}

func (h *Hit) Draw(screen *ebiten.Image, camX float64, camY float64) {
	if h.TextImage == nil {
		return
	}
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(h.X-camX, h.Y-camY)
	screen.DrawImage(h.TextImage, &op)
}
