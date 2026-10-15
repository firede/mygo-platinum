package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

type menuApp struct {
	opened, saved int
	wrap          bool
}

var appIcon = []string{
	"....##....",
	"...#oo#...",
	"..#oooo#..",
	".#oo++oo#.",
	"#oo+ww+oo#",
	"#oo+ww+oo#",
	".#oo++oo#.",
	"..#oooo#..",
	"...#oo#...",
	"....##....",
}

func (a *menuApp) menus() []Menu {
	return []Menu{
		{Title: "About", Icon: appIcon, Items: []MenuItem{{Title: "About This Desk…"}}},
		{Title: "File", Items: []MenuItem{
			{Title: "New", Key: ui.KeyN, Action: func() { a.opened++ }},
			{Title: "Open…", Key: ui.KeyO},
			SeparatorItem(),
			{Title: "Save", Key: ui.KeyS, Action: func() { a.saved++ }},
			{Title: "Print…", Disabled: true},
			{Title: "Recent", Submenu: []MenuItem{{Title: "Notes.txt"}, {Title: "Letter.txt"}}},
		}},
		{Title: "Edit", Items: []MenuItem{
			{Title: "Undo", Key: ui.KeyZ},
			SeparatorItem(),
			{Title: "Wrap Lines", Checked: a.wrap, Action: func() { a.wrap = !a.wrap }},
		}},
	}
}

func (a *menuApp) view(c *ui.Context) {
	Use(c, nil)
	ui.Column(c).Fill().Children(func() {
		MenuBar(c, a.menus())
	})
}

func TestMenuBarClickChoose(t *testing.T) {
	freeze(t)
	a := &menuApp{}
	tt := ui.NewTester(a.view, 300, 200)
	r, _ := tt.Find("File")
	tt.ClickAt(r.X+5, r.Y+5)
	tt.Frame()
	n, ok := tt.Find("New")
	if !ok {
		t.Fatalf("New not shown; texts %q", tt.Texts())
	}
	tt.ClickAt(n.X+5, n.Y+5)
	tt.Frame()
	tt.Frame()
	if a.opened != 1 {
		t.Fatalf("New chosen %d times", a.opened)
	}
	if _, ok := tt.Find("Save"); ok {
		t.Fatal("menu still open after a choice")
	}
}

func TestMenuBarDragChoose(t *testing.T) {
	freeze(t)
	a := &menuApp{}
	tt := ui.NewTester(a.view, 300, 200)
	r, _ := tt.Find("File")
	tt.Press(r.X+5, r.Y+5)
	tt.Frame()
	s, ok := tt.Find("Save")
	if !ok {
		t.Fatal("menu did not open on press")
	}
	tt.Move(s.X+5, s.Y+5)
	tt.Release(s.X+5, s.Y+5)
	tt.Frame()
	tt.Frame()
	if a.saved != 1 {
		t.Fatalf("Save chosen %d times", a.saved)
	}
}

func TestMenuBarKeyboard(t *testing.T) {
	freeze(t)
	a := &menuApp{}
	tt := ui.NewTester(a.view, 300, 200)
	tt.Key(ui.Cmd, ui.KeyS)
	tt.Frame()
	if a.saved != 1 {
		t.Fatalf("Cmd-S chose Save %d times", a.saved)
	}
	tt.Key(ui.Ctrl, ui.KeyF2) // the first menu
	tt.Frame()
	tt.Key(0, ui.KeyRight) // File, New highlighted
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	tt.Frame()
	if a.opened != 1 {
		t.Fatalf("keyboard chose New %d times", a.opened)
	}
	tt.Key(ui.Ctrl, ui.KeyF2)
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if _, ok := tt.Find("About This Desk…"); ok {
		t.Fatal("Escape did not close the menu")
	}
}

func TestSnapshotMenuOpen(t *testing.T) {
	freeze(t)
	a := &menuApp{wrap: true}
	tt := ui.NewTester(a.view, 260, 170)
	r, _ := tt.Find("File")
	tt.Press(r.X+5, r.Y+5)
	tt.Frame()
	s, _ := tt.Find("Recent")
	tt.Move(s.X+5, s.Y+5)
	tt.Frame()
	compareGolden(t, "menubar-open", tt.Image())
}

// TestMenuBarAccessibility presses a title and an item as assistive
// technology does, through their elements rather than the pointer.
func TestMenuBarAccessibility(t *testing.T) {
	freeze(t)
	a := &menuApp{}
	tt := ui.NewTester(a.view, 300, 200)
	if err := tt.Click("File"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if _, ok := tt.Find("Save"); !ok {
		t.Fatalf("pressing the title opened no menu: %q", tt.Texts())
	}
	tt.Frame()
	if err := tt.Click("Save"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	tt.Frame()
	if a.saved != 1 {
		t.Fatalf("pressing Save saved %d times", a.saved)
	}
}

// TestMenuBarEmptyMenu presses a title with no items, as a clock in the
// menu bar: nothing opens, nothing breaks, and its action runs.
func TestMenuBarEmptyMenu(t *testing.T) {
	freeze(t)
	clicks := 0
	menus := []Menu{{Title: "File", Items: []MenuItem{{Title: "Open"}}}, {Title: "9:41 AM", Right: true, Action: func() { clicks++ }}}
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Fill().Children(func() { MenuBar(c, menus) })
	}
	tt := ui.NewTester(view, 300, 100)
	tt.Frame()
	if err := tt.Click("9:41 AM"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if tt.HasText("Open") {
		t.Fatal("a menu opened")
	}
	if clicks != 1 {
		t.Fatalf("the title's action ran %d times", clicks)
	}
}

func TestContextMenu(t *testing.T) {
	freeze(t)
	renamed := 0
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() {
			row := ui.Box(c).Size(120, 30).Focusable().Label("Row")
			ContextMenu(c, row, []MenuItem{{Title: "Rename", Action: func() { renamed++ }}, {Title: "Delete", Disabled: true}})
		})
	}
	tt := ui.NewTester(view, 260, 160)
	if err := tt.RightClick("Row"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	x, y := findCenter(t, tt, "Rename")
	compareGolden(t, "context-menu", tt.Image())
	tt.ClickAt(x, y)
	tt.Frame()
	tt.Frame()
	if renamed != 1 {
		t.Fatalf("Rename ran %d times", renamed)
	}
	// From the keyboard.
	if err := tt.Click("Row"); err != nil {
		t.Fatal(err)
	}
	tt.Key(ui.Shift, ui.KeyF10)
	tt.Frame()
	if _, ok := tt.Find("Rename"); !ok {
		t.Fatalf("Shift-F10 did not open the menu; focused %v", tt.Focused("Row"))
	}
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	tt.Frame()
	if renamed != 2 {
		t.Fatalf("Shift-F10 and Return ran Rename %d times", renamed)
	}
}

// TestShortcutText checks the keys of shortcuts as menus write them: by
// the symbols of a Mac's keyboard, or by name, as other systems do.
func TestShortcutText(t *testing.T) {
	t.Cleanup(func() { MacKeys = true })
	items := []MenuItem{
		{Key: ui.KeyN},
		{Key: ui.KeyS, Mods: ui.Cmd | ui.Shift},
		{Key: ui.KeyBackspace},
		{Key: ui.KeyEnter, Mods: ui.Cmd | ui.Alt},
		{},
	}
	for _, tc := range []struct {
		mac  bool
		want []string
	}{
		{true, []string{"⌘N", "↑⌘S", "⌘⌫", "~⌘↵", ""}},
		{false, []string{"Ctrl+N", "Ctrl+Shift+S", "Ctrl+Backspace", "Ctrl+Alt+Enter", ""}},
	} {
		MacKeys = tc.mac
		for i, it := range items {
			if got := it.shortcutText(); got != tc.want[i] {
				t.Errorf("MacKeys %v: item %d shows %q, want %q", tc.mac, i, got, tc.want[i])
			}
		}
	}
}
