package main

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
	"github.com/firede/mygo-platinum/examples/desktop/locales"
)

// catalog is the Controls application: a list of the library's controls
// by the names of their functions, and the controls of the one chosen,
// live, in their states.
type catalog struct {
	chosen int
	// What the controls show and change.
	on, off, mixed bool
	radio          string
	slider         float64
	open           bool
	tab, pick      int
	text, area     string
	list           int
	scroll         float64
	front, back    pt.WindowState
	alert, dialog  bool
	toggle         bool
	segment        int
	number, steps  float64
	query          string
	tree           *pt.TreeNode
	table          *pt.TableState
	icon           int
	split          float64
	scrollView     ui.ScrollState
	day            time.Time
	combo          string
	color          ui.Color
	sheet          *pt.SheetState
	find           bool
	findState      pt.FindState
	ruler          pt.RulerState
	columns        []*pt.TreeNode
	docTab         int
	font           pt.FontChoice
	pageSetup      bool
	pageState      pt.PageSetupState
	progress       bool
	banner         bool
	page           int
	doc            *pt.RichText
	dock           pt.PopupWindowState
	opening        bool
	saving         bool
	files          pt.FileDialogState
	zoomFrom       ui.Rect
	zoomStart      time.Time
}

// catalogPage is a page of the catalog: the functions it shows, and what
// builds their controls.
type catalogPage struct {
	name  string
	build func(c *ui.Context, k *catalog, d *desk, m locales.Messages)
}

func newCatalogWindow(d *desk) *win {
	k := &catalog{
		on: true, mixed: true, radio: "two", slider: 40, open: true,
		front: pt.WindowState{X: 30, Y: 8, W: 170, H: 90}, back: pt.WindowState{X: 6, Y: 30, W: 150, H: 80},
		toggle: true, segment: 1, number: 2, steps: 4,
		table: pt.NewTableState(), icon: 1, split: 120,
		day: d.now(), color: pt.Lavender.Mid,
		sheet: pt.NewSheetState(), ruler: pt.RulerState{Left: 9, Right: 270},
		font: pt.FontChoice{Family: "Sans", Size: 12}, page: 2,
		pageState: pt.PageSetupState{Scale: 100, Margin: 36},
		dock:      pt.PopupWindowState{X: 8, W: 150, H: 70, Open: true},
	}
	for r, row := range [][]string{{"", "1"}, {"", "2"}, {"", "3"}, {"", "=SUM(B1:B3)"}} {
		for col, v := range row {
			k.sheet.Cells[pt.Cell{Col: col, Row: r}] = v
		}
	}
	w := &win{app: appCatalog, catalog: k}
	w.frame = pt.WindowState{W: 600, H: 380}
	w.content = func(c *ui.Context, w *win) { catalogView(c, d, w.catalog) }
	return w
}

func catalogView(c *ui.Context, d *desk, k *catalog) {
	m := d.say()
	pages := catalogPages()
	names := make([]string, len(pages))
	for i, p := range pages {
		names[i] = p.name
	}
	ui.Row(c).Fill().Children(func() {
		ui.Column(c).Width(170).AlignSelf(ui.Stretch).AlignItems(ui.Stretch).Children(func() {
			pt.ListBox(c, &k.chosen, names).Label(d.appTitle(appCatalog)).Element().Grow(1)
		})
		p := pages[k.chosen]
		ui.Column(c).Grow(1).AlignSelf(ui.Stretch).Justify(ui.Start).Padding(12, 14).Gap(10).AlignItems(ui.Stretch).Clip().Children(func() {
			pt.Label(c, p.name)
			pt.Separator(c)
			ui.Column(c.Key(p.name)).Grow(1).Justify(ui.Start).Gap(8).AlignItems(ui.Start).Children(func() {
				p.build(c, k, d, m)
			})
		})
	})
}

// row lays out fn's controls side by side.
func row(c *ui.Context, fn func()) {
	ui.Row(c).Gap(10).AlignItems(ui.Center).Children(fn)
}

// box lays out fn's controls over a gray area w×h, as on a desktop, for
// windows.
func box(c *ui.Context, w, h float32, fn func()) {
	ui.Box(c).Size(w, h).Background(pt.Gray5).Clip().Children(fn)
}

// catalogPages are the pages of the catalog, in the order of the README's
// table of controls.
func catalogPages() []catalogPage {
	return []catalogPage{
		{"Button", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			row(c, func() {
				pt.Button(c, m.Cancel()).Cancel()
				pt.Button(c, m.Ok()).Default()
			})
			pt.Button(c, m.CatDisabled()).Disabled(true)
		}},
		{"Checkbox", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			pt.Checkbox(c, &k.on, m.CatOne())
			pt.Checkbox(c, &k.off, m.CatTwo())
			pt.Checkbox(c, &k.mixed, m.CatThree()).Mixed()
			pt.Checkbox(c, &k.on, m.CatDisabled()).Disabled(true)
		}},
		{"Radio, RadioGroup", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			pt.RadioGroup(c, func() {
				pt.Radio(c, &k.radio, "one", m.CatOne())
				pt.Radio(c, &k.radio, "two", m.CatTwo())
				pt.Radio(c, &k.radio, "three", m.CatDisabled()).Disabled(true)
			})
		}},
		{"Slider", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			pt.Slider(c, &k.slider, 0, 100).Width(220)
			pt.Slider(c, &k.slider, 0, 100).Width(220).Ticks(5).Step(25)
		}},
		{"ProgressBar", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			pt.ProgressBar(c, 0.62).Width(220)
			pt.ProgressBar(c, pt.Indeterminate).Width(220)
		}},
		{"Disclosure", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			pt.Disclosure(c, &k.open, m.CatOne())
			if k.open {
				pt.TextIn(c, m.CatText(), pt.AppFont, pt.Black)
			}
		}},
		{"GroupBox, Separator", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			ui.Column(c).Width(240).AlignItems(ui.Stretch).Gap(8).Children(func() {
				pt.GroupBox(c, m.CatOne(), func() {
					pt.TextIn(c, m.CatText(), pt.AppFont, pt.Black)
				})
				pt.Separator(c)
			})
		}},
		{"TabControl", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			tabs := []string{m.CatOne(), m.CatTwo(), m.CatThree()}
			pt.TabControl(c, &k.tab, tabs, func(i int) {
				pt.TextIn(c, tabs[i], pt.AppFont, pt.Black)
			}).Element().Size(260, 90)
		}},
		{"PopupMenu", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			pt.PopupMenu(c, &k.pick, m.CatOne()+":", []string{m.CatOne(), m.CatTwo(), m.CatThree()})
			pt.PopupMenu(c, &k.pick, "", []string{m.CatOne(), m.CatTwo(), m.CatThree()}).Disabled(true)
		}},
		{"TextField, TextArea", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			pt.TextField(c, &k.text).Width(220).Placeholder(m.Untitled())
			pt.TextField(c, &k.text).Width(220).Password()
			pt.TextArea(c, &k.area).Element().Size(220, 70)
		}},
		{"ListBox", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			pt.ListBox(c, &k.list, []string{m.Letter(), m.Recipes(), m.ToDo(), m.ReadMe(), m.Untitled()}).Element().Size(180, 80)
		}},
		{"ScrollBar", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			row(c, func() {
				pt.ScrollBar(c, &k.scroll, 20, 100).Element().Height(100)
				pt.ScrollBar(c, &k.scroll, 20, 100).Horizontal().Element().Width(160)
			})
		}},
		{"Window", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			box(c, 300, 140, func() {
				pt.Window(c.Key("back"), &k.back, m.Untitled(), pt.WindowOptions{}, func() {})
				pt.Window(c.Key("front"), &k.front, m.CatWindow(), pt.WindowOptions{Active: true}, func() {})
			})
		}},
		{"MenuBar", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			ui.Column(c).Width(300).AlignItems(ui.Stretch).Children(func() {
				pt.MenuBar(c, []pt.Menu{
					{Title: m.MenuFile(), Items: []pt.MenuItem{{Title: m.Open(), Key: ui.KeyO}, pt.Divider(), {Title: m.Save(), Key: ui.KeyS, Disabled: true}}},
					{Title: m.MenuEdit(), Items: []pt.MenuItem{{Title: m.Copy(), Key: ui.KeyC}, {Title: m.Paste(), Key: ui.KeyV}}},
				})
			})
		}},
		{"Alert", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			if pt.Button(c, m.CatShow()).Clicked() {
				k.alert = true
			}
			pt.Alert(c, &k.alert, pt.AlertNote, m.CatMessage(), m.CatText(), m.Cancel(), m.Ok())
		}},
		{"Dialog", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			if pt.Button(c, m.CatShow()).Clicked() {
				k.dialog = true
			}
			pt.Dialog(c, &k.dialog, m.CatWindow(), 260, func() {
				pt.Checkbox(c, &k.on, m.CatOne())
				ui.Row(c).Justify(ui.End).Children(func() {
					if pt.Button(c, m.Ok()).Default().Clicked() {
						k.dialog = false
					}
				})
			})
		}},
		{"Label, TextIn, Paragraph", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			pt.Label(c, m.CatText())
			pt.TextIn(c, m.CatText(), pt.AppFont, pt.Black)
			pt.Paragraph(c, m.CatText()+" "+m.CatText(), pt.AppFont, 240)
		}},
		{"BevelButton, IconButton", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			row(c, func() {
				pt.BevelButton(c, m.CatOne())
				pt.BevelButton(c, m.CatTwo()).Toggle(&k.toggle)
				pt.BevelButton(c, m.CatDisabled()).Disabled(true)
			})
			row(c, func() {
				pt.BevelButton(c, m.CatThree()).Menu([]pt.MenuItem{{Title: m.CatOne()}, {Title: m.CatTwo()}})
				pt.IconButton(c, pt.IconBold, m.CatOne())
				pt.IconButton(c, pt.IconItalic, m.CatTwo())
			})
		}},
		{"Segmented", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			pt.Segmented(c, &k.segment, []string{m.CatOne(), m.CatTwo(), m.CatThree()})
		}},
		{"Toolbar, ToolbarSeparator", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			ui.Column(c).Width(300).AlignItems(ui.Stretch).Children(func() {
				pt.Toolbar(c, []pt.ToolItem{
					{Label: m.Save(), Icon: iconSave, IconOnly: true},
					{Separator: true},
					{Label: m.CatOne(), Icon: pt.IconBold, IconOnly: true, On: &k.toggle},
					{Label: m.CatTwo(), Icon: pt.IconItalic, IconOnly: true},
					{Separator: true},
					{Label: m.CatThree(), Icon: pt.IconAlign[0]},
				})
			})
		}},
		{"Stepper, NumberField", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			pt.Stepper(c, &k.steps, 0, 10, 1)
			pt.NumberField(c, &k.number, 1, 99, 1, 0)
		}},
		{"SearchField", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			pt.SearchField(c, &k.query).Width(200)
		}},
		{"ChasingArrows", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			pt.ChasingArrows(c)
		}},
		{"Placard, WindowHeader", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			pt.Placard(c, m.Items(3))
			ui.Column(c).Width(260).AlignItems(ui.Stretch).Children(func() {
				pt.WindowHeader(c, m.Items(3), m.Available(m.SizeMB(47.7)))
			})
		}},
		{"Badge, Tag", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			row(c, func() {
				pt.Badge(c, "3")
				pt.Tag(c, m.CatOne())
				pt.Tag(c, m.CatTwo())
			})
		}},
		{"ImageWell", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			pt.ImageWell(c, iconPuzzle.normal)
		}},
		{"Balloon", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			b := pt.Button(c, m.CatOne())
			pt.Balloon(c, b.Element(), m.CatText())
		}},
		{"ContextMenu", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			hint := m.CatContext()
			target := ui.Box(c).Size(260, 50).Focusable().Label(hint)
			target.Draw(func(p *ui.Painter, r ui.Rect) {
				pn := pt.NewPen(p, r)
				pt.DrawWell(pn)
				pn.TextCentered(0, 0, pn.W, pn.H, hint, pt.AppFont, pt.Gray8)
			})
			pt.ContextMenu(c, target, []pt.MenuItem{{Title: m.Cut()}, {Title: m.Copy()}, {Title: m.Paste(), Disabled: true}})
		}},
		{"TreeView", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			tree := d.fileTree()
			tree[0].Open = true
			pt.TreeView(c, tree, &k.tree).Element().Size(240, 120)
		}},
		{"Table", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			items := []string{m.Rent(), m.Groceries(), m.LampOil(), m.Books()}
			costs := []int{1200, 340, 18, 65}
			cols := []pt.Column{{Title: m.Item(), Width: 150}, {Title: m.Cost(), Width: 100, Right: true}}
			pt.Table(c, k.table, cols, len(items), func(r, col int) string {
				if col == 0 {
					return items[r]
				}
				return fmt.Sprint(costs[r])
			}).Element().Size(270, 100)
		}},
		{"IconView", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			pt.IconView(c, []pt.IconItem{
				{Name: m.Calculator(), Icon: iconCalculator.normal},
				{Name: m.Jotter(), Icon: iconJotter.normal},
				{Name: m.Scribe(), Icon: iconScribe.normal},
				{Name: m.Puzzle(), Icon: iconPuzzle.normal},
			}, &k.icon).Element().Size(280, 110)
		}},
		{"Split, ScrollView", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			ui.Box(c).Size(300, 120).Children(func() {
				pt.Split(c, &k.split, false, func() {
					ui.Column(c).Padding(6).Children(func() {
						pt.TextIn(c, m.CatOne(), pt.AppFont, pt.Black)
					})
				}, func() {
					pt.ScrollView(c, &k.scrollView, func() {
						for i := range 12 {
							pt.TextIn(c, fmt.Sprintf("%d. %s", i+1, m.CatText()), pt.AppFont, pt.Black)
						}
					}).Grow(1)
				}).Fill()
			})
		}},
		{"Calendar", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			pt.Calendar(c, &k.day)
		}},
		{"DatePicker, TimeField", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			pt.DatePicker(c, &k.day)
			pt.TimeField(c, &k.day, 5)
		}},
		{"ComboBox", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			pt.ComboBox(c, &k.combo, []string{m.Slow(), m.Medium(), m.Fast()})
		}},
		{"ColorWell", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			pt.ColorWell(c, &k.color, m.HighlightColor())
		}},
		{"Sheet, FormulaBar", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			ui.Column(c).Width(300).AlignItems(ui.Stretch).Gap(4).Children(func() {
				for r, s := range []string{m.Rent(), m.Groceries(), m.Books(), m.Total()} {
					k.sheet.Cells[pt.Cell{Col: 0, Row: r}] = s
				}
				pt.FormulaBar(c, k.sheet)
				pt.Sheet(c, k.sheet, 20, 6).Element().Size(300, 110)
			})
		}},
		{"FindReplace", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			if pt.Button(c, m.CatShow()).Clicked() {
				k.find = true
			}
			pt.FindReplace(c, &k.find, &k.findState)
		}},
		{"Ruler", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			ui.Column(c).Width(300).AlignItems(ui.Stretch).Children(func() {
				pt.Ruler(c, &k.ruler)
			})
		}},
		{"ColumnView", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			pt.ColumnView(c, d.fileTree(), &k.columns).Element().Size(330, 110)
		}},
		{"DocumentTabs", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			docs := []string{m.Letter(), m.Recipes(), m.ToDo()}
			pt.DocumentTabs(c, &k.docTab, docs, func(i int) {
				pt.TextIn(c, docs[i], pt.AppFont, pt.Black)
			}).Element().Size(280, 90)
		}},
		{"FontPicker", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			pt.FontPicker(c, &k.font, []string{"Sans", "Serif", "Mono"})
		}},
		{"PageSetup", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			if pt.Button(c, m.CatShow()).Clicked() {
				k.pageSetup = true
			}
			pt.PageSetup(c, &k.pageSetup, &k.pageState)
		}},
		{"ProgressDialog", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			if pt.Button(c, m.CatShow()).Clicked() {
				k.progress = true
			}
			pt.ProgressDialog(c, &k.progress, m.CatWindow(), m.CatMessage(), 0.4, m.CatText())
		}},
		{"Banner", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			ui.Column(c).Width(320).AlignItems(ui.Stretch).Gap(8).Children(func() {
				pt.Banner(c, &k.banner, pt.AlertNote, m.CatMessage())
				ui.Row(c).Children(func() {
					if pt.Button(c, m.CatShow()).Clicked() {
						k.banner = true
					}
				})
			})
		}},
		{"Toasts", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			row(c, func() {
				if pt.Button(c, m.CatOne()).Clicked() {
					c.Toast(m.CatMessage())
				}
				if pt.Button(c, m.CatTwo()).Clicked() {
					c.AddToast(ui.Toast{Title: m.CatMessage(), Type: "caution", Action: m.Ok(), OnAction: func() {}})
				}
				if pt.Button(c, m.CatThree()).Clicked() {
					c.AddToast(ui.Toast{Title: m.CatMessage(), Description: m.CatText(), Type: "error"})
				}
			})
		}},
		{"EmptyState, Card", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			ui.Column(c).Width(260).AlignItems(ui.Stretch).Children(func() {
				pt.Card(c, func() {
					ui.Box(c).Height(80).AlignSelf(ui.Stretch).Children(func() {
						pt.EmptyState(c, m.CatEmpty())
					})
				})
			})
		}},
		{"PathPopup, Breadcrumb", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			path := []string{m.PlatinumDisk(), m.Documents(), m.Letter()}
			pt.PathPopup(c, path)
			pt.Breadcrumb(c, path)
		}},
		{"Pagination", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			pt.Pagination(c, &k.page, 5)
		}},
		{"StatusBar", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			ui.Column(c).Width(300).AlignItems(ui.Stretch).Children(func() {
				pt.StatusBar(c, m.Items(3), m.Available(m.SizeMB(47.7)))
			})
		}},
		{"RichTextArea, FormatBar, SizeMenu", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			if k.doc == nil {
				k.doc = pt.NewRichText("")
				k.doc.SetRuns([]pt.Run{{Text: m.CatOne(), Style: pt.Bold.Scaled(2)}, {Text: "\n" + m.CatText()}})
			}
			ui.Column(c).Width(320).AlignItems(ui.Stretch).Children(func() {
				pt.FormatBar(c, k.doc)
				pt.RichTextArea(c, k.doc).Element().Height(90)
			})
		}},
		{"PopupWindow, SpringOpen", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			box(c, 300, 130, func() {
				pt.PopupWindow(c, &k.dock, m.Documents(), pt.WindowOptions{Active: true}, func() {
					ui.Column(c).Fill().Background(pt.White).Padding(4).Children(func() {
						pt.TextIn(c, m.Letter(), pt.AppFont, pt.Black)
					})
				})
			})
		}},
		{"OpenDialog, SaveDialog", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			row(c, func() {
				if pt.Button(c, m.Open()).Clicked() {
					k.opening = true
				}
				if pt.Button(c, m.SaveAsMenu()).Clicked() {
					k.files.Name = m.Untitled()
					k.saving = true
				}
			})
			tree := d.fileTree()
			pt.OpenDialog(c, &k.opening, &k.files, tree)
			pt.SaveDialog(c, &k.saving, &k.files, tree, m.SaveAsLabel())
		}},
		{"ZoomRects", func(c *ui.Context, k *catalog, d *desk, m locales.Messages) {
			b := pt.Button(c, m.CatShow())
			if b.Clicked() {
				x, y := b.Element().LocalToWindow(0, 0)
				k.zoomFrom, k.zoomStart = ui.Rect{X: x, Y: y, W: 60, H: 20}, time.Now()
			}
			if !k.zoomStart.IsZero() {
				x, y := b.Element().LocalToWindow(0, 0)
				if !pt.ZoomRects(c, k.zoomFrom, ui.Rect{X: x - 40, Y: y + 30, W: 280, H: 160}, k.zoomStart) {
					k.zoomStart = time.Time{}
				}
			}
		}},
	}
}
