package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// TestRadioGroupArrows checks that the arrows choose within a radio
// group, as they move the focus, passing over a disabled button and going
// round.
func TestRadioGroupArrows(t *testing.T) {
	size := "s"
	view := func(c *ui.Context) {
		Use(c, nil)
		RadioGroup(c, func() {
			Radio(c, &size, "s", "Small")
			Radio(c, &size, "m", "Medium")
			Radio(c, &size, "l", "Large").Disabled(true)
		})
	}
	tt := ui.NewTester(view, 200, 80)
	tt.Key(0, ui.KeyTab)
	for _, want := range []string{"m", "s"} {
		tt.Key(0, ui.KeyDown)
		tt.Frame()
		if size != want {
			t.Fatalf("Down chose %q, want %q", size, want)
		}
	}
}
