package actors

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type HealthBar struct {
	X, Y            float64
	W, H            float64
	Points          int
	MaxPoints       int
	HealthBarColor  color.RGBA
	HealthLostColor color.RGBA
	FontColor       color.RGBA
	FontSize        int
	DrawOptions     *text.DrawOptions
	TextFont        *text.GoXFace
	FixedSize       bool
	FixedPos        bool
}

func (h *HealthBar) SetDrawOptions() {
	h.DrawOptions = &text.DrawOptions{}
	h.DrawOptions.ColorScale.SetR(float32(h.FontColor.R) / 256.0)
	h.DrawOptions.ColorScale.SetG(float32(h.FontColor.G) / 256.0)
	h.DrawOptions.ColorScale.SetB(float32(h.FontColor.B) / 256.0)
	h.DrawOptions.ColorScale.SetA(float32(h.FontColor.A) / 256.0)
	if h.FixedPos {
		h.DrawOptions.GeoM.Reset()
		h.DrawOptions.GeoM.Translate(h.X, h.Y)
	}
}

func (h *HealthBar) Update(x, y float64, points, maxPoints int) {
	h.X, h.Y = x, y
	h.Points, h.MaxPoints = points, maxPoints
}

func (h *HealthBar) Draw(screen *ebiten.Image, camX, camY float64) {
	healthBarMsg := fmt.Sprintf("%d/%d", h.Points, h.MaxPoints)
	msgPadding := 4
	width := h.W
	if !h.FixedSize {
		width = float64(len(healthBarMsg))*float64(h.FontSize) + float64(msgPadding)
	}
	w1 := float32(width * float64(h.Points) / float64(h.MaxPoints))
	w2 := float32(width * float64(h.MaxPoints-h.Points) / float64(h.MaxPoints))
	h1 := float32(h.H)
	h2 := float32(h.H)
	x1, y1 := float32(h.X-width/float64(msgPadding))-float32(camX), float32(h.Y)-float32(camY)
	x2, y2 := x1+w1, y1
	if !h.FixedPos {
		h.DrawOptions.GeoM.Reset()
		h.DrawOptions.GeoM.Translate(float64(h.X-width/float64(msgPadding)+float64(msgPadding)/2)-camX, float64(h.Y)-camY)
	}
	if h.FixedPos {
		x1, y1 = float32(h.X), float32(h.Y)
		x2, y2 = x1+w1, y1
	}
	vector.FillRect(screen, x1, y1, w1, h1, h.HealthBarColor, false)
	vector.FillRect(screen, x2, y2, w2, h2, h.HealthLostColor, false)
	text.Draw(screen, healthBarMsg, h.TextFont, h.DrawOptions)
}
