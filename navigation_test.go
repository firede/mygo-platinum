package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestPathsAndPages(t *testing.T) {
	freeze(t)
	path := []string{"Platinum Disk", "Documents", "Letters"}
	popped, crumbed, page := -1, -1, 2
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Gap(10).AlignItems(ui.Start).Children(func() {
			if i := PathPopup(c, path); i >= 0 {
				popped = i
			}
			if i := Breadcrumb(c, path); i >= 0 {
				crumbed = i
			}
			Pagination(c, &page, 5)
			Card(c, func() {
				TextIn(c, "A card of controls", AppFont, Black)
			})
			ui.Box(c).Size(300, 70).Children(func() { EmptyState(c, "No documents") })
			StatusBar(c, "3 items", "1.2 MB available").Width(300)
		})
	}
	tt := ui.NewTester(view, 340, 300)
	tt.Frame()
	compareGolden(t, "paths-pages", tt.Image())
	// The path pop-up, the first control, from the keyboard: Down opens
	// it on its folder, two more go up to the disk.
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyDown)
	tt.Frame()
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	tt.Frame()
	if popped != 0 {
		t.Fatalf("path pop-up chose %d", popped)
	}
	if err := tt.Click("Documents"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if crumbed != 1 {
		t.Fatalf("breadcrumb chose %d", crumbed)
	}
	if err := tt.Click("Next page"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if page != 3 {
		t.Fatalf("next page is %d", page)
	}
}
