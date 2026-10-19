package platinum

import (
	"math"
	"sync"

	"github.com/egoist/mygo/ui"
)

// Font is one of the two bitmap faces of the library, from platinum-fonts:
// SystemFont, Platinum Sans Bold, for menus, buttons, labels and titles,
// and AppFont, Platinum Sans Small, for documents, lists and the names of
// icons. East Asian characters are drawn with the faces SetEastAsian
// gives, and other characters the faces lack in the system's font.
type Font int

// The two faces of the library.
const (
	SystemFont Font = iota
	AppFont
)

func (f Font) face() *face { return faces().regular[f] }

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

// MeasureStyled returns the width of s on one line in a style, as
// TextStyled draws it: bold glyphs, the room italic takes at the end of the
// run, and the scale; without the tracking after the last glyph, as Measure.
func (f Font) MeasureStyled(s string, style TextStyle) int {
	bold := style&Bold != 0
	k := style.Scale()
	w := 0
	for _, r := range s {
		if g, ok := f.glyph(r, bold); ok {
			w += g.advance * k
		} else {
			w += (f.advance(r) + boolInt(bold)) * k
		}
	}
	if w == 0 {
		return 0
	}
	if style&Italic != 0 {
		w += ItalicSlant(f) * k
	}
	return w - f.face().Tracking*k
}

// advance returns how far the pen moves after r: its glyph's advance, or
// for a character no face draws, the advance of the system's font that
// draws it instead.
func (f Font) advance(r rune) int {
	if g, ok := f.glyph(r, false); ok {
		return g.advance
	}
	return fallback(f, r).advance
}

// The characters no face draws, drawn in the system's font: a character of
// a script the faces do not cover shows as itself rather than as a box.
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
	g, ok := f.glyph(r, false)
	if !ok {
		fg := fallback(f, r)
		if fg.glyphs != nil {
			pn.P.Glyphs(fg.glyphs, float32(pn.X+x), float32(pn.Y+y+f.Ascent()), pn.unfade(c))
			return fg.advance
		}
		g, _ = f.glyph(f.face().Fallback(), false) // the box
	}
	if r != ' ' {
		pn.P.ImageRegion(g.bmp,
			ui.Rect{X: float32(g.x), Y: 0, W: float32(g.w), H: float32(g.rows)},
			ui.Rect{X: float32(pn.X + x + g.left), Y: float32(pn.Y + y + f.Ascent() - g.up), W: float32(g.w), H: float32(g.rows)},
			ui.ImageOptions{Tint: pn.unfade(c)})
	}
	return g.advance
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

// Has reports whether the font's own face draws r; East Asian characters
// draw with the faces SetEastAsian gives, and others in the system's font,
// or as a box where it has none.
func (f Font) Has(r rune) bool { return f.face().Has(r) }

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
// Characters no face draws show as the face's box.
func (pn Pen) TextScaled(x, y int, s string, f Font, c ui.Color, k int) int {
	for _, r := range s {
		g, ok := f.glyph(r, false)
		if !ok {
			g, _ = f.glyph(f.face().Fallback(), false)
		}
		if r != ' ' {
			pn.P.ImageRegion(g.bmp,
				ui.Rect{X: float32(g.x), Y: 0, W: float32(g.w), H: float32(g.rows)},
				ui.Rect{X: float32(pn.X + x + g.left*k), Y: float32(pn.Y + y + (f.Ascent()-g.up)*k), W: float32(g.w * k), H: float32(g.rows * k)},
				ui.ImageOptions{Tint: pn.unfade(c)})
		}
		x += g.advance * k
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

// TextIn builds a line of text in font f and color ink, sized to it. Made
// taller, as tall as a control beside it, it centers its capitals on its
// middle, as the control centers its own text.
func TextIn(c *ui.Context, text string, f Font, ink ui.Color) ui.Element {
	el := ui.Box(c).Size(float32(f.Measure(text)), float32(f.Height())).Role(ui.RoleText).Label(text)
	el.Draw(func(p *ui.Painter, r ui.Rect) {
		pn := NewPen(p, r)
		if pn.H > f.Height() {
			pn.TextMiddle(0, 0, pn.H, text, f, ink)
			return
		}
		pn.Text(0, 0, text, f, ink)
	})
	return el
}

// TextStyled draws s as Text does, in style: bold in the bold faces, each
// glyph drawn again a pixel to its right and a pixel wider, as classic
// bitmap text was made bold; italic slanted to the right, each pair of
// rows a pixel right of the pair below, with the slant's room added after
// the run, East Asian characters staying upright; underlined with a line a
// pixel under the baseline; and at the style's scale, each pixel of all
// that k by k. Characters no face draws are drawn upright in the system's
// font. It returns the x after it.
func (pn Pen) TextStyled(x, y int, s string, f Font, c ui.Color, style TextStyle) int {
	face := f.face()
	bold, italic := style&Bold != 0, style&Italic != 0
	k := style.Scale()
	start := x
	for _, r := range s {
		g, ok := f.glyph(r, bold)
		if !ok {
			// In the system's font, at its own size, on the baseline.
			pn.drawGlyph(x, y+(k-1)*face.Ascent, r, f, c)
			if bold {
				pn.drawGlyph(x+1, y+(k-1)*face.Ascent, r, f, c)
			}
			x += (f.advance(r) + boolInt(bold)) * k
			continue
		}
		// East Asian scripts have no italic: their glyphs stay upright.
		slant := italic && !g.eastAsian
		if r != ' ' {
			// The rows drawn at a time, from the top of the cell: all of
			// the glyph's, or pairs, slanted.
			top, end := face.Ascent-g.up, face.Ascent-g.up+g.rows
			first, step := top, g.rows
			if slant {
				first, step = -(-top+1)/2*2, 2
			}
			for row := first; row < end; row += step {
				from := max(row, top)
				h := min(row+step, end) - from
				if h <= 0 {
					continue
				}
				shift := 0
				if slant {
					shift = (face.Ascent - row) / 2
				}
				src := ui.Rect{X: float32(g.x), Y: float32(from - top), W: float32(g.w), H: float32(h)}
				dst := ui.Rect{X: float32(pn.X + x + (g.left+shift)*k), Y: float32(pn.Y + y + from*k), W: float32(g.w * k), H: float32(h * k)}
				pn.P.ImageRegion(g.bmp, src, dst, ui.ImageOptions{Tint: pn.unfade(c)})
			}
		}
		x += g.advance * k
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
