package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestPopupWindow(t *testing.T) {
	freeze(t)
	files := PopupWindowState{X: 10, W: 200, H: 90, Open: true}
	tools := PopupWindowState{X: 230, W: 150, H: 80}
	var torn, closed bool
	front := "Files"
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Box(c).Fill().Background(Gray5).Children(func() {
			for _, p := range []struct {
				st    *PopupWindowState
				title string
			}{{&files, "Files"}, {&tools, "Tools"}} {
				w := PopupWindow(c.Key(p.title), p.st, p.title, WindowOptions{Active: front == p.title}, func() {
					ui.Column(c).Fill().Background(White).Padding(6).Children(func() {
						Label(c, "Read Me")
						Label(c, "Recipes")
					})
				})
				if w.Activated() {
					front = p.title
				}
				torn = torn || w.TornOff()
				closed = closed || w.Closed()
			}
		})
	}
	tt := ui.NewTester(view, 400, 140)
	tt.Frame()
	compareGolden(t, "popup-window", tt.Image())
	// A click on the closed tab opens its window.
	if err := tt.Click("Tools"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tools.Open || front != "Tools" {
		t.Fatalf("the click left it open %v, %s in front", tools.Open, front)
	}
	// Dragging the tab sideways moves the window, without closing it.
	r, _ := tt.Find("Tools")
	x, y := r.X+popupTabInset+20, r.Y+PopupTabHeight/2 // on the tab
	tt.Press(x, y)
	tt.Move(x-20, y)
	tt.Move(x-30, y)
	tt.Release(x-30, y)
	tt.Frame()
	if tools.X != 200 || !tools.Open {
		t.Fatalf("after the drag the window is at %d, open %v", tools.X, tools.Open)
	}
	// Dragging it up tears it off.
	r, _ = tt.Find("Tools")
	x, y = r.X+popupTabInset+20, r.Y+PopupTabHeight/2
	tt.Press(x, y)
	tt.Move(x, y-20)
	tt.Move(x, y-40)
	tt.Release(x, y-40)
	tt.Frame()
	tt.Frame()
	if !torn || !tools.Open {
		t.Fatalf("dragging the tab up tore it off %v, open %v", torn, tools.Open)
	}
	// The keyboard closes it, from its tab.
	for i := 0; i < 6 && !tt.Focused("Tools"); i++ {
		tt.Key(0, ui.KeyTab)
	}
	tt.Key(0, ui.KeySpace)
	tt.Frame()
	if tools.Open {
		t.Fatal("Space did not close the window")
	}
	compareGolden(t, "popup-window-focus", tt.Image())
	// The tab held down.
	r, _ = tt.Find("Files")
	tt.Press(r.X+popupTabInset+20, r.Y+PopupTabHeight/2)
	tt.Frame()
	compareGolden(t, "popup-window-pressed", tt.Image())
	tt.Release(r.X+popupTabInset+20, r.Y+PopupTabHeight/2)
	tt.Frame()
	if files.Open {
		t.Fatal("a click on the open window's tab did not close it")
	}
	// Command-W closes it, for the app to remove.
	files.Open = true
	tt.Frame()
	for i := 0; i < 6 && !tt.Focused("Files"); i++ {
		tt.Key(0, ui.KeyTab)
	}
	tt.Key(ui.Cmd, ui.KeyW)
	tt.Frame()
	if !closed {
		t.Fatal("Command-W did not close the window")
	}
}
