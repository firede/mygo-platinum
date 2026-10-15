package platinum

import (
	"github.com/egoist/mygo/ui"
)

// PageSetupState is what a page setup dialog chooses, kept by the app.
type PageSetupState struct {
	// Paper is an index of Papers; Landscape turns the page.
	Paper     int
	Landscape bool
	// Scale is the size of the printing, in percent; Margin the margin
	// around it, in points.
	Scale, Margin float64
}

// Papers are the sizes of paper a page setup dialog offers, in points.
var Papers = []struct {
	Name string
	W, H float64
}{{"US Letter", 612, 792}, {"US Legal", 612, 1008}, {"A4", 595, 842}, {"A5", 420, 595}, {"Envelope #10", 297, 684}}

// PageSetup shows a page setup dialog while *open: the paper, the
// orientation as two bevel buttons of pictures, the scale and margins, and
// a picture of the page as chosen. It edits a copy of *st and reports
// whether OK put it in *st.
func PageSetup(c *ui.Context, open *bool, st *PageSetupState) bool {
	ok := false
	draft := ui.Local(c.Root(), "pagesetup", func() PageSetupState { return *st })
	if !*open {
		*draft = *st
		return false
	}
	Dialog(c, open, say(c, "pageSetup"), 360, func() {
		d := draft
		ui.Row(c).Gap(16).AlignItems(ui.Start).Children(func() {
			ui.Column(c).Gap(8).Children(func() {
				names := make([]string, len(Papers))
				for i, p := range Papers {
					names[i] = p.Name
				}
				PopupMenu(c, &d.Paper, say(c, "paperLabel"), names)
				ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
					Label(c, say(c, "orientationLabel"))
					portrait, landscape := !d.Landscape, d.Landscape
					if IconButton(c, portraitIcon, say(c, "portrait")).Toggle(&portrait).Clicked() {
						d.Landscape = false
					}
					if IconButton(c, landscapeIcon, say(c, "landscape")).Toggle(&landscape).Clicked() {
						d.Landscape = true
					}
				})
				ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
					Label(c, say(c, "scaleLabel"))
					NumberField(c, &d.Scale, 25, 400, 5, 0).Label(say(c, "scale"))
					Label(c, "%")
				})
				ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
					Label(c, say(c, "marginsLabel"))
					NumberField(c, &d.Margin, 0, 144, 6, 0).Label(say(c, "margins"))
					Label(c, "pt")
				})
			})
			page := *d
			ui.Box(c).Size(90, 110).Role(ui.RoleImage).Label(say(c, "page")).Draw(func(p *ui.Painter, r ui.Rect) {
				DrawPagePreview(NewPen(p, r), page)
			})
		})
		ui.Row(c).Gap(10).Justify(ui.End).Children(func() {
			if Button(c, say(c, "cancel")).Cancel().Clicked() {
				*open = false
			}
			if Button(c, say(c, "ok")).Default().Clicked() {
				*st = *d
				*open = false
				ok = true
			}
		})
	})
	return ok
}

var (
	portraitIcon = []string{
		"######..",
		"#....##.",
		"#.....#.",
		"#.###.#.",
		"#.....#.",
		"#.###.#.",
		"#.....#.",
		"#######.",
	}
	landscapeIcon = []string{
		"##########",
		"#........#",
		"#.######.#",
		"#........#",
		"#.######.#",
		"##########",
	}
)

// DrawPagePreview draws a picture of a page set up as st, centered in pn's
// box, with a shadow, its margins dotted and gray lines for its text.
func DrawPagePreview(pn Pen, st PageSetupState) {
	paper := Papers[min(max(st.Paper, 0), len(Papers)-1)]
	pw, ph := paper.W, paper.H
	if st.Landscape {
		pw, ph = ph, pw
	}
	k := min(float64(pn.W-6)/pw, float64(pn.H-6)/ph)
	w, h := int(pw*k), int(ph*k)
	x, y := (pn.W-w)/2, (pn.H-h)/2
	pn.Rect(x+2, y+2, w, h, Gray7)
	pn.Rect(x, y, w, h, White)
	pn.Frame(x, y, w, h, Black)
	m := int(st.Margin * k)
	if m > 0 && 2*m < w && 2*m < h {
		for i := x + m; i < x+w-m; i += 2 {
			pn.Dot(i, y+m, Gray5)
			pn.Dot(i, y+h-m-1, Gray5)
		}
		for j := y + m; j < y+h-m; j += 2 {
			pn.Dot(x+m, j, Gray5)
			pn.Dot(x+w-m-1, j, Gray5)
		}
	}
	// Lines of text, as long as the scale lets them run.
	lineW := min(int(float64(w-2*m-4)*100/max(st.Scale, 1)), w-2*m-4)
	for j := y + m + 3; j < y+h-m-3; j += 3 {
		pn.HLine(x+m+2, j, lineW, Gray4)
	}
}
