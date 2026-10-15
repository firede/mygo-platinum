package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestProgressDialogAndBanner(t *testing.T) {
	freeze(t)
	working, banner := true, true
	cancelled := 0
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Fill().Padding(6).Gap(8).AlignItems(ui.Stretch).Children(func() {
			if ProgressDialog(c, &working, "Copying", "Copying 12 items to “Backup”…", 0.4, "About a minute left") {
				cancelled++
			}
			// A banner across the top of the content, which goes under it.
			Banner(c, &banner, AlertNote, "The disk was backed up.")
			TextIn(c, "28 items, 3.2 MB available", AppFont, Black)
		})
	}
	tt := ui.NewTester(view, 420, 240)
	tt.Frame()
	compareGolden(t, "progress-dialog-banner", tt.Image())
	tt.Frame()
	if got := tt.Announcements(); len(got) != 1 || got[0] != "The disk was backed up." {
		t.Fatalf("announced %q", got)
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if cancelled != 1 || working {
		t.Fatalf("Escape cancelled %d times, open %v", cancelled, working)
	}
	if err := tt.Click("Close"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if banner {
		t.Fatal("the close box did not close the banner")
	}
}

// TestSnapshotToasts draws the toasts MyGo keeps in the Platinum look: a
// note, a caution with an action, and a stop with a description.
func TestSnapshotToasts(t *testing.T) {
	freeze(t)
	added, undone := false, 0
	view := func(c *ui.Context) {
		Use(c, nil)
		if !added {
			added = true
			c.AddToast(ui.Toast{Title: "Saved"})
			c.AddToast(ui.Toast{Title: "Note deleted", Type: "caution", Action: "Undo", OnAction: func() { undone++ }})
			c.AddToast(ui.Toast{Title: "The disk is full", Description: "Empty the Wastebasket", Type: "error"})
		}
		ui.Box(c).Fill().Background(Gray5)
		Toasts(c)
	}
	tt := ui.NewTester(view, 420, 160)
	tt.Frame()
	compareGolden(t, "toasts", tt.Image())
	if err := tt.Click("Undo"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if undone != 1 {
		t.Fatalf("Undo ran %d times", undone)
	}
}
