package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestFormulas(t *testing.T) {
	st := NewSheetState()
	st.Cells[Cell{0, 0}] = "2"
	st.Cells[Cell{0, 1}] = "3.5"
	st.Cells[Cell{0, 2}] = "=A1*A2+1"
	st.Cells[Cell{1, 0}] = "=SUM(A1:A3)"
	st.Cells[Cell{1, 1}] = "=AVERAGE(A1:A2)"
	st.Cells[Cell{1, 2}] = "=(A1+1)/0"
	st.Cells[Cell{2, 0}] = "=C1"
	st.Cells[Cell{2, 1}] = "=MAX(A1, 10, A2)"
	st.Cells[Cell{2, 2}] = "=A1+"
	st.Cells[Cell{3, 0}] = "text"
	st.Cells[Cell{3, 1}] = "=D1+1"
	for cell, want := range map[string]string{
		"A3": "8", "B1": "13.5", "B2": "2.75", "B3": "#DIV/0!", "C1": "#CYCLE!",
		"C2": "10", "C3": "#SYNTAX!", "D1": "text", "D2": "1",
	} {
		c, _ := ParseCell(cell)
		if got := st.Value(c); got != want {
			t.Errorf("%s = %q, want %q", cell, got, want)
		}
	}
	if (Cell{27, 9}).String() != "AB10" {
		t.Errorf("column 27 is %s", Cell{27, 9})
	}
}

func TestSheetEditing(t *testing.T) {
	freeze(t)
	st := NewSheetState()
	st.Cells[Cell{0, 0}] = "12"
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(8).AlignItems(ui.Stretch).Children(func() {
			FormulaBar(c, st)
			Sheet(c, st, 50, 8).Element().Size(330, 140)
		})
	}
	tt := ui.NewTester(view, 350, 200)
	r, _ := tt.Find("Sheet")
	tt.ClickAt(r.X+2+SheetRowHeadW+10, r.Y+2+SheetCellH+SheetCellH+5) // A2
	tt.Frame()
	if st.Cursor != (Cell{0, 1}) {
		t.Fatalf("clicked %v", st.Cursor)
	}
	tt.Type("30")
	tt.Key(0, ui.KeyEnter)
	tt.Type("=SUM(A1:A2)")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if st.Value(Cell{0, 2}) != "42" {
		t.Fatalf("A3 is %q", st.Value(Cell{0, 2}))
	}
	tt.Key(0, ui.KeyUp)
	tt.Key(ui.Shift, ui.KeyRight)
	tt.Frame()
	compareGolden(t, "sheet", tt.Image())
	tt.Key(0, ui.KeyBackspace)
	tt.Frame()
	if st.Cells[Cell{0, 2}] != "" {
		t.Fatal("Delete did not clear the range")
	}
}

// TestSheetTabLeaves tabs across a sheet and out of it to the next control,
// and back in and out the other way: the sheet keeps Tab only while there
// is a cell to move to.
func TestSheetTabLeaves(t *testing.T) {
	freeze(t)
	st := NewSheetState()
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(8).Gap(8).AlignItems(ui.Start).Children(func() {
			Button(c, "Before")
			Sheet(c, st, 4, 3).Label("Sheet").Element().Size(200, 80)
			Button(c, "After")
		})
	}
	tt := ui.NewTester(view, 240, 180)
	tt.Frame()
	for i := 0; i < 6 && !tt.Focused("Sheet"); i++ {
		tt.Key(0, ui.KeyTab)
	}
	tt.Type("7")
	for i := 0; i < 3; i++ {
		tt.Key(0, ui.KeyTab)
		tt.Frame()
	}
	if !tt.Focused("After") || st.Cells[Cell{0, 0}] != "7" {
		t.Fatalf("after three tabs, the focus is not on After, or A1 is %q", st.Cells[Cell{0, 0}])
	}
	tt.Key(ui.Shift, ui.KeyTab) // back into the sheet, at C1
	tt.Key(ui.Shift, ui.KeyTab)
	tt.Key(ui.Shift, ui.KeyTab)
	tt.Key(ui.Shift, ui.KeyTab)
	tt.Frame()
	if !tt.Focused("Before") {
		t.Fatal("Shift-Tab from the first column did not leave the sheet")
	}
}
