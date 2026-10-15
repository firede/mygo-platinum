package main

import (
	"image"

	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
)

// A picture to share the desk with, as the card a link to it shows: w×h
// DIPs drawn at scale.
type sharePicture struct {
	name  string
	w, h  int
	scale float32
}

// sharePictures are the pictures a page of the desk is shared with: the
// card of a link (1200×630) and the preview of a repository (1280×640).
var sharePictures = []sharePicture{
	{"card", 400, 210, 3},
	{"preview", 426, 213, 3},
}

// shareScene returns the view of a desk of the size of pic: the calculator
// open, at 42, behind a window of the desk's name, four times the size of
// the system font, and what it is.
func shareScene(pic sharePicture) func(c *ui.Context) {
	d := scene("")
	d.size = [2]int{pic.w, pic.h - pt.MenuBarHeight}
	d.windows = d.windows[:0]
	k := d.launch(appCalculator)
	k.frame.X, k.frame.Y = 262, 76
	for _, key := range []string{"6", "×", "7", "="} {
		k.calc.press(key)
	}
	d.covered = true
	const word, big = "Platinum", 4
	welcome := pt.WindowState{X: 18, Y: 40, W: 282}
	return func(c *ui.Context) {
		d.view(c)
		ui.Box(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).Children(func() {
			pt.Window(c.Key("welcome"), &welcome, "", pt.WindowOptions{Active: true, NoGrow: true, NoZoom: true, NoCollapse: true}, func() {
				ui.Column(c).Padding(14).Gap(9).Children(func() {
					f := pt.SystemFont
					ui.Box(c).Size(float32(f.Measure(word)*big), float32((f.Ascent()+1)*big)).Draw(func(p *ui.Painter, r ui.Rect) {
						pt.NewPen(p, r).TextScaled(0, 0, word, f, pt.Black, big)
					})
					pt.TextIn(c, "Classic desktop controls for MyGo", pt.SystemFont, pt.Black)
					pt.TextIn(c, "English · 简体中文 · 繁體中文 · 日本語 · 한국어", pt.AppFont, pt.Gray9)
				})
			})
		})
	}
}

// renderShare draws pic.
func renderShare(pic sharePicture) *image.RGBA {
	tt := ui.NewTester(shareScene(pic), pic.w, pic.h)
	tt.SetScale(pic.scale)
	tt.Frame()
	return tt.Image()
}
