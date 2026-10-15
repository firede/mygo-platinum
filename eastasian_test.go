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
