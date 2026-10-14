package platinum

import (
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSplitAndScrollView(t *testing.T) {
	pos := 80.0
	var scroll ui.ScrollState
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() {
			Split(c, &pos, false, func() {
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
