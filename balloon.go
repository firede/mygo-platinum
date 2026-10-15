package platinum

import (
	"github.com/egoist/mygo/ui"
)

// Balloon gives anchor a help balloon: a rounded white balloon with a tail
// pointing at it, holding text in the regular face, which shows as the
// pointer rests on anchor or the keyboard focus comes to it.
//
// The counterpart among MyGo's own controls is Element.Tooltip.
func Balloon(c *ui.Context, anchor ui.Element, text string) {
	ui.TooltipBase(c, anchor, func(tip ui.Element) {
		width := min(AppFont.Measure(text), 200)
		e := &editor{font: AppFont, multi: true, wrap: width, text: []rune(text)}
		e.layout()
		lineH := AppFont.Height() + 2
		h := len(e.lines)*lineH + 12 + balloonTail
		// Under anchor, its tail's point under anchor's middle: MyGo puts
		// a tip above, centered, where a balloon's tail rises from its
		// top left.
		tip.Size(float32(width+20), float32(h)).AttachTo(anchor, ui.AnchorBottom, ui.AnchorTopLeft).
			Margin(2, 0, 0, 0).Left(-balloonTailX)
		tip.Draw(func(p *ui.Painter, r ui.Rect) {
			pn := NewPen(p, r)
			DrawBalloon(pn)
			for i, l := range e.lines {
				pn.Text(10, balloonTail+6+i*lineH, string(e.text[l.start:l.end]), AppFont, Black)
			}
		})
	})
}

const (
	balloonTail  = 8  // the height of a balloon's tail
	balloonTailX = 13 // where its point is, from the balloon's left
)

// DrawBalloon draws a help balloon filling pn's box, its tail at the top
// left, balloonTail tall.
func DrawBalloon(pn Pen) {
	w, h := pn.W, pn.H-balloonTail
	body := pn.Sub(0, balloonTail, w-1, h-1)
	// A shadow, then the balloon.
	pn.Sub(1, balloonTail+1, w-1, h-1).FillRound(0, 0, w-1, h-1, 4, Gray9.Alpha(0.5))
	body.RoundFrame(0, 0, body.W, body.H, 4, Black, White)
	body.FillRound(1, 1, body.W-2, body.H-2, 3, White)
	// The tail: a wedge from the body up to a point over the anchor.
	// Its left edge stands straight up from the body; its right edge
	// slants down and away to meet the body.
	for y := 0; y < balloonTail; y++ {
		x0, x1 := 12, 12+y
		pn.HLine(x0+1, y, x1-x0, White)
		pn.Dot(x0, y, Black)
		pn.Dot(x1+1, y, Black)
	}
	pn.HLine(13, balloonTail, balloonTail, White)
}
