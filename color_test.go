package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestColorWell(t *testing.T) {
	freeze(t)
	col := Lavender.Mid
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() { ColorWell(c, &col, "Highlight") })
	}
	tt := ui.NewTester(view, 340, 300)
	if err := tt.Click("Highlight"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	p, ok := tt.Find("Palette")
	if !ok {
		t.Fatal("the picker did not open")
	}
	tt.ClickAt(p.X+2+swatchSize/2, p.Y+2+2*swatchSize+swatchSize/2) // the third row's red
	tt.Frame()
	compareGolden(t, "colorpicker", tt.Image())
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if col != pickerPalette[16] {
		t.Fatalf("chose %v, want %v", col, pickerPalette[16])
	}
}
