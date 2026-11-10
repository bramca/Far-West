package actors

import (
	"image/color"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// TelegraphShape is the shape of the area an enemy is about to shoot into.
type TelegraphShape int

const (
	// TelegraphLine is a straight line, used by single projectile weapons.
	TelegraphLine TelegraphShape = iota
	// TelegraphCone is a wide triangle, used by the shotgun.
	TelegraphCone
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

// telegraphForGun returns the telegraph shape and the half angle of the area a
// gun covers. The shotgun fans out its pellets, the other guns fire straight.
func telegraphForGun(gun *Gun) (TelegraphShape, float64) {
	if gun == nil {
		return TelegraphLine, 0
	}

	if gun.GunType != Shotgun {
		return TelegraphLine, 0
	}

	if gun.BulletSpread <= 0 || gun.BulletCount <= 0 {
		return TelegraphCone, math.Pi / 12
	}

	// the pellets are spread by BulletSpread steps of half a pi, so the
	// outermost pellet sits at roughly this angle from the middle
	halfAngle := float64(gun.BulletCount) * math.Pi / float64(gun.BulletSpread) / 2

	return TelegraphCone, max(math.Pi/12, halfAngle*1.15)
}

// drawTelegraph renders the windup indicator of an enemy attack. The bright
// part grows toward the target as the shot gets closer, so the player can read
// both the direction and the timing at a glance.
func drawTelegraph(screen *ebiten.Image, cx, cy, angle, progress float64, shape TelegraphShape, halfAngle, length float64) {
	progress = min(1, max(0, progress))
	if length <= 0 {
		return
	}

	dim := color.RGBA{10, 10, 10, uint8(1 + 26*progress)}
	bright := color.RGBA{255, 65, 65, uint8(80 + 80*progress)}
	width := float32(1.1 + 1.4*progress)

	switch shape {
	case TelegraphCone:
		left := angle - halfAngle
		right := angle + halfAngle
		leftX := float32(cx + math.Cos(left)*length)
		leftY := float32(cy + math.Sin(left)*length)
		rightX := float32(cx + math.Cos(right)*length)
		rightY := float32(cy + math.Sin(right)*length)
		reach := length * progress
		reachLeftX := float32(cx + math.Cos(left)*reach)
		reachLeftY := float32(cy + math.Sin(left)*reach)
		reachRightX := float32(cx + math.Cos(right)*reach)
		reachRightY := float32(cy + math.Sin(right)*reach)

		var path vector.Path
		path.MoveTo(float32(cx), float32(cy))
		path.LineTo(leftX, leftY)
		path.LineTo(rightX, rightY)
		path.Close()

		fill := &vector.DrawPathOptions{AntiAlias: true}
		fill.ColorScale.ScaleWithColor(dim)
		vector.FillPath(screen, &path, nil, fill)

		var pathReach vector.Path
		pathReach.MoveTo(float32(cx), float32(cy))
		pathReach.LineTo(reachLeftX, reachLeftY)
		pathReach.LineTo(reachRightX, reachRightY)
		pathReach.Close()

		fill = &vector.DrawPathOptions{AntiAlias: true}
		fill.ColorScale.ScaleWithColor(bright)
		vector.FillPath(screen, &pathReach, nil, fill)

		vector.StrokeLine(screen, float32(cx), float32(cy), leftX, leftY, width, dim, true)
		vector.StrokeLine(screen, float32(cx), float32(cy), rightX, rightY, width, dim, true)
		vector.StrokeLine(screen, leftX, leftY, rightX, rightY, 1, dim, true)
	default:
		tipX := float32(cx + math.Cos(angle)*length)
		tipY := float32(cy + math.Sin(angle)*length)
		vector.StrokeLine(screen, float32(cx), float32(cy), tipX, tipY, 1, dim, true)

		reach := length * progress
		reachX := float32(cx + math.Cos(angle)*reach)
		reachY := float32(cy + math.Sin(angle)*reach)
		vector.StrokeLine(screen, float32(cx), float32(cy), reachX, reachY, width, bright, true)
		vector.FillCircle(screen, reachX, reachY, float32(1.5+2*progress), bright, true)
	}
}
