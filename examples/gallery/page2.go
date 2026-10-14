package main

import (
	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
)

// iconNew is a 16×16 page with a folded corner.
var iconNew = []string{
	"................",
	"...#######......",
	"...#wwwww##.....",
	"...#wwwww#w#....",
	"...#wwwww####...",
	"...#wwwwwwww#...",
	"...#wwwwwwww#...",
	"...#wwwwwwww#...",
	"...#wwwwwwww#...",
	"...#wwwwwwww#...",
	"...#wwwwwwww#...",
	"...#wwwwwwww#...",
	"...#wwwwwwww#...",
	"...##########...",
	"................",
	"................",
}

// page2 shows the controls beyond the classic set: bevel buttons, a
// toolbar, little arrows and fields, status controls and help, rich text
// and pop-up windows.
func (g *gallery) page2(c *ui.Context) {
	ui.Column(c).Grow(1).AlignItems(ui.Stretch).Children(func() {
		pt.Toolbar(c, []pt.ToolItem{
			{Label: "New", Icon: iconNew, IconOnly: true},
			{Separator: true},
			{Label: "Bold", Icon: pt.IconBold, IconOnly: true, On: &g.bold},
			{Label: "Italic", Icon: pt.IconItalic, IconOnly: true, On: &g.italic},
			{Separator: true},
			{Label: "Align", Icon: pt.IconAlign[0]},
		})
		ui.Row(c).Padding(12, 12, 0, 12).Gap(12).AlignItems(ui.Start).Children(func() {
			ui.Column(c).Width(232).Gap(12).AlignItems(ui.Stretch).Children(func() {
				pt.GroupBox(c, "Bevel buttons", func() {
					ui.Row(c).Gap(6).Children(func() {
						pt.BevelButton(c, "Plain")
						pt.BevelButton(c, "Toggle").Toggle(&g.toggle)
						pt.BevelButton(c, "Off").Disabled(true)
					})
					ui.Row(c).Gap(6).Children(func() {
						pt.BevelButton(c, "Page").Icon(iconNew)
						pt.BevelButton(c, "Style").Menu([]pt.MenuItem{{Title: "Plain"}, {Title: "Bold"}})
					})
					pt.Segmented(c, &g.segment, []string{"Left", "Center", "Right"})
				})
				pt.GroupBox(c, "Badges and tags", func() {
					ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
						pt.Badge(c, "3")
						pt.Badge(c, "New")
						pt.Tag(c, "draft")
						pt.Tag(c, "letters")
					})
				})
			})
			ui.Column(c).Width(232).Gap(12).AlignItems(ui.Stretch).Children(func() {
				pt.GroupBox(c, "Numbers and search", func() {
					ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
						pt.Label(c, "Copies:")
						pt.NumberField(c, &g.copies, 1, 99, 1, 0).Label("Copies")
					})
					ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
						pt.Label(c, "Level:")
						pt.Stepper(c, &g.level, 0, 10, 1).Label("Level")
					})
					pt.SearchField(c, &g.query).Width(200).Label("Search")
				})
				pt.GroupBox(c, "Status", func() {
					ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
						pt.ChasingArrows(c).Label("Working")
						pt.TextIn(c, "Connecting…", pt.AppFont, pt.Black)
					})
					pt.Placard(c, "28 items, 3.2 MB available")
				})
			})
			ui.Column(c).Width(232).Gap(12).AlignItems(ui.Stretch).Children(func() {
				pt.GroupBox(c, "Image well and help", func() {
					ui.Row(c).Gap(10).AlignItems(ui.Start).Children(func() {
						pt.ImageWell(c, nil).Label("Picture")
						ui.Box(c).Size(150, 50).Draw(func(p *ui.Painter, r ui.Rect) {
							pn := pt.NewPen(p, r)
							pt.DrawBalloon(pn)
							pn.Text(10, 14, "Drop a picture in", pt.AppFont, pt.Black)
							pn.Text(10, 27, "the well to use it.", pt.AppFont, pt.Black)
						})
					})
				})
				pt.GroupBox(c, "Context menu", func() {
					target := ui.Box(c).Height(40).AlignSelf(ui.Stretch).Focusable().Label("Right-click here")
					target.Draw(func(p *ui.Painter, r ui.Rect) {
						pn := pt.NewPen(p, r)
						pt.DrawWell(pn)
						pn.TextCentered(0, 0, pn.W, pn.H, "Right-click here", pt.AppFont, pt.Gray8)
					})
					pt.ContextMenu(c, target, []pt.MenuItem{{Title: "Cut"}, {Title: "Copy"}, {Title: "Paste", Disabled: true}})
				})
			})
		})
		ui.Row(c).Grow(1).Padding(12).Gap(12).AlignItems(ui.Stretch).Children(func() {
			ui.Column(c).Width(476).AlignItems(ui.Stretch).Children(func() {
				pt.GroupBox(c, "Rich text", func() {
					if g.letter == nil {
						g.letter = pt.NewRichText("")
						g.letter.SetRuns([]pt.Run{
							{Text: "Minutes", Style: pt.Bold.Scaled(2)},
							{Text: "\nWe met at noon in the "},
							{Text: "east", Style: pt.Italic},
							{Text: " room and agreed on the "},
							{Text: "budget", Style: pt.Bold},
							{Text: ".\n"},
							{Text: "Next: the 14th", Style: pt.Underline},
						})
						g.letter.SetAligns([]pt.Align{pt.AlignCenter, pt.AlignLeft, pt.AlignRight})
					}
					pt.FormatBar(c, g.letter)
					pt.RichTextArea(c, g.letter).Label("Minutes").Element().Size(452, 86)
				})
			})
			ui.Column(c).Width(232).AlignItems(ui.Stretch).Children(func() {
				pt.GroupBox(c, "Pop-up windows", func() {
					desk := ui.Box(c).Height(118).AlignSelf(ui.Stretch).Clip()
					desk.Draw(func(p *ui.Painter, r ui.Rect) {
						pn := pt.NewPen(p, r)
						pn.Rect(0, 0, pn.W, pn.H, pt.Gray6)
						pn.Frame(0, 0, pn.W, pn.H, pt.Black)
					})
					desk.Children(func() {
						if g.docks[0].W == 0 {
							g.docks = [2]pt.PopupWindowState{{X: 6, W: 120, H: 80, Open: true}, {X: 104, W: 104, H: 60}}
						}
						// Back to front: the one in front last.
						for _, i := range []int{1, 0} {
							title := []string{"Projects", "Mail"}[i]
							pt.PopupWindow(c.Key(title), &g.docks[i], title, pt.WindowOptions{Active: i == 0}, func() {
								ui.Column(c).Fill().Background(pt.White).Padding(4).Children(func() {
									if i == 0 {
										pt.TextIn(c, "Platinum", pt.AppFont, pt.Black)
										pt.TextIn(c, "Desk", pt.AppFont, pt.Black)
									}
								})
							})
						}
					})
				})
			})
		})
	})
}
