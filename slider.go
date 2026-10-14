package platinum

import "github.com/egoist/mygo/ui"

// Slider metrics, from docs/spec.md.
const (
	SliderHeight     = 18 // the height of a horizontal slider
	SliderThumbW     = 11 // the width of its thumb
	SliderThumbH     = 16 // the height of its thumb
	SliderTrackH     = 6  // the height of its track
	SliderMinWidth   = 60 // the narrowest a slider gets
	sliderTrackTop   = 4
	sliderThumbInset = SliderThumbW / 2
)

// SliderWidget is a slider being built.
type SliderWidget struct {
	inactive bool
	el       ui.Element
	value    *float64
	lo, hi   float64
	disabled bool
	focused  bool
	ticks    int
	th       *Theme
}

// Slider builds a horizontal slider setting *value between lo and hi: a
// sunken track and a pointed thumb that the pointer drags, and the arrows,
// Page Up and Down, Home and End move while it has the keyboard focus.
// Label it for assistive technology.
func Slider(c *ui.Context, value *float64, lo, hi float64) *SliderWidget {
	s := &SliderWidget{value: value, lo: lo, hi: hi, th: ThemeOf(c), inactive: inactive(c)}
	s.el = ui.SliderBase(c, value, lo, hi).FocusRing(false).
		Height(SliderHeight).Width(160).MinWidth(SliderMinWidth).
		Padding(0, sliderThumbInset, 0, sliderThumbInset)
	s.focused = s.el.FocusVisible()
	s.el.Draw(s.draw)
	return s
}

// Width sets the slider's width.
func (s *SliderWidget) Width(w int) *SliderWidget {
	s.el.Width(float32(w))
	return s
}

// Step makes the slider move by step, as the keys do, and stop only on its
// multiples.
func (s *SliderWidget) Step(step float64) *SliderWidget {
	s.el.Step(step)
	return s
}

// Ticks draws n tick marks under the track, evenly spaced from end to end.
func (s *SliderWidget) Ticks(n int) *SliderWidget {
	s.ticks = n
	if n > 1 {
		s.el.Height(SliderHeight + 5)
	}
	return s
}

// Label names the slider for assistive technology and tests.
func (s *SliderWidget) Label(l string) *SliderWidget {
	s.el.Label(l)
	return s
}

// Disabled dims the slider and makes it take no input while d is true.
func (s *SliderWidget) Disabled(d bool) *SliderWidget {
	s.disabled = d
	s.el.Disabled(d)
	return s
}

// Changed reports whether the user moved the slider since the last frame.
func (s *SliderWidget) Changed() bool { return s.el.Changed() }

// Element returns the slider's element.
func (s *SliderWidget) Element() ui.Element { return s.el }

func (s *SliderWidget) draw(p *ui.Painter, r ui.Rect) {
	pn := NewPen(p, r).Faded(s.disabled)
	f := 0.0
	if s.hi > s.lo {
		f = (*s.value - s.lo) / (s.hi - s.lo)
	}
	st := StateNormal
	if s.disabled {
		st = StateDisabled
	} else if s.inactive {
		st = StateInactive
	}
	DrawSlider(pn, f, s.ticks, st, s.th.Accent)
	if s.focused && !s.disabled {
		DrawFocusRing(pn.Sub(0, 0, pn.W, SliderHeight), 0, s.th.Accent.Dark)
	}
}

var sliderThumb = [SliderThumbH]string{
	"ooooooooooo",
	"o---------o",
	"o---------o",
	"o---------o",
	"o---------o",
	"o---------o",
	"o---------o",
	"o---------o",
	"o---------o",
	"o---------o",
	"o---------o",
	".o-------o.",
	"..o-----o..",
	"...o---o...",
	"....o-o....",
	".....o.....",
}

// DrawSlider draws a horizontal slider at fraction f (0 to 1) of its track
// in pn's box, with n tick marks under it when n > 1.
func DrawSlider(pn Pen, f float64, n int, s State, a Accent) {
	f = min(max(f, 0), 1)
	w := pn.W
	tx, tw := sliderThumbInset, w-2*sliderThumbInset
	// The track: sunken, its part left of the thumb filled with the accent.
	ty := sliderTrackTop
	pn.Bevel(tx-2, ty, tw+4, SliderTrackH, Gray7, White)
	// The track is sunken along its whole length; only the thumb carries
	// the accent.
	empty, line := Gray4, Gray11
	switch s {
	case StateDisabled:
		empty, line = Gray3, Gray6
	case StateInactive:
		empty, line = Gray3, Gray8
	}
	pn.Frame(tx-1, ty+1, tw+2, SliderTrackH-2, line)
	pos := tx + int(f*float64(tw-1)+0.5)
	pn.Rect(tx, ty+2, tw, SliderTrackH-4, empty)
	if n > 1 {
		for i := 0; i < n; i++ {
			x := tx + i*(tw-1)/(n-1)
			pn.VLine(x, SliderThumbH+2, 4, line)
		}
	}
	DrawSliderThumb(pn.Sub(pos-sliderThumbInset, 0, SliderThumbW, SliderThumbH), s, a)
}

// DrawSliderThumb draws the pointed thumb of a slider in pn's corner.
func DrawSliderThumb(pn Pen, s State, a Accent) {
	frame, face, tl, br := Black, thumbFace(a), White, a.Dark
	switch s {
	case StatePressed:
		face, tl, br = a.Mid, a.Dark, a.Light
	case StateDisabled:
		frame, face, tl, br = Gray6, Gray2, Gray1, Gray4
	case StateInactive:
		frame, face, tl, br = Gray8, Gray2, Gray2, Gray2
	}
	for y, row := range sliderThumb {
		for x := 0; x < len(row); x++ {
			switch row[x] {
			case 'o':
				pn.Dot(x, y, frame)
			case '-':
				c := face
				if y == 1 || row[x-1] == 'o' && y < 11 {
					c = tl
				} else if row[x+1] == 'o' || y >= 11 && row[x+1] != '-' {
					c = br
				}
				pn.Dot(x, y, c)
			}
		}
	}
	// Three ridges to grip it by.
	if s != StateDisabled && s != StateInactive {
		for i := 0; i < 3; i++ {
			pn.HLine(3, 4+2*i, 5, a.Dark)
			pn.HLine(4, 5+2*i, 5, White)
		}
	}
}

// thumbFace is the face of thumbs, between an accent's light and mid
// colors.
func thumbFace(a Accent) ui.Color { return a.Light.Mix(a.Mid, 0.5) }
