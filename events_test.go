package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// TestOnChange checks that MyGo's OnChange, given the element of a control
// of the library, runs as the user changes it — as for MyGo's own — and
// that the disclosure triangle and document tabs report their changes.
func TestOnChange(t *testing.T) {
	sel, open, doc := 0, false, 0
	listChanged, opened, docChanged := 0, 0, 0
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Gap(8).Children(func() {
			ListBox(c, &sel, []string{"Apples", "Pears"}).Element().Size(120, 40).OnChange(func() { listChanged++ })
			if Disclosure(c, &open, "Details").Changed() {
				opened++
			}
			tabs := DocumentTabs(c, &doc, []string{"Letter", "Notes"}, func(int) {})
			tabs.Element().Size(200, 60)
			if tabs.Changed() {
				docChanged++
			}
		})
	}
	tt := ui.NewTester(view, 260, 200)
	for _, label := range []string{"Pears", "Details", "Notes"} {
		if err := tt.Click(label); err != nil {
			t.Fatal(err)
		}
		tt.Frame()
	}
	if listChanged != 1 {
		t.Errorf("the list's OnChange ran %d times", listChanged)
	}
	if opened != 1 || !open {
		t.Errorf("the disclosure triangle reported %d changes, open %v", opened, open)
	}
	if docChanged != 1 || doc != 1 {
		t.Errorf("the document tabs reported %d changes, the document %d", docChanged, doc)
	}
}
