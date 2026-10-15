package main

import (
	"time"

	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
)

func (g *gallery) office() {
	g.day = time.Date(2026, 10, 14, 9, 41, 0, 0, time.UTC)
	g.color = pt.Lavender.Mid
	g.sheet = pt.NewSheetState()
	for r, row := range [][]string{{"Month", "Sales"}, {"July", "120"}, {"August", "145"}, {"September", "98"}, {"Total", "=SUM(B2:B4)"}} {
		for col, v := range row {
			g.sheet.Cells[pt.Cell{Col: col, Row: r}] = v
		}
	}
	g.sheet.Cursor, g.sheet.Anchor = pt.Cell{Col: 1, Row: 4}, pt.Cell{Col: 1, Row: 4}
	g.ruler = pt.RulerState{Left: 16, Right: 400}
}

// page4 shows dates, colors, a combo box, a sheet and a ruler.
func (g *gallery) page4(c *ui.Context) {
	ui.Row(c).Grow(1).Padding(12).Gap(12).AlignItems(ui.Start).Children(func() {
		ui.Column(c).Width(232).Gap(12).AlignItems(ui.Stretch).Children(func() {
			pt.GroupBox(c, "Dates and times", func() {
				pt.Calendar(c, &g.day)
				ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
					pt.DatePicker(c, &g.day)
				})
				pt.TimeField(c, &g.day, 5)
			})
		})
		ui.Column(c).Width(482).Gap(12).AlignItems(ui.Stretch).Children(func() {
			pt.GroupBox(c, "Combo box and color well", func() {
				ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
					pt.Combobox(c, &g.combo, fonts).Label("Font")
					pt.Label(c, "Color:")
					pt.ColorWell(c, &g.color).Label("Highlight color")
				})
			})
			pt.GroupBox(c, "Sheet", func() {
				pt.FormulaBar(c, g.sheet)
				pt.Sheet(c, g.sheet, 50, 10).Element().Size(456, 130).Label("Sales")
			})
			pt.GroupBox(c, "Ruler", func() {
				pt.Ruler(c, &g.ruler)
			})
		})
	})
}
