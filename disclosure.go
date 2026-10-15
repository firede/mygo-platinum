package platinum

import "github.com/egoist/mygo/ui"

// DisclosureSize is the box of a disclosure triangle.
const DisclosureSize = 12

// DisclosureWidget is a disclosure triangle being built.
type DisclosureWidget struct {
	inactive bool
	el       ui.Element
	open     *bool
	label    string
	disabled bool
	pressed  bool
	focused  bool
	changed  bool
	th       *Theme
}

// Disclosure builds a disclosure triangle that shows and hides what it
// labels: it points right while *open is false and down while it is true,
// and a click, Space or Enter turns it. Left and Right close and open it
// too. Build what it shows after it, while *open.
//
// The counterpart among MyGo's own controls is ui.Collapsible.
func Disclosure(c *ui.Context, open *bool, label string) *DisclosureWidget {
	d := &DisclosureWidget{open: open, label: label, th: ThemeOf(c), inactive: inactive(c)}
	d.el = ui.ButtonBase(c).FocusRing(false).Role(ui.RoleDisclosure).Label(label).Expanded(*open)
	d.pressed = d.el.Pressed()
	d.focused = d.el.FocusVisible()
	w := DisclosureSize
	if label != "" {
		w += CheckGap + SystemFont.Measure(label)
	}
	d.el.Size(float32(w), CheckHeight)
	was := *open
	if d.el.Clicked() {
		*open = !*open
	}
	if d.el.Shortcut(0, ui.KeyRight) {
		*open = true
	}
	if d.el.Shortcut(0, ui.KeyLeft) {
		*open = false
	}
	d.changed = *open != was
	d.el.Draw(d.draw)
	report(d.el, d.changed)
	return d
}

// Disabled dims the triangle and makes it take no input while d is true.
func (d *DisclosureWidget) Disabled(dis bool) *DisclosureWidget {
	d.disabled = dis
	d.el.Disabled(dis)
	return d
}

// Element returns the triangle's element.
func (d *DisclosureWidget) Element() ui.Element { return d.el }

func (d *DisclosureWidget) draw(p *ui.Painter, r ui.Rect) {
	pn := NewPen(p, r).Faded(d.disabled)
	s := StateNormal
	switch {
	case d.disabled:
		s = StateDisabled
	case d.inactive:
		s = StateInactive
	case d.pressed:
		s = StatePressed
	}
	box := pn.Sub(0, (pn.H-DisclosureSize)/2, DisclosureSize, DisclosureSize)
	DrawDisclosure(box, *d.open, s)
	if d.focused && !d.disabled {
		DrawFocusRing(box, 0, d.th.Accent.Dark)
	}
	if d.label != "" {
		pn.TextMiddle(DisclosureSize+CheckGap, 0, pn.H, d.label, SystemFont, labelColor(s))
	}
}

var triangleRight = []string{
	"o......",
	"oo.....",
	"o-o....",
	"o--o...",
	"o---o..",
	"o----o.",
	"o---o..",
	"o--o...",
	"o-o....",
	"oo.....",
	"o......",
}

var triangleDown = []string{
	"ooooooooooo",
	".o-------o.",
	"..o-----o..",
	"...o---o...",
	"....o-o....",
	".....o.....",
}

// DrawDisclosure draws a disclosure triangle in pn's 12×12 corner, pointing
// down when open.
func DrawDisclosure(pn Pen, open bool, s State) {
	line, fill, hi := Black, Gray6, Gray4
	switch s {
	case StatePressed:
		fill, hi = Black, Black
	case StateDisabled:
		line, fill, hi = Gray6, Gray3, Gray3
	case StateInactive:
		line, fill, hi = Gray8, Gray3, Gray3
	}
	rows, x, y := triangleRight, 3, 0
	if open {
		rows, x, y = triangleDown, 0, 3
	}
	for j, row := range rows {
		for i := 0; i < len(row); i++ {
			switch row[i] {
			case 'o':
				pn.Dot(x+i, y+j, line)
			case '-':
				c := fill
				// A highlight along the upper edge of the fill.
				if (!open && j < len(rows)/2 && row[i+1] == 'o') || (open && j == 1) {
					c = hi
				}
				pn.Dot(x+i, y+j, c)
			}
		}
	}
}

// Changed reports whether the user opened or closed the triangle since
// the last frame.
func (d *DisclosureWidget) Changed() bool { return report(d.el, d.changed) }
