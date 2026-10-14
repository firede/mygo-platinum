package platinum_test

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
)

// The examples build views as an app does, and render them without a
// window; they are compiled with the tests, so they keep up with the API.

type app struct {
	theme                *pt.Theme
	sound, details, bold bool
	size                 string
	volume, copies       float64
	sort, tab, font      int
	name, query, note    string
	confirm, find        bool
	frame                pt.WindowState
	split                float64
	scroll               ui.ScrollState
	table                *pt.TableState
	folders              []*pt.TreeNode
	folder               *pt.TreeNode
	icons                []pt.IconItem
	icon                 int
}

func render(view func(c *ui.Context)) { ui.Render(view, 400, 300, 1) }

func Example() {
	a := &app{}
	render(func(c *ui.Context) {
		pt.Use(c, a.theme) // nil is the default theme
		ui.Column(c).Fill().Padding(12).Gap(8).Children(func() {
			pt.Checkbox(c, &a.sound, "Play sounds")
			if pt.Button(c, "OK").Default().Clicked() {
				fmt.Println("OK")
			}
		})
	})
	// Output:
}

func ExampleButton() {
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		ui.Row(c).Gap(10).Children(func() {
			if pt.Button(c, "Cancel").Cancel().Clicked() {
				fmt.Println("cancelled")
			}
			if pt.Button(c, "Save").Default().Clicked() {
				fmt.Println("saved")
			}
		})
	})
	// Output:
}

func ExampleCheckbox() {
	a := &app{}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.Checkbox(c, &a.sound, "Play alert sounds")
		pt.Checkbox(c, &a.bold, "Some of them").Mixed()
	})
	// Output:
}

func ExampleRadio() {
	a := &app{size: "small"}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.RadioGroup(c, func() {
			pt.Radio(c, &a.size, "small", "Small")
			pt.Radio(c, &a.size, "large", "Large")
		})
	})
	// Output:
}

func ExampleSlider() {
	a := &app{}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.Slider(c, &a.volume, 0, 100).Width(160).Ticks(5).Label("Volume")
	})
	// Output:
}

func ExampleProgressBar() {
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.ProgressBar(c, 0.4).Width(200)
		pt.ProgressBar(c, pt.Indeterminate).Width(200)
	})
	// Output:
}

func ExampleDisclosure() {
	a := &app{}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.Disclosure(c, &a.details, "Details")
		if a.details {
			pt.TextIn(c, "More about it.", pt.AppFont, pt.Black)
		}
	})
	// Output:
}

func ExampleGroupBox() {
	a := &app{}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.GroupBox(c, "Sound", func() {
			pt.Checkbox(c, &a.sound, "Beep on errors")
			pt.Separator(c)
		})
	})
	// Output:
}

func ExampleTabControl() {
	a := &app{}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.TabControl(c, &a.tab, []string{"General", "Sharing"}, func(i int) {
			pt.Label(c, []string{"General settings", "Who may connect"}[i])
		}).Element().Size(300, 120)
	})
	// Output:
}

func ExamplePopupMenu() {
	a := &app{}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		if pt.PopupMenu(c, &a.sort, "Sort by:", []string{"Name", "Date", "-", "Size"}).Changed() {
			fmt.Println("sorted")
		}
	})
	// Output:
}

func ExampleTextField() {
	a := &app{}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		if pt.TextField(c, &a.name).Width(180).Placeholder("Name").Label("Name").Submitted() {
			fmt.Println("renamed to", a.name)
		}
		pt.TextField(c, &a.query).Width(180).Password().Label("Password")
		pt.TextArea(c, &a.note).Label("Note").Element().Size(300, 120)
	})
	// Output:
}

func ExampleListBox() {
	a := &app{}
	fonts := []string{"Sans", "Serif", "Mono"}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		lb := pt.ListBox(c, &a.font, fonts)
		lb.Element().Size(160, 96).Label("Fonts")
		if lb.Activated() {
			fmt.Println("chose", fonts[a.font])
		}
	})
	// Output:
}

func ExampleScrollBar() {
	top := 0.0
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.ScrollBar(c, &top, 10, 100).Step(1).Element().Height(160).Label("Rows")
	})
	// Output:
}

func ExampleWindow() {
	a := &app{frame: pt.WindowState{X: 20, Y: 20, W: 240, H: 160}}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		ui.Box(c).Fill().Children(func() {
			w := pt.Window(c, &a.frame, "Notes", pt.WindowOptions{Active: true}, func() {
				pt.Label(c, "Inside the window")
			})
			if w.Closed() {
				fmt.Println("closed")
			}
		})
	})
	// Output:
}

func ExampleMenuBar() {
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		ui.Column(c).Fill().Children(func() {
			pt.MenuBar(c, []pt.Menu{
				{Title: "File", Items: []pt.MenuItem{
					{Title: "New", Key: ui.KeyN, Action: func() { fmt.Println("new") }},
					pt.Divider(),
					{Title: "Recent", Submenu: []pt.MenuItem{{Title: "Notes"}}},
					{Title: "Print…", Disabled: true},
				}},
			})
		})
	})
	// Output:
}

func ExampleAlert() {
	a := &app{confirm: true}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		switch pt.Alert(c, &a.confirm, pt.AlertCaution,
			"Save changes to “Notes” before closing?", "Your changes will be lost if you don’t save them.",
			"Don’t Save", "Cancel", "Save") {
		case 0:
			fmt.Println("discarded")
		case 2:
			fmt.Println("saved")
		}
	})
	// Output:
}

func ExampleDialog() {
	a := &app{find: true}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.Dialog(c, &a.find, "Find", 320, func() {
			pt.TextField(c, &a.query).Width(280).Label("Find")
			if pt.Button(c, "Find").Default().Clicked() {
				fmt.Println("finding", a.query)
			}
		})
	})
	// Output:
}

var boldIcon = []string{
	"######..",
	".##..##.",
	".#####..",
	".##..##.",
	"######..",
}

func ExampleBevelButton() {
	a := &app{}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.BevelButton(c, "Bold").Icon(boldIcon).Toggle(&a.bold)
		pt.BevelButton(c, "Style").Menu([]pt.MenuItem{{Title: "Plain"}, {Title: "Bold"}})
		pt.IconButton(c, boldIcon, "Bold")
		pt.Segmented(c, &a.tab, []string{"Left", "Center", "Right"})
	})
	// Output:
}

func ExampleToolbar() {
	a := &app{}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.Toolbar(c, []pt.ToolItem{
			{Label: "New", Action: func() { fmt.Println("new") }},
			{Separator: true},
			{Label: "Bold", Icon: boldIcon, IconOnly: true, On: &a.bold},
		})
	})
	// Output:
}

func ExampleNumberField() {
	a := &app{copies: 1}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.NumberField(c, &a.copies, 1, 99, 1, 0).Label("Copies")
		pt.Stepper(c, &a.volume, 0, 10, 1).Label("Level")
		pt.SearchField(c, &a.query).Width(160).Label("Search")
	})
	// Output:
}

func ExampleChasingArrows() {
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.ChasingArrows(c).Label("Connecting")
		pt.Placard(c, "12 items")
		pt.Badge(c, "3")
		pt.Tag(c, "draft")
		pt.ImageWell(c, nil).Label("Picture")
	})
	// Output:
}

func ExampleBalloon() {
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		b := pt.Button(c, "Erase")
		pt.Balloon(c, b.Element(), "Erases the disk. You cannot undo it.")
	})
	// Output:
}

func ExampleContextMenu() {
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		row := ui.Box(c).Size(200, 20).Focusable().Label("Row")
		pt.ContextMenu(c, row, []pt.MenuItem{{Title: "Rename", Action: func() { fmt.Println("rename") }}})
	})
	// Output:
}

func ExampleTreeView() {
	a := &app{folders: []*pt.TreeNode{
		{Title: "Documents", Open: true, Children: []*pt.TreeNode{{Title: "Letters"}}},
		{Title: "Pictures"},
	}}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.TreeView(c, a.folders, &a.folder).Element().Size(200, 120).Label("Folders")
	})
	// Output:
}

func ExampleDrawTree() {
	roots := []*pt.TreeNode{{Title: "Documents", Open: true, Children: []*pt.TreeNode{{Title: "Letters"}}}}
	render(func(c *ui.Context) {
		ui.Box(c).Size(160, 60).Draw(func(p *ui.Painter, r ui.Rect) {
			pt.DrawTree(pt.NewPen(p, r), roots, 0, roots[0].Children[0], true, pt.Lavender)
		})
	})
	// Output:
}

func ExampleTable() {
	a := &app{table: pt.NewTableState()}
	rows := [][]string{{"Kiwi", "0.50"}, {"Fig", "2.50"}}
	cols := []pt.Column{{Title: "Name", Width: 100}, {Title: "Price", Width: 60, Right: true}}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		t := pt.Table(c, a.table, cols, len(rows), func(r, col int) string { return rows[r][col] })
		t.Element().Size(200, 100).Label("Prices")
		if t.Changed() {
			fmt.Println(len(a.table.Selected), "selected")
		}
	})
	// Output:
}

func ExampleIconView() {
	a := &app{icons: []pt.IconItem{{Name: "Alpha"}, {Name: "Beta"}}, icon: -1}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		if pt.IconView(c, a.icons, &a.icon).Activated() {
			fmt.Println("open", a.icons[a.icon].Name)
		}
	})
	// Output:
}

func ExampleSplit() {
	a := &app{split: 120}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.Split(c, &a.split, false, func() {
			pt.Label(c, "Folders")
		}, func() {
			pt.ScrollView(c, &a.scroll, func() {
				for i := 0; i < 30; i++ {
					pt.TextIn(c, fmt.Sprint("Line ", i), pt.AppFont, pt.Black)
				}
			}).Grow(1)
		}).Size(360, 200)
	})
	// Output:
}

func ExampleEditCommand() {
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		ui.Column(c).Fill().Children(func() {
			pt.MenuBar(c, []pt.Menu{{Title: "Edit", Items: []pt.MenuItem{
				{Title: "Copy", Key: ui.KeyC, Action: func() { pt.EditCommand(c, "copy") }},
				{Title: "Paste", Key: ui.KeyV, Action: func() { pt.EditCommand(c, "paste") }},
			}}})
		})
	})
	// Output:
}

func ExampleNewPen() {
	render(func(c *ui.Context) {
		ui.Box(c).Size(60, 24).Draw(func(p *ui.Painter, r ui.Rect) {
			pn := pt.NewPen(p, r)
			pt.DrawButton(pn, "Drawn", pt.StateNormal)
		})
	})
	// Output:
}

func ExampleDatePicker() {
	day := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	at := day
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		if pt.DatePicker(c, &day).Changed() {
			fmt.Println("due", day.Format("Jan 2"))
		}
		pt.Calendar(c, &day)
		pt.TimeField(c, &at, 5)
	})
	// Output:
}

func ExampleComboBox() {
	font := ""
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.ComboBox(c, &font, []string{"Sans", "Serif", "Mono"}).Label("Font")
	})
	// Output:
}

func ExampleColorWell() {
	col := pt.Lavender.Mid
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		if pt.ColorWell(c, &col, "Highlight color").Changed() {
			fmt.Println("now", col)
		}
	})
	// Output:
}

func ExampleSheet() {
	st := pt.NewSheetState()
	st.Cells[pt.Cell{Col: 0, Row: 0}] = "12"
	st.Cells[pt.Cell{Col: 0, Row: 1}] = "=A1*2"
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		ui.Column(c).Fill().Children(func() {
			pt.FormulaBar(c, st)
			pt.Sheet(c, st, 100, 26).Element().Grow(1).FillWidth()
		})
	})
	fmt.Println(st.Value(pt.Cell{Col: 0, Row: 1}))
	// Output: 24
}

func ExampleFindReplace() {
	text := "One cat, two cats."
	open := true
	st := &pt.FindState{Query: "cat", With: "dog"}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		area := pt.TextArea(c, &text)
		switch pt.FindReplace(c, &open, st) {
		case pt.FindNext:
			_, from := area.Selection()
			if a, b := pt.Find(text, st.Query, from, st.Options); a >= 0 {
				area.Select(a, b)
			}
		case pt.ReplaceEvery:
			var n int
			text, n = pt.ReplaceAll(text, st.Query, st.With, st.Options)
			st.Note = fmt.Sprint(n, " replaced")
		}
	})
	fmt.Println(pt.ReplaceAll(text, "cat", "dog", pt.FindOptions{WholeWord: true}))
	// Output: One dog, two cats. 1
}

func ExampleRuler() {
	margins := &pt.RulerState{Left: 16, Right: 300}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.Ruler(c, margins)
	})
	// Output:
}

func ExampleProgressDialog() {
	copying, done := true, true
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		if pt.ProgressDialog(c, &copying, "Copying", "Copying 12 items…", 0.4, "About a minute left") {
			fmt.Println("stopped")
		}
		pt.Banner(c, &done, pt.AlertNote, "The disk was backed up.")
	})
	// Output:
}

func ExamplePathPopup() {
	path := []string{"Disk", "Documents", "Letters"}
	page := 0
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		if i := pt.PathPopup(c, path); i >= 0 {
			path = path[:i+1]
		}
		if i := pt.Breadcrumb(c, path); i >= 0 {
			path = path[:i+1]
		}
		pt.Pagination(c, &page, 12)
		pt.Card(c, func() { pt.Label(c, "In a card") })
		pt.EmptyState(c, "No documents")
		pt.StatusBar(c, "3 items", "1.2 MB available")
	})
	// Output:
}

func ExampleColumnView() {
	roots := []*pt.TreeNode{{Title: "Documents", Children: []*pt.TreeNode{{Title: "Letters"}}}, {Title: "Read Me"}}
	var path []*pt.TreeNode
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.ColumnView(c, roots, &path).Element().Size(420, 160).Label("Browser")
	})
	// Output:
}

func ExampleDocumentTabs() {
	docs := []string{"Letter", "Notes"}
	sel := 0
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		tabs := pt.DocumentTabs(c, &sel, docs, func(i int) {
			pt.TextIn(c, "The text of "+docs[i], pt.AppFont, pt.Black)
		})
		if i := tabs.Closed(); i >= 0 {
			docs = append(docs[:i:i], docs[i+1:]...)
		}
	})
	// Output:
}

func ExampleFontPicker() {
	choice := pt.FontChoice{Family: "Serif", Size: 12}
	page := pt.PageSetupState{Scale: 100, Margin: 36}
	setup := false
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.FontPicker(c, &choice, []string{"Sans", "Serif", "Mono"})
		if pt.PageSetup(c, &setup, &page) {
			fmt.Println("paper", pt.Papers[page.Paper].Name)
		}
	})
	// Output:
}

func ExampleZoomRects() {
	opened := time.Date(2026, 10, 14, 9, 41, 0, 0, time.UTC) // when the window opened
	icon := ui.Rect{X: 600, Y: 40, W: 32, H: 32}
	window := ui.Rect{X: 40, Y: 40, W: 300, H: 200}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.ZoomRects(c, icon, window, opened)
	})
	// Output:
}

func ExampleRichTextArea() {
	doc := pt.NewRichText("")
	doc.SetRuns([]pt.Run{{Text: "Dear "}, {Text: "Ada", Style: pt.Bold}, {Text: ","}})
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		ui.Column(c).Fill().Children(func() {
			pt.FormatBar(c, doc)
			pt.RichTextArea(c, doc).Label("Letter").Element().Grow(1).FillWidth()
		})
	})
	fmt.Println(doc.Text, len(doc.Runs()))
	// Output: Dear Ada, 3
}

func ExampleRichText_SetSize() {
	doc := pt.NewRichText("Minutes\nWe met at noon.")
	doc.SetAligns([]pt.Align{pt.AlignCenter})
	// The menu bar of the app offers a Size menu for the document.
	sizes := pt.SizeMenu(doc)
	fmt.Println(sizes.Title, sizes.Items[1].Title, doc.Aligns())
	// Output: Size 22 Point [1 0]
}

func ExamplePopupWindow() {
	dock := pt.PopupWindowState{X: 40, W: 260, H: 160, Open: true}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		ui.Box(c).Fill().Children(func() {
			w := pt.PopupWindow(c, &dock, "Projects", pt.WindowOptions{Active: true}, func() {
				pt.EmptyState(c, "No projects")
			})
			if w.TornOff() {
				// Show it as a Window from now on.
			}
		})
	})
	fmt.Println(dock.Open)
	// Output: true
}

func ExampleSaveDialog() {
	docs := &pt.TreeNode{Title: "Documents", Children: []*pt.TreeNode{{Title: "Letter"}}}
	disk := &pt.TreeNode{Title: "Disk", Children: []*pt.TreeNode{docs}}
	open := true
	st := &pt.FileDialogState{Path: []*pt.TreeNode{disk, docs}, Name: "untitled"}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		ui.Box(c).Fill().Children(func() {
			if pt.SaveDialog(c, &open, st, []*pt.TreeNode{disk}, "Save as:") {
				// Write the document as st.Name in st.Path's last folder.
			}
		})
	})
	fmt.Println(open, st.Name, st.Path[len(st.Path)-1].Title)
	// Output: true untitled Documents
}

func ExampleLabel() {
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			pt.Label(c, "Name:") // a label in the bold face, beside its field
			pt.TextIn(c, "Ada Lovelace", pt.AppFont, pt.Black)
		})
	})
	// Output:
}

func ExampleTextIn() {
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.TextIn(c, "28 items, 3.2 MB available", pt.AppFont, pt.Gray9)
	})
	// Output:
}

func ExampleParagraph() {
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.Paragraph(c, "The document could not be saved, because the disk is full.", pt.AppFont, 200)
	})
	// Output:
}

func ExampleTextArea() {
	notes := "Dear diary,\ntoday the controls lined up on whole pixels."
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.TextArea(c, &notes).Label("Notes").Element().Size(240, 80)
	})
	fmt.Println(len(notes) > 0)
	// Output: true
}

func ExampleSearchField() {
	query := ""
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.SearchField(c, &query).Width(180).Label("Search")
	})
	fmt.Printf("%q\n", query)
	// Output: ""
}

func ExampleStepper() {
	copies := 2.0
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
			pt.TextIn(c, fmt.Sprint(copies), pt.AppFont, pt.Black)
			pt.Stepper(c, &copies, 1, 99, 1).Label("Copies")
		})
	})
	fmt.Println(copies)
	// Output: 2
}

func ExampleSegmented() {
	align := 1
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.Segmented(c, &align, []string{"Left", "Center", "Right"}).Label("Alignment")
	})
	fmt.Println(align)
	// Output: 1
}

func ExampleIconButton() {
	page := []string{"######..", "#....##.", "#.....#.", "#######."}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		if pt.IconButton(c, page, "New Page").Clicked() {
			// Make a page.
		}
	})
	// Output:
}

func ExampleToolbarSeparator() {
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		ui.Row(c).Gap(4).Children(func() {
			pt.BevelButton(c, "Cut")
			pt.ToolbarSeparator(c)
			pt.BevelButton(c, "Paste")
		})
	})
	// Output:
}

func ExampleRadioGroup() {
	size := 1
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.RadioGroup(c, func() {
			pt.Radio(c, &size, 0, "Small")
			pt.Radio(c, &size, 1, "Medium")
			pt.Radio(c, &size, 2, "Large")
		}).Label("Size")
	})
	fmt.Println(size)
	// Output: 1
}

func ExampleSeparator() {
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		ui.Column(c).Gap(6).AlignItems(ui.Stretch).Children(func() {
			pt.Label(c, "Above")
			pt.Separator(c)
			pt.Label(c, "Below")
		})
	})
	// Output:
}

func ExampleBadge() {
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		ui.Row(c).Gap(6).Children(func() {
			pt.TextIn(c, "Mail", pt.AppFont, pt.Black)
			pt.Badge(c, "3")
		})
	})
	// Output:
}

func ExampleTag() {
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		ui.Row(c).Gap(4).Children(func() {
			pt.Tag(c, "draft")
			pt.Tag(c, "letters")
		})
	})
	// Output:
}

func ExamplePlacard() {
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.Placard(c, "12 words, 2 lines")
	})
	// Output:
}

func ExampleStatusBar() {
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		ui.Column(c).Fill().AlignItems(ui.Stretch).Children(func() {
			ui.Box(c).Grow(1)
			pt.StatusBar(c, "12 items", "3 selected", "48 MB available")
		})
	})
	// Output:
}

func ExampleWindowHeader() {
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		ui.Column(c).Fill().AlignItems(ui.Stretch).Children(func() {
			pt.WindowHeader(c, "4 items", "47.7 MB available")
		})
	})
	// Output:
}

func ExampleImageWell() {
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.ImageWell(c, nil).Label("Picture") // empty until a picture is dropped
	})
	// Output:
}

func ExampleCard() {
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.Card(c, func() {
			pt.Label(c, "Backup")
			pt.TextIn(c, "Last backed up today.", pt.AppFont, pt.Black)
		})
	})
	// Output:
}

func ExampleEmptyState() {
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		ui.Column(c).Fill().Children(func() { pt.EmptyState(c, "No messages") })
	})
	// Output:
}

func ExampleBanner() {
	shown := true
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		ui.Box(c).Fill().Children(func() {
			pt.Banner(c, &shown, pt.AlertNote, "The disk was backed up.")
		})
	})
	fmt.Println(shown)
	// Output: true
}

func ExampleBreadcrumb() {
	path := []string{"Platinum Disk", "Documents", "Letters"}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		if i := pt.Breadcrumb(c, path); i >= 0 {
			path = path[:i+1] // go up to the folder chosen
		}
	})
	fmt.Println(len(path))
	// Output: 3
}

func ExamplePagination() {
	page := 2
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.Pagination(c, &page, 12)
	})
	fmt.Println(page)
	// Output: 2
}

func ExampleScrollView() {
	var st ui.ScrollState
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.ScrollView(c, &st, func() {
			for i := 1; i <= 40; i++ {
				pt.TextIn(c, fmt.Sprintf("Line %d", i), pt.AppFont, pt.Black)
			}
		}).Size(200, 120)
	})
	// Output:
}

func ExampleCalendar() {
	day := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.Calendar(c, &day).Label("Calendar")
	})
	fmt.Println(day.Format("Jan 2"))
	// Output: Oct 14
}

func ExampleTimeField() {
	t := time.Date(2026, 10, 14, 9, 41, 0, 0, time.UTC)
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		pt.TimeField(c, &t, 1).Label("Time")
	})
	fmt.Println(t.Format("3:04 PM"))
	// Output: 9:41 AM
}

func ExampleFormatBar() {
	doc := pt.NewRichText("Minutes")
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		ui.Column(c).Fill().AlignItems(ui.Stretch).Children(func() {
			pt.FormatBar(c, doc) // bold, italic, underline and alignment
			pt.RichTextArea(c, doc).Label("Document").Element().Grow(1)
		})
	})
	// Output:
}

func ExampleFormulaBar() {
	sheet := pt.NewSheetState()
	sheet.Cells[pt.Cell{Col: 0, Row: 0}] = "=1+2"
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		ui.Column(c).Fill().AlignItems(ui.Stretch).Children(func() {
			pt.FormulaBar(c, sheet)
			pt.Sheet(c, sheet, 20, 6).Label("Sheet").Element().Grow(1)
		})
	})
	fmt.Println(sheet.Value(pt.Cell{Col: 0, Row: 0}))
	// Output: 3
}

func ExamplePageSetup() {
	open := true
	setup := pt.PageSetupState{Scale: 100, Margin: 36}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		ui.Box(c).Fill().Children(func() {
			if pt.PageSetup(c, &open, &setup) {
				// Lay the document out again for the paper chosen.
			}
		})
	})
	fmt.Println(pt.Papers[setup.Paper].Name)
	// Output: US Letter
}

func ExampleOpenDialog() {
	docs := &pt.TreeNode{Title: "Documents", Children: []*pt.TreeNode{{Title: "Letter"}}}
	disk := &pt.TreeNode{Title: "Disk", Children: []*pt.TreeNode{docs}}
	open := true
	st := &pt.FileDialogState{}
	render(func(c *ui.Context) {
		pt.Use(c, nil)
		ui.Box(c).Fill().Children(func() {
			if f := pt.OpenDialog(c, &open, st, []*pt.TreeNode{disk}); f != nil {
				// Open f.
			}
		})
	})
	fmt.Println(open, len(st.Path))
	// Output: true 0
}
