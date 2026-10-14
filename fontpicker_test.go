package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestFontPickerAndPageSetup(t *testing.T) {
	freeze(t)
	choice := FontChoice{Family: "Serif", Size: 12, Italic: true, Underline: true}
	open := true
	page := PageSetupState{Paper: 0, Scale: 100, Margin: 36}
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() {
			FontPicker(c, &choice, []string{"Sans", "Serif", "Mono"})
			PageSetup(c, &open, &page)
		})
	}
	tt := ui.NewTester(view, 420, 420)
	tt.Frame()
	compareGolden(t, "pagesetup", tt.Image())
	if err := tt.Click("Landscape"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("OK"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !page.Landscape || open {
		t.Fatalf("page %+v, open %v", page, open)
	}
	if err := tt.Click("Bold"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	compareGolden(t, "fontpicker", tt.Image())
	if !choice.Bold {
		t.Fatal("Bold is not checked")
	}
}
