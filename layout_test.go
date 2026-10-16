package platinum

import (
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSplitAndScrollView(t *testing.T) {
	pos := float32(80)
	var scroll ui.ScrollState
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() {
			Split(c, &pos, func() {
				TextIn(c, "Left pane", AppFont, Black)
			}, func() {
				ScrollView(c, &scroll, func() {
					for i := 0; i < 20; i++ {
						TextIn(c, fmt.Sprintf("Line %d", i+1), AppFont, Black)
					}
				}).Grow(1)
			}).Size(260, 100)
		})
	}
	tt := ui.NewTester(view, 290, 130)
	d, _ := tt.Find("Divider")
	tt.Press(d.X+3, d.Y+20)
	tt.Move(d.X+33, d.Y+20)
	tt.Release(d.X+33, d.Y+20)
	tt.Frame()
	if pos != 110 {
		t.Fatalf("dragged the divider to %v", pos)
	}
	tt.Key(0, ui.KeyLeft)
	tt.Frame()
	if pos != 102 {
		t.Fatalf("Left moved it to %v", pos)
	}
	// The scroll bar follows the content, and scrolls it.
	tt.Scroll(200, 60, 0, 40)
	tt.Frame()
	if scroll.Y <= 0 {
		t.Fatalf("scrolled to %v", scroll.Y)
	}
	compareGolden(t, "split-scrollview", tt.Image())
}

// TestSnapshotSplitVertical draws a split whose panes are stacked, and
// moves its divider with the arrows, Up and Down.
func TestSnapshotSplitVertical(t *testing.T) {
	pos := float32(40)
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() {
			SplitVertical(c, &pos, func() {
				TextIn(c, "Top pane", AppFont, Black)
			}, func() {
				TextIn(c, "Bottom pane", AppFont, Black)
			}).Size(160, 120)
		})
	}
	tt := ui.NewTester(view, 180, 140)
	tt.Frame()
	compareGolden(t, "split-vertical", tt.Image())
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyDown)
	tt.Frame()
	if pos != 48 {
		t.Fatalf("Down moved the divider to %v", pos)
	}
}

// TestSplitPercent shares the room by a percentage as the split grows,
// and turns drags and keys into one.
func TestSplitPercent(t *testing.T) {
	p, w := float32(25), float32(207) // 200 to share beside the divider
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() {
			SplitPercent(c, &p, func() {
				ui.Box(c).Grow(1).Label("One")
			}, func() {
				ui.Box(c).Grow(1).Label("Two")
			}).Size(w, 100)
		})
	}
	tt := ui.NewTester(view, 440, 130)
	tt.Frame()
	width := func() float32 {
		tt.Frame()
		r, _ := tt.Find("One")
		return r.W
	}
	if got := width(); got != 50 {
		t.Fatalf("25%% of 200 is %v wide", got)
	}
	w = 407
	if got := width(); got != 100 {
		t.Fatalf("25%% of 400 is %v wide", got)
	}
	d, _ := tt.Find("Divider")
	tt.Press(d.X+3, d.Y+20)
	tt.Move(d.X+33, d.Y+20)
	tt.Release(d.X+33, d.Y+20)
	tt.Frame()
	if p != 32.5 || width() != 130 {
		t.Fatalf("dragged by 30 to %v%%, %v wide", p, width())
	}
	tt.Key(0, ui.KeyRight)
	tt.Frame()
	if p != 34.5 {
		t.Fatalf("Right moved it to %v%%", p)
	}
	p = 0
	if got := width(); got != SplitMin {
		t.Fatalf("0%% is %v wide, not the least", got)
	}
}

// A split too small for its first pane's width shrinks the pane and keeps
// the width, for when it grows again.
func TestSplitKeepsItsWidth(t *testing.T) {
	pos, w := float32(150), float32(307)
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() {
			Split(c, &pos, func() {
				ui.Box(c).Grow(1).Label("One")
			}, func() {
				ui.Box(c).Grow(1).Label("Two")
			}).Size(w, 100)
		})
	}
	tt := ui.NewTester(view, 340, 130)
	width := func() float32 {
		tt.Frame()
		tt.Frame()
		r, _ := tt.Find("One")
		return r.W
	}
	if got := width(); got != 150 {
		t.Fatalf("the first pane is %v wide", got)
	}
	w = 157 // 150 to share: the second pane keeps its 40
	if got := width(); got != 110 || pos != 150 {
		t.Fatalf("shrunk: the first pane is %v wide, pos %v", got, pos)
	}
	w = 307
	if got := width(); got != 150 {
		t.Fatalf("grown again: the first pane is %v wide", got)
	}
}

// TestSnapshotSplitPercent draws a split whose panes are stacked, shared
// by a percentage.
func TestSnapshotSplitPercent(t *testing.T) {
	p := float32(30)
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() {
			SplitVerticalPercent(c, &p, func() {
				TextIn(c, "30%", AppFont, Black)
			}, func() {
				TextIn(c, "70%", AppFont, Black)
			}).Size(160, 127)
		})
	}
	tt := ui.NewTester(view, 180, 147)
	tt.Frame()
	compareGolden(t, "split-percent", tt.Image())
}
