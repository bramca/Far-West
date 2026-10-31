package world

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Boat is the player's moored ship and the dock that connects it to the island.
type Boat struct {
	X, Y        float64
	W, H        float64
	Sprite      *ebiten.Image
	DrawOptions *ebiten.DrawImageOptions
	Scale       float64
	Side        DockSide

	DockX, DockY float64
	DockW, DockH float64
}

// CenterX returns the horizontal center of the boat.
func (b *Boat) CenterX() float64 {
	return b.X + b.W*b.Scale/2
}

// CenterY returns the vertical center of the boat.
func (b *Boat) CenterY() float64 {
	return b.Y + b.H*b.Scale/2
}

// Draw renders the dock and then the boat that is moored at its end.
func (b *Boat) Draw(screen *ebiten.Image, camX, camY float64) {
	b.drawDock(screen, camX, camY)

	b.DrawOptions.GeoM.Reset()
	b.DrawOptions.GeoM.Scale(b.Scale, b.Scale)
	b.DrawOptions.GeoM.Translate(b.X-camX, b.Y-camY)
	screen.DrawImage(b.Sprite, b.DrawOptions)
}

func (b *Boat) drawDock(screen *ebiten.Image, camX, camY float64) {
	if b.DockW <= 0 || b.DockH <= 0 {
		return
	}

	x := float32(b.DockX - camX)
	y := float32(b.DockY - camY)
	w := float32(b.DockW)
	h := float32(b.DockH)
	plankColor := color.RGBA{155, 104, 54, 255}
	borderColor := color.RGBA{62, 39, 20, 255}

	vector.FillRect(screen, x, y, w, h, color.RGBA{105, 66, 32, 255}, false)
	vector.StrokeRect(screen, x, y, w, h, 2, borderColor, false)

	// plank lines run across the walkway and make the dock read as timber
	if b.Side == DockNorth || b.Side == DockSouth {
		for plankY := y + 12; plankY < y+h; plankY += 12 {
			vector.StrokeLine(screen, x+2, plankY, x+w-2, plankY, 2, plankColor, false)
		}
		return
	}

	for plankX := x + 12; plankX < x+w; plankX += 12 {
		vector.StrokeLine(screen, plankX, y+2, plankX, y+h-2, 2, plankColor, false)
	}
}
