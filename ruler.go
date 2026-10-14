package platinum

import (
	"github.com/egoist/mygo/ui"
)

// Ruler metrics, from docs/spec.md.
const (
	RulerHeight = 22 // the height of a ruler
	rulerUnit   = 9  // pixels between ticks: an eighth of an inch at 72 dpi
)

// RulerState is the margins a ruler sets, in pixels from its left, and
// the first-line indent, from the left margin.
type RulerState struct {
	Left, Right int
	First       int
	drag        int // 1 the left marker, 2 the right one, 3 the first-line one
}

// Ruler builds a ruler over a document: a scale from the left margin,
// ticked every 9 pixels and numbered every 72 (an inch at 72 dpi), with
// markers of the left and right margins under it and of the first-line
// indent on it, which the pointer drags in steps of 9. The text between
// the margins is the document's width; the first line of each paragraph
// starts First further in.
func Ruler(c *ui.Context, st *RulerState) ui.Element {
	el := ui.Box(c).Height(RulerHeight).AlignSelf(ui.Stretch).Role(ui.RoleSlider).Label("Ruler")
	dim := inactive(c)
	el.Range(0, 1000, float64(st.Right-st.Left))
	var width float32
	el.HandleInput(func(ev ui.InputEvent) bool {
		x := round(ev.X)
		switch ev.Kind {
		case ui.InputPointerDown:
			switch {
			case ev.Y < rulerFirstBottom+2 && abs(x-st.Left-st.First) <= 4:
				st.drag = 3
			case abs(x-st.Left) <= 5:
				st.drag = 1
			case abs(x-st.Right) <= 5:
				st.drag = 2
			default:
				return false
			}
			return true
		case ui.InputPointerMove:
			if st.drag == 1 {
				st.Left = min(max(x/rulerUnit*rulerUnit, 0), st.Right-4*rulerUnit)
				return true
			}
			if st.drag == 2 {
				// In steps from the left margin, where the scale starts.
				st.Right = st.Left + max(min((x-st.Left+rulerUnit/2)/rulerUnit*rulerUnit, round(width)-1-st.Left), 4*rulerUnit)
				st.First = min(st.First, st.Right-st.Left-4*rulerUnit)
				return true
			}
			if st.drag == 3 {
				st.First = min(max((x-st.Left)/rulerUnit*rulerUnit, 0), st.Right-st.Left-4*rulerUnit)
				return true
			}
		case ui.InputPointerUp:
			st.drag = 0
		}
		return false
	})
	el.Draw(func(p *ui.Painter, r ui.Rect) {
		width = r.W
		DrawRuler(NewPen(p, r).Dimmed(dim), st.Left, st.Right, st.First)
	})
	return el
}

var (
	marginLeft  = []string{"#......", "##.....", "###....", "####...", "#####..", "######.", "#######"}
	marginRight = []string{"......#", ".....##", "....###", "...####", "..#####", ".######", "#######"}
)

// DrawRuler draws a ruler filling pn's box, with its margin markers at left
// and right, and its first-line indent marker first in from left.
func DrawRuler(pn Pen, left, right, first int) {
	w, h := pn.W, pn.H
	pn.Rect(0, 0, w, h, Gray2)
	pn.HLine(0, h-1, w, Black)
	// The scale: a band 12 tall, white between the margins, ticks rising
	// from its bottom, the whole units numbered right of their ticks.
	const top, bottom = 2, 14
	pn.Rect(0, top, w, bottom-top, Gray1)
	pn.Rect(left, top, right-left, bottom-top, White)
	pn.HLine(0, top-1, w, Gray7)
	pn.HLine(0, bottom, w, Gray7)
	// The scale starts at the left margin, an inch every 72 pixels.
	for k := -(left / rulerUnit); left+k*rulerUnit < w; k++ {
		x := left + k*rulerUnit
		th := 2
		switch {
		case k%8 == 0:
			th = bottom - top
		case k%4 == 0:
			th = 5
		}
		pn.VLine(x, bottom-th, th, Black)
		if k%8 == 0 && k > 0 {
			pn.Text(x+3, top+1, itoa(k/8), AppFont, Black)
		}
	}
	pn.Bitmap(left, bottom+1, marginLeft[1:], Black)
	pn.Bitmap(right-6, bottom+1, marginRight[1:], Black)
	// The first-line indent: a small triangle pointing down on the scale.
	pn.Bitmap(left+first-2, top, firstIndent, Black)
}

var firstIndent = []string{"#####", ".###.", "..#.."}

const rulerFirstBottom = 5 // the bottom of the first-line marker's reach

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for ; n > 0; n /= 10 {
		s = string(rune('0'+n%10)) + s
	}
	return s
}
