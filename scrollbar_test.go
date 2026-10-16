package platinum

import (
	"testing"
	"time"

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

// An arrow held down repeats by the time of the events and frames: the
// Tester's clock, which only the test moves, in the tests of an app that
// sets no clock of its own.
func TestScrollBarRepeatsByTheFramesClock(t *testing.T) {
	defer SetClock(SetClock(nil))
	v := 0.0
	tt := ui.NewTester(func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() {
			ScrollBar(c, &v, 100, 1000).Step(10).Element().Height(120).Label("Bar")
		})
	}, 60, 160)
	r, _ := tt.Find("Bar")
	tt.Press(r.X+8, r.Y+r.H-8)
	tt.Frame()
	for _, step := range []struct {
		d    time.Duration
		want float64
	}{{0, 10}, {scrollDelay - time.Millisecond, 10}, {time.Millisecond, 20}, {2 * scrollRepeat, 40}} {
		tt.Advance(step.d)
		if v != step.want {
			t.Fatalf("held %v more: value %g, want %g", step.d, v, step.want)
		}
	}
	tt.Release(r.X+8, r.Y+r.H-8)
	tt.Advance(time.Second)
	if v != 40 {
		t.Fatalf("the arrow repeated after it was let go: %g", v)
	}
}
