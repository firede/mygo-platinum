package main

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
)

var miniFolder = []string{
	"###.......",
	"#..######.",
	"#.......#.",
	"#.......#.",
	"#.......#.",
	"#########.",
}

type planet struct {
	name   string
	moons  int
	radius int
}

var planets = []planet{
	{"Mercury", 0, 2440}, {"Venus", 0, 6052}, {"Earth", 1, 6371}, {"Mars", 2, 3390},
	{"Jupiter", 95, 69911}, {"Saturn", 146, 58232}, {"Uranus", 28, 25362}, {"Neptune", 16, 24622},
}

func (g *gallery) lists() {
	g.tree = []*pt.TreeNode{
		{Title: "Documents", Icon: miniFolder, Open: true, Children: []*pt.TreeNode{
			{Title: "Letters", Icon: miniFolder, Children: []*pt.TreeNode{{Title: "To Ada"}}},
			{Title: "Recipes"},
			{Title: "To Do"},
		}},
		{Title: "Pictures", Icon: miniFolder, Children: []*pt.TreeNode{{Title: "Beach"}}},
		{Title: "Read Me"},
	}
	g.treeSel = g.tree[0].Children[1]
	g.table = pt.NewTableState()
	g.table.SortColumn = 2
	g.table.Selected[2] = true
	g.split = 150
}

// page3 shows lists, tables and layout: a tree view, a table, an icon
// view, a split with a scroll view, and a utility window.
func (g *gallery) page3(c *ui.Context) {
	ui.Row(c).Grow(1).Padding(12).Gap(12).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(232).Gap(12).AlignItems(ui.Stretch).Children(func() {
			pt.GroupBox(c, "Tree view", func() {
				pt.TreeView(c, g.tree, &g.treeSel).Element().Size(206, 116).Label("Files")
			})
			pt.GroupBox(c, "Icon view", func() {
				pt.IconView(c, g.icons, &g.iconSel).Element().Size(206, 130).Label("Icons")
			})
		})
		ui.Column(c).Width(482).Gap(12).AlignItems(ui.Stretch).Children(func() {
			pt.GroupBox(c, "Table", func() {
				cols := []pt.TableColumn{{Title: "Planet", Width: 150, Icon: func(int) []string { return planetIcon }}, {Title: "Moons", Width: 120, Right: true}, {Title: "Radius (km)", Width: 160, Right: true}}
				pt.Table(c, g.table, cols, len(planets), func(r, col int) string {
					p := planets[r]
					switch col {
					case 0:
						return p.name
					case 1:
						return fmt.Sprint(p.moons)
					}
					return fmt.Sprint(p.radius)
				}).Element().Size(456, 116).Label("Planets")
			})
			pt.GroupBox(c, "Split and scroll view", func() {
				ui.Box(c).Height(130).AlignSelf(ui.Stretch).Background(pt.Gray5).Children(func() {
					pt.Split(c, &g.split, func() {
						ui.Column(c).Padding(6).Children(func() {
							pt.TextIn(c, "A pane", pt.AppFont, pt.Black)
						})
					}, func() {
						pt.ScrollView(c, &g.scroll2, func() {
							for i, p := range planets {
								pt.TextIn(c, fmt.Sprintf("%d. %s", i+1, p.name), pt.AppFont, pt.Black)
							}
						}).Grow(1)
					}).Size(300, 130)
					pt.Window(c, &g.palette, "Palette", pt.WindowOptions{Utility: true, Active: true}, func() {
						ui.Row(c).Padding(6).Gap(4).Children(func() {
							pt.IconButton(c, pt.IconBold, "Bold")
							pt.IconButton(c, pt.IconItalic, "Italic")
						})
					})
				})
			})
		})
	})
}

// planetIcon is a small round planet, for the rows of the table.
var planetIcon = []string{
	"....####....",
	"..##++++##..",
	".#+ww+++++#.",
	".#+w++++++#.",
	"#++++++++++#",
	"#++++++++++#",
	"#++++++++o+#",
	"#+++++++oo+#",
	".#+++++ooo#.",
	".#++++oooo#.",
	"..##oooo##..",
	"....####....",
}
