package platinum

import (
	"math"

	"github.com/egoist/mygo/ui"
)

// Pen draws on whole pixels of the Platinum grid, one DIP each: it rounds
// the box it draws in to whole DIPs, so that every line it draws is one
// pixel wide at a scale of 1 and a whole block of pixels at any whole
// scale. Controls draw with it, and so can controls of your own.
type Pen struct {
	P *ui.Painter
	// X and Y are the corner of the box on the window, and W and H its
	// size, in whole DIPs. The pen's coordinates start at the corner.
	X, Y, W, H int
	// faded is set for a pen drawing an element MyGo paints at half
	// opacity, as it paints disabled ones (see Faded).
	faded bool
	// dim draws as in a window in the background (see Dimmed).
	dim bool
}

// Dimmed returns the pen for drawing a control in a window in the
// background, while d: Black and the grays darker than Gray8 come out
// Gray8, and colors as a light gray of their lightness, as Platinum dims
// the controls of a window that is not in front.
func (pn Pen) Dimmed(d bool) Pen {
	pn.dim = d
	return pn
}

func dimColor(c ui.Color) ui.Color {
	if c.R == c.G && c.G == c.B {
		if c.R < Gray8.R {
			return Gray8.Alpha(float32(c.A) / 255)
		}
		return c
	}
	l := (299*int(c.R) + 587*int(c.G) + 114*int(c.B)) / 1000
	v := uint8(l + (255-l)/2)
	return ui.RGB(v, v, v).Alpha(float32(c.A) / 255)
}

// Faded returns the pen for drawing an element that MyGo paints at half
// opacity, as it paints the elements it disables: it draws each color so
// that, half over the window's face (Gray2) beneath, it comes out as the
// color asked for. Colors lighter than the face can come out a little
// darker than asked: white comes out Gray1.
func (pn Pen) Faded(f bool) Pen {
	pn.faded = f
	return pn
}

// unfade returns the color to draw for c to come out as c through MyGo's
// half opacity over Gray2.
func (pn Pen) unfade(c ui.Color) ui.Color {
	if pn.dim && c.A > 0 {
		c = dimColor(c)
	}
	if !pn.faded || c.A == 0 {
		return c
	}
	ch := func(v, bg uint8) uint8 { return uint8(min(max(2*int(v)-int(bg), 0), 255)) }
	out := ui.RGB(ch(c.R, Gray2.R), ch(c.G, Gray2.G), ch(c.B, Gray2.B))
	if c.A < 255 {
		out = out.Alpha(float32(c.A) / 255)
	}
	return out
}

// NewPen returns a pen drawing in r, rounded to whole DIPs.
func NewPen(p *ui.Painter, r ui.Rect) Pen {
	x := round(r.X)
	y := round(r.Y)
	return Pen{P: p, X: x, Y: y, W: round(r.X+r.W) - x, H: round(r.Y+r.H) - y}
}

func round(v float32) int { return int(math.Floor(float64(v) + 0.5)) }

// Rect fills w×h pixels at x, y.
func (pn Pen) Rect(x, y, w, h int, c ui.Color) {
	if w <= 0 || h <= 0 || c.A == 0 {
		return
	}
	c = pn.unfade(c)
	pn.P.Fill(ui.Rect{X: float32(pn.X + x), Y: float32(pn.Y + y), W: float32(w), H: float32(h)}, c, 0)
}

// Dot fills the pixel at x, y.
func (pn Pen) Dot(x, y int, c ui.Color) { pn.Rect(x, y, 1, 1, c) }

// HLine draws a line w pixels long rightward from x, y.
func (pn Pen) HLine(x, y, w int, c ui.Color) { pn.Rect(x, y, w, 1, c) }

// VLine draws a line h pixels long downward from x, y.
func (pn Pen) VLine(x, y, h int, c ui.Color) { pn.Rect(x, y, 1, h, c) }

// Frame outlines the w×h box at x, y with a line one pixel wide.
func (pn Pen) Frame(x, y, w, h int, c ui.Color) {
	pn.HLine(x, y, w, c)
	pn.HLine(x, y+h-1, w, c)
	pn.VLine(x, y+1, h-2, c)
	pn.VLine(x+w-1, y+1, h-2, c)
}

// Bevel draws a ring one pixel wide just inside the w×h box at x, y: tl
// along its top and left edges, br along its bottom and right ones. The
// two corners where they meet, top right and bottom left, are left
// unpainted, for the face under them to show, as the Platinum bevels have
// them.
func (pn Pen) Bevel(x, y, w, h int, tl, br ui.Color) {
	pn.HLine(x, y, w-1, tl)
	pn.VLine(x, y+1, h-2, tl)
	pn.HLine(x+1, y+h-1, w-1, br)
	pn.VLine(x+w-1, y+1, h-2, br)
}

// RoundFrame outlines the w×h box at x, y with a line one pixel wide whose
// corners are cut in steps of r pixels (0 to 4), and fills it with face
// unless face is transparent.
func (pn Pen) RoundFrame(x, y, w, h, r int, line, face ui.Color) {
	for i := 0; i < h; i++ {
		d := Inset(i, h, r)
		if i == 0 || i == h-1 {
			pn.HLine(x+d, y+i, w-2*d, line)
			continue
		}
		pn.Dot(x+d, y+i, line)
		pn.Dot(x+w-1-d, y+i, line)
		pn.HLine(x+d+1, y+i, w-2*d-2, face)
	}
	// Close the steps of the corners where they jump by more than a pixel.
	in := cornerInset(r)
	for i := 1; i < len(in) && i < h/2; i++ {
		for k := in[i] + 1; k < in[i-1]; k++ {
			pn.Dot(x+k, y+i, line)
			pn.Dot(x+w-1-k, y+i, line)
			pn.Dot(x+k, y+h-1-i, line)
			pn.Dot(x+w-1-k, y+h-1-i, line)
		}
	}
}

// FillRound fills the w×h box at x, y, its corners cut as RoundFrame cuts
// them.
func (pn Pen) FillRound(x, y, w, h, r int, c ui.Color) {
	for i := 0; i < h; i++ {
		d := Inset(i, h, r)
		pn.HLine(x+d, y+i, w-2*d, c)
	}
}

// cornerInset returns how far each row of a corner of radius r starts in
// from the side of its box.
func cornerInset(r int) []int {
	switch r {
	case 1:
		return []int{1}
	case 2:
		return []int{2, 1}
	case 3:
		return []int{3, 1, 1}
	case 4:
		return []int{4, 2, 1, 1}
	}
	return nil
}

// Inset returns how far row i of a box h tall, with corners of radius r,
// starts in from its sides, for shapes of one's own rounded as the
// library's are.
func Inset(i, h, r int) int {
	in := cornerInset(r)
	if i < len(in) {
		return in[i]
	}
	if h-1-i < len(in) {
		return in[h-1-i]
	}
	return 0
}

// Pattern fills the w×h box at x, y with an 8×8 pattern: bit 7-x of row y
// of pat paints fg, the others bg (unless bg is transparent). The pattern
// is anchored at the window's corner, so that neighbors line up.
func (pn Pen) Pattern(x, y, w, h int, pat [8]byte, fg, bg ui.Color) {
	if bg.A != 0 {
		pn.Rect(x, y, w, h, bg)
	}
	for j := 0; j < h; j++ {
		row := pat[(pn.Y+y+j)&7]
		run := -1
		for i := 0; i <= w; i++ {
			on := i < w && row&(0x80>>((pn.X+x+i)&7)) != 0
			if on && run < 0 {
				run = i
			} else if !on && run >= 0 {
				pn.HLine(x+run, y+j, i-run, fg)
				run = -1
			}
		}
	}
}

// Clip draws what fn draws within the w×h box at x, y only.
func (pn Pen) Clip(x, y, w, h int, fn func()) {
	pn.P.Clip(ui.Rect{X: float32(pn.X + x), Y: float32(pn.Y + y), W: float32(w), H: float32(h)}, 0, fn)
}

// Sub returns a pen drawing in the w×h box at x, y of this one.
func (pn Pen) Sub(x, y, w, h int) Pen {
	return Pen{P: pn.P, X: pn.X + x, Y: pn.Y + y, W: w, H: h, faded: pn.faded, dim: pn.dim}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
