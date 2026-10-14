package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestStepperAndNumberField(t *testing.T) {
	v := 5.0
	n := 12.0
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Gap(8).Children(func() {
			Stepper(c, &v, 0, 10, 1).Label("Count")
			NumberField(c, &n, 6, 72, 2, 0).Label("Size")
		})
	}
	tt := ui.NewTester(view, 200, 100)
	r, _ := tt.Find("Count")
	tt.ClickAt(r.X+6, r.Y+4) // the up half
	tt.Frame()
	if v != 6 {
		t.Fatalf("up made %v", v)
	}
	tt.ClickAt(r.X+6, r.Y+r.H-4) // the down half
	tt.Frame()
	tt.Key(0, ui.KeyDown)
	tt.Frame()
	if v != 4 {
		t.Fatalf("down twice made %v", v)
	}
	// Type a number into the field and submit it.
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("30")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if n != 30 {
		t.Fatalf("typing 30 made %v", n)
	}
	tt.Key(0, ui.KeyUp)
	tt.Frame()
	if n != 32 {
		t.Fatalf("Up made %v", n)
	}
	// Out of range text is put back in order.
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("500")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	tt.Frame()
	if n != 32 {
		t.Fatalf("500 made %v", n)
	}
}
