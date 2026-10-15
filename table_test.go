package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestTableSortSelectResize(t *testing.T) {
	st := NewTableState()
	var order []int
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() {
			tw := Table(c, st, produceCols, len(produce), produceCell)
			tw.Element().Size(240, 130).Label("Produce")
			order = tw.Order()
		})
	}
	tt := ui.NewTester(view, 270, 160)
	// Sort by price, then turn the order over.
	r, _ := tt.Find("Price")
	tt.ClickAt(r.X+20, r.Y+8)
	tt.Frame()
	if produce[order[0]].name != "Banana" {
		t.Fatalf("sorted by price, first is %s", produce[order[0]].name)
	}
	tt.ClickAt(r.X+20, r.Y+8)
	tt.Frame()
	if produce[order[0]].name != "Date" || !st.Descending {
		t.Fatalf("descending, first is %s", produce[order[0]].name)
	}
	// Select a row, extend with Shift-Down, add one with Command-click.
	first, _ := tt.Find(produce[order[0]].name)
	tt.ClickAt(first.X+5, first.Y+5)
	tt.Frame()
	tt.Key(ui.Shift, ui.KeyDown)
	tt.Frame()
	if len(st.Selected) != 2 {
		t.Fatalf("Shift-Down selected %d", len(st.Selected))
	}
	fifth, _ := tt.Find(produce[order[4]].name)
	tt.ClickAtWith(ui.Cmd, fifth.X+5, fifth.Y+5)
	tt.Frame()
	if len(st.Selected) != 3 || !st.Selected[order[4]] {
		t.Fatalf("Command-click: %v", st.Selected)
	}
	// Resize the Name column by its right edge.
	n, _ := tt.Find("Name")
	edge := n.X + n.W
	tt.Press(edge-1, n.Y+8)
	tt.Move(edge+19, n.Y+8)
	tt.Release(edge+19, n.Y+8)
	tt.Frame()
	if st.Widths[0] != 100 {
		t.Fatalf("Name is %d wide", st.Widths[0])
	}
	compareGolden(t, "table", tt.Image())
}

// TestTableLess sorts by a column of its own order.
func TestTableLess(t *testing.T) {
	sizes := []string{"2 K", "300 bytes", "1 MB"}
	bytes := []int{2048, 300, 1 << 20}
	st := NewTableState()
	st.SortColumn = 0
	var order []int
	view := func(c *ui.Context) {
		Use(c, nil)
		cols := []TableColumn{{Title: "Size", Less: func(a, b int) bool { return bytes[a] < bytes[b] }}}
		order = Table(c, st, cols, 3, func(r, _ int) string { return sizes[r] }).Order()
	}
	ui.NewTester(view, 200, 100).Frame()
	if order[0] != 1 || order[1] != 0 || order[2] != 2 {
		t.Fatalf("sorted %v", order)
	}
}

// TestTableShrinks selects with Shift after the model lost rows.
func TestTableShrinks(t *testing.T) {
	rows := 9
	st := NewTableState()
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() {
			Table(c, st, produceCols, rows, produceCell).Element().Size(240, 130).Label("Produce")
		})
	}
	tt := ui.NewTester(view, 270, 160)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyEnd)
	tt.Frame()
	rows = 3
	tt.Frame()
	tt.Key(ui.Shift, ui.KeyUp)
	tt.Frame()
	for r := range st.Selected {
		if r >= rows {
			t.Fatalf("row %d selected of %d", r, rows)
		}
	}
}
