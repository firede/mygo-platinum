package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestRuler(t *testing.T) {
	st := &RulerState{Left: 16, Right: 240}
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(8).AlignItems(ui.Stretch).Children(func() { Ruler(c, st) })
	}
	tt := ui.NewTester(view, 300, 50)
	r, _ := tt.Find("Ruler")
	tt.Press(r.X+16, r.Y+17)
	tt.Move(r.X+40, r.Y+17)
	tt.Release(r.X+40, r.Y+17)
	tt.Frame()
	if st.Left != 36 { // in steps of 9, an eighth of an inch
		t.Fatalf("dragged the left margin to %d", st.Left)
	}
	// The first-line indent, on the scale, moves with the margin and drags
	// on its own.
	tt.Press(r.X+37, r.Y+3)
	tt.Move(r.X+60, r.Y+3)
	tt.Release(r.X+60, r.Y+3)
	tt.Frame()
	if st.First != 18 || st.Left != 36 {
		t.Fatalf("dragged the first-line indent to %d, the margin to %d", st.First, st.Left)
	}
	compareGolden(t, "ruler", tt.Image())
}

// TestFirstLineIndent starts the first line of each paragraph of a text
// area further in, wrapping it sooner.
func TestFirstLineIndent(t *testing.T) {
	freeze(t)
	text := "A first paragraph long enough to wrap.\nA second one."
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(8).AlignItems(ui.Stretch).Children(func() {
			TextArea(c, &text).Label("Text").Indent(24).Element().Size(160, 70)
		})
	}
	tt := ui.NewTester(view, 180, 90)
	tt.Frame()
	compareGolden(t, "first-line-indent", tt.Image())
}
