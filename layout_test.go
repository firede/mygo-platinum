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
