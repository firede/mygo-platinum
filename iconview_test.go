package platinum

import (
	"fmt"
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
				if iv.Submitted() {
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
	// The choice scrolled into view: find it where it is now.
	x, y = findCenter(t, tt, "Gamma")
	tt.ClickAt(x, y)
	tt.ClickAt(x, y)
	tt.Frame()
	if sel != 2 || opened != 1 {
		t.Fatalf("double click chose %d and activated %d times", sel, opened)
	}
	compareGolden(t, "iconview-utility", tt.Image())
}

func testIconViewItems(n int) []IconItem {
	items := make([]IconItem, n)
	ic := testIcon()
	for i := range items {
		items[i] = IconItem{Name: fmt.Sprintf("Icon %d", i+1), Icon: ic}
	}
	return items
}

// Content drawn by either API stays inside its box, even when the box is
// narrower than one cell or cuts through the last icon and its name.
func TestIconViewClipping(t *testing.T) {
	for _, mode := range []string{"widget", "drawing"} {
		for _, size := range []struct{ w, h float32 }{{280, 110}, {48, 48}} {
			t.Run(fmt.Sprintf("%s-%gx%g", mode, size.w, size.h), func(t *testing.T) {
				items, sel := testIconViewItems(4), -1
				tt := ui.NewTester(func(c *ui.Context) {
					Use(c, nil)
					ui.Box(c).Fill().Background(Gray5).Children(func() {
						if mode == "widget" {
							IconView(c, items, &sel).Element().Absolute().Left(10).Top(10).Size(size.w, size.h).Label("Icons")
						} else {
							ui.Box(c).Absolute().Left(10).Top(10).Size(size.w, size.h).Label("Icons").Draw(func(p *ui.Painter, r ui.Rect) {
								DrawIconView(NewPen(p, r), items, -1, false, Lavender)
							})
						}
					})
				}, 320, 180)
				r, _ := tt.Find("Icons")
				img := tt.Image()
				for y := range img.Bounds().Dy() {
					for x := range img.Bounds().Dx() {
						if float32(x) >= r.X && float32(x) < r.X+r.W && float32(y) >= r.Y && float32(y) < r.Y+r.H {
							continue
						}
						p := img.RGBAAt(x, y)
						if p.R != Gray5.R || p.G != Gray5.G || p.B != Gray5.B || p.A != 255 {
							t.Fatalf("content escaped %+v at (%d,%d): %v", r, x, y, p)
						}
					}
				}
				if mode == "widget" {
					for x := int(r.X) + 1; x < int(r.X+r.W)-1; x++ {
						p := img.RGBAAt(x, int(r.Y+r.H)-2)
						if p.R != 0 || p.G != 0 || p.B != 0 {
							t.Fatalf("content overwrote the bottom outline at x=%d: %v", x, p)
						}
					}
				}
			})
		}
	}
}

func TestIconViewScrolling(t *testing.T) {
	items, sel := testIconViewItems(12), -1
	changed, opened := 0, 0
	tt := ui.NewTester(func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() {
			iv := IconView(c, items, &sel)
			iv.Element().Size(280, 110).Label("Icons")
			if iv.Changed() {
				changed++
			}
			if iv.Submitted() {
				opened++
			}
		})
	}, 320, 150)
	r, _ := tt.Find("Icons")
	row, _ := tt.Find(items[3].Name)
	tt.Scroll(r.X+40, r.Y+40, 0, 64)
	settled(t, tt, "the icons scrolled by the wheel")
	after, ok := tt.Find(items[3].Name)
	if !ok || after.Y != row.Y-64 {
		t.Fatalf("the wheel moved row 2 from %+v to %+v", row, after)
	}
	if changed != 0 || sel != -1 {
		t.Fatalf("scrolling chose %d and reported %d changes", sel, changed)
	}
	// Hit testing follows the offset: this is the first icon of row 2.
	tt.ClickAt(r.X+40, r.Y+20)
	settled(t, tt, "an icon chosen after scrolling")
	if sel != 3 || changed != 1 {
		t.Fatalf("click after scrolling chose %d and reported %d changes", sel, changed)
	}
	tt.Key(0, ui.KeyEnter)
	if opened != 1 {
		t.Fatalf("Enter activated %d times", opened)
	}
	tt.Scroll(r.X+40, r.Y+40, 0, -10000)
	// The down arrow uses the same offset as the wheel, without choosing.
	tt.ClickAt(r.X+r.W-9, r.Y+r.H-9)
	settled(t, tt, "the icons scrolled by the arrow")
	after, _ = tt.Find(items[3].Name)
	if after.Y != row.Y-16 || sel != 3 || changed != 1 || opened != 1 {
		t.Fatalf("scroll arrow: row=%+v selection=%d changes=%d activations=%d", after, sel, changed, opened)
	}
	// Drag the thumb to the end, then select and double-click a last-row icon.
	pos, _, _ := scrollThumb(int(r.H)-2, 16, 106, 264)
	tt.Press(r.X+r.W-9, r.Y+1+float32(pos)+4)
	tt.Move(r.X+r.W-9, r.Y+r.H)
	tt.Release(r.X+r.W-9, r.Y+r.H)
	settled(t, tt, "the icons scrolled by the thumb")
	last, ok := tt.Find(items[9].Name)
	if !ok || last.H != IconCellH || last.Y+last.H > r.Y+r.H-2 {
		t.Fatalf("the last row is not fully visible: %+v", last)
	}
	tt.ClickAt(last.X+40, last.Y+20)
	tt.ClickAt(last.X+40, last.Y+20)
	if sel != 9 || opened != 2 {
		t.Fatalf("double click after scrolling: selection=%d activations=%d", sel, opened)
	}
}

func TestIconViewKeyboardScroll(t *testing.T) {
	items, sel := testIconViewItems(12), -1
	changes := 0
	tt := ui.NewTester(func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() {
			iv := IconView(c, items, &sel)
			iv.Element().Size(260, 110).Label("Icons")
			if iv.Changed() {
				changes++
			}
		})
	}, 300, 150)
	tt.Key(0, ui.KeyTab)
	for _, step := range []struct {
		key  ui.Key
		want int
	}{{ui.KeyRight, 0}, {ui.KeyDown, 2}, {ui.KeyDown, 4}, {ui.KeyEnd, 11}, {ui.KeyUp, 9},
		{ui.KeyHome, 0}, {ui.KeyPageDown, 2}, {ui.KeyPageUp, 0}} {
		before := changes
		tt.Key(0, step.key)
		settled(t, tt, "the icon chosen by the keyboard")
		if sel != step.want {
			t.Fatalf("key %v chose %d, want %d (the scroll bar leaves two columns)", step.key, sel, step.want)
		}
		if changes != before+1 {
			t.Fatalf("one key reported %d selection changes", changes-before)
		}
		r, _ := tt.Find("Icons")
		cell, ok := tt.Find(items[sel].Name)
		if !ok || cell.H != IconCellH || cell.Y < r.Y+2 || cell.Y+cell.H > r.Y+r.H-2 || cell.X+cell.W > r.X+r.W-ScrollBarWidth-1 {
			t.Fatalf("the selected cell is clipped: %+v in %+v", cell, r)
		}
	}
}

func TestIconViewResizeAndShrink(t *testing.T) {
	items, sel := testIconViewItems(12), -1
	w, h := float32(260), float32(110)
	tt := ui.NewTester(func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() {
			IconView(c, items, &sel).Element().Size(w, h).Label("Icons")
		})
	}, 340, 350)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyEnd)
	items = items[:1]
	tt.Frame()
	settled(t, tt, "the icons removed")
	r, _ := tt.Find("Icons")
	first, ok := tt.Find(items[0].Name)
	if sel != 0 || !ok || first.Y != r.Y+6 {
		t.Fatalf("shrinking did not reset the offset: selection=%d cell=%+v", sel, first)
	}
	items = nil
	tt.Frame()
	tt.Key(0, ui.KeyEnd)
	if sel != -1 {
		t.Fatalf("the empty view kept selection %d", sel)
	}
	items = testIconViewItems(12)
	tt.Frame()
	tt.Key(0, ui.KeyEnd)
	w, h = 280, 300 // all four rows fit without a bar, in three columns
	tt.Frame()
	settled(t, tt, "the icons resized to fit")
	r, _ = tt.Find("Icons")
	first, ok = tt.Find(items[0].Name)
	third, _ := tt.Find(items[2].Name)
	if !ok || first.Y != r.Y+6 || third.Y != first.Y || third.X != first.X+2*IconCellW {
		t.Fatalf("resizing did not remove the bar and restore three columns: %+v %+v", first, third)
	}
	// Shrink again at a width where the bar reduces three columns to two.
	w, h = 260, 110
	tt.Frame()
	tt.Key(0, ui.KeyHome)
	tt.Key(0, ui.KeyDown)
	settled(t, tt, "the icons resized to scroll")
	if sel != 2 {
		t.Fatalf("after resizing, Down chose %d instead of the next row's first icon", sel)
	}
}
