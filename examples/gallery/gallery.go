package main

import (
	"time"

	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
)

// gallery is the state of the controls the gallery shows.
type gallery struct {
	theme      *pt.Theme
	sound      bool
	blink      bool
	mixed      bool
	size       string
	volume     float64
	speed      float64
	sort       int
	viewAs     int
	name       string
	search     string
	note       string
	font       int
	tab        int
	details    bool
	scroll     float64
	front      pt.WindowState
	back       pt.WindowState
	alertOpen  bool
	accentPick int
	// The second page.
	page          int
	bold, italic  bool
	opening       bool
	saving        bool
	fileDlg       pt.FileDialogState
	letter        *pt.RichText
	docks         [2]pt.PopupWindowState
	toggle        bool
	segment       int
	copies, level float64
	query         string
	// The third page.
	tree    []*pt.TreeNode
	treeSel *pt.TreeNode
	table   *pt.TableState
	icons   []pt.IconItem
	iconSel int
	split   float32
	scroll2 ui.ScrollState
	palette pt.WindowState
	// The fourth page.
	day   time.Time
	color ui.Color
	combo string
	sheet *pt.SheetState
	ruler pt.RulerState
	// The fifth page.
	columns    []*pt.TreeNode
	fontChoice pt.FontChoice
	pageNo     int
	docTab     int
}

func newGallery() *gallery {
	g := &gallery{
		theme:   pt.DefaultTheme(),
		bold:    true,
		toggle:  true,
		segment: 1,
		copies:  2,
		level:   4,
		sound:   true,
		size:    "medium",
		volume:  60,
		speed:   0.3,
		name:    "Untitled Folder",
		note:    "Dear diary, today the controls\nall lined up on whole pixels.",
		font:    2,
		tab:     0,
		front:   pt.WindowState{X: 10, Y: 6, W: 190, H: 104},
		back:    pt.WindowState{X: 70, Y: 50, W: 150, H: 84},
		palette: pt.WindowState{X: 330, Y: 14, W: 100, H: 60},
		iconSel: 1,
	}
	g.lists()
	g.office()
	g.fontChoice = pt.FontChoice{Family: "Serif", Size: 12, Italic: true}
	g.pageNo = 2
	g.combo = "Mono"
	g.icons = galleryIcons()
	return g
}

var fonts = []string{"Sans", "Serif", "Mono", "Script", "Display", "Rounded", "Slab", "Pixel"}

var logo = []string{
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

func (g *gallery) menus() []pt.Menu {
	return []pt.Menu{
		{Title: "About", Icon: logo, Items: []pt.MenuItem{{Title: "About the Gallery…"}}},
		{Title: "File", Items: []pt.MenuItem{
			{Title: "New Window", Key: ui.KeyN},
			{Title: "Open…", Key: ui.KeyO, Action: func() { g.opening = true }},
			{Title: "Save As…", Key: ui.KeyS, Mods: ui.Cmd | ui.Shift, Action: func() { g.saving = true }},
			{Title: "Close", Key: ui.KeyW},
		}},
		{Title: "Edit", Items: []pt.MenuItem{{Title: "Undo", Key: ui.KeyZ}}},
		{Title: "Appearance", Items: g.accentItems()},
		{Title: "Page", Items: []pt.MenuItem{
			{Title: "Classic Controls", Key: ui.Key1, Checked: g.page == 0, Action: func() { g.page = 0 }},
			{Title: "More Controls", Key: ui.Key2, Checked: g.page == 1, Action: func() { g.page = 1 }},
			{Title: "Lists and Layout", Key: ui.Key3, Checked: g.page == 2, Action: func() { g.page = 2 }},
			{Title: "Forms and Office", Key: ui.Key4, Checked: g.page == 3, Action: func() { g.page = 3 }},
			{Title: "Documents and Navigation", Key: ui.Key5, Checked: g.page == 4, Action: func() { g.page = 4 }},
		}},
	}
}

func (g *gallery) accentItems() []pt.MenuItem {
	var items []pt.MenuItem
	for _, a := range pt.Accents() {
		items = append(items, pt.MenuItem{Title: a.Name, Checked: g.theme.Accent.Name == a.Name, Action: func() { g.theme.Accent = a }})
	}
	return items
}

func (g *gallery) view(c *ui.Context) {
	pt.Use(c, g.theme)
	ui.Column(c).Fill().AlignItems(ui.Stretch).Children(func() {
		pt.MenuBar(c, g.menus())
		switch g.page {
		case 1:
			g.page2(c)
			return
		case 2:
			g.page3(c)
			return
		case 3:
			g.page4(c)
			return
		case 4:
			g.page5(c)
			return
		}
		ui.Row(c).Grow(1).Padding(12).Gap(12).AlignItems(ui.Start).Children(func() {
			ui.Column(c).Width(232).Gap(12).AlignItems(ui.Stretch).Children(func() { g.left(c) })
			ui.Column(c).Width(232).Gap(12).AlignItems(ui.Stretch).Children(func() { g.middle(c) })
			ui.Column(c).Width(232).Gap(12).AlignItems(ui.Stretch).Children(func() { g.right(c) })
		})
	})
	pt.Alert(c, &g.alertOpen, pt.AlertNote, "This is the gallery of the Platinum controls.", "Every pixel of it is drawn on a whole-pixel grid.", "OK")
	// The file dialogs, over the gallery's tree of files.
	pt.OpenDialog(c, &g.opening, &g.fileDlg, g.tree)
	if g.fileDlg.Name == "" {
		g.fileDlg.Name = "untitled"
	}
	pt.SaveDialog(c, &g.saving, &g.fileDlg, g.tree, "Save as:")
}

func (g *gallery) left(c *ui.Context) {
	pt.GroupBox(c, "Push buttons", func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			pt.Button(c, "Cancel").Cancel()
			if pt.Button(c, "OK").Default().Clicked() {
				g.alertOpen = true
			}
		})
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			pt.Button(c, "Disabled").Disabled(true)
			ui.Box(c).Size(70, 20).Draw(func(p *ui.Painter, r ui.Rect) {
				pt.DrawButton(pt.NewPen(p, r), "Pressed", pt.StatePressed)
			})
		})
	})
	pt.GroupBox(c, "Check boxes and radios", func() {
		pt.Checkbox(c, &g.sound, "Play sounds")
		pt.Checkbox(c, &g.blink, "Blink the caret")
		pt.Checkbox(c, &g.mixed, "Some of them").Mixed()
		pt.Checkbox(c, &g.blink, "Not now").Disabled(true)
		pt.Separator(c)
		pt.RadioGroup(c, func() {
			pt.Radio(c, &g.size, "small", "Small")
			pt.Radio(c, &g.size, "medium", "Medium")
			pt.Radio(c, &g.size, "large", "Large").Disabled(true)
		})
	})
	pt.GroupBox(c, "Disclosure", func() {
		pt.Disclosure(c, &g.details, "Show details")
		if g.details {
			pt.TextIn(c, "Here they are.", pt.AppFont, pt.Black)
		}
	})
	g.keyboardFocus(c)
}

// keyboardFocus shows how the controls look with the keyboard focus, which
// shows only as the keyboard moves it.
func (g *gallery) keyboardFocus(c *ui.Context) {
	pt.GroupBox(c, "Keyboard focus", func() {
		a := g.theme.Accent
		ui.Box(c).Size(206, 34).Draw(func(p *ui.Painter, r ui.Rect) {
			pn := pt.NewPen(p, r)
			b := pn.Sub(3, 4, 58, 20)
			pt.DrawButton(b, "Button", pt.StateNormal)
			pt.DrawFocusRing(b, 3, a.Dark)
			cb := pn.Sub(76, 8, 12, 12)
			pt.DrawCheckbox(cb, pt.CheckOn, pt.StateNormal)
			pt.DrawFocusRing(cb, 0, a.Dark)
			rb := pn.Sub(102, 8, 12, 12)
			pt.DrawRadio(rb, true, pt.StateNormal)
			pt.DrawFocusCircle(rb, a.Dark)
			l := pn.Sub(132, 3, 70, 30)
			pt.DrawContentRing(l)
			pt.DrawListBox(l.Sub(1, 1, l.W-2, l.H-2), []string{"Focused", "Row"}, 0, 0, true, false, a)
			pt.DrawFocusRing(l, 0, a.Dark)
		})
	})
}

func (g *gallery) middle(c *ui.Context) {
	pt.GroupBox(c, "Sliders and progress", func() {
		pt.Slider(c, &g.volume, 0, 100).Width(200).Label("Volume")
		pt.Slider(c, &g.speed, 0, 1).Width(200).Ticks(5).Step(0.25).Label("Speed")
		pt.ProgressBar(c, 0.62).Width(200).Label("Copying")
		pt.ProgressBar(c, pt.Indeterminate).Width(200).Label("Waiting")
	})
	pt.GroupBox(c, "Pop-up menus", func() {
		pt.PopupMenu(c, &g.sort, "Sort by:", []string{"Name", "Date Modified", "Size", "Kind"})
		pt.PopupMenu(c, &g.viewAs, "", []string{"as Icons", "as List"}).Disabled(true)
	})
	pt.GroupBox(c, "Text", func() {
		pt.TextField(c, &g.name).Width(200).Label("Name")
		pt.TextField(c, &g.search).Width(200).Placeholder("Search").Label("Search")
		pt.TextArea(c, &g.note).Label("Note").Element().Size(200, 74)
	})
}

func (g *gallery) right(c *ui.Context) {
	pt.GroupBox(c, "List and scroll bar", func() {
		ui.Row(c).Gap(8).Children(func() {
			pt.ListBox(c, &g.font, fonts).Element().Size(150, 82)
			pt.ScrollBar(c, &g.scroll, 20, 100).Element().Height(82).Label("Scroll")
		})
	})
	pt.TabControl(c, &g.tab, []string{"General", "Sharing", "Memory"}, func(i int) {
		pt.TextIn(c, []string{"General settings go here.", "Who may connect.", "How much to keep."}[i], pt.AppFont, pt.Black)
	}).Element().Height(84)
	ui.Box(c).Height(150).Background(pt.Gray5).Children(func() {
		pt.Window(c, &g.back, "Behind", pt.WindowOptions{}, func() {})
		pt.Window(c, &g.front, "Window", pt.WindowOptions{Active: true}, func() {
			ui.Row(c).Padding(10).Gap(10).Children(func() {
				for _, k := range []pt.AlertKind{pt.AlertNote, pt.AlertCaution, pt.AlertStop} {
					ui.Box(c).Size(32, 32).Draw(func(p *ui.Painter, r ui.Rect) { pt.DrawAlertIcon(pt.NewPen(p, r), k) })
				}
			})
		})
	})
}
