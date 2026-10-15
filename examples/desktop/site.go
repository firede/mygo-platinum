package main

import (
	"image"
	"path/filepath"

	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
)

// sitePictures returns the pictures of the desk's web page by the names of
// their files: the pictures it is shared with; the desk in each of its
// languages, for screens too small to run it; and, drawn once at a scale
// of 1 for the page to scale without smoothing, the pattern of its
// desktop, the icons of its menu bar and the frames of a window, buttons,
// a pop-up menu, a text field and a progress bar, which the page's style sheet stretches around its
// own content (border-image).
func sitePictures() map[string]*image.RGBA {
	pics := map[string]*image.RGBA{}
	for _, pic := range sharePictures {
		pics[pic.name+".png"] = renderShare(pic)
	}
	for _, lang := range languages {
		pics["scene-"+lang+".png"] = renderScene(2, lang)
	}
	a := pt.DefaultTheme().Accent
	pat := patterns[0]
	pics["pattern.png"] = drawn(8, 8, 1, func(pn pt.Pen) {
		pn.Pattern(0, 0, 8, 8, pat.bits, ui.RGB(pat.fg.R, pat.fg.G, pat.fg.B), ui.RGB(pat.bg.R, pat.bg.G, pat.bg.B))
	})
	// The icons of the menu bar, dimmed: before the desk starts, its menus
	// wait, as under a modal dialog.
	pics["logo.png"] = drawn(len(deskLogo[0]), len(deskLogo), 1, func(pn pt.Pen) { pt.DrawIcon(pn.Dimmed(true), deskLogo, a, false) })
	pics["desk.png"] = drawn(len(miniDesk[0]), len(miniDesk), 1, func(pn pt.Pen) { pt.DrawIcon(pn.Dimmed(true), miniDesk, a, false) })
	pics["icon-32.png"] = drawn(32, 32, 1, func(pn pt.Pen) {
		pt.DrawIcon(pn.Sub(1, 1, 30, 30), scaleRows(deskLogo, 3), a, false)
	})
	pics["icon-180.png"] = drawn(180, 180, 1, func(pn pt.Pen) {
		pn.Rect(0, 0, 180, 180, pt.Gray2)
		pt.DrawIcon(pn.Sub(30, 30, 120, 120), scaleRows(deskLogo, 12), a, false)
	})
	pics["window.png"] = drawn(64, 64, 1, func(pn pt.Pen) {
		pt.DrawWindowFrame(pn, "", pt.WindowOptions{Active: true, NoClose: true, NoZoom: true, NoCollapse: true, NoGrow: true}, false, pt.PartNone)
	})
	pics["button.png"] = drawn(24, 20, 1, func(pn pt.Pen) { pt.DrawButton(pn, "", pt.StateNormal) })
	pics["button-pressed.png"] = drawn(24, 20, 1, func(pn pt.Pen) { pt.DrawButton(pn, "", pt.StatePressed) })
	const ring = pt.DefaultRing
	pics["button-default.png"] = drawn(24+2*ring, 20+2*ring, 1, func(pn pt.Pen) {
		pt.DrawDefaultRing(pn, 0, 0, pn.W, pn.H, pt.StateNormal)
		pt.DrawButton(pn.Sub(ring, ring, 24, 20), "", pt.StateNormal)
	})
	pics["button-default-pressed.png"] = drawn(24+2*ring, 20+2*ring, 1, func(pn pt.Pen) {
		pt.DrawDefaultRing(pn, 0, 0, pn.W, pn.H, pt.StateNormal)
		pt.DrawButton(pn.Sub(ring, ring, 24, 20), "", pt.StatePressed)
	})
	pics["popup.png"] = drawn(40, pt.PopupHeight, 1, func(pn pt.Pen) { pt.DrawPopupButton(pn, "", pt.StateNormal) })
	pics["popup-pressed.png"] = drawn(40, pt.PopupHeight, 1, func(pn pt.Pen) { pt.DrawPopupButton(pn, "", pt.StatePressed) })
	pics["field.png"] = drawn(12, 12, 1, func(pn pt.Pen) { pt.DrawTextFrame(pn, false) })
	pics["button-default-disabled.png"] = drawn(24+2*ring, 20+2*ring, 1, func(pn pt.Pen) {
		pt.DrawDefaultRing(pn, 0, 0, pn.W, pn.H, pt.StateDisabled)
		pt.DrawButton(pn.Sub(ring, ring, 24, 20), "", pt.StateDisabled)
	})
	pics["progress.png"] = drawn(24, 12, 1, func(pn pt.Pen) { pt.DrawProgress(pn, 0, a) })
	pics["progress-full.png"] = drawn(24, 12, 1, func(pn pt.Pen) { pt.DrawProgress(pn, 1, a) })
	return pics
}

// writeSite writes the pictures of the desk's web page to dir.
func writeSite(dir string) {
	for name, img := range sitePictures() {
		savePNG(filepath.Join(dir, name), img)
	}
}

// drawn returns a w×h picture of what draw draws, at scale, on nothing.
func drawn(w, h int, scale float32, draw func(pn pt.Pen)) *image.RGBA {
	tt := ui.NewTester(func(c *ui.Context) {
		c.Root().Background(ui.Transparent)
		ui.Box(c).Fill().Draw(func(p *ui.Painter, r ui.Rect) { draw(pt.NewPen(p, r)) })
	}, w, h)
	tt.SetScale(scale)
	tt.Frame()
	return tt.Image()
}

// scaleRows returns the rows of an icon with each of its pixels k×k.
func scaleRows(rows []string, k int) []string {
	var out []string
	for _, row := range rows {
		var b []byte
		for i := range len(row) {
			for range k {
				b = append(b, row[i])
			}
		}
		for range k {
			out = append(out, string(b))
		}
	}
	return out
}
