package platinum

import (
	"image"
	"image/color"
	"testing"

	"github.com/egoist/mygo/ui"
)

// TestPixelExactRendering checks what the library rests on: rectangles on
// whole DIPs and pixelated bitmaps come out as exact pixels at a scale of
// 1, and as exact blocks at a scale of 2, with no blending at their edges.
func TestPixelExactRendering(t *testing.T) {
	checker := image.NewRGBA(image.Rect(0, 0, 2, 2))
	checker.SetRGBA(0, 0, color.RGBA{0, 0, 0, 255})
	checker.SetRGBA(1, 1, color.RGBA{0, 0, 0, 255})
	bmp := ui.NewBitmap(checker).Pixelated()
	view := func(c *ui.Context) {
		ui.Box(c).Fill().Background(Gray2).Draw(func(p *ui.Painter, r ui.Rect) {
			pn := NewPen(p, r)
			pn.Rect(10, 10, 20, 10, Black)
			pn.Dot(40, 10, White)
			p.ImageRegion(bmp, ui.Rect{W: 2, H: 2}, ui.Rect{X: 50, Y: 10, W: 4, H: 4}, ui.ImageOptions{})
		})
	}
	for _, scale := range []int{1, 2} {
		img := ui.Render(view, 64, 32, float32(scale))
		at := func(x, y int) color.RGBA { return img.RGBAAt(x*scale, y*scale) }
		for _, c := range []struct {
			x, y int
			want ui.Color
		}{
			{9, 10, Gray2}, {10, 10, Black}, {29, 19, Black}, {30, 19, Gray2}, {29, 20, Gray2},
			{40, 10, White}, {41, 10, Gray2}, {39, 10, Gray2},
			{50, 10, Black}, {51, 11, Black}, {52, 10, Gray2}, {52, 12, Black}, {53, 13, Black}, {50, 12, Gray2},
		} {
			if got := at(c.x, c.y); got != rgba(c.want) {
				t.Errorf("scale %d: pixel %d,%d is %v, want %v", scale, c.x, c.y, got, rgba(c.want))
			}
		}
		if scale == 2 {
			// The far corner of a block is the block's color too.
			if got := img.RGBAAt(2*29+1, 2*19+1); got != rgba(Black) {
				t.Errorf("scale 2: the block of 29,19 ends in %v", got)
			}
		}
	}
}

// TestFallbackCharacters measures and draws characters the faces lack in
// the system's font: they take room of their own and leave ink, rather
// than all taking the box's width. (Their pixels are the system's, so no
// golden image holds them.)
func TestFallbackCharacters(t *testing.T) {
	box := AppFont.face().Glyph(0xFFFD).Advance
	if w := AppFont.Measure("中文"); w <= 2*box-AppFont.face().Tracking || w > 40 {
		t.Errorf("two CJK characters measure %d, the box %d", w, box)
	}
	if AppFont.Has('中') {
		t.Error("the App face draws 中 itself?")
	}
	img := ui.Render(func(c *ui.Context) {
		ui.Box(c).Fill().Background(White).Draw(func(p *ui.Painter, r ui.Rect) {
			NewPen(p, r).Text(4, 4, "中", AppFont, Black)
		})
	}, 30, 20, 1)
	ink := 0
	for y := 4; y < 15; y++ {
		for x := 4; x < 16; x++ {
			if c := img.RGBAAt(x, y); c.R < 128 {
				ink++
			}
		}
	}
	if ink < 10 {
		t.Errorf("中 left %d dark pixels", ink)
	}
}

// TestAccentedLetters draws the letters of Latin-1 in both faces, in the
// faces' own pixels — a mark a row clear over a small letter, over a
// capital in the leading above the cell, a cedilla under the baseline —
// and the symbols of currency and text that go with them.
func TestAccentedLetters(t *testing.T) {
	freeze(t)
	for _, f := range []Font{SystemFont, AppFont} {
		for _, r := range "ÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓÔÕÖØÙÚÛÜÝÞßàáâãäåæçèéêëìíîïðñòóôõöøùúûüýþÿŒœŁłĄąĘęĆćŃńŚśŹźŻżČčĚěŇňŘřŠšŽžŐőŰűĞğŞşŮůĎďŤťİıĀāĒēĪīŪūȘșȚțĶķĻļŅņ" {
			if !f.Has(r) {
				t.Errorf("the %s face lacks %q", f.face().Name, r)
			}
		}
	}
	name := "Café"
	notes := "ÉLÈVE, Über, Ångström, Æble, Þór\nfaçade, naïve, señor, Straße, smørrebrød, cœur, ðað"
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(8).Gap(8).AlignItems(ui.Start).Children(func() {
			ui.Box(c).Size(384, 140).Draw(func(p *ui.Painter, r ui.Rect) {
				pn := NewPen(p, r)
				pn.Text(0, 46, "€5 £5 ¥5 ¢5 ® ™ ±1 §2 ¶ ¿Qué? ¡Sí! «oui» ‹a› a·b 5µm ¬ ´ ¨ ¸", SystemFont, Black)
				pn.Text(0, 60, "€5 £5 ¥5 ¢5 ® ™ ±1 §2 ¶ ¿Qué? ¡Sí! «oui» ‹a› a·b 5µm ¬ ´ ¨ ¸", AppFont, Black)
				pn.Text(0, 76, "Łódź, Żółć, Český Krumlov, Říčany, Győr, Ğüneş, Ąę, ďť İı", SystemFont, Black)
				pn.Text(0, 92, "Łódź, Żółć, Český Krumlov, Říčany, Győr, Ğüneş, Ąę, ĺŕ ďť ĎŤ İı", AppFont, Black)
				pn.Text(0, 108, "Iași, Țară; Rīga, Ķemeri, Liepāja, Ņ", SystemFont, Black)
				pn.Text(0, 124, "Iași, Țară; Rīga, Ķemeri, Liepāja, Ņ", AppFont, Black)
				pn.Text(0, 4, "ÀÁÂÃÄÅÆ ÇÈÉÊË ÌÍÎÏ Ñ ÒÓÔÕÖØŒ ÙÚÛÜ Ý", SystemFont, Black)
				pn.Text(0, 18, "àáâãäåæ çèéêë ìíîï ñ òóôõöøœ ùúûü ýÿ ß Ðð Þþ", SystemFont, Black)
				pn.Text(0, 32, "ÀÁÂÃÄÅÆ ÇÈÉÊË ÌÍÎÏ Ñ ÒÓÔÕÖØŒ ÙÚÛÜ Ý àáâãäåæ çèéêë ìíîï ñ òóôõöøœ ùúûü ýÿ ß Ðð Þþ", AppFont, Black)
			})
			Button(c, "Résumé")
			TextField(c, &name).Label("Name").Width(120)
			TextArea(c, &notes).Label("Notes").Element().Size(300, 40)
		})
	}
	tt := ui.NewTester(view, 400, 278)
	tt.Frame()
	compareGolden(t, "accents", tt.Image())
}

// TestTextWholePixelsAt2x draws accented letters, with their marks over
// the cell, and rich text at four sizes, styled, at scale 1 and 2: every
// pixel at 2 must be a 2×2 block of the one at 1.
func TestTextWholePixelsAt2x(t *testing.T) {
	freeze(t)
	doc := NewRichText("")
	doc.SetRuns([]Run{{Text: "Big", Style: Bold.Scaled(2)}, {Text: " ital", Style: Italic}, {Text: " ŁÓDŹ", Style: Underline.Scaled(3)}})
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(8).Gap(6).AlignItems(ui.Start).Children(func() {
			ui.Box(c).Size(300, 36).Draw(func(p *ui.Painter, r ui.Rect) {
				pn := NewPen(p, r)
				pn.Text(0, 4, "ÉLÈVE Åå Ğüneş Ąę Ťť", SystemFont, Black)
				pn.Text(0, 20, "ÉLÈVE Åå Ğüneş Ąę Ťť ȘȚ", AppFont, Black)
			})
			RichTextArea(c, doc).Label("Document").Element().Size(300, 80)
		})
	}
	one := ui.Render(view, 320, 140, 1)
	two := ui.Render(view, 320, 140, 2)
	for y := 0; y < 140; y++ {
		for x := 0; x < 320; x++ {
			c := one.RGBAAt(x, y)
			for dy := 0; dy < 2; dy++ {
				for dx := 0; dx < 2; dx++ {
					if d := two.RGBAAt(2*x+dx, 2*y+dy); d != c {
						t.Fatalf("at 2x, (%d,%d) is %v; at 1x (%d,%d) is %v", 2*x+dx, 2*y+dy, d, x, y, c)
					}
				}
			}
		}
	}
}
