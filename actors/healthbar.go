package actors

import (
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
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
	TextFont        *text.GoXFace
	FixedSize       bool
	FixedPos        bool

	textValid  bool
	textPoints int
	textMax    int
	textFont   *text.GoXFace
	textColor  color.RGBA
	textMsg    string
	textImage  *ebiten.Image
}

func (h *HealthBar) Update(x, y float64, points, maxPoints int) {
	h.X, h.Y = x, y
	h.Points, h.MaxPoints = points, maxPoints
}

func (h *HealthBar) Draw(screen *ebiten.Image, camX, camY float64) {
	msgPadding := 4
	if !h.textValid || h.textPoints != h.Points || h.textMax != h.MaxPoints ||
		h.textFont != h.TextFont || h.textColor != h.FontColor {
		h.textMsg = strconv.Itoa(h.Points) + "/" + strconv.Itoa(h.MaxPoints)
		h.textPoints = h.Points
		h.textMax = h.MaxPoints
		h.textFont = h.TextFont
		h.textColor = h.FontColor
		if h.FontColor.A == 0 {
			h.textImage = nil
		} else {
			h.textImage = cachedTextImage(h.textMsg, h.TextFont, h.FontColor)
		}
		h.textValid = true
	}

	width := h.W
	if !h.FixedSize {
		width = float64(len(h.textMsg))*float64(h.FontSize) + float64(msgPadding)
	}
	w1 := float32(width * float64(h.Points) / float64(h.MaxPoints))
	w2 := float32(width * float64(h.MaxPoints-h.Points) / float64(h.MaxPoints))
	h1 := float32(h.H)
	h2 := float32(h.H)
	x1, y1 := float32(h.X-width/float64(msgPadding))-float32(camX), float32(h.Y)-float32(camY)
	x2, y2 := x1+w1, y1
	if h.FixedPos {
		x1, y1 = float32(h.X), float32(h.Y)
		x2, y2 = x1+w1, y1
	}
	fillRect(screen, x1, y1, w1, h1, h.HealthBarColor)
	fillRect(screen, x2, y2, w2, h2, h.HealthLostColor)

	if h.textImage == nil {
		return
	}
	textX, textY := h.X, h.Y
	if !h.FixedPos {
		textX = float64(h.X-width/float64(msgPadding)+float64(msgPadding)/2) - camX
		textY = float64(h.Y) - camY
	}
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(textX, textY)
	screen.DrawImage(h.textImage, &op)
}
