package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestDocumentTabs(t *testing.T) {
	freeze(t)
	docs := []string{"Letter", "Notes", "Budget"}
	sel := 1
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() {
			dt := DocumentTabs(c, &sel, docs, func(i int) {
				TextIn(c, "The text of "+docs[i], AppFont, Black)
			})
			dt.Element().Size(320, 90)
			if i := dt.Closed(); i >= 0 {
				docs = append(docs[:i:i], docs[i+1:]...)
			}
		})
	}
	tt := ui.NewTester(view, 340, 120)
	compareGolden(t, "document-tabs", tt.Image())
	// Close the first from its context menu.
	if err := tt.RightClick("Letter"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	x, y := findCenter(t, tt, "Close")
	tt.ClickAt(x, y)
	tt.Frame()
	tt.Frame()
	if len(docs) != 2 || docs[0] != "Notes" {
		t.Fatalf("docs %v", docs)
	}
}

// TestTabControlEmpty builds a tab control without tabs.
func TestTabControlEmpty(t *testing.T) {
	tab := 0
	view := func(c *ui.Context) {
		Use(c, nil)
		TabControl(c, &tab, nil, func(int) { t.Error("a page built without tabs") })
	}
	ui.NewTester(view, 200, 100).Frame()
}
