package platinum

import "github.com/egoist/mygo/ui"

// Check box and radio button metrics, from docs/spec.md.
const (
	CheckSize   = 12 // the box of a check box, and the circle of a radio button
	CheckHeight = 16 // the height of a check box or radio button with its label
	CheckGap    = 5  // between the box and the label
)

// CheckboxWidget is a check box being built.
type CheckboxWidget struct {
	inactive bool
	el       ui.Element
	label    string
	checked  bool
	mixed    bool
	disabled bool
	pressed  bool
	focused  bool
	th       *Theme
}

// Checkbox builds a check box that toggles *checked, labeled label: a small
// beveled box that shows a cross while checked. A click on the box or the
// label toggles it, as does Space while it has the keyboard focus.
func Checkbox(c *ui.Context, checked *bool, label string) *CheckboxWidget {
	b := &CheckboxWidget{label: label, th: ThemeOf(c), inactive: inactive(c)}
	b.el = ui.CheckboxBase(c, checked).FocusRing(false).Label(label)
	b.pressed = b.el.Pressed()
	b.focused = b.el.FocusVisible()
	b.el.Size(float32(checkWidth(label)), CheckHeight)
	b.el.Draw(func(p *ui.Painter, r ui.Rect) {
		b.checked = *checked
		b.draw(p, r)
	})
	return b
}

func checkWidth(label string) int {
	if label == "" {
		return CheckSize
	}
	return CheckSize + CheckGap + SystemFont.Measure(label)
}

// Mixed shows the box partly checked, a dash in place of the cross, for a
// box over others of which some are checked.
func (b *CheckboxWidget) Mixed() *CheckboxWidget {
	b.mixed = true
	b.el.Mixed()
	return b
}

// Disabled dims the check box and makes it take no input while d is true.
func (b *CheckboxWidget) Disabled(d bool) *CheckboxWidget {
	b.disabled = d
	b.el.Disabled(d)
	return b
}

// Changed reports whether the user toggled the box since the last frame.
func (b *CheckboxWidget) Changed() bool { return b.el.Changed() }

// Element returns the check box's element.
func (b *CheckboxWidget) Element() ui.Element { return b.el }

func (b *CheckboxWidget) draw(p *ui.Painter, r ui.Rect) {
	pn := NewPen(p, r).Faded(b.disabled)
	y := (pn.H - CheckSize) / 2
	s := StateNormal
	switch {
	case b.disabled:
		s = StateDisabled
	case b.inactive:
		s = StateInactive
	case b.pressed:
		s = StatePressed
	}
	mark := CheckNone
	switch {
	case b.mixed:
		mark = CheckMixed
	case b.checked:
		mark = CheckOn
	}
	box := pn.Sub(0, y, CheckSize, CheckSize)
	DrawCheckbox(box, mark, s)
	if b.focused && !b.disabled {
		DrawFocusRing(box, 0, b.th.Accent.Dark)
	}
	if b.label != "" {
		pn.TextMiddle(CheckSize+CheckGap, 0, pn.H, b.label, SystemFont, labelColor(s))
	}
}

func labelColor(s State) ui.Color {
	if s == StateDisabled || s == StateInactive {
		return Gray8
	}
	return Black
}

// CheckMark is what a check box shows in its box.
type CheckMark int

// The marks of a check box: none, a cross, a dash.
const (
	CheckNone CheckMark = iota
	CheckOn
	CheckMixed
)

var crossMark = [8]string{
	"#......#",
	"##....##",
	".##..##.",
	"..####..",
	"..####..",
	".##..##.",
	"##....##",
	"#......#",
}

// DrawCheckbox draws the box of a check box in pn's corner.
func DrawCheckbox(pn Pen, mark CheckMark, s State) {
	frame, face, tl, br, ink := Black, Gray2, White, Gray7, Black
	switch s {
	case StatePressed:
		face, tl, br = Gray4, Gray7, Gray1
	case StateDisabled:
		frame, tl, br, ink = Gray6, Gray2, Gray2, Gray6
	case StateInactive:
		frame, tl, br, ink = Gray8, Gray2, Gray2, Gray8
	}
	n := CheckSize
	pn.Frame(0, 0, n, n, frame)
	pn.Rect(1, 1, n-2, n-2, face)
	pn.Bevel(1, 1, n-2, n-2, tl, br)
	switch mark {
	case CheckOn:
		pn.Bitmap(2, 2, crossMark[:], ink)
	case CheckMixed:
		pn.Rect(3, 5, 6, 2, ink)
	}
}

// Bitmap draws a small picture given as rows of '#' (ink) and other
// characters (paper) with its corner at x, y.
func (pn Pen) Bitmap(x, y int, rows []string, c ui.Color) {
	for j, row := range rows {
		run := -1
		for i := 0; i <= len(row); i++ {
			on := i < len(row) && row[i] == '#'
			if on && run < 0 {
				run = i
			} else if !on && run >= 0 {
				pn.HLine(x+run, y+j, i-run, c)
				run = -1
			}
		}
	}
}

// RadioWidget is a radio button being built.
type RadioWidget struct {
	inactive bool
	el       ui.Element
	label    string
	on       func() bool
	disabled bool
	pressed  bool
	focused  bool
	th       *Theme
}

// Radio builds a radio button labeled label that sets *selected to value: a
// small beveled circle with a dot while *selected is value. Put the buttons
// of one choice in a RadioGroup, which makes them one stop of Tab that the
// arrows move within.
func Radio[T comparable](c *ui.Context, selected *T, value T, label string) *RadioWidget {
	b := &RadioWidget{label: label, th: ThemeOf(c), on: func() bool { return *selected == value }, inactive: inactive(c)}
	b.el = ui.RadioBase(c, selected, value).FocusRing(false).Label(label)
	b.pressed = b.el.Pressed()
	b.focused = b.el.FocusVisible()
	b.el.Size(float32(checkWidth(label)), CheckHeight)
	b.el.Draw(b.draw)
	return b
}

// RadioGroup lays out the radio buttons fn builds in a column, as one stop
// of Tab whose buttons the arrow keys move between and choose.
func RadioGroup(c *ui.Context, fn func()) ui.Element {
	return ui.RadioGroup(c, fn).Gap(2)
}

// Disabled dims the radio button and makes it take no input while d is true.
func (b *RadioWidget) Disabled(d bool) *RadioWidget {
	b.disabled = d
	b.el.Disabled(d)
	return b
}

// Changed reports whether the user chose the button since the last frame.
func (b *RadioWidget) Changed() bool { return b.el.Changed() }

// Element returns the radio button's element.
func (b *RadioWidget) Element() ui.Element { return b.el }

func (b *RadioWidget) draw(p *ui.Painter, r ui.Rect) {
	pn := NewPen(p, r).Faded(b.disabled)
	y := (pn.H - CheckSize) / 2
	s := StateNormal
	switch {
	case b.disabled:
		s = StateDisabled
	case b.inactive:
		s = StateInactive
	case b.pressed:
		s = StatePressed
	}
	box := pn.Sub(0, y, CheckSize, CheckSize)
	DrawRadio(box, b.on(), s)
	if b.focused && !b.disabled {
		DrawFocusCircle(box, b.th.Accent.Dark)
	}
	if b.label != "" {
		pn.TextMiddle(CheckSize+CheckGap, 0, pn.H, b.label, SystemFont, labelColor(s))
	}
}

// DrawFocusCircle draws the focus ring of a radio button around pn's 12×12
// circle: a ring two pixels wide, a pixel outside it.
func DrawFocusCircle(pn Pen, c ui.Color) {
	for y := -4; y < 16; y++ {
		for x := -4; x < 16; x++ {
			dx, dy := float64(x)+0.5-6, float64(y)+0.5-6
			d := dx*dx + dy*dy
			if d >= 7.2*7.2 && d < 9.2*9.2 {
				pn.Dot(x, y, c)
			}
		}
	}
}

// radioMask is the circle of a radio button: 'o' is its outline, '-' its
// inside.
var radioMask = [12]string{
	"....oooo....",
	"..oo----oo..",
	".o--------o.",
	".o--------o.",
	"o----------o",
	"o----------o",
	"o----------o",
	"o----------o",
	".o--------o.",
	".o--------o.",
	"..oo----oo..",
	"....oooo....",
}

var radioDot = []string{
	".##.",
	"####",
	"####",
	".##.",
}

// DrawRadio draws the circle of a radio button in pn's corner.
func DrawRadio(pn Pen, on bool, s State) {
	frame, face, tl, br, ink := Black, Gray2, White, Gray7, Black
	switch s {
	case StatePressed:
		face, tl, br = Gray4, Gray7, Gray1
	case StateDisabled:
		frame, tl, br, ink = Gray6, Gray2, Gray2, Gray6
	case StateInactive:
		frame, tl, br, ink = Gray8, Gray2, Gray2, Gray8
	}
	inside := func(x, y int) bool {
		return x >= 0 && y >= 0 && x < 12 && y < 12 && radioMask[y][x] == '-'
	}
	for y, row := range radioMask {
		for x := 0; x < len(row); x++ {
			switch row[x] {
			case 'o':
				pn.Dot(x, y, frame)
			case '-':
				c := face
				// The pixels along the outline take the bevel: light toward
				// the top left, dark toward the bottom right.
				if !inside(x-1, y) || !inside(x, y-1) || !inside(x-1, y-1) {
					if x+y < 11 {
						c = tl
					} else if x+y > 11 {
						c = br
					}
				}
				if !inside(x+1, y) || !inside(x, y+1) || !inside(x+1, y+1) {
					if x+y > 11 {
						c = br
					} else if x+y < 11 && c == face {
						c = tl
					}
				}
				pn.Dot(x, y, c)
			}
		}
	}
	if on {
		pn.Bitmap(4, 4, radioDot, ink)
	}
}
