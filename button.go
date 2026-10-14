package platinum

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Push button metrics, from docs/spec.md.
const (
	ButtonHeight   = 20 // the height of a push button
	ButtonMinWidth = 58 // the narrowest a push button gets
	ButtonPadding  = 12 // the room on each side of its label, at least
	DefaultRing    = 4  // how far a default button's ring reaches outside it
)

// ButtonWidget is a push button being built; its methods set it up and ask
// what the user did to it. Set it up before asking.
type ButtonWidget struct {
	c        *ui.Context
	el       ui.Element
	label    string
	width    int
	def      bool
	cancel   bool
	disabled bool
	inactive bool
	pressed  bool
	focused  bool
	flash    bool
	th       *Theme
	flashTo  *time.Time
}

// Button builds a push button showing label: a rounded, beveled gray
// button that turns dark while pressed. Space and Enter press it while it
// has the keyboard focus.
func Button(c *ui.Context, label string) *ButtonWidget {
	b := &ButtonWidget{c: c, label: label, th: ThemeOf(c), inactive: inactive(c)}
	b.el = ui.ButtonBase(c).FocusRing(false).Label(label)
	flashUntil := ui.Local(b.el, "flash", func() time.Time { return time.Time{} })
	b.pressed = b.el.Pressed()
	b.focused = b.el.FocusVisible()
	// A frozen clock shows the moment after the flash.
	if now := frameNow(c); now.Before(*flashUntil) && !frozen() {
		b.flash = true
		c.After(flashUntil.Sub(now))
	}
	b.width = max(ButtonMinWidth, SystemFont.Measure(label)+2*ButtonPadding)
	b.el.Size(float32(b.width), ButtonHeight)
	b.el.Draw(b.draw)
	b.flashTo = flashUntil
	return b
}

// Width sets the width of the button, which is otherwise as wide as its
// label needs, and at least ButtonMinWidth.
func (b *ButtonWidget) Width(w int) *ButtonWidget {
	b.width = w
	b.resize()
	return b
}

// Default makes the button the default one of its window or dialog: it
// shows a thick ring around it, and Return and Enter press it wherever the
// keyboard focus is, unless the focused control takes them.
func (b *ButtonWidget) Default() *ButtonWidget {
	b.def = true
	b.resize()
	return b
}

// Cancel makes Escape and Command-period press the button, as a dialog's
// Cancel button.
func (b *ButtonWidget) Cancel() *ButtonWidget {
	b.cancel = true
	return b
}

// Disabled dims the button and makes it take no clicks or keys while d is
// true.
func (b *ButtonWidget) Disabled(d bool) *ButtonWidget {
	b.disabled = d
	b.el.Disabled(d)
	return b
}

// Clicked reports whether the user clicked the button since the last
// frame: a press and release inside it, Space or Enter while it has the
// focus, its keyboard equivalent when it is the default or cancel button,
// or assistive technology pressing it.
func (b *ButtonWidget) Clicked() bool {
	if b.disabled {
		return false
	}
	if b.el.Clicked() {
		return true
	}
	if b.def && (b.c.Shortcut(0, ui.KeyEnter)) || b.cancel && (b.c.Shortcut(0, ui.KeyEscape) || b.c.Shortcut(ui.Cmd, ui.KeyPeriod)) {
		// Flash the button pressed, as it is when pressed by the keyboard.
		*b.flashTo = frameNow(b.c).Add(FlashDuration)
		b.c.After(FlashDuration)
		return true
	}
	return false
}

// Flash shows the button pressed for a moment, as when its keyboard
// equivalent presses it: for a key of your own that stands for it.
func (b *ButtonWidget) Flash() *ButtonWidget {
	*b.flashTo = frameNow(b.c).Add(FlashDuration)
	b.flash = true
	b.c.After(FlashDuration)
	return b
}

// FlashDuration is how long a button shows pressed when the keyboard
// presses it.
const FlashDuration = 130 * time.Millisecond

// Element returns the button's element, to lay it out or to ask more of it.
func (b *ButtonWidget) Element() ui.Element { return b.el }

func (b *ButtonWidget) resize() {
	w, h := b.width, ButtonHeight
	if b.def {
		w += 2 * DefaultRing
		h += 2 * DefaultRing
	}
	b.el.Size(float32(w), float32(h))
}

func (b *ButtonWidget) draw(p *ui.Painter, r ui.Rect) {
	pn := NewPen(p, r).Faded(b.disabled)
	x, y, w, h := 0, 0, pn.W, pn.H
	if b.def {
		ring := StateNormal
		switch {
		case b.disabled:
			ring = StateDisabled
		case b.inactive:
			ring = StateInactive
		}
		DrawDefaultRing(pn, 0, 0, w, h, ring)
		x, y, w, h = DefaultRing, DefaultRing, w-2*DefaultRing, h-2*DefaultRing
	}
	state := StateNormal
	switch {
	case b.disabled:
		state = StateDisabled
	case b.inactive:
		state = StateInactive
	case b.pressed || b.flash:
		state = StatePressed
	}
	DrawButton(pn.Sub(x, y, w, h), b.label, state)
	if b.focused && !b.disabled {
		DrawFocusRing(pn.Sub(x, y, w, h), 3, b.th.Accent.Dark)
	}
}

// State is how a control shows: normal, pressed or disabled.
type State int

// The states a control draws in.
const (
	StateNormal State = iota
	StatePressed
	StateDisabled
	// StateInactive is a control in a window in the background: its
	// outline and text gray, its bevels flat.
	StateInactive
)

// DrawButton draws a push button with label filling pn's box, for controls
// of your own that look like one.
func DrawButton(pn Pen, label string, s State) {
	w, h := pn.W, pn.H
	frame, face, tl1, br1, tl2, br2, text := Black, Gray2, White, Gray7, Gray1, Gray5, Black
	switch s {
	case StatePressed:
		face, tl1, br1, tl2, br2, text = Gray9, Gray11, Gray7, Gray10, Gray8, White
	case StateDisabled:
		frame, tl1, br1, tl2, br2, text = Gray6, Gray1, Gray4, Gray2, Gray2, Gray8
	case StateInactive:
		frame, tl1, br1, tl2, br2, text = Gray8, Gray2, Gray2, Gray2, Gray2, Gray8
	}
	pn.RoundFrame(0, 0, w, h, 3, frame, face)
	// Two rings of bevel inside the frame, which follow its corners.
	pn.HLine(3, 1, w-6, tl1)
	pn.VLine(1, 3, h-6, tl1)
	pn.Dot(2, 2, tl1)
	pn.HLine(3, h-2, w-6, br1)
	pn.VLine(w-2, 3, h-6, br1)
	pn.Dot(w-3, h-3, br1)
	pn.HLine(3, 2, w-7, tl2)
	pn.VLine(2, 3, h-7, tl2)
	pn.HLine(4, h-3, w-7, br2)
	pn.VLine(w-3, 4, h-7, br2)
	pn.TextCentered(0, 0, w, h, label, SystemFont, text)
}

// DrawDefaultRing draws the ring of a default button around the w×h box at
// x, y of pn, DefaultRing pixels wide, in state s (normal, disabled or
// inactive).
func DrawDefaultRing(pn Pen, x, y, w, h int, s State) {
	line, face, tl, br := Black, Gray4, White, Gray7
	switch s {
	case StateDisabled:
		line, face, tl, br = Gray6, Gray2, Gray1, Gray4
	case StateInactive:
		line, face, tl, br = Gray8, Gray2, Gray2, Gray2
	}
	pn.RoundFrame(x, y, w, h, 4, line, face)
	// The ring's own bevel, inside its outline.
	pn.HLine(x+4, y+1, w-8, tl)
	pn.VLine(x+1, y+4, h-8, tl)
	pn.Dot(x+2, y+2, tl)
	pn.Dot(x+3, y+2, tl)
	pn.Dot(x+2, y+3, tl)
	pn.HLine(x+4, y+h-2, w-8, br)
	pn.VLine(x+w-2, y+4, h-8, br)
	pn.Dot(x+w-3, y+h-3, br)
	pn.Dot(x+w-4, y+h-3, br)
	pn.Dot(x+w-3, y+h-4, br)
	// The ring's inner edge, against the button: flat in the background.
	edge := Gray5
	if s == StateInactive {
		edge = Gray2
	}
	pn.Rect(x+3, y+3, w-6, h-6, edge)
}

// DrawFocusRing draws the ring of the keyboard focus around pn's box: two
// pixels of c, a pixel outside it, its corners cut by radius r.
func DrawFocusRing(pn Pen, r int, c ui.Color) {
	w, h := pn.W+6, pn.H+6
	o := pn.Sub(-3, -3, w, h)
	o.RoundFrame(0, 0, w, h, min(r+1, 4), c, ui.Transparent)
	o.RoundFrame(1, 1, w-2, h-2, r, c, ui.Transparent)
}
