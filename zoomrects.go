package platinum

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// ZoomDuration is how long zoom rectangles take to go from where a window
// comes from to where it opens.
const ZoomDuration = 180 * time.Millisecond

// zoomSteps is how many outlines show at once, trailing each other.
const zoomSteps = 4

// ZoomRects shows the outlines that grow from the box from, as the icon a
// window opens from, to the box to, where it opens, over ZoomDuration from
// start — the classic desktop's sign of where a window came from. It draws
// over the window, in its coordinates, and asks for frames until it is
// done; with the desktop asking for less motion, or the time frozen by
// SetClock, it shows nothing. It reports whether it is still going.
func ZoomRects(c *ui.Context, from, to ui.Rect, start time.Time) bool {
	// Frozen time, for screenshots, shows the moment after: nothing.
	if c.Preferences().ReduceMotion || frozen() {
		return false
	}
	now := frameNow(c)
	t := float32(now.Sub(start)) / float32(ZoomDuration)
	if t >= 1 {
		return false
	}
	// A start taken from the wall clock while the frame was built is a
	// little after the frame's time: the rectangles start from there.
	t = max(t, 0)
	c.AnimationFrame()
	ui.Overlay(c, func() {
		ui.Box(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).PassThrough().Role(ui.RoleNone).
			Draw(func(p *ui.Painter, r ui.Rect) {
				DrawZoomRects(Pen{P: p}, from, to, t)
			})
	})
	return true
}

// DrawZoomRects draws the outlines of zoom rectangles at t (0 to 1) of the
// way from from to to: zoomSteps outlines, the newest t of the way, the
// others trailing behind it, in a gray that shows on light and dark.
func DrawZoomRects(pn Pen, from, to ui.Rect, t float32) {
	lerp := func(a, b, f float32) float32 { return a + (b-a)*f }
	for i := 0; i < zoomSteps; i++ {
		f := t - float32(i)*0.12
		if f < 0 {
			continue
		}
		x, y := round(lerp(from.X, to.X, f)), round(lerp(from.Y, to.Y, f))
		w, h := round(lerp(from.W, to.W, f)), round(lerp(from.H, to.H, f))
		// Dotted, as an outline being dragged.
		for k := 0; k < w; k += 2 {
			pn.Dot(x+k, y, Gray10)
			pn.Dot(x+k, y+h-1, Gray10)
		}
		for k := 0; k < h; k += 2 {
			pn.Dot(x, y+k, Gray10)
			pn.Dot(x+w-1, y+k, Gray10)
		}
	}
}
