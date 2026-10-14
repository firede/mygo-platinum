package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestToolbarOverflow(t *testing.T) {
	freeze(t)
	bold := false
	ran := 0
	items := []ToolItem{
		{Label: "One", Action: func() { ran++ }},
		{Label: "Two"},
		{Separator: true},
		{Label: "Three"},
		{Label: "Bold", On: &bold},
	}
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Width(150).Children(func() { Toolbar(c, items) })
	}
	tt := ui.NewTester(view, 150, 140)
	tt.Frame()
	if _, ok := tt.Find("Bold"); ok && func() bool { r, _ := tt.Find("Bold"); return r.Y < 32 }() {
		t.Fatal("Bold should not fit")
	}
	compareGolden(t, "toolbar-overflow", tt.Image())
	if err := tt.Click("More"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	x, y := findCenter(t, tt, "Bold")
	tt.ClickAt(x, y)
	tt.Frame()
	tt.Frame()
	if !bold {
		t.Fatal("the overflow menu did not toggle Bold")
	}
	if err := tt.Click("One"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if ran != 1 {
		t.Fatalf("One ran %d times", ran)
	}
}
