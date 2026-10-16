package platinum

import (
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"
)

type winApp struct {
	front, back WindowState
	closed      int
	on          bool
}

func (a *winApp) view(c *ui.Context) {
	Use(c, nil)
	ui.Box(c).Fill().Background(Gray5).Children(func() {
		Window(c, &a.back, "Background", WindowOptions{}, func() {})
		w := Window(c, &a.front, "Untitled", WindowOptions{Active: true, ZoomTo: [4]int{2, 2, 296, 216}}, func() {
			ui.Column(c).Padding(10).Children(func() {
				Checkbox(c, &a.on, "Inside")
			})
		})
		if w.Closed() {
			a.closed++
		}
	})
}

func newWinApp() *winApp {
	return &winApp{
		front: WindowState{X: 30, Y: 30, W: 180, H: 120},
		back:  WindowState{X: 120, Y: 90, W: 160, H: 110},
	}
}

func TestSnapshotWindows(t *testing.T) {
	a := newWinApp()
	tt := ui.NewTester(a.view, 300, 220)
	compareGolden(t, "window-frames", tt.Image())
}

func TestWindowDragResizeBoxes(t *testing.T) {
	a := newWinApp()
	tt := ui.NewTester(a.view, 300, 220)
	// Drag the title bar.
	tt.Press(60, 38)
	tt.Move(70, 48)
	tt.Frame()
	tt.Move(80, 58)
	tt.Release(80, 58)
	tt.Frame()
	if a.front.X != 50 || a.front.Y != 50 {
		t.Fatalf("dragged to %d,%d", a.front.X, a.front.Y)
	}
	// Resize from the corner.
	cx, cy := float32(a.front.X+a.front.W-4), float32(a.front.Y+a.front.H-4)
	tt.Press(cx, cy)
	tt.Move(cx+30, cy+20)
	tt.Release(cx+30, cy+20)
	tt.Frame()
	if a.front.W != 210 || a.front.H != 140 {
		t.Fatalf("resized to %dx%d", a.front.W, a.front.H)
	}
	// Press the close box, drag off it, release: nothing happens.
	bx, by := float32(a.front.X+titleBoxInset+5), float32(a.front.Y+stripeTop+5)
	tt.Press(bx, by)
	tt.Move(bx+40, by)
	tt.Release(bx+40, by)
	tt.Frame()
	if a.closed != 0 {
		t.Fatal("close box acted though released off it")
	}
	tt.ClickAt(bx, by)
	tt.Frame()
	if a.closed != 1 {
		t.Fatalf("close box clicked, closed %d", a.closed)
	}
	// The collapse box rolls it up; a double click on the title unrolls it.
	tt.ClickAt(float32(a.front.X+a.front.W-titleBoxInset-5), by)
	tt.Frame()
	if !a.front.Collapsed {
		t.Fatal("collapse box did not roll up the window")
	}
	tt.ClickAt(float32(a.front.X+60), by)
	tt.ClickAtWith(0, float32(a.front.X+60), by)
	tt.Frame()
	// The zoom box.
	if a.front.Collapsed {
		tt.ClickAt(float32(a.front.X+a.front.W-titleBoxInset-5), by)
		tt.Frame()
	}
	tt.ClickAt(float32(a.front.X+a.front.W-titleBoxInset-TitleBoxSize-titleBoxGap-5), by)
	tt.Frame()
	if !a.front.Zoomed || a.front.W != 296 {
		t.Fatalf("zoom box gave %+v", a.front)
	}
}

func TestWindowBoxesUpdateGeometryImmediately(t *testing.T) {
	state := WindowState{X: 30, Y: 30, W: 180, H: 120}
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Box(c).Fill().Background(Gray5).Children(func() {
			Window(c, &state, "Window", WindowOptions{Active: true, ZoomTo: [4]int{2, 2, 296, 216}}, func() {
				ui.Box(c).Size(60, 30).Label("Content")
			})
		})
	}
	tt := ui.NewTester(view, 300, 220)
	check := func(height int, content bool) {
		t.Helper()
		r, ok := tt.Find("Window")
		want := ui.Rect{X: float32(state.X), Y: float32(state.Y), W: float32(state.W), H: float32(height)}
		if !ok || r != want {
			t.Fatalf("window bounds after click: got %+v, want %+v", r, want)
		}
		if _, visible := tt.Find("Content"); visible != content {
			t.Fatalf("content visible: got %v, want %v", visible, content)
		}
	}
	// Inspect the click's own frame: an extra Frame would hide the stale
	// geometry that otherwise persists until another input or timer fires.
	for i := 0; i < 3; i++ {
		tt.ClickAt(float32(state.X+state.W-titleBoxInset-5), float32(state.Y+stripeTop+5))
		check(TitleBarHeight, false)
		tt.ClickAt(float32(state.X+state.W-titleBoxInset-5), float32(state.Y+stripeTop+5))
		check(state.H, true)
	}
	for i := 0; i < 2; i++ {
		tt.ClickAt(float32(state.X+state.W-titleBoxInset-TitleBoxSize-titleBoxGap-5), float32(state.Y+stripeTop+5))
		check(state.H, true)
	}
}

// TestBackgroundWindowFirstClick clicks a check box in a window in the
// background: the click brings the window to the front, and the box stays
// as it was until the next click.
func TestBackgroundWindowFirstClick(t *testing.T) {
	on := false
	front := 1
	frames := [2]WindowState{{X: 10, Y: 10, W: 160, H: 100}, {X: 120, Y: 60, W: 160, H: 100}}
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Box(c).Fill().Children(func() {
			for i := range frames {
				w := Window(c.Key(i), &frames[i], fmt.Sprint("Window ", i), WindowOptions{Active: i == front}, func() {
					if i == 0 {
						ui.Column(c).Padding(8).Children(func() { Checkbox(c, &on, "Option") })
					}
				})
				if w.Activated() {
					front = i
				}
			}
		})
	}
	tt := ui.NewTester(view, 300, 200)
	x, y := findCenter(t, tt, "Option")
	tt.ClickAt(x-20, y)
	tt.Frame()
	if on || front != 0 {
		t.Fatalf("first click: checked %v, front %d", on, front)
	}
	tt.ClickAt(x-20, y)
	tt.Frame()
	if !on {
		t.Fatal("the second click did not check the box")
	}
}

// A window in the background is dragged by its title bar at once, as it
// comes to the front: only its content keeps the first click.
func TestBackgroundWindowTitleDrags(t *testing.T) {
	front := 1
	frames := [2]WindowState{{X: 10, Y: 10, W: 160, H: 100}, {X: 120, Y: 60, W: 160, H: 100}}
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Box(c).Fill().Children(func() {
			for i := range frames {
				w := Window(c.Key(i), &frames[i], fmt.Sprint("Window ", i), WindowOptions{Active: i == front}, func() {})
				if w.Activated() {
					front = i
				}
			}
		})
	}
	tt := ui.NewTester(view, 300, 200)
	tt.Press(30, 16)
	for i := 1; i <= 4; i++ {
		tt.Move(30+float32(5*i), 16+float32(3*i))
		tt.Frame()
	}
	tt.Release(50, 28)
	tt.Frame()
	if front != 0 || frames[0].X != 30 || frames[0].Y != 22 {
		t.Fatalf("dragged from the background: front %d, window at %d,%d", front, frames[0].X, frames[0].Y)
	}
}

// TestBackgroundWindowTakesDrops drags a value from the window in front to
// one in the background: told a drag is under way (Dragging), the window
// in the background lets it drop in its content, and stays behind.
func TestBackgroundWindowTakesDrops(t *testing.T) {
	front := 1
	frames := [2]WindowState{{X: 10, Y: 10, W: 160, H: 120}, {X: 180, Y: 10, W: 110, H: 120}}
	dragging, seen := false, false
	got := ""
	view := func(c *ui.Context) {
		Use(c, nil)
		dragging, seen = seen, false
		ui.Box(c).Fill().Children(func() {
			for i := range frames {
				w := Window(c.Key(i), &frames[i], fmt.Sprint("Window ", i), WindowOptions{Active: i == front, Dragging: dragging}, func() {
					if i == 0 {
						target := ui.Box(c).Size(120, 70).Label("Target")
						if v, ok := ui.Drop[string](target); ok {
							got = v
						}
						return
					}
					src := ui.Box(c).Size(60, 30).Label("Source").Drag("note")
					seen = seen || src.Dragging()
				})
				if w.Activated() {
					front = i
				}
			}
		})
	}
	tt := ui.NewTester(view, 300, 150)
	tt.Frame()
	x0, y0 := findCenter(t, tt, "Source")
	x1, y1 := findCenter(t, tt, "Target")
	tt.Press(x0, y0)
	for i := 1; i <= 4; i++ {
		f := float32(i) / 4
		tt.Move(x0+(x1-x0)*f, y0+(y1-y0)*f)
		tt.Frame()
	}
	tt.Release(x1, y1)
	tt.Frame()
	if got != "note" || front != 1 {
		t.Fatalf("dropped %q, front %d", got, front)
	}
}

// TestWindowPathMenu Command-presses a folder window's title: a menu of
// the folder and those it is in drops from it, and releasing on one
// chooses it.
func TestWindowPathMenu(t *testing.T) {
	freeze(t)
	frame := WindowState{X: 20, Y: 10, W: 220, H: 120}
	chosen := -1
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Box(c).Fill().Children(func() {
			w := Window(c, &frame, "Letters", WindowOptions{Active: true, Path: []string{"Platinum Disk", "Documents"}}, func() {})
			if i := w.PathChosen(); i >= 0 {
				chosen = i
			}
		})
	}
	tt := ui.NewTester(view, 260, 150)
	tt.Frame()
	x, y := float32(20+110), float32(10+9)
	tt.ClickAtWith(ui.Cmd, x, y) // a quick click leaves it open
	tt.Frame()
	r, ok := tt.Find("Platinum Disk")
	if !ok {
		t.Fatal("no path menu")
	}
	tt.Move(r.X+r.W/2, r.Y+r.H/2)
	tt.Frame()
	compareGolden(t, "window-path-menu", tt.Image())
	tt.ClickAt(r.X+r.W/2, r.Y+r.H/2)
	for i := 0; i < 3; i++ {
		tt.Frame()
	}
	if chosen != 0 {
		t.Fatalf("chose %d from the path menu", chosen)
	}
	if frame.X != 20 || frame.Y != 10 {
		t.Fatal("the Command-press moved the window")
	}
}

func TestWindowHeader(t *testing.T) {
	freeze(t)
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(8).AlignItems(ui.Stretch).Children(func() {
			WindowHeader(c, "4 items", "47.7 MB available")
		})
	}
	tt := ui.NewTester(view, 260, 34)
	tt.Frame()
	compareGolden(t, "window-header", tt.Image())
	if !tt.HasText("4 items, 47.7 MB available") {
		t.Fatal("the header does not tell assistive technology what it says")
	}
}
