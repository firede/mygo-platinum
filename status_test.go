package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSnapshotSmallStatus(t *testing.T) {
	freeze(t)
	snap(t, "stepper-status", 420, 80, func(c *ui.Context) {
		v, n := 3.0, 12.0
		q := "pix"
		ui.Column(c).Padding(8).Gap(8).Children(func() {
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				Stepper(c, &v, 0, 10, 1).Label("Count")
				ui.Box(c).Size(13, 23).Draw(func(p *ui.Painter, r ui.Rect) { DrawStepper(NewPen(p, r), 1, StateNormal) })
				ui.Box(c).Size(13, 23).Draw(func(p *ui.Painter, r ui.Rect) { DrawStepper(NewPen(p, r), 0, StateDisabled) })
				NumberField(c, &n, 6, 72, 1, 0).Label("Size")
				SearchField(c, &q).Width(120)
				ChasingArrows(c).Label("Busy")
			})
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				Placard(c, "12 items")
				Badge(c, "3")
				Badge(c, "New")
				Tag(c, "draft")
				for s := 0; s < 4; s++ {
					ui.Box(c).Size(16, 16).Draw(func(p *ui.Painter, r ui.Rect) { DrawChasingArrows(NewPen(p, r), s*2) })
				}
			})
		})
	})
}

func TestSnapshotWellAndBalloon(t *testing.T) {
	snap(t, "well-balloon", 260, 80, func(c *ui.Context) {
		ui.Row(c).Padding(8).Gap(12).Children(func() {
			ImageWell(c, nil).Label("Picture")
			ui.Box(c).Size(150, 50).Draw(func(p *ui.Painter, r ui.Rect) {
				pn := NewPen(p, r)
				DrawBalloon(pn)
				pn.Text(10, balloonTail+6, "Turns the page.", AppFont, Black)
			})
		})
	})
}
