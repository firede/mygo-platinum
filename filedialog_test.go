package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func fileTree() []*TreeNode {
	letter, recipes := &TreeNode{Title: "Letter"}, &TreeNode{Title: "Recipes"}
	docs := &TreeNode{Title: "Documents", Children: []*TreeNode{letter, recipes}}
	empty := &TreeNode{Title: "Empty Folder", Children: []*TreeNode{}}
	disk := &TreeNode{Title: "Platinum Disk", Children: []*TreeNode{docs, empty, {Title: "Read Me"}}}
	return []*TreeNode{disk, {Title: "Note"}}
}

func TestOpenDialog(t *testing.T) {
	freeze(t)
	roots := fileTree()
	open := true
	st := &FileDialogState{}
	var got *TreeNode
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Box(c).Fill().Children(func() {
			if n := OpenDialog(c, &open, st, roots); n != nil {
				got = n
			}
		})
	}
	tt := ui.NewTester(view, 400, 350)
	tt.Frame()
	// Into the disk with a double click, into Documents with Return.
	x, y := findCenter(t, tt, "Platinum Disk")
	tt.ClickAt(x, y)
	tt.ClickAt(x, y)
	tt.Frame()
	if len(st.Path) != 1 {
		t.Fatalf("a double click on the disk left the path %d deep", len(st.Path))
	}
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if len(st.Path) != 2 || st.Path[1].Title != "Documents" {
		t.Fatalf("Return did not open Documents: %d deep", len(st.Path))
	}
	tt.Key(0, ui.KeyDown)
	tt.Frame()
	compareGolden(t, "open-dialog", tt.Image())
	// Command-Up goes back up, and down again.
	tt.Key(ui.Cmd, ui.KeyUp)
	tt.Frame()
	if len(st.Path) != 1 {
		t.Fatal("Command-Up did not go up")
	}
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if got == nil || got.Title != "Recipes" || open {
		t.Fatalf("opened %v, still open %v", got, open)
	}
}

func TestSaveDialog(t *testing.T) {
	freeze(t)
	roots := fileTree()
	open := true
	st := &FileDialogState{Path: []*TreeNode{roots[0], roots[0].Children[0]}, Name: "untitled"}
	saved := 0
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Box(c).Fill().Children(func() {
			if SaveDialog(c, &open, st, roots, "Save as:") {
				saved++
			}
		})
	}
	tt := ui.NewTester(view, 400, 390)
	tt.Frame()
	compareGolden(t, "save-dialog", tt.Image())
	// Typing replaces the name, all of it selected.
	tt.Type("Menu")
	tt.Frame()
	if st.Name != "Menu" {
		t.Fatalf("typing made the name %q", st.Name)
	}
	// The files are dimmed, not to be chosen; typing a file's name and
	// saving over it asks first.
	x, y := findCenter(t, tt, "Recipes")
	tt.ClickAt(x, y)
	tt.Frame()
	if st.sel >= 0 || st.Name != "Menu" {
		t.Fatalf("a click on a file chose row %d, name %q", st.sel, st.Name)
	}
	x, y = findCenter(t, tt, "Save as:")
	tt.ClickAt(x+60, y+22) // the name's field, under its prompt
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("Recipes")
	tt.Frame()
	if err := tt.Click("Save"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if saved != 0 || !st.replacing {
		t.Fatalf("saved %d times without asking", saved)
	}
	compareGolden(t, "save-dialog-replace", tt.Image())
	if err := tt.Click("Replace"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if saved != 1 || open {
		t.Fatalf("saved %d times, still open %v", saved, open)
	}
}

// TestFileDialogsFocus tabs round both dialogs: every stop has a name, and
// Tab comes back to the first.
func TestFileDialogsFocus(t *testing.T) {
	freeze(t)
	for _, save := range []bool{false, true} {
		roots := fileTree()
		open := true
		st := &FileDialogState{Name: "untitled"}
		view := func(c *ui.Context) {
			Use(c, nil)
			ui.Box(c).Fill().Children(func() {
				if save {
					SaveDialog(c, &open, st, roots, "Save as:")
				} else {
					OpenDialog(c, &open, st, roots)
				}
			})
		}
		tt := ui.NewTester(view, 400, 390)
		tt.Frame()
		first, back := "", false
		for i := 0; i < 20 && !back; i++ {
			tt.Key(0, ui.KeyTab)
			tt.Frame()
			named := ""
			for _, s := range tt.Texts() {
				if tt.Focused(s) {
					named = s
					break
				}
			}
			if named == "" {
				t.Errorf("save %v: focus stop %d has no name", save, i+1)
			}
			if i == 0 {
				first = named
			} else if named == first {
				back = true
			}
		}
		if !back {
			t.Errorf("save %v: Tab did not come back to %q", save, first)
		}
	}
}
