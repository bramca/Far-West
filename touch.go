package farwest

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Touch controls use the same state as the keyboard and gamepad handlers, so
// the gameplay code does not need to know where the input came from.

const (
	touchStickRadius   = 60.0
	touchStickDeadzone = 8.0
	touchHintRadius    = 46.0

	touchContextW = 200.0
	touchContextH = 36.0

	touchDeadEyeRadius = 40.0
	touchDodgeRadius   = 40.0
	touchReloadRadius  = 40.0
	touchPauseSize     = 36.0

	storePanelW = 560.0
	storePanelH = 440.0
	storePanelY = 80.0
	storeRowH   = 40.0

	wantedPanelW = 520.0
	wantedPanelH = 420.0
)

type touchRole int

const (
	touchRoleNone touchRole = iota
	touchRoleConsumed
	touchRoleMove
	touchRoleAim
)

type touchAction int

const (
	touchActionNone touchAction = iota
	touchActionInteract
	touchActionSail
)

type touchPoint struct {
	x, y float64
}

// refreshTouches samples the touch state once per tick and clears the
// touch-derived flags that the gameplay code reads later in Update.
func (g *Game) refreshTouches() {
	g.touchIDs = ebiten.AppendTouchIDs(g.touchIDs[:0])
	g.justTouchIDs = inpututil.AppendJustPressedTouchIDs(g.justTouchIDs[:0])

	if len(g.touchIDs) > 0 {
		g.usingTouch = true
	}

	g.touchFireHeld = false
	g.touchFireJustPressed = false
	g.touchAimActive = false
	g.touchMoveActive = false
	g.touchInteract = false
	g.touchSail = false
	g.touchDeadEye = false
	g.touchDodge = false
	g.touchReload = false

	if g.touchRoles == nil {
		g.touchRoles = map[ebiten.TouchID]touchRole{}
	}
	if g.touchAnchors == nil {
		g.touchAnchors = map[ebiten.TouchID]touchPoint{}
	}

	active := make(map[ebiten.TouchID]struct{}, len(g.touchIDs))
	for _, id := range g.touchIDs {
		active[id] = struct{}{}
	}
	for id := range g.touchRoles {
		if _, ok := active[id]; !ok {
			delete(g.touchRoles, id)
			delete(g.touchAnchors, id)
		}
	}
}

func (g *Game) isTouchJustPressed(id ebiten.TouchID) bool {
	for _, justID := range g.justTouchIDs {
		if justID == id {
			return true
		}
	}

	return false
}

// updateTouchMenus handles the parts of the game that are not the gameplay
// world: title screens, the store and the wanted poster.
func (g *Game) updateTouchMenus() {
	if len(g.justTouchIDs) == 0 {
		return
	}

	switch g.mode {
	case ModeTitle, ModeGameOver, ModePause:
		buttonsJustPressed["FBR"] = true
		for _, id := range g.justTouchIDs {
			g.touchRoles[id] = touchRoleConsumed
		}

		return
	}

	if g.mode != ModeGame {
		return
	}

	if g.storeOpen {
		for _, id := range g.justTouchIDs {
			g.touchRoles[id] = touchRoleConsumed
			if g.touchStoreTap(id) {
				return
			}
		}

		return
	}

	if g.wantedPosterOpen {
		for _, id := range g.justTouchIDs {
			g.touchRoles[id] = touchRoleConsumed
			g.touchWantedPosterTap(id)

			return
		}
	}
}

// updateTouchGameplay converts touches into movement, aim, fire and the
// on-screen buttons. It must run after the gamepad pass so touch axes win
// while a finger is down.
func (g *Game) updateTouchGameplay() {
	if g.mode != ModeGame || g.storeOpen || g.wantedPosterOpen || g.player == nil {
		return
	}

	contextRect, contextAction, _ := g.touchContextInfo()
	nearLoot := g.nearestLoot() != nil
	if g.touchCompareActive && !nearLoot {
		g.touchCompareActive = false
	}

	for _, id := range g.touchIDs {
		x, y := ebiten.TouchPositionF(id)
		pt := image.Pt(int(x), int(y))
		justPressed := g.isTouchJustPressed(id)
		role := g.touchRoles[id]

		if role == touchRoleNone {
			switch {
			case pt.In(touchPauseRect()):
				role = touchRoleConsumed
				if justPressed {
					g.mode = ModePause
				}
			case pt.In(touchDeadEyeRect()):
				role = touchRoleConsumed
				if justPressed {
					g.touchDeadEye = true
				}
			case pt.In(touchDodgeRect()):
				role = touchRoleConsumed
				if justPressed {
					g.touchDodge = true
				}
			case pt.In(touchReloadRect()):
				role = touchRoleConsumed
				if justPressed {
					g.touchReload = true
				}
			case nearLoot && pt.In(touchCompareRect()):
				role = touchRoleConsumed
				if justPressed {
					g.touchCompareActive = !g.touchCompareActive
				}
			case contextAction != touchActionNone && pt.In(contextRect):
				role = touchRoleConsumed
				if justPressed {
					if contextAction == touchActionSail {
						g.touchSail = true
					} else {
						g.touchInteract = true
					}
				}
			default:
				if index, ok := g.gunBeltIndexAt(x, y); ok {
					role = touchRoleConsumed
					if justPressed {
						g.player.DrawWeapon(index)
					}
				} else if x < ScreenWidth/2 {
					role = touchRoleMove
					g.touchAnchors[id] = touchPoint{x: x, y: y}
				} else {
					role = touchRoleAim
					g.touchAnchors[id] = touchPoint{x: x, y: y}
				}
			}
			g.touchRoles[id] = role
		}

		switch g.touchRoles[id] {
		case touchRoleMove:
			anchor := g.touchAnchors[id]
			dx, dy := x-anchor.x, y-anchor.y
			mag := math.Hypot(dx, dy)
			g.touchMoveActive = true
			g.touchMoveAnchor = anchor
			g.touchMoveKnob = anchor
			if mag > touchStickDeadzone {
				scale := math.Min(mag, touchStickRadius) / mag
				g.xLeftAxis = max(-1, min(1, dx*scale/touchStickRadius))
				g.yLeftAxis = max(-1, min(1, dy*scale/touchStickRadius))
				g.touchMoveKnob = touchPoint{x: anchor.x + dx*scale, y: anchor.y + dy*scale}
			}
		case touchRoleAim:
			anchor := g.touchAnchors[id]
			dx, dy := x-anchor.x, y-anchor.y
			g.touchAimActive = true
			g.touchFireHeld = true
			g.touchAimAnchor = anchor
			g.touchAimKnob = anchor
			if justPressed {
				g.touchFireJustPressed = true
			}
			if mag := math.Hypot(dx, dy); mag > touchStickDeadzone {
				g.player.AimAngle = math.Atan2(dy, dx)
				scale := math.Min(mag, touchStickRadius) / mag
				g.touchAimKnob = touchPoint{x: anchor.x + dx*scale, y: anchor.y + dy*scale}
			}
		}
	}
}

// touchContextInfo returns the context-sensitive action button and its label.
func (g *Game) touchContextInfo() (image.Rectangle, touchAction, string) {
	rect := touchContextRect()
	if g.player == nil || g.player.Dead {
		return rect, touchActionNone, ""
	}

	if g.nearestClosedChest() != nil {
		return rect, touchActionInteract, "OPEN CHEST"
	}
	if g.nearestLoot() != nil {
		return rect, touchActionInteract, "TAKE GUN"
	}
	if g.nearestJail() != nil {
		return rect, touchActionInteract, "WANTED POSTER"
	}
	if g.nearestStore() != nil {
		return rect, touchActionInteract, "SHOP"
	}
	if g.nearestBoat() != nil {
		if g.aliveEnemies() != 0 {
			return rect, touchActionSail, "CLEAR BANDITS"
		}

		return rect, touchActionSail, "NEXT ISLAND"
	}

	return rect, touchActionNone, ""
}

func (g *Game) gunBeltIndexAt(x, y float64) (int, bool) {
	best := -1
	bestDistance := 42.0
	for i, gun := range g.player.GunBelt {
		if gun == nil || gun.Sprite == nil {
			continue
		}
		width := float64(gun.Sprite.Bounds().Dx()) * gun.Scale
		height := float64(gun.Sprite.Bounds().Dy()) * gun.Scale
		distance := math.Hypot(x-(gun.X+width/2), y-(gun.Y+height/2))
		if distance < bestDistance {
			best = i
			bestDistance = distance
		}
	}

	return best, best >= 0
}

func (g *Game) touchStoreTap(id ebiten.TouchID) bool {
	x, y := ebiten.TouchPosition(id)
	pt := image.Pt(x, y)

	if pt.In(storeCloseRect()) {
		g.closeStore()

		return true
	}

	for i := range g.store.Stock {
		if !pt.In(storeRowRect(i)) {
			continue
		}

		if g.storeSelection == i {
			g.buyStoreItem(&g.store.Stock[i])
		} else {
			g.storeSelection = i
		}

		return true
	}

	if !pt.In(storePanelRect()) {
		g.closeStore()

		return true
	}

	// A tap on the panel but outside every row is ignored on purpose so the
	// player does not close the shop by accident.
	return false
}

func (g *Game) touchWantedPosterTap(id ebiten.TouchID) {
	x, y := ebiten.TouchPosition(id)
	pt := image.Pt(x, y)

	if pt.In(wantedCloseRect()) {
		g.closeWantedPoster()

		return
	}

	claimable := g.bounty != nil &&
		(g.bounty.KnockedOut || g.bounty.Dead) &&
		!g.bounty.Escaped && !g.bountyClaimed

	if claimable && pt.In(wantedPanelRect()) {
		g.claimBountyReward()
	}

	g.closeWantedPoster()
}

func touchContextRect() image.Rectangle {
	x := (ScreenWidth - touchContextW) / 2
	y := float64(ScreenHeight - 180)

	return image.Rect(int(x), int(y), int(x+touchContextW), int(y+touchContextH))
}

func touchCompareRect() image.Rectangle {
	x := (ScreenWidth - touchContextW) / 2
	y := float64(ScreenHeight - 224)

	return image.Rect(int(x), int(y), int(x+touchContextW), int(y+touchContextH))
}

func touchDeadEyeCenter() (float64, float64) {
	return ScreenWidth - 80, ScreenHeight - 96
}

func touchDeadEyeRect() image.Rectangle {
	cx, cy := touchDeadEyeCenter()
	r := float64(touchDeadEyeRadius)

	return image.Rect(int(cx-r), int(cy-r), int(cx+r), int(cy+r))
}

func touchDodgeCenter() (float64, float64) {
	return ScreenWidth - 272, ScreenHeight - 96
}

func touchDodgeRect() image.Rectangle {
	cx, cy := touchDodgeCenter()
	r := float64(touchDodgeRadius)

	return image.Rect(int(cx-r), int(cy-r), int(cx+r), int(cy+r))
}

func touchReloadCenter() (float64, float64) {
	return ScreenWidth - 176, ScreenHeight - 96
}

func touchReloadRect() image.Rectangle {
	cx, cy := touchReloadCenter()
	r := float64(touchReloadRadius)

	return image.Rect(int(cx-r), int(cy-r), int(cx+r), int(cy+r))
}

func touchPauseRect() image.Rectangle {
	cx := ScreenWidth / 2
	y := 14.0

	return image.Rect(int(float64(cx)-touchPauseSize/2), int(y), int(float64(cx)+touchPauseSize/2), int(y+touchPauseSize))
}

func storePanelRect() image.Rectangle {
	x := (float64(ScreenWidth) - storePanelW) / 2

	return image.Rect(int(x), int(storePanelY), int(x+storePanelW), int(storePanelY+storePanelH))
}

func storeCloseRect() image.Rectangle {
	x := (float64(ScreenWidth) - storePanelW) / 2

	return image.Rect(int(x+storePanelW-40), int(storePanelY+10), int(x+storePanelW-12), int(storePanelY+40))
}

func storeRowRect(index int) image.Rectangle {
	x := (float64(ScreenWidth) - storePanelW) / 2
	y := storePanelY + 60 + float64(index)*storeRowH

	return image.Rect(int(x+10), int(y-6), int(x+storePanelW-10), int(y+storeRowH-10))
}

func wantedPanelRect() image.Rectangle {
	x := (float64(ScreenWidth) - wantedPanelW) / 2
	y := (float64(ScreenHeight) - wantedPanelH) / 2

	return image.Rect(int(x), int(y), int(x+wantedPanelW), int(y+wantedPanelH))
}

func wantedCloseRect() image.Rectangle {
	x := (float64(ScreenWidth) - wantedPanelW) / 2
	y := (float64(ScreenHeight) - wantedPanelH) / 2

	return image.Rect(int(x+wantedPanelW-40), int(y+10), int(x+wantedPanelW-12), int(y+40))
}

// drawTouchControls renders the on-screen sticks and buttons once the player
// has used a touch, so desktop builds stay uncluttered.
func (g *Game) drawTouchControls(screen *ebiten.Image) {
	if !g.usingTouch || g.mode != ModeGame || g.storeOpen || g.wantedPosterOpen {
		return
	}

	g.drawTouchButton(screen, touchPauseRect(), "II", false)

	g.drawTouchStick(screen, "MOVE", touchRoleMove)
	g.drawTouchStick(screen, "AIM / FIRE", touchRoleAim)

	if g.nearestLoot() != nil {
		g.drawTouchButton(screen, touchCompareRect(), "COMPARE GUN", g.touchCompareActive)
	}

	if _, action, label := g.touchContextInfo(); action != touchActionNone {
		g.drawTouchButton(screen, touchContextRect(), label, false)
	}

	cx, cy := touchDodgeCenter()
	ready := g.player != nil && g.player.CanDodge()
	fill := color.RGBA{80, 65, 25, 105}
	if ready {
		fill = color.RGBA{150, 110, 25, 150}
	}
	g.drawTouchCircleButton(screen, cx, cy, touchDodgeRadius, "DODGE", fill, color.RGBA{245, 205, 55, 150})

	if g.player != nil && g.player.DodgeCooldown > 0 {
		progress := 1 - float64(g.player.DodgeCooldownTimer)/float64(g.player.DodgeCooldown)
		x := float32(cx - 20)
		y := float32(cy + 12)
		vector.FillRect(screen, x, y, 40, 3, color.RGBA{40, 40, 40, 120}, false)
		vector.FillRect(screen, x, y, float32(40*progress), 3, color.RGBA{110, 200, 220, 180}, false)
	}

	cx, cy = touchReloadCenter()
	ready = g.player != nil && g.player.CanReload()
	fill = color.RGBA{80, 65, 25, 105}
	if ready {
		fill = color.RGBA{150, 110, 25, 150}
	}
	g.drawTouchCircleButton(screen, cx, cy, touchReloadRadius, "RELOAD", fill, color.RGBA{245, 205, 55, 150})

	if g.player != nil && g.player.Reloading {
		progress := 1 - float64(g.player.ReloadTimer)/float64(g.player.ReloadDuration)
		x := float32(cx - 20)
		y := float32(cy + 12)
		vector.FillRect(screen, x, y, 40, 3, color.RGBA{40, 40, 40, 120}, false)
		vector.FillRect(screen, x, y, float32(40*progress), 3, color.RGBA{230, 190, 60, 180}, false)
	}

	cx, cy = touchDeadEyeCenter()
	fill = color.RGBA{90, 25, 25, 105}
	if g.deadEyeActive {
		fill = color.RGBA{215, 55, 40, 150}
	}
	g.drawTouchCircleButton(screen, cx, cy, touchDeadEyeRadius, "D.EYE", fill, color.RGBA{245, 205, 55, 150})
}

func (g *Game) drawTouchCircleButton(screen *ebiten.Image, cx, cy, radius float64, label string, fill, stroke color.RGBA) {
	vector.DrawFilledCircle(screen, float32(cx), float32(cy), float32(radius), fill, true)
	vector.StrokeCircle(screen, float32(cx), float32(cy), float32(radius), 2, stroke, true)

	x := cx - float64(len(label)*g.hudFontSize)/2
	y := cy - float64(g.hudFontSize)/2
	g.drawSmallText(screen, label, x, y, color.RGBA{240, 230, 190, 220})
}

func (g *Game) drawTouchButton(screen *ebiten.Image, rect image.Rectangle, label string, active bool) {
	fill := color.RGBA{18, 16, 12, 105}
	if active {
		fill = color.RGBA{70, 58, 30, 145}
	}
	vector.FillRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), fill, false)
	vector.StrokeRect(screen, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), 2, color.RGBA{245, 205, 55, 145}, false)

	x := float64(rect.Min.X) + (float64(rect.Dx())-float64(len(label)*g.hudFontSize))/2
	y := float64(rect.Min.Y) + (float64(rect.Dy())-float64(g.hudFontSize))/2
	g.drawSmallText(screen, label, x, y, color.RGBA{240, 230, 190, 215})
}

func (g *Game) drawTouchStick(screen *ebiten.Image, label string, role touchRole) {
	hintX := 150.0
	if role == touchRoleAim {
		hintX = ScreenWidth - 150
	}
	center := touchPoint{x: hintX, y: ScreenHeight - 220}
	knob := center
	active := false

	if role == touchRoleMove && g.touchMoveActive {
		center = g.touchMoveAnchor
		knob = g.touchMoveKnob
		active = true
	}
	if role == touchRoleAim && g.touchAimActive {
		center = g.touchAimAnchor
		knob = g.touchAimKnob
		active = true
	}

	baseAlpha := uint8(12)
	knobAlpha := uint8(25)
	if active {
		baseAlpha = 30
		knobAlpha = 85
	}
	radius := float32(touchHintRadius)
	if active {
		radius = touchStickRadius
	}
	vector.DrawFilledCircle(screen, float32(center.x), float32(center.y), radius, color.RGBA{220, 220, 210, baseAlpha}, true)
	vector.StrokeCircle(screen, float32(center.x), float32(center.y), radius, 2, color.RGBA{245, 205, 55, 40}, true)
	if active {
		vector.DrawFilledCircle(screen, float32(knob.x), float32(knob.y), 22, color.RGBA{245, 205, 55, knobAlpha}, true)
	}

	labelY := center.y + float64(touchHintRadius) + float64(g.hudFontSize) + 2
	g.drawSmallText(screen, label, center.x-float64(len(label)*g.hudFontSize)/2, labelY, color.RGBA{220, 210, 180, 100})
}
