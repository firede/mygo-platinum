package platinum

import (
	"math"
	"sync"

	"github.com/egoist/mygo/ui"
	"github.com/firede/mygo-platinum/internal/pixfont"
)

// Font is one of the two bitmap faces of the library, both drawn for it:
// SystemFont, bold, for menus, buttons, labels and titles, and AppFont,
// regular, for documents, lists and the names of icons.
type Font int

// The two faces of the library.
const (
	SystemFont Font = iota
	AppFont
)

func (f Font) face() *pixfont.Face {
	if f == AppFont {
		return pixfont.App
	}
	return pixfont.System
}

// Height returns the height of the font's lines, without leading.
func (f Font) Height() int { return f.face().Height }

// Ascent returns the rows of the font above its baseline.
func (f Font) Ascent() int { return f.face().Ascent }

// CapHeight returns the height of the font's capitals.
func (f Font) CapHeight() int { return f.face().CapHeight }

// Measure returns the width of s on one line, in pixels.
func (f Font) Measure(s string) int {
	w := 0
	for _, r := range s {
		w += f.advance(r)
	}
	if w > 0 {
		w -= f.face().Tracking // no tracking after the last glyph
	}
	return w
}

// advance returns how far the pen moves after r: its glyph's advance, or
// for a character the face lacks, as of a script it does not cover, the
// advance of the system's font that draws it instead.
func (f Font) advance(r rune) int {
	face := f.face()
	if g, ok := face.Glyphs[r]; ok {
		return g.Advance
	}
	return fallback(f, r).advance
}

// The characters the faces lack, drawn in the system's font: a CJK
// character or an accented letter shows as itself rather than as a box.
type fallbackGlyph struct {
	glyphs  []ui.Glyph
	advance int
}

var (
	fallbackMu    sync.Mutex
	fallbackCache = map[[2]int]fallbackGlyph{}
)

func (f Font) systemFont() ui.Font {
	if f == SystemFont {
		return ui.Font{Size: 12, Weight: 700}
	}
	return ui.Font{Size: 11}
}

func fallback(f Font, r rune) fallbackGlyph {
	key := [2]int{int(f), int(r)}
	fallbackMu.Lock()
	g, ok := fallbackCache[key]
	fallbackMu.Unlock()
	if ok {
		return g
	}
	glyphs := ui.Shape(string(r), f.systemFont())
	adv := float32(0)
	for _, gl := range glyphs {
		adv += gl.Advance
	}
	g = fallbackGlyph{glyphs: glyphs, advance: int(math.Ceil(float64(adv))) + f.face().Tracking}
	if len(glyphs) == 0 || adv <= 0 {
		// Nothing draws it: the face's box.
		g = fallbackGlyph{advance: f.face().Glyph(r).Advance}
	}
	fallbackMu.Lock()
	fallbackCache[key] = g
	fallbackMu.Unlock()
	return g
}

// drawGlyph draws r with its cell's top at x, y, and returns its advance.
func (pn Pen) drawGlyph(x, y int, r rune, f Font, c ui.Color) int {
	face := f.face()
	g, ok := face.Glyphs[r]
	if !ok {
		fg := fallback(f, r)
		if fg.glyphs == nil {
			g = face.Glyph(r) // the box
		} else {
			pn.P.Glyphs(fg.glyphs, float32(pn.X+x), float32(pn.Y+y+face.Ascent), pn.unfade(c))
			return fg.advance
		}
	}
	if r != ' ' {
		// The cell, with the rows over it a mark over a capital takes.
		h := float32(face.Above + face.Height)
		pn.P.ImageRegion(atlas(f),
			ui.Rect{X: float32(g.X), Y: 0, W: float32(g.Width), H: h},
			ui.Rect{X: float32(pn.X + x), Y: float32(pn.Y + y - face.Above), W: float32(g.Width), H: h},
			ui.ImageOptions{Tint: pn.unfade(c)})
	}
	return g.Advance
}

// Truncate returns s cut with an ellipsis to fit in width pixels, or s as it
// is when it fits.
func (f Font) Truncate(s string, width int) string {
	if f.Measure(s) <= width {
		return s
	}
	face := f.face()
	ell := face.Glyph('…').Advance
	w := 0
	for i, r := range s {
		a := f.advance(r)
		if w+a+ell-face.Tracking > width {
			return s[:i] + "…"
		}
		w += a
	}
	return s
}

// Has reports whether the font draws r itself; others draw in the
// system's font, or as a box where it has none.
func (f Font) Has(r rune) bool { return f.face().Has(r) }

var (
	atlasOnce sync.Once
	atlases   [2]*ui.Bitmap
)

func atlas(f Font) *ui.Bitmap {
	atlasOnce.Do(func() {
		atlases[SystemFont] = ui.NewBitmap(pixfont.System.Atlas).Pixelated()
		atlases[AppFont] = ui.NewBitmap(pixfont.App.Atlas).Pixelated()
	})
	return atlases[f]
}

// Text draws s with its cell's top at x, y, in c, and returns the x after
// it.
func (pn Pen) Text(x, y int, s string, f Font, c ui.Color) int {
	for _, r := range s {
		x += pn.drawGlyph(x, y, r, f, c)
	}
	return x
}

// TextCentered draws s centered in the w×h box at x, y, its capitals
// centered on the box's middle.
func (pn Pen) TextCentered(x, y, w, h int, s string, f Font, c ui.Color) {
	tw := f.Measure(s)
	ty := y + (h-f.CapHeight())/2 - (f.Ascent() - f.CapHeight())
	pn.Text(x+(w-tw)/2, ty, s, f, c)
}

// TextMiddle draws s from x, its capitals centered on the middle of the h
// pixels from y.
func (pn Pen) TextMiddle(x, y, h int, s string, f Font, c ui.Color) int {
	ty := y + (h-f.CapHeight())/2 - (f.Ascent() - f.CapHeight())
	return pn.Text(x, ty, s, f, c)
}

// TextScaled draws s as Text does, each pixel of its glyphs a k×k block,
// for large type such as a clock's, and returns the x after it.
func (pn Pen) TextScaled(x, y int, s string, f Font, c ui.Color, k int) int {
	face := f.face()
	bmp := atlas(f)
	h := float32(face.Above + face.Height)
	for _, r := range s {
		g := face.Glyph(r)
		if r != ' ' {
			pn.P.ImageRegion(bmp,
				ui.Rect{X: float32(g.X), Y: 0, W: float32(g.Width), H: h},
				ui.Rect{X: float32(pn.X + x), Y: float32(pn.Y + y - face.Above*k), W: float32(g.Width * k), H: h * float32(k)},
				ui.ImageOptions{Tint: pn.unfade(c)})
		}
		x += g.Advance * k
	}
	return x
}

// Paragraph builds text that wraps at width pixels, in font f.
func Paragraph(c *ui.Context, text string, f Font, width int) ui.Element {
	e := &editor{font: f, multi: true, wrap: width, text: []rune(text)}
	e.layout()
	lineH := f.Height() + 2
	el := ui.Box(c).Size(float32(width), float32(len(e.lines)*lineH)).Role(ui.RoleText).Label(text)
	el.Draw(func(p *ui.Painter, r ui.Rect) {
		pn := NewPen(p, r)
		for i, l := range e.lines {
			pn.Text(0, i*lineH, string(e.text[l.start:l.end]), f, Black)
		}
	})
	return el
}

// Label builds a line of text in the bold face, sized to it.
func Label(c *ui.Context, text string) ui.Element {
	return TextIn(c, text, SystemFont, Black)
}

// TextIn builds a line of text in font f and color ink, sized to it.
func TextIn(c *ui.Context, text string, f Font, ink ui.Color) ui.Element {
	el := ui.Box(c).Size(float32(f.Measure(text)), float32(f.Height())).Role(ui.RoleText).Label(text)
	el.Draw(func(p *ui.Painter, r ui.Rect) {
		NewPen(p, r).Text(0, 0, text, f, ink)
	})
	return el
}

// TextStyled draws s as Text does, in style: bold as classic bitmap text
// was made bold, each glyph drawn again a pixel to its right and a pixel wider;
// italic slanted to the right, each pair of rows a pixel right of the pair
// below, with the slant's room added after the run; underlined with a line
// a pixel under the baseline; and at the style's scale, each pixel of all
// that k by k. It returns the x after it.
func (pn Pen) TextStyled(x, y int, s string, f Font, c ui.Color, style TextStyle) int {
	face := f.face()
	bmp := atlas(f)
	bold, italic := style&Bold != 0, style&Italic != 0
	k := style.Scale()
	start := x
	for _, r := range s {
		if !face.Has(r) {
			// In the system's font, at its own size, on the baseline.
			pn.drawGlyph(x, y+(k-1)*face.Ascent, r, f, c)
			if bold {
				pn.drawGlyph(x+1, y+(k-1)*face.Ascent, r, f, c)
			}
			x += (f.advance(r) + boolInt(bold)) * k
			continue
		}
		g := face.Glyph(r)
		if r != ' ' {
			// The rows drawn at a time: all, the rows over the cell a mark
			// takes with them, or pairs, slanted.
			first, step := -face.Above, face.Above+face.Height
			if italic {
				first, step = -(face.Above+1)/2*2, 2
			}
			for b := 0; b <= boolInt(bold); b++ {
				for row := first; row < face.Height; row += step {
					from := max(row, -face.Above)
					h := min(row+step, face.Height) - from
					shift := 0
					if italic {
						shift = (face.Ascent - row) / 2
					}
					src := ui.Rect{X: float32(g.X), Y: float32(face.Above + from), W: float32(g.Width), H: float32(h)}
					dst := ui.Rect{X: float32(pn.X + x + (shift+b)*k), Y: float32(pn.Y + y + from*k), W: float32(g.Width * k), H: float32(h * k)}
					pn.P.ImageRegion(bmp, src, dst, ui.ImageOptions{Tint: pn.unfade(c)})
				}
			}
		}
		x += (g.Advance + boolInt(bold)) * k
	}
	if italic {
		x += ItalicSlant(f) * k
	}
	if style&Underline != 0 {
		pn.Rect(start, y+(face.Ascent+1)*k, x-start-face.Tracking*k, k, c)
	}
	return x
}

// ItalicSlant returns the room italic text of f takes beyond its glyphs'
// advances, at the end of each run: how far its top row leans right.
func ItalicSlant(f Font) int { return f.face().Ascent / 2 }

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
