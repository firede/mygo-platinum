package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestTreeViewKeys(t *testing.T) {
	roots := sampleTree()
	var sel *TreeNode
	activated := 0
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() {
			tv := TreeView(c, roots, &sel)
			tv.Element().Size(180, 100).Label("Files")
			if tv.Submitted() {
				activated++
			}
		})
	}
	tt := ui.NewTester(view, 220, 130)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyDown) // Documents
	tt.Key(0, ui.KeyDown) // Letters
	tt.Frame()
	if sel == nil || sel.Title != "Letters" {
		t.Fatalf("Down twice chose %v", sel)
	}
	tt.Key(0, ui.KeyRight) // open Letters
	tt.Frame()
	tt.Key(0, ui.KeyRight) // into it
	tt.Frame()
	if sel.Title != "To Ada" {
		t.Fatalf("Right twice chose %q", sel.Title)
	}
	tt.Key(0, ui.KeyLeft) // to Letters
	tt.Key(0, ui.KeyLeft) // close it
	tt.Frame()
	if sel.Title != "Letters" || roots[0].Children[0].Open {
		t.Fatalf("Left twice: %q, open %v", sel.Title, roots[0].Children[0].Open)
	}
	tt.TypeKey(0, ui.KeyP, "p")
	tt.Frame()
	if sel.Title != "Pictures" {
		t.Fatalf("typing p chose %q", sel.Title)
	}
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if activated != 1 {
		t.Fatalf("Return activated %d times", activated)
	}
	// A click on a triangle opens its item.
	r, _ := tt.Find("Pictures")
	tt.ClickAt(r.X+treeTriLeft+6, r.Y+8)
	tt.Frame()
	if !roots[1].Open {
		t.Fatal("a click on the triangle did not open Pictures")
	}
	compareGolden(t, "treeview", tt.Image())
}
