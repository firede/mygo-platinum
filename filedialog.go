package platinum

import (
	"github.com/egoist/mygo/ui"
)

// FileDialogState is where an open or save dialog is, kept by the app
// between showings: the folders from the outermost down to the one it
// shows, and the name typed to save as.
type FileDialogState struct {
	// Path holds the folders opened, from one of the roots in; empty, the
	// dialog shows the roots, as the desktop. The dialog finds them again
	// by title in the tree it is given each frame.
	Path []*TreeNode
	// Name is the name to save as.
	Name string
	// NewFolder, if set, makes a folder in the one shown (nil: the
	// desktop) for the save dialog's New Folder button, and returns its
	// title, for the dialog to go into it.
	NewFolder func(in *TreeNode) string
	sel       int
	replacing bool // asking whether to replace a file of the name
	shown     bool // the save dialog has selected its name since it opened
}

// The size of a file dialog's list, and the width of its buttons.
const (
	FileListW   = 220
	FileListH   = 150
	FileButtonW = 100
	// the dialog: the list and its scroll bar, the buttons 12 from it, the
	// padding and the frame
	fileDialogW = FileListW + 12 + FileButtonW + 2*16
)

// resolve finds the folders of st.Path again in roots by their titles, so
// that the app may build its tree afresh each frame; it keeps the path as
// far as its folders are still there.
func (st *FileDialogState) resolve(roots []*TreeNode) {
	items := roots
	for i, n := range st.Path {
		var found *TreeNode
		for _, it := range items {
			if it.Title == n.Title && it.Children != nil {
				found = it
				break
			}
		}
		if found == nil {
			st.Path = st.Path[:i]
			return
		}
		st.Path[i] = found
		items = found.Children
	}
}

// folder returns the items the dialog shows: the children of the last of
// st.Path, or roots.
func (st *FileDialogState) folder(roots []*TreeNode) []*TreeNode {
	if len(st.Path) == 0 {
		return roots
	}
	return st.Path[len(st.Path)-1].Children
}

// fileModal shows a modal dialog box of the classic desktop's kind while
// *open: no title bar, a raised double frame, centered across the window a
// fifth of the way down, the controls fn builds inside it. Escape closes
// it unless its controls take it.
func fileModal(c *ui.Context, open *bool, label string, fn func()) {
	if !*open {
		return
	}
	ui.Overlay(c, func() {
		activeWithin(c, func() {
			back := ui.Box(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).Modal()
			if back.OverlayShortcut(0, ui.KeyEscape) {
				*open = false
			}
			cw, ch := c.Size()
			back.Children(func() {
				ui.Column(c).Absolute().Left(float32((round(cw) - fileDialogW) / 2)).Top(float32(round(ch) / 5)).Width(fileDialogW).
					Padding(16).Gap(8).Role(ui.RoleDialog).Label(label).
					Draw(func(p *ui.Painter, r ui.Rect) { DrawAlertFrame(NewPen(p, r)) }).
					Children(fn)
			})
		})
	})
}

// fileBody builds what both kinds of file dialog show: the folder shown as
// a path pop-up, which goes up to a folder it lists; under it, the list of
// what is in the folder, each with its icon, which a double click or
// Return goes into or opens (files dimmed and not to be chosen, when
// saving); beside them the button column — the disk and its name, Eject,
// Desktop and buttons, a line, and the bottom buttons — and under the list
// what under does. Command-Up goes up a folder. It returns the item a
// double click or Return opened that is not a folder, or nil.
func fileBody(c *ui.Context, st *FileDialogState, roots []*TreeNode, saving bool, buttons, bottom, under func()) *TreeNode {
	st.resolve(roots)
	items := st.folder(roots)
	st.sel = min(st.sel, len(items)-1)
	enter := func(n *TreeNode) {
		st.Path = append(st.Path, n)
		st.sel = 0
	}
	// The folder shown, as a path pop-up: a disk's own icon at a disk.
	names := []string{"Desktop"}
	for _, n := range st.Path {
		names = append(names, n.Title)
	}
	icon := iconFolder16
	switch len(st.Path) {
	case 0:
		icon = iconDesktop16
	case 1:
		icon = iconDisk16
	}
	pb := BevelButton(c, names[len(names)-1]).Icon(icon)
	chosen := ui.Local(pb.Element(), "path", func() int { return -1 })
	var menu []MenuItem
	for i := len(names) - 1; i >= 0; i-- {
		menu = append(menu, MenuItem{Title: names[i], Checked: i == len(names)-1, Action: func() { *chosen = i }})
	}
	pb.Menu(menu)
	if *chosen >= 0 {
		st.Path, st.sel = st.Path[:*chosen], 0
		*chosen = -1
		items = st.folder(roots)
	}
	var opened *TreeNode
	ui.Row(c).Gap(12).AlignItems(ui.Stretch).Children(func() {
		ui.Column(c).Gap(6).Children(func() {
			titles := make([]string, len(items))
			icons := make([][]string, len(items))
			dim := make([]bool, len(items))
			for i, n := range items {
				titles[i] = n.Title
				icons[i] = iconDocument16
				switch {
				case n.Children != nil && len(st.Path) == 0:
					icons[i] = iconDisk16
				case n.Children != nil:
					icons[i] = iconFolder16
				default:
					dim[i] = saving // a save dialog shows its files, not to choose
				}
			}
			if saving && st.sel >= 0 && st.sel < len(dim) && dim[st.sel] {
				st.sel = -1
			}
			lb := ListBox(c, &st.sel, titles).Icons(icons).Dim(dim).ScrollBarAlways().Label("Files")
			lb.Element().Size(FileListW, FileListH)
			if !saving {
				lb.Element().AutoFocus()
			}
			if lb.Activated() && st.sel >= 0 && st.sel < len(items) {
				if n := items[st.sel]; n.Children != nil {
					enter(n)
				} else {
					opened = n
				}
			}
			if lb.Element().Shortcut(ui.Cmd, ui.KeyUp) && len(st.Path) > 0 {
				st.Path = st.Path[:len(st.Path)-1]
				st.sel = 0
			}
			under()
		})
		ui.Column(c).Gap(8).Width(FileButtonW).AlignItems(ui.Start).Children(func() {
			// The disk the folder is on, and its name.
			disk := "Desktop"
			if len(st.Path) > 0 {
				disk = st.Path[0].Title
			}
			ui.Row(c).Gap(4).Height(16).AlignItems(ui.Center).Children(func() {
				ui.Box(c).Size(16, 16).Role(ui.RoleImage).Label("Disk").Draw(func(p *ui.Painter, r ui.Rect) {
					DrawIcon(NewPen(p, r), iconDisk16, ThemeOf(c).Accent, false)
				})
				TextIn(c, AppFont.Truncate(disk, FileButtonW-20), AppFont, Black)
			})
			Button(c, "Eject").Width(FileButtonW).Disabled(true) // nothing here ejects
			if Button(c, "Desktop").Width(FileButtonW).Disabled(len(st.Path) == 0).Clicked() {
				st.Path, st.sel = nil, 0
			}
			buttons()
			ui.Box(c).Size(FileButtonW, 1).Role(ui.RoleNone).Draw(func(p *ui.Painter, r ui.Rect) {
				pn := NewPen(p, r)
				pn.HLine(0, 0, pn.W, Gray4) // the line between the buttons
			})
			ui.Box(c).Grow(1)
			bottom()
		})
	})
	return opened
}

// defaultButton builds the default button of a file dialog, its ring
// outside the column so that its face lines up with the buttons above.
func defaultButton(c *ui.Context, label string, disabled bool) bool {
	b := Button(c, label).Width(FileButtonW).Default().Disabled(disabled)
	b.Element().Margin(0, 0, 0, -DefaultRing)
	return b.Clicked()
}

// OpenDialog shows a dialog choosing a file to open while *open, as the
// classic desktop's standard one: a modal box with no title bar; the
// folder it shows as a path pop-up; a list of what is in it, each with a
// small icon, with a scroll bar; and beside it the disk's icon and name,
// Eject, Desktop, and Cancel and Open under a line. The files are a tree:
// roots are what is on the desktop, and an item whose Children is not nil
// is a folder, which opening goes into. It returns the file the user
// opened, and closes, or nil.
func OpenDialog(c *ui.Context, open *bool, st *FileDialogState, roots []*TreeNode) *TreeNode {
	var got *TreeNode
	fileModal(c, open, "Open File", func() {
		items := st.folder(roots)
		opened := fileBody(c, st, roots, false, func() {}, func() {
			if Button(c, "Cancel").Width(FileButtonW).Cancel().Clicked() {
				*open = false
			}
			var sel *TreeNode
			if st.sel >= 0 && st.sel < len(items) {
				sel = items[st.sel]
			}
			if defaultButton(c, "Open", sel == nil) && sel != nil {
				if sel.Children != nil {
					st.Path = append(st.Path, sel)
					st.sel = 0
				} else {
					got = sel
				}
			}
		}, func() {})
		if opened != nil {
			got = opened
		}
	})
	if got != nil {
		*open = false
	}
	return got
}

// SaveDialog shows a dialog choosing where to save while *open, as the
// classic desktop's standard one: as OpenDialog's, with the files dimmed
// and not to be chosen, and New Folder among the buttons (when
// st.NewFolder is set); under the list, the prompt and a field of the name,
// selected to type over; Cancel and Save at the bottom of the buttons.
// Saving over a file of the same name asks first. It reports whether the
// user saved, and closes: st.Path's last folder (or the desktop) and
// st.Name say where.
func SaveDialog(c *ui.Context, open *bool, st *FileDialogState, roots []*TreeNode, prompt string) bool {
	saved := false
	fileModal(c, open, "Save File", func() {
		items := st.folder(roots)
		exists := false
		for _, n := range items {
			exists = exists || n.Title == st.Name && n.Children == nil
		}
		fileBody(c, st, roots, true, func() {
			if Button(c, "New Folder").Width(FileButtonW).Disabled(st.NewFolder == nil).Clicked() {
				var in *TreeNode
				if len(st.Path) > 0 {
					in = st.Path[len(st.Path)-1]
				}
				if title := st.NewFolder(in); title != "" {
					st.Path = append(st.Path, &TreeNode{Title: title, Children: []*TreeNode{}})
					st.sel = 0
				}
			}
		}, func() {
			if Button(c, "Cancel").Width(FileButtonW).Cancel().Clicked() {
				*open = false
			}
			if defaultButton(c, "Save", st.Name == "") {
				if exists {
					st.replacing = true
				} else {
					saved = true
				}
			}
		}, func() {
			Label(c, prompt)
			// The name has the keyboard, all of it selected to type over.
			f := TextField(c, &st.Name).Width(FileListW).Label(prompt)
			f.Element().AutoFocus()
			if !st.shown {
				f.Select(0, len([]rune(st.Name)))
				st.shown = true
			}
		})
	})
	if st.replacing {
		switch Alert(c, &st.replacing, AlertCaution, "Replace existing “"+st.Name+"”?", "", "Cancel", "Replace") {
		case 1:
			saved = true
		}
	}
	if saved {
		*open = false
	}
	if !*open {
		st.shown = false
	}
	return saved
}

// The 16×16 icons of the file dialogs, as rows for DrawIcon.
var (
	iconFolder16 = []string{
		"................",
		"................",
		".####...........",
		"#wwww#..........",
		"###############.",
		"#wwwwwwwwwwwww#.",
		"#w++++++++++++#.",
		"#w++++++++++++#.",
		"#w++++++++++++#.",
		"#w++++++++++++#.",
		"#w++++++++++++#.",
		"#w++++++++++++#.",
		"#+++++++++++++#.",
		"###############.",
		"................",
		"................",
	}
	iconDocument16 = []string{
		"..########......",
		"..#wwwwww##.....",
		"..#wwwwww#w#....",
		"..#wwwwww####...",
		"..#wwwwwwwww#...",
		"..#w::::::ww#...",
		"..#wwwwwwwww#...",
		"..#w:::::::w#...",
		"..#wwwwwwwww#...",
		"..#w::::::ww#...",
		"..#wwwwwwwww#...",
		"..#w:::::::w#...",
		"..#wwwwwwwww#...",
		"..#wwwwwwwww#...",
		"..###########...",
		"................",
	}
	iconDisk16 = []string{
		"................",
		"................",
		"................",
		"................",
		"................",
		"################",
		"#wwwwwwwwwwwwww#",
		"#w::::::::::::w#",
		"#wwwwwwwwwwwwww#",
		"#w#####wwwwww+w#",
		"#wwwwwwwwwwwwww#",
		"################",
		".##..........##.",
		"................",
		"................",
		"................",
	}
	iconDesktop16 = []string{
		"................",
		"..############..",
		"..#wwwwwwwwww#..",
		"..#w++++++++w#..",
		"..#w++++++++w#..",
		"..#w++++++++w#..",
		"..#w++++++++w#..",
		"..#wwwwwwwwww#..",
		"..############..",
		"......####......",
		"....########....",
		"................",
		"................",
		"................",
		"................",
		"................",
	}
)
