package actors

import (
	"image/color"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type textImageKey struct {
	msg   string
	face  *text.GoXFace
	color color.RGBA
}

var (
	textImageCacheMu sync.Mutex
	textImageCache   = map[textImageKey]*ebiten.Image{}
)

const textImageCacheLimit = 512

func cachedTextImage(msg string, face *text.GoXFace, clr color.RGBA) *ebiten.Image {
	if msg == "" || face == nil {
		return nil
	}

	key := textImageKey{msg: msg, face: face, color: clr}
	textImageCacheMu.Lock()
	defer textImageCacheMu.Unlock()

	if img, ok := textImageCache[key]; ok {
		return img
	}

	width, height := text.Measure(msg, face, 0)
	img := ebiten.NewImage(int(math.Ceil(width))+1, int(math.Ceil(height))+1)
	op := &text.DrawOptions{}
	op.ColorScale.SetWithColor(clr)
	text.Draw(img, msg, face, op)

	if len(textImageCache) >= textImageCacheLimit {
		clear(textImageCache)
	}
	textImageCache[key] = img

	return img
}
