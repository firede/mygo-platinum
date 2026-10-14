package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

var fruits = []string{"Almond", "Apricot", "Banana", "Cherry", "Date", "Elderberry", "Fig", "Grape", "Kiwi", "Lemon"}

func TestListBoxKeyboard(t *testing.T) {
	sel := 0
	activated := 0
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() {
			lb := ListBox(c, &sel, fruits)
			lb.Element().Size(140, 98)
			if lb.Activated() {
				activated++
			}
		})
	}
	tt := ui.NewTester(view, 200, 140)
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyDown)
	tt.Frame()
	if sel != 2 {
		t.Fatalf("Down twice chose %d", sel)
	}
	tt.Key(0, ui.KeyEnd)
	tt.Frame()
	if sel != 9 || !tt.HasText("Lemon") {
		t.Fatalf("End chose %d; rows %q", sel, tt.Texts())
	}
	tt.TypeKey(0, ui.KeyC, "c")
	tt.Frame()
	if sel != 3 {
		t.Fatalf("typing c chose %d", sel)
	}
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if activated != 1 {
		t.Fatalf("Return activated %d times", activated)
	}
	r, _ := tt.Find("Cherry")
	tt.ClickAt(r.X+10, r.Y+20) // the row below: Date
	tt.Frame()
	if sel != 4 {
		t.Fatalf("click chose %d", sel)
	}
	compareGolden(t, "listbox-focused", tt.Image())
}

func TestSnapshotListBoxStates(t *testing.T) {
	snap(t, "listbox-states", 330, 80, func(c *ui.Context) {
		board(c, 100, 66,
			func(pn Pen) { DrawListBox(pn, fruits, 0, 1, true, false, Lavender) },
			func(pn Pen) { DrawListBox(pn, fruits, 2, 3, false, false, Lavender) },
			func(pn Pen) { DrawListBox(pn, fruits, 0, 1, false, true, Lavender) },
		)
	})
}
