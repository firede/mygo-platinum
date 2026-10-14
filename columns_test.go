package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestColumnView(t *testing.T) {
	freeze(t)
	roots := sampleTree()
	var path []*TreeNode
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() {
			ColumnView(c, roots, &path).Element().Size(420, 100).Label("Browser")
		})
	}
	tt := ui.NewTester(view, 450, 130)
	x, y := findCenter(t, tt, "Documents")
	tt.ClickAt(x, y)
	tt.Frame()
	if len(path) != 1 || path[0].Title != "Documents" {
		t.Fatalf("path %v", path)
	}
	tt.Key(0, ui.KeyRight) // into Documents: Letters
	tt.Frame()
	tt.Frame()
	tt.Key(0, ui.KeyRight) // into Letters: To Ada
	tt.Frame()
	tt.Frame()
	if len(path) != 3 || path[2].Title != "To Ada" {
		t.Fatalf("Right twice: %d deep", len(path))
	}
	compareGolden(t, "columnview", tt.Image())
	tt.Key(0, ui.KeyLeft)
	tt.Frame()
	tt.Frame()
	if len(path) != 2 {
		t.Fatalf("Left left %d columns chosen", len(path))
	}
}
