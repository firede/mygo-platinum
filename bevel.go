package platinum

import (
	"github.com/egoist/mygo/ui"
)

// Bevel button metrics, from docs/spec.md.
const (
	BevelHeight  = 22 // the height of a bevel button with a label alone
	bevelPadding = 8  // the room on each side of its content
	bevelGap     = 4  // between its icon and its label
	bevelArrowW  = 9  // the room its pop-up arrow takes
)

type bevelState struct {
	t    *menuTracker
	rect ui.Rect
}

// BevelWidget is a bevel button being built.
type BevelWidget struct {
	inactive bool
	c        *ui.Context
	el       ui.Element
	st       *bevelState
	label    string
	icon     []string
	on       *bool
	menu     []MenuItem
	noArrow  bool // a menu button that shows no arrow, as a toolbar's »
	disabled bool
	pressed  bool
	focused  bool
	clicked  bool
	w, h     int
	th       *Theme
}

// BevelButton builds a bevel button: a square, beveled button that shows
// a label, an icon above or beside it (Icon), stays pressed in as a toggle
// (Toggle), or opens a menu (Menu). Space and Enter press it while it has
// the keyboard focus.
func BevelButton(c *ui.Context, label string) *BevelWidget {
	b := &BevelWidget{c: c, label: label, th: ThemeOf(c), inactive: inactive(c)}
	b.el = ui.ButtonBase(c).FocusRing(false).Label(label)
	b.st = ui.Local(b.el, "bevel", func() bevelState { return bevelState{} })
	b.pressed = b.el.Pressed()
	b.focused = b.el.FocusVisible()
	b.resize()
	b.el.Draw(b.draw)
	return b
}

// IconButton builds a bevel button showing icon alone, labeled label for
// assistive technology, as the buttons of a toolbar.
func IconButton(c *ui.Context, icon []string, label string) *BevelWidget {
	b := BevelButton(c, label)
	b.icon = icon
	b.label = ""
	b.resize()
	b.el.Label(label)
	return b
}

// Icon shows a small picture in the button, left of its label, given as
// rows for DrawIcon.
func (b *BevelWidget) Icon(rows []string) *BevelWidget {
	b.icon = rows
	b.resize()
	return b
}

// Toggle makes the button stay pressed in while *on, and turn *on over as
// it is clicked.
func (b *BevelWidget) Toggle(on *bool) *BevelWidget {
	b.on = on
	b.el.Role(ui.RoleToggleButton)
	if !b.disabled && b.el.Clicked() {
		*on = !*on
		b.clicked = true
	}
	b.el.Checked(*on)
	return b
}

// Menu makes the button open a menu of items below it as it is pressed,
// with a small arrow at its right.
func (b *BevelWidget) Menu(items []MenuItem) *BevelWidget {
	b.menu = items
	b.el.Role(ui.RoleMenuButton)
	b.resize()
	b.trackMenu()
	return b
}

// Size sets the button's size, which is otherwise as its content needs.
func (b *BevelWidget) Size(w, h int) *BevelWidget {
	b.w, b.h = w, h
	b.resize()
	return b
}

// Disabled dims the button and makes it take no input while d is true.
func (b *BevelWidget) Disabled(d bool) *BevelWidget {
	b.disabled = d
	b.el.Disabled(d)
	return b
}

// Clicked reports whether the user clicked the button since the last
// frame. A toggle has turned over by then.
func (b *BevelWidget) Clicked() bool {
	if b.clicked {
		return report(b.el, true)
	}
	if b.disabled || b.menu != nil || !b.el.Clicked() {
		return false
	}
	b.clicked = true
	return true
}

// Element returns the button's element.
func (b *BevelWidget) Element() ui.Element { return b.el }

func (b *BevelWidget) contentSize() (int, int) {
	return bevelContentSize(b.label, b.icon, b.menu != nil && !b.noArrow)
}

func bevelContentSize(label string, icon []string, arrow bool) (int, int) {
	w, h := 0, SystemFont.CapHeight()
	if label != "" {
		w = SystemFont.Measure(label)
	}
	if icon != nil {
		iw, ih := iconWidth(icon), len(icon)
		if w > 0 {
			w += bevelGap
		}
		w += iw
		h = max(h, ih)
	}
	if arrow {
		w += bevelArrowW
	}
	return w, h
}

func (b *BevelWidget) resize() {
	cw, ch := b.contentSize()
	w, h := cw+2*bevelPadding, max(BevelHeight, ch+10)
	if b.label == "" && b.icon != nil && b.menu == nil {
		// Icon buttons are square, three pixels round their icon: 22×22
		// for the 16×16 icons of a toolbar.
		w = max(BevelHeight, max(iconWidth(b.icon), len(b.icon))+6)
		h = w
	}
	if b.w > 0 {
		w = b.w
	}
	if b.h > 0 {
		h = b.h
	}
	b.el.Size(float32(w), float32(h))
}

func (b *BevelWidget) trackMenu() {
	st := b.st
	cw, ch := b.c.Size()
	if b.el.Clicked() || b.el.Shortcut(0, ui.KeyDown) {
		if st.t == nil && !b.disabled {
			g := layoutMenu(b.menu, round(st.rect.X), round(st.rect.Y+st.rect.H)-1, round(st.rect.W))
			st.t = newTracker(b.menu, fitMenu(g, round(cw), round(ch)))
			st.t.keyboard = true
			st.t.key(frameNow(b.c), ui.KeyDown)
		}
	}
	if st.t != nil {
		if it, _, ok := st.t.done(frameNow(b.c)); ok {
			st.t = nil
			if it.Action != nil {
				it.Action()
				report(b.el, true)
			}
		} else if st.t.closed {
			st.t = nil
		}
	}
	if st.t != nil {
		trackerOverlay(b.c, st.t, b.th.Accent, nil, nil)
		b.el.Expanded(true)
	}
}

func (b *BevelWidget) draw(p *ui.Painter, r ui.Rect) {
	resized(b.c, b.st.rect, r)
	b.st.rect = r
	pn := NewPen(p, r).Faded(b.disabled)
	s := StateNormal
	switch {
	case b.disabled:
		s = StateDisabled
	case b.inactive:
		s = StateInactive
	case b.pressed || b.st.t != nil:
		s = StatePressed
	}
	on := b.on != nil && *b.on
	DrawBevelButton(pn, s, on)
	ink := Black
	if s == StateDisabled || s == StateInactive {
		ink = Gray8
	}
	cw, ch := b.contentSize()
	x := (pn.W - cw) / 2
	if s == StatePressed || on {
		x++ // the content sinks with the button
	}
	dy := 0
	if s == StatePressed || on {
		dy = 1
	}
	if b.icon != nil {
		ih := len(b.icon)
		DrawIcon(pn.Sub(x, (pn.H-ih)/2+dy, iconWidth(b.icon), ih).Dimmed(s == StateInactive), b.icon, b.th.Accent, false)
		x += iconWidth(b.icon) + bevelGap
	}
	if b.label != "" {
		pn.TextMiddle(x, dy, pn.H, b.label, SystemFont, ink)
	}
	_ = ch
	if b.menu != nil && !b.noArrow {
		pn.Bitmap(pn.W-bevelPadding-5, (pn.H-3)/2+dy, []string{"#####", ".###.", "..#.."}, ink)
	}
	if b.focused && !b.disabled {
		DrawFocusRing(pn, 0, b.th.Accent.Dark)
	}
}

// DrawBevelButton draws the face of a bevel button filling pn's box, pushed
// in while pressed or on.
func DrawBevelButton(pn Pen, s State, on bool) {
	w, h := pn.W, pn.H
	frame, face, tl1, br1, tl2, br2 := Gray10, Gray2, White, Gray7, Gray1, Gray5
	switch {
	case s == StatePressed:
		face, tl1, br1, tl2, br2 = Gray6, Gray9, Gray4, Gray8, Gray5
	case on:
		face, tl1, br1, tl2, br2 = Gray4, Gray7, Gray1, Gray6, Gray3
	}
	switch s {
	case StateDisabled:
		frame, tl1, br1, tl2, br2 = Gray6, Gray1, Gray4, Gray2, Gray2
		if on {
			face = Gray3
		}
	case StateInactive:
		frame, tl1, br1, tl2, br2 = Gray8, Gray2, Gray2, Gray2, Gray2
		if on {
			face = Gray3
		}
	}
	// Square corners and a dark gray outline set bevel buttons apart from
	// push buttons, which are rounded and outlined in black.
	pn.Rect(1, 1, w-2, h-2, face)
	pn.Frame(0, 0, w, h, frame)
	pn.Bevel(1, 1, w-2, h-2, tl1, br1)
	pn.Bevel(2, 2, w-4, h-4, tl2, br2)
}

// SegmentedWidget is a segmented control being built.
type SegmentedWidget struct {
	el ui.Element
}

// Segmented builds a row of bevel buttons joined into one control, of which
// one, *selected, is pressed in: a choice of a few views or modes. The
// arrows move the choice while it has the keyboard focus.
func Segmented(c *ui.Context, selected *int, labels []string) *SegmentedWidget {
	th := ThemeOf(c)
	parts := ui.SegmentedBase(c, selected, len(labels))
	w := &SegmentedWidget{el: parts.Track}
	parts.Track.Children(func() {
		for i, l := range labels {
			seg := parts.Segment(i).FocusRing(false).Label(l)
			sw := SystemFont.Measure(l) + 2*bevelPadding + 4
			seg.Size(float32(sw), BevelHeight)
			if i > 0 {
				seg.Margin(0, 0, 0, -1) // neighbors share their outline
			}
			on := i == *selected
			pressed := seg.Pressed()
			focused := seg.FocusVisible()
			seg.Draw(func(p *ui.Painter, r ui.Rect) {
				pn := NewPen(p, r)
				s := StateNormal
				if pressed {
					s = StatePressed
				}
				DrawBevelButton(pn, s, on)
				dy := 0
				if on || pressed {
					dy = 1
				}
				pn.TextCentered(0, dy, pn.W, pn.H, l, SystemFont, Black)
				if focused {
					pn.Rect(5, pn.H-5, pn.W-10, 2, th.Accent.Dark)
				}
			})
		}
	})
	return w
}

// Changed reports whether the user chose another segment since the last
// frame.
func (w *SegmentedWidget) Changed() bool { return w.el.Changed() }

// Element returns the control's element.
func (w *SegmentedWidget) Element() ui.Element { return w.el }

// Label names the segmented control.
func (w *SegmentedWidget) Label(s string) *SegmentedWidget { w.el.Label(s); return w }
