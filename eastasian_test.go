package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/firede/platinum-fonts/cjk12/ja"
	"github.com/firede/platinum-fonts/cjk12/zhhans"
)

// useEastAsian sets the East Asian faces for the test, and unsets them
// after it.
func useEastAsian(t *testing.T) {
	SetEastAsian(zhhans.Regular, zhhans.Bold)
	t.Cleanup(func() { SetEastAsian(nil, nil) })
}

// TestEastAsianMeasure checks that the faces SetEastAsian gives measure
// East Asian characters: 12 pixels each, 13 in bold, the System face's.
func TestEastAsianMeasure(t *testing.T) {
	useEastAsian(t)
	if w := AppFont.Measure("中文"); w != 23 {
		t.Errorf("中文 measures %d in the App face, want 23", w)
	}
	if w := SystemFont.Measure("中文"); w != 25 {
		t.Errorf("中文 measures %d in the System face, want 25", w)
	}
	if w := AppFont.Measure("A中"); w != AppFont.Measure("A")+12 {
		t.Errorf("A中 measures %d", w)
	}
	SetEastAsian(nil, nil)
	if g, ok := AppFont.glyph('中', false); ok {
		t.Errorf("中 has a glyph %v with no East Asian faces", g)
	}
}

// TestEastAsianPixels draws characters with the faces SetEastAsian gives
// and compares each pixel with the face's own: black where its glyph is
// ink, white elsewhere, nothing in between.
func TestEastAsianPixels(t *testing.T) {
	useEastAsian(t)
	for _, tc := range []struct {
		f    Font
		bold bool
	}{{AppFont, false}, {SystemFont, true}} {
		const text = "中文鬱"
		img := ui.Render(func(c *ui.Context) {
			Use(c, nil)
			ui.Box(c).Fill().Background(White).Draw(func(p *ui.Painter, r ui.Rect) {
				NewPen(p, r).Text(4, 4, text, tc.f, Black)
			})
		}, 60, 24, 1)
		src := zhhans.Regular()
		if tc.bold {
			src = zhhans.Bold()
		}
		x := 4
		for _, r := range text {
			g, _ := src.Glyph(r)
			base := 4 + tc.f.Ascent()
			for y := g.Bounds.Min.Y - 1; y <= g.Bounds.Max.Y; y++ {
				for gx := g.Bounds.Min.X - 1; gx <= g.Bounds.Max.X; gx++ {
					c := img.RGBAAt(x+gx, base+y)
					want := uint8(0xFF)
					if g.At(gx, y) {
						want = 0
					}
					if c.R != want || c.G != want || c.B != want {
						t.Fatalf("%v: %q at %d, %d is %v, want %d", tc.f, r, gx, y, c, want)
					}
				}
			}
			x += g.Advance
		}
	}
}

// TestSnapshotEastAsian draws East Asian text in both faces, styled, in
// another language's forms, and scaled, beside Latin text.
func TestSnapshotEastAsian(t *testing.T) {
	freeze(t)
	useEastAsian(t)
	view := func(c *ui.Context) {
		ui.Column(c).Fill().Background(Gray2).Padding(4, 6).Gap(6).AlignItems(ui.Start).Children(func() {
			Label(c, "System 系统设置：显示")
			TextIn(c, "App 文稿，3 个项目。", AppFont, Black)
			ui.Box(c).Size(188, 44).Draw(func(p *ui.Painter, r ui.Rect) {
				pn := NewPen(p, r)
				x := pn.TextStyled(0, 0, "粗体 ", AppFont, Black, Bold)
				x = pn.TextStyled(x, 0, "斜体 ", AppFont, Black, Italic)
				pn.TextStyled(x+4, 0, "下划线", AppFont, Black, Underline)
				pn.TextScaled(0, 18, "12:30 中", SystemFont, Black, 2)
			})
		})
	}
	snap(t, "eastasian", 200, 90, view)
	SetEastAsian(ja.Regular, ja.Bold)
	snap(t, "eastasian-ja", 200, 90, view)
}

// TestEastAsianCaret checks a line of text with East Asian characters, at
// each scale: its glyphs inside the line, and the caret reaching as far up and down as
// they do and as the Latin glyphs do, as it does without them.
func TestEastAsianCaret(t *testing.T) {
	check := func(text string, k int) {
		t.Helper()
		e := &editor{font: AppFont, multi: true, rich: true, text: []rune(text)}
		for range e.text {
			e.styles = append(e.styles, TextStyle(0).Scaled(k))
		}
		e.caret = len(e.text)
		e.layout()
		c := e.caretRect(0, 0, TextLineHeight, 0, 0)
		lineH := e.lineHeight(0, TextLineHeight)
		base := e.lineTextTop(0, TextLineHeight) + AppFont.Ascent()*k
		fc := AppFont.face()
		up, down := fc.Ascent, fc.Height-fc.Ascent
		for _, r := range text {
			if g, ok := AppFont.glyph(r, false); ok && g.eastAsian {
				up, down = max(up, g.up), max(down, g.rows-g.up)
			}
		}
		up, down = up*k, down*k
		if base-up < 0 || base+down > lineH {
			t.Errorf("%q at %d×: the glyphs reach %d to %d, out of the line 0 to %d", text, k, base-up, base+down, lineH)
		}
		if int(c.Y) > base-up || int(c.Y+c.H) < base+down {
			t.Errorf("%q at %d×: the caret runs %v to %v, the glyphs %d to %d", text, k, c.Y, c.Y+c.H, base-up, base+down)
		}
	}
	for k := 1; k <= MaxScale; k++ {
		check("Hg", k)
	}
	useEastAsian(t)
	for k := 1; k <= MaxScale; k++ {
		check("搅匀，用黄油煎。", k)
		check("Hg", k)
	}
}
