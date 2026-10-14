package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSnapshotScrollBar(t *testing.T) {
	snap(t, "scrollbar-states", 260, 130, func(c *ui.Context) {
		board(c, 16, 116,
			func(pn Pen) { DrawScrollBar(pn, false, 0, 10, 40, ScrollNone, true, Lavender) },
			func(pn Pen) { DrawScrollBar(pn, false, 15, 10, 40, ScrollThumb, true, Lavender) },
			func(pn Pen) { DrawScrollBar(pn, false, 30, 10, 40, ScrollDownArrow, true, Lavender) },
			func(pn Pen) { DrawScrollBar(pn, false, 10, 10, 40, ScrollPageDown, true, Lavender) },
			func(pn Pen) { DrawScrollBar(pn, false, 0, 10, 5, ScrollNone, false, Lavender) },
		)
		ui.Box(c).Absolute().Left(120).Top(10).Size(130, 16).Draw(func(p *ui.Painter, r ui.Rect) {
			DrawScrollBar(NewPen(p, r), true, 20, 10, 40, ScrollUpArrow, true, Lavender)
		})
	})
}

func TestScrollBarArrowsAndDrag(t *testing.T) {
	v := 0.0
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() {
			ScrollBar(c, &v, 10, 100).Step(1).Element().Height(120).Label("Bar")
		})
	}
	tt := ui.NewTester(view, 60, 140)
	r, _ := tt.Find("Bar")
	tt.ClickAt(r.X+8, r.Y+r.H-8) // the down arrow
	tt.Frame()
	if v != 1 {
		t.Fatalf("down arrow scrolled to %v", v)
	}
	tt.ClickAt(r.X+8, r.Y+r.H-30) // the track below the thumb
	tt.Frame()
	if v != 10 {
		t.Fatalf("page down scrolled to %v", v)
	}
	// Drag the thumb to the bottom.
	pos, _, _ := scrollThumb(int(r.H), v, 10, 100)
	tt.Press(r.X+8, r.Y+float32(pos)+4)
	tt.Move(r.X+8, r.Y+r.H)
	tt.Release(r.X+8, r.Y+r.H)
	tt.Frame()
	if v != 90 {
		t.Fatalf("thumb dragged to %v", v)
	}
}
