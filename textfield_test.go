package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestTextFieldEditing(t *testing.T) {
	freeze(t)
	name := "Hello"
	submitted := 0
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Gap(8).Children(func() {
			if TextField(c, &name).Width(160).Label("Name").Submitted() {
				submitted++
			}
		})
	}
	tt := ui.NewTester(view, 200, 60)
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !tt.Focused("Name") {
		t.Fatal("Tab did not focus the field")
	}
	tt.Key(0, ui.KeyEnd)
	tt.Type(", world")
	tt.Frame()
	if name != "Hello, world" {
		t.Fatalf("typed %q", name)
	}
	tt.Key(ui.Alt|ui.Shift, ui.KeyLeft) // select "world"
	tt.Key(ui.Cmd, ui.KeyX)
	tt.Frame()
	if name != "Hello, " || tt.Clipboard() != "world" {
		t.Fatalf("cut left %q, clipboard %q", name, tt.Clipboard())
	}
	tt.Key(ui.Cmd, ui.KeyZ)
	tt.Frame()
	if name != "Hello, world" {
		t.Fatalf("undo gave %q", name)
	}
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("Ada")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if name != "Ada" || submitted != 1 {
		t.Fatalf("got %q, submitted %d", name, submitted)
	}
	tt.Key(0, ui.KeyBackspace)
	tt.Frame()
	if name != "Ad" {
		t.Fatalf("backspace gave %q", name)
	}
	compareGolden(t, "textfield-focused", tt.Image())
}

func TestTextAreaLines(t *testing.T) {
	freeze(t)
	text := "The quick brown fox jumps over the lazy dog, again and again."
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() {
			TextArea(c, &text).Label("Note").Element().Size(150, 76)
		})
	}
	tt := ui.NewTester(view, 180, 100)
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	tt.Type("Line two")
	tt.Frame()
	tt.Frame()
	if text != "The quick brown fox jumps over the lazy dog, again and again.\nLine two" {
		t.Fatalf("text %q", text)
	}
	tt.Key(0, ui.KeyUp)
	tt.Key(ui.Shift, ui.KeyDown)
	tt.Frame()
	compareGolden(t, "textarea-focused", tt.Image())
}

func TestSnapshotTextFieldStates(t *testing.T) {
	freeze(t)
	snap(t, "textfield-states", 360, 40, func(c *ui.Context) {
		a, b, d := "Untitled Disk", "", "Read only"
		ui.Row(c).Padding(8).Gap(10).Children(func() {
			TextField(c, &a).Width(100)
			TextField(c, &b).Width(100).Placeholder("Search")
			TextField(c, &d).Width(100).Disabled(true)
		})
	})
}

func TestSearchFieldClears(t *testing.T) {
	q := ""
	changed := 0
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() {
			if SearchField(c, &q).Width(150).Label("Find").Changed() {
				changed++
			}
		})
	}
	tt := ui.NewTester(view, 200, 60)
	tt.Key(0, ui.KeyTab)
	tt.Type("kiwi")
	tt.Frame()
	if q != "kiwi" || changed == 0 {
		t.Fatalf("typed %q, changed %d", q, changed)
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if q != "" {
		t.Fatalf("Escape left %q", q)
	}
	tt.Type("fig")
	tt.Frame()
	r, _ := tt.Find("Find")
	tt.ClickAt(r.X+r.W-8, r.Y+r.H/2)
	tt.Frame()
	if q != "" {
		t.Fatalf("the clear button left %q", q)
	}
}

// TestPasswordNeverCopies tries every way out of a password field: the
// keys, the menus' edit commands and EditCommand.
func TestPasswordNeverCopies(t *testing.T) {
	secret := "hunter2"
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() {
			TextField(c, &secret).Password().Label("Password")
		})
	}
	tt := ui.NewTester(view, 200, 60)
	tt.SetClipboard("before")
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Key(ui.Cmd, ui.KeyC)
	tt.Key(ui.Cmd, ui.KeyX)
	tt.Command("copy")
	tt.Command("cut")
	tt.Frame()
	if tt.Clipboard() != "before" || secret != "hunter2" {
		t.Fatalf("clipboard %q, password %q", tt.Clipboard(), secret)
	}
}
