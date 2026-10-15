package platinum

import (
	"math"
	"time"

	"github.com/egoist/mygo/ui"
)

// Status control metrics, from docs/spec.md.
const (
	ChasingArrowsSize = 16 // the box of the chasing arrows
	PlacardHeight     = 16 // the height of a placard
	BadgeHeight       = 14 // the height of a badge or tag
	chasingSteps      = 8  // frames of a turn of the chasing arrows
	chasingRate       = 10 // frames a second
)

// ChasingArrows builds the chasing arrows: two arrows running round a
// circle, which show that work of unknown length goes on, in less room than
// a progress bar.
//
// The counterpart among MyGo's own controls is ui.Spinner.
func ChasingArrows(c *ui.Context) ui.Element {
	el := ui.Box(c).Size(ChasingArrowsSize, ChasingArrowsSize).Role(ui.RoleProgress)
	el.Draw(func(p *ui.Painter, r ui.Rect) {
		t := paintNow(p)
		if !frozen() {
			p.After(time.Second / chasingRate)
		}
		step := int(t.UnixMilli()*chasingRate/1000) % chasingSteps
		DrawChasingArrows(NewPen(p, r), step)
	})
	return el
}

// DrawChasingArrows draws the chasing arrows in pn's 16×16 corner, turned
// step eighths of a turn.
func DrawChasingArrows(pn Pen, step int) {
	turn := float64(step) / chasingSteps * 2 * math.Pi
	for y := 0; y < ChasingArrowsSize; y++ {
		for x := 0; x < ChasingArrowsSize; x++ {
			dx, dy := float64(x)+0.5-8, float64(y)+0.5-8
			d := math.Hypot(dx, dy)
			a := math.Mod(math.Atan2(dy, dx)-turn+4*math.Pi, 2*math.Pi)
			// Each half of the circle: an arc, and the arrowhead at its end.
			h := math.Mod(a, math.Pi)
			switch {
			case h < 0.55:
				// The gap before the next arrow.
			case h < 1.15:
				// The arrowhead, wider than the arc.
				w := (h - 0.55) / 0.6 * 3.2
				if math.Abs(d-5.5) <= w {
					pn.Dot(x, y, Black)
				}
			default:
				if d >= 4.3 && d <= 6.7 {
					pn.Dot(x, y, Black)
				}
			}
		}
	}
}

// Placard builds a placard: a small raised panel of text in the regular
// face, as the information along the bottom of a window, beside its
// horizontal scroll bar.
func Placard(c *ui.Context, text string) ui.Element {
	w := AppFont.Measure(text) + 12
	el := ui.Box(c).Size(float32(w), PlacardHeight).Role(ui.RoleStatus).Label(text)
	el.Draw(func(p *ui.Painter, r ui.Rect) {
		pn := NewPen(p, r)
		DrawPlacard(pn)
		pn.TextMiddle(6, 0, pn.H, AppFont.Truncate(text, pn.W-8), AppFont, Black)
	})
	return el
}

// DrawPlacard draws the panel of a placard filling pn's box: a raised
// band, with no outline.
func DrawPlacard(pn Pen) {
	pn.Rect(0, 0, pn.W, pn.H, Gray2)
	pn.Bevel(0, 0, pn.W, pn.H, White, Gray7)
}

// Badge builds a badge: a short text, as a count, inverted — white on
// black in a square box — as the classic desktop marks what is selected.
func Badge(c *ui.Context, text string) ui.Element {
	return label(c, text, Black, White)
}

// Tag builds a tag: a short text in a light gray box with a black outline,
// as a keyword.
func Tag(c *ui.Context, text string) ui.Element {
	return label(c, text, Gray1, Black)
}

func label(c *ui.Context, text string, face, ink ui.Color) ui.Element {
	w := max(AppFont.Measure(text)+8, BadgeHeight)
	el := ui.Box(c).Size(float32(w), BadgeHeight).Role(ui.RoleStatus).Label(text)
	el.Draw(func(p *ui.Painter, r ui.Rect) {
		pn := NewPen(p, r)
		DrawLabelBox(pn, face)
		pn.TextCentered(0, 0, pn.W, pn.H, text, AppFont, ink)
	})
	return el
}

// DrawLabelBox draws the square box of a badge or tag filling pn's box:
// face inside a black outline.
func DrawLabelBox(pn Pen, face ui.Color) {
	pn.Rect(0, 0, pn.W, pn.H, face)
	pn.Frame(0, 0, pn.W, pn.H, Black)
}

// StatusBar builds a row of placards along the bottom of a window, one for
// each of texts, the last one taking the room left.
func StatusBar(c *ui.Context, texts ...string) ui.Element {
	return ui.Row(c).Height(PlacardHeight).AlignSelf(ui.Stretch).Role(ui.RoleStatus).Children(func() {
		for i, t := range texts {
			p := Placard(c, t)
			if i == len(texts)-1 {
				p.Grow(1)
			}
		}
	})
}
