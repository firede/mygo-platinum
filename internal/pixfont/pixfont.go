// Package pixfont parses the hand-drawn bitmap faces used by the Platinum
// widgets and packs them into an atlas.
//
// A face is described in a plain text format so that its glyphs can be read
// and edited as pictures:
//
//	@height 12
//	@ascent 9
//	== A
//	..###..
//	.##.##.
//
// Every glyph starts with a line "== X", where X is the character, or a
// hexadecimal code point written as U+XXXX, or "space N" for a blank glyph N
// pixels wide. The rows that follow draw it from the top of the cell: '#' is
// ink and any other character is paper. Blank lines are ignored, so a blank
// row is written as dots; rows missing at the bottom are blank, and the glyph
// is as wide as its longest row.
//
// The combining marks, U+0300 to U+036F, are drawn as their shape alone and
// are not glyphs of their own: the accented letters of Latin-1 are made of
// a letter and a mark, the mark centered a row clear of the letter's top
// (an i losing its dot), or, for the cedilla, under the baseline. Marks over
// capitals rise above the cell, into the leading, as Above says, and sit
// on the capital, at most MaxAbove rows over the cell.
package pixfont

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"
)

// Glyph is one character of a face, placed in the face's atlas.
type Glyph struct {
	// X is where the glyph starts in the atlas; it is Width pixels wide and
	// the face's Height tall.
	X, Width int
	// Advance is how far the pen moves after the glyph.
	Advance int
}

// Face is a parsed bitmap face.
type Face struct {
	Name string
	// Height is the height of every glyph cell, Ascent the rows above the
	// baseline, and Tracking the blank columns added after every glyph.
	Height, Ascent, Tracking int
	// XHeight and CapHeight describe the design, in rows above the
	// baseline.
	XHeight, CapHeight int
	// Above is the rows over the cell that marks over capitals take: the
	// atlas has them over every glyph's cell.
	Above  int
	Glyphs map[rune]Glyph
	// Atlas holds every glyph in a row, white ink on transparent paper,
	// each cell Above rows down from its top.
	Atlas *image.RGBA
	// Fallback is drawn for characters the face lacks.
	Fallback rune
	rows     map[rune][]string
}

// Parse reads a face from its text description.
func Parse(name, src string) (*Face, error) {
	f := &Face{Name: name, Tracking: 1, Glyphs: map[rune]Glyph{}, rows: map[rune][]string{}}
	var order []rune
	var cur rune = -1
	widths := map[rune]int{}
	sc := bufio.NewScanner(strings.NewReader(src))
	line := 0
	for sc.Scan() {
		line++
		s := strings.TrimRight(sc.Text(), " \t\r")
		switch {
		case s == "":
			continue
		case strings.HasPrefix(s, "@"):
			fields := strings.Fields(s[1:])
			if len(fields) != 2 {
				return nil, fmt.Errorf("%s:%d: bad directive %q", name, line, s)
			}
			n, err := strconv.Atoi(fields[1])
			if err != nil {
				return nil, fmt.Errorf("%s:%d: %v", name, line, err)
			}
			switch fields[0] {
			case "height":
				f.Height = n
			case "ascent":
				f.Ascent = n
			case "tracking":
				f.Tracking = n
			case "xheight":
				f.XHeight = n
			case "capheight":
				f.CapHeight = n
			default:
				return nil, fmt.Errorf("%s:%d: unknown directive %q", name, line, fields[0])
			}
		case strings.HasPrefix(s, "== "):
			spec := s[3:]
			var r rune
			switch {
			case strings.HasPrefix(spec, "space "):
				w, err := strconv.Atoi(strings.TrimSpace(spec[6:]))
				if err != nil {
					return nil, fmt.Errorf("%s:%d: %v", name, line, err)
				}
				r = ' '
				widths[r] = w
			case strings.HasPrefix(spec, "U+"):
				n, err := strconv.ParseInt(spec[2:], 16, 32)
				if err != nil {
					return nil, fmt.Errorf("%s:%d: %v", name, line, err)
				}
				r = rune(n)
			default:
				rs := []rune(spec)
				if len(rs) != 1 {
					return nil, fmt.Errorf("%s:%d: bad glyph name %q", name, line, spec)
				}
				r = rs[0]
			}
			if _, dup := f.rows[r]; dup {
				return nil, fmt.Errorf("%s:%d: glyph %q defined twice", name, line, r)
			}
			f.rows[r] = nil
			order = append(order, r)
			cur = r
		default:
			if cur < 0 {
				return nil, fmt.Errorf("%s:%d: row outside a glyph", name, line)
			}
			f.rows[cur] = append(f.rows[cur], s)
		}
	}
	if f.Height <= 0 || f.Ascent <= 0 {
		return nil, fmt.Errorf("%s: missing @height or @ascent", name)
	}
	// The marks are not glyphs; the letters made with them are.
	marks := map[rune][]string{}
	kept := order[:0]
	for _, r := range order {
		if isMark(r) {
			marks[r] = f.rows[r]
			delete(f.rows, r)
		} else {
			kept = append(kept, r)
		}
	}
	order = kept
	over := map[rune][]string{} // the rows over the cell, the last nearest it
	for _, comp := range composites {
		r, base, mark := comp[0], comp[1], comp[2]
		if _, has := f.rows[r]; has || f.rows[base] == nil || marks[mark] == nil {
			continue
		}
		rows, above := f.compose(f.rows[base], marks[mark], mark == 0x327 || mark == 0x328 || mark == 0x326, mark == 0x328, base == 'i' || base == 'j')
		f.rows[r] = rows
		over[r] = above
		f.Above = max(f.Above, len(above))
		order = append(order, r)
	}
	x := 0
	for _, r := range order {
		rows := f.rows[r]
		if len(rows) > f.Height {
			return nil, fmt.Errorf("%s: glyph %q is %d rows tall, more than %d", name, r, len(rows), f.Height)
		}
		w := widths[r]
		for _, row := range append(rows, over[r]...) {
			w = max(w, len([]rune(row)))
		}
		f.Glyphs[r] = Glyph{X: x, Width: w, Advance: w + f.Tracking}
		x += w + 1
	}
	f.Atlas = image.NewRGBA(image.Rect(0, 0, max(x, 1), f.Above+f.Height))
	ink := func(x, y int, row string) {
		for i, ch := range []rune(row) {
			if ch == '#' {
				f.Atlas.SetRGBA(x+i, y, color.RGBA{255, 255, 255, 255})
			}
		}
	}
	for _, r := range order {
		g := f.Glyphs[r]
		for y, row := range f.rows[r] {
			ink(g.X, f.Above+y, row)
		}
		for y, row := range over[r] {
			ink(g.X, f.Above-len(over[r])+y, row)
		}
	}
	if _, ok := f.Glyphs[0xFFFD]; ok {
		f.Fallback = 0xFFFD
	} else {
		f.Fallback = '?'
	}
	return f, nil
}

func isMark(r rune) bool { return r >= 0x300 && r <= 0x36F }

// composites are the letters made of a letter and a mark: the letter, its
// base and its mark.
var composites = func() [][3]rune {
	var out [][3]rune
	for _, set := range []struct {
		mark  rune
		bases string
		made  string
	}{
		{0x300, "AEIOUaeiou", "ÀÈÌÒÙàèìòù"},
		{0x301, "AEIOUYaeiouy", "ÁÉÍÓÚÝáéíóúý"},
		{0x302, "AEIOUaeiou", "ÂÊÎÔÛâêîôû"},
		{0x303, "ANOano", "ÃÑÕãñõ"},
		{0x308, "AEIOUaeiouy", "ÄËÏÖÜäëïöüÿ"},
		{0x30A, "Aa", "Åå"},
		{0x327, "CcSs", "ÇçŞş"},
		// Latin Extended-A, for Central European and Turkish text.
		{0x301, "CNSZLRcnszlr", "ĆŃŚŹĹŔćńśźĺŕ"},
		{0x30C, "CENRSZTcenrsz", "ČĚŇŘŠŽŤčěňřšž"},
		{0x30C, "D", "Ď"},
		{0x307, "ZEIze", "ŻĖİżė"},
		{0x30B, "OUou", "ŐŰőű"},
		{0x306, "GAga", "ĞĂğă"},
		{0x30A, "Uu", "Ůů"},
		{0x328, "AEae", "ĄĘąę"},
		// Romanian and Latvian.
		{0x304, "AEIUaeiu", "ĀĒĪŪāēīū"},
		{0x326, "SsTtKkLlNn", "ȘșȚțĶķĻļŅņ"},
	} {
		made := []rune(set.made)
		for i, b := range []rune(set.bases) {
			out = append(out, [3]rune{made[i], b, set.mark})
		}
	}
	return out
}()

// compose draws a letter of base with mark over it — or under it, below —
// centered on it, or under its right side, right; dotless takes the dot off an i or j first. It returns the
// rows of the cell, and those over it where the mark rises above.
func (f *Face) compose(base, mark []string, below, right, dotless bool) (rows, above []string) {
	bw, mw := 0, 0
	for _, r := range base {
		bw = max(bw, len([]rune(r)))
	}
	for _, r := range mark {
		mw = max(mw, len([]rune(r)))
	}
	w := max(bw, mw)
	cells := make([][]rune, f.Height)
	for y := range cells {
		cells[y] = []rune(strings.Repeat(".", w))
	}
	put := func(y, x int, row string, set [][]rune) {
		for i, ch := range []rune(row) {
			if ch == '#' && y >= 0 && y < len(set) && x+i < w {
				set[y][x+i] = '#'
			}
		}
	}
	top := f.Height
	for y, r := range base {
		if strings.Contains(r, "#") && !(dotless && y < f.Ascent-f.XHeight) {
			top = min(top, y)
		}
	}
	for y, r := range base {
		if dotless && y < f.Ascent-f.XHeight {
			continue
		}
		put(y, (w-bw)/2, r, cells)
	}
	if below {
		x := (w - mw) / 2
		if right { // an ogonek, under the letter's right side
			x = max((w-bw)/2+bw-mw, 0)
		}
		for i, r := range mark {
			put(f.Ascent+i, x, r, cells)
		}
	} else {
		// A row clear of the letter's top, where the leading has room; over
		// a capital, on it.
		first := max(top-1-len(mark), -MaxAbove)
		var over [][]rune
		if first < 0 {
			over = make([][]rune, -first)
			for y := range over {
				over[y] = []rune(strings.Repeat(".", w))
			}
		}
		for i, r := range mark {
			y := first + i
			if y < 0 {
				put(y+len(over), (w-mw)/2, r, over)
			} else {
				put(y, (w-mw)/2, r, cells)
			}
		}
		for _, r := range over {
			above = append(above, string(r))
		}
	}
	for _, r := range cells {
		rows = append(rows, string(r))
	}
	return rows, above
}

// MaxAbove is as many rows as a mark rises over the cell: what the leading
// over a line of text holds.
const MaxAbove = 2

// MustParse is Parse for faces compiled into the program.
func MustParse(name, src string) *Face {
	f, err := Parse(name, src)
	if err != nil {
		panic(err)
	}
	return f
}

// Glyph returns the glyph for r, or the fallback glyph.
func (f *Face) Glyph(r rune) Glyph {
	if g, ok := f.Glyphs[r]; ok {
		return g
	}
	return f.Glyphs[f.Fallback]
}

// Has reports whether the face draws r itself.
func (f *Face) Has(r rune) bool {
	_, ok := f.Glyphs[r]
	return ok
}

// Measure returns the advance of s: the width it takes on a line, including
// the tracking after its last glyph.
func (f *Face) Measure(s string) int {
	w := 0
	for _, r := range s {
		w += f.Glyph(r).Advance
	}
	return w
}

// Rows returns the drawing of r as parsed, for tests.
func (f *Face) Rows(r rune) []string { return f.rows[r] }
