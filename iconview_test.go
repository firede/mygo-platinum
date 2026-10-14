package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestIconViewAndUtility(t *testing.T) {
	ic := testIcon()
	items := []IconItem{{"Alpha", ic}, {"Beta", ic}, {"Gamma", ic}, {"Delta", ic}}
	sel := -1
	opened := 0
	frame := WindowState{X: 10, Y: 10, W: 280, H: 120}
	tools := WindowState{X: 220, Y: 100, W: 90, H: 60}
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Box(c).Fill().Background(Gray5).Children(func() {
			Window(c, &frame, "Icons", WindowOptions{Active: true}, func() {
				iv := IconView(c, items, &sel)
				iv.Element().Grow(1).FillWidth().Label("Icons")
				if iv.Activated() {
					opened++
				}
			})
			Window(c, &tools, "Tools", WindowOptions{Utility: true, Active: true}, func() {})
		})
	}
	tt := ui.NewTester(view, 320, 180)
	tt.Frame()
	x, y := findCenter(t, tt, "Beta")
	tt.ClickAt(x, y)
	tt.Frame()
	if sel != 1 {
		t.Fatalf("click chose %d", sel)
	}
	tt.Key(0, ui.KeyRight)
	tt.Frame()
	if sel != 2 {
		t.Fatalf("Right chose %d", sel)
	}
	tt.ClickAt(x, y)
	tt.ClickAt(x, y)
	tt.Frame()
	if opened != 1 {
		t.Fatalf("double click activated %d times", opened)
	}
	compareGolden(t, "iconview-utility", tt.Image())
}
