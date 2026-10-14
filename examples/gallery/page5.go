package main

import (
	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
)

// page5 shows documents and navigation: a column view, document tabs, a
// font picker, paths and pages, a card, an empty state and a status bar.
func (g *gallery) page5(c *ui.Context) {
	if g.columns == nil {
		g.columns = []*pt.TreeNode{g.tree[0], g.tree[0].Children[0], g.tree[0].Children[0].Children[0]}
		g.tree[0].Children[0].Open = true
	}
	ui.Column(c).Grow(1).AlignItems(ui.Stretch).Children(func() {
		ui.Row(c).Grow(1).Padding(12).Gap(12).AlignItems(ui.Start).Children(func() {
			ui.Column(c).Width(470).Gap(12).AlignItems(ui.Stretch).Children(func() {
				pt.GroupBox(c, "Column view", func() {
					pt.ColumnView(c, g.tree, &g.columns).Element().Size(444, 96).Label("Browser")
				})
				pt.GroupBox(c, "Font", func() {
					pt.FontPicker(c, &g.fontChoice, fonts)
				})
			})
			ui.Column(c).Width(246).Gap(12).AlignItems(ui.Stretch).Children(func() {
				pt.GroupBox(c, "Path", func() {
					pt.PathPopup(c, []string{"Platinum Disk", "Documents", "Letters"})
				})
				pt.DocumentTabs(c, &g.docTab, []string{"Letter", "Notes"}, func(i int) {
					pt.TextIn(c, []string{"Dear Ada,", "Buy lamp oil."}[i], pt.AppFont, pt.Black)
				}).Element().Height(80)
				pt.Card(c, func() {
					ui.Box(c).Height(70).AlignSelf(ui.Stretch).Children(func() {
						pt.EmptyState(c, "No messages")
					})
				})
			})
		})
		pt.StatusBar(c, "12 items", "3 selected", "48 MB available")
	})
}
