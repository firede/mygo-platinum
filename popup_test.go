package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestPopupMenu(t *testing.T) {
	freeze(t)
	sel := 0
	items := []string{"Name", "Date", "-", "Size"}
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(20).Children(func() {
			PopupMenu(c, &sel, "Sort by:", items)
		})
	}
	tt := ui.NewTester(view, 260, 160)
	r, ok := tt.Find("Sort by:")
	if !ok {
		t.Fatal("pop-up not found")
	}
	// Press, drag to Size, release.
	tt.Press(r.X+r.W-10, r.Y+10)
	tt.Frame()
	s, ok := tt.Find("Size")
	if !ok {
		t.Fatalf("menu not open: %q", tt.Texts())
	}
	tt.Move(s.X+10, s.Y+5)
	tt.Release(s.X+10, s.Y+5)
	tt.Frame()
	tt.Frame()
	if sel != 3 {
		t.Fatalf("selected %d, want 3", sel)
	}
	// Click to open, then click an item.
	tt.ClickAt(r.X+r.W-10, r.Y+10)
	tt.Frame()
	d, ok := tt.Find("Date")
	if !ok {
		t.Fatal("click did not leave the menu open")
	}
	compareGolden(t, "popup-open", tt.Image())
	tt.ClickAt(d.X+10, d.Y+5)
	tt.Frame()
	tt.Frame()
	if sel != 1 {
		t.Fatalf("selected %d, want 1", sel)
	}
	// The keyboard: focus with Tab, open with Space, move down, choose.
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeySpace)
	tt.Frame()
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	tt.Frame()
	if sel != 3 {
		t.Fatalf("keyboard selected %d, want 3", sel)
	}
}

func TestSnapshotPopupStates(t *testing.T) {
	snap(t, "popup-states", 420, 34, func(c *ui.Context) {
		board(c, 120, 20,
			func(pn Pen) { DrawPopupButton(pn, "Name", StateNormal) },
			func(pn Pen) { DrawPopupButton(pn, "Pressed", StatePressed) },
			func(pn Pen) { DrawPopupButton(pn, "Disabled", StateDisabled) },
		)
	})
}
