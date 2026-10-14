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
		ui.Box(c).Fill().Children(func() {
			if ProgressDialog(c, &working, "Copying", "Copying 12 items to “Backup”…", 0.4, "About a minute left") {
				cancelled++
			}
			Banner(c, &banner, AlertNote, "The disk was backed up.")
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
