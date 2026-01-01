package actors

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type Hit struct {
	X, Y        float64
	Color       color.RGBA
	Msg         string
	TextFont    *text.GoXFace
	DrawOptions *text.DrawOptions
	Duration    int
}

func (h *Hit) SetDrawOptions() {
	drawOptions := text.DrawOptions{
		DrawImageOptions: ebiten.DrawImageOptions{},
	}
	drawOptions.ColorScale.SetR(float32(h.Color.R) / 256.0)
	drawOptions.ColorScale.SetG(float32(h.Color.G) / 256.0)
	drawOptions.ColorScale.SetB(float32(h.Color.B) / 256.0)
	drawOptions.ColorScale.SetA(float32(h.Color.G) / 256.0)
	h.DrawOptions = &drawOptions
}

func (h *Hit) Update() {
	h.Y -= 1
	h.Duration -= 1
}

func (h *Hit) Draw(screen *ebiten.Image, camX float64, camY float64) {
	h.DrawOptions.GeoM.Translate(h.X-camX, h.Y-camY)
	text.Draw(screen, h.Msg, h.TextFont, h.DrawOptions)
	h.DrawOptions.GeoM.Reset()
}
