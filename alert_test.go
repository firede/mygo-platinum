package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestAlertButtons(t *testing.T) {
	freeze(t)
	open := true
	got := -2
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Box(c).Fill().Children(func() {
			if i := Alert(c, &open, AlertCaution, "Save changes to the document “Notes” before closing?",
				"If you don’t save, your changes will be lost.", "Don’t Save", "Cancel", "Save"); i >= 0 {
				got = i
			}
		})
	}
	tt := ui.NewTester(view, 400, 200)
	compareGolden(t, "alert-caution", tt.Image())
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if got != 1 || open {
		t.Fatalf("Escape chose %d, open %v", got, open)
	}
	open = true
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if got != 2 || open {
		t.Fatalf("Return chose %d", got)
	}
	open = true
	tt.Frame()
	if err := tt.Click("Don’t Save"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if got != 0 {
		t.Fatalf("click chose %d", got)
	}
}

func TestSnapshotAlertIcons(t *testing.T) {
	snap(t, "alert-icons", 140, 44, func(c *ui.Context) {
		board(c, 32, 32,
			func(pn Pen) { DrawAlertIcon(pn, AlertNote) },
			func(pn Pen) { DrawAlertIcon(pn, AlertCaution) },
			func(pn Pen) { DrawAlertIcon(pn, AlertStop) },
		)
	})
}

func TestDialogMovable(t *testing.T) {
	freeze(t)
	open := true
	name := ""
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Box(c).Fill().Children(func() {
			Dialog(c, &open, "Rename", 260, func() {
				Label(c, "New name:")
				TextField(c, &name).Width(220).Label("New name")
				ui.Row(c).Gap(10).Justify(ui.End).Children(func() {
					if Button(c, "Cancel").Cancel().Clicked() {
						open = false
					}
					if Button(c, "Rename").Default().Clicked() {
						open = false
					}
				})
			})
		})
	}
	tt := ui.NewTester(view, 360, 240)
	tt.Frame()
	compareGolden(t, "dialog-modal", tt.Image())
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if open {
		t.Fatal("Escape did not close the dialog")
	}
}
