package platinum

import (
	"math"
	"strconv"
	"time"

	"github.com/egoist/mygo/ui"
)

// Little arrows metrics, from docs/spec.md.
const (
	StepperWidth  = 13 // the width of the little arrows
	StepperHeight = 23 // their height
)

type stepperState struct {
	part    int // 1 up, -1 down, 0 none: the half pressed
	inside  bool
	next    time.Time
	changed bool
	picked  float64 // the value the pointer set between builds
}

// StepperWidget is a pair of little arrows being built.
type StepperWidget struct {
	inactive bool
	el       ui.Element
	st       *stepperState
	value    *float64
	lo, hi   float64
	step     float64
	disabled bool
	changed  bool
	focused  bool
	th       *Theme
}

// Stepper builds the little arrows: two small beveled halves, up and down,
// that add step to *value and take it away, between lo and hi, again and
// again while held. Up and Down do the same while it has the keyboard
// focus. Label it for assistive technology.
func Stepper(c *ui.Context, value *float64, lo, hi, step float64) *StepperWidget {
	s := &StepperWidget{value: value, lo: lo, hi: hi, step: step, th: ThemeOf(c), inactive: inactive(c)}
	s.el = ui.Box(c).Size(StepperWidth, StepperHeight).Focusable().FocusRing(false).Role(ui.RoleStepper)
	s.el.Range(lo, hi, *value).Step(step)
	s.st = ui.Local(s.el, "stepper", func() stepperState { return stepperState{} })
	st := s.st
	if st.changed {
		*value = st.picked
	}
	s.changed, st.changed = st.changed, false
	now := frameNow(c)
	if st.part != 0 {
		if st.inside && !now.Before(st.next) {
			s.add(float64(st.part))
			st.next = now.Add(scrollRepeat)
		}
		c.After(scrollRepeat)
	}
	if s.el.Shortcut(0, ui.KeyUp) {
		s.add(1)
	}
	if s.el.Shortcut(0, ui.KeyDown) {
		s.add(-1)
	}
	s.focused = s.el.FocusVisible()
	s.el.HandleInput(func(ev ui.InputEvent) bool {
		if s.disabled {
			return false
		}
		half := 1
		if ev.Y >= StepperHeight/2 {
			half = -1
		}
		switch ev.Kind {
		case ui.InputPointerDown:
			if ev.Button != 0 {
				return false
			}
			st.part, st.inside = half, true
			s.add(float64(half))
			st.next = wallNow().Add(scrollDelay)
			return false // and take the focus
		case ui.InputPointerMove:
			if st.part != 0 {
				st.inside = half == st.part && ev.X >= 0 && ev.X < StepperWidth && ev.Y >= 0 && ev.Y < StepperHeight
				return true
			}
		case ui.InputPointerUp:
			st.part, st.inside = 0, false
		}
		return false
	})
	s.el.Draw(func(p *ui.Painter, r ui.Rect) {
		pn := NewPen(p, r).Faded(s.disabled)
		pressed := 0
		if st.inside {
			pressed = st.part
		}
		state := StateNormal
		switch {
		case s.disabled:
			state = StateDisabled
		case s.inactive:
			state = StateInactive
		}
		DrawStepper(pn, pressed, state)
		if s.focused && !s.disabled {
			DrawFocusRing(pn, 1, s.th.Accent.Dark)
		}
	})
	return s
}

func (s *StepperWidget) add(n float64) {
	v := min(max(*s.value+n*s.step, s.lo), s.hi)
	if s.step > 0 {
		v = math.Round(v/s.step) * s.step
		v = min(max(v, s.lo), s.hi)
	}
	if v != *s.value {
		*s.value = v
		s.changed = true
		s.st.changed = true
		s.st.picked = v
	}
}

// Label names the arrows for assistive technology and tests.
func (s *StepperWidget) Label(l string) *StepperWidget {
	s.el.Label(l)
	return s
}

// Disabled dims the arrows and makes them take no input while d is true.
func (s *StepperWidget) Disabled(d bool) *StepperWidget {
	s.disabled = d
	s.el.Disabled(d)
	return s
}

// Changed reports whether the user changed the value since the last frame.
func (s *StepperWidget) Changed() bool { return s.changed }

// Element returns the arrows' element.
func (s *StepperWidget) Element() ui.Element { return s.el }

// DrawStepper draws the little arrows in pn's 13×23 corner, the half
// pressed (1 up, -1 down) darkened, in state s (normal, disabled or
// inactive).
func DrawStepper(pn Pen, pressed int, s State) {
	w, h := StepperWidth, StepperHeight
	frame, ink := Black, Black
	disabled := s == StateDisabled || s == StateInactive
	switch s {
	case StateDisabled:
		frame, ink = Gray6, Gray6
	case StateInactive:
		frame, ink = Gray8, Gray8
	}
	pn.RoundFrame(0, 0, w, h, 2, frame, Gray2)
	mid := h / 2
	pn.HLine(1, mid, w-2, frame)
	for i, half := range []int{1, -1} {
		y0, y1 := 1, mid
		if i == 1 {
			y0, y1 = mid+1, h-1
		}
		face, tl, br := Gray2, White, Gray7
		if pressed == half {
			face, tl, br = Gray6, Gray9, Gray4
		}
		if disabled {
			tl, br = Gray2, Gray2
		}
		pn.Rect(2, y0+1, w-4, y1-y0-2, face)
		pn.Bevel(1, y0, w-2, y1-y0, tl, br)
		arrow := []string{"..#..", ".###.", "#####"}
		if half < 0 {
			arrow = []string{"#####", ".###.", "..#.."}
		}
		pn.Bitmap((w-5)/2, y0+(y1-y0-3)/2, arrow, ink)
	}
}

type numberState struct {
	text    string
	shown   float64
	editing bool
}

// NumberFieldWidget is a number field being built.
type NumberFieldWidget struct {
	el      ui.Element
	field   ui.Element
	arrows  *StepperWidget
	changed bool
}

// NumberField builds a text field of a number between lo and hi with little
// arrows beside it stepping it by step. Typing changes the number once the
// text is a number; Return or leaving the field puts the text back in
// order. digits sets how many decimals it shows.
func NumberField(c *ui.Context, value *float64, lo, hi, step float64, digits int) *NumberFieldWidget {
	return numberField(c, value, lo, hi, step, digits, 0, 56)
}

// numberField is NumberField with the number padded with zeros to pad
// digits, in a field width wide.
func numberField(c *ui.Context, value *float64, lo, hi, step float64, digits, pad, width int) *NumberFieldWidget {
	w := &NumberFieldWidget{}
	w.el = ui.Row(c).Gap(3).AlignItems(ui.Center)
	st := ui.Local(w.el, "number", func() numberState { return numberState{} })
	format := func(v float64) string {
		s := strconv.FormatFloat(v, 'f', digits, 64)
		for len(s) < pad {
			s = "0" + s
		}
		return s
	}
	if !st.editing && (st.text == "" || st.shown != *value) {
		st.text, st.shown = format(*value), *value
	}
	w.el.Children(func() {
		f := TextField(c, &st.text).Width(width)
		f.Element().Role(ui.RoleTextField)
		w.field = f.Element()
		focused := f.Element().Focused()
		if f.Changed() {
			st.editing = true
			if v, err := strconv.ParseFloat(st.text, 64); err == nil && v >= lo && v <= hi {
				*value, st.shown = v, v
				w.changed = true
			}
		}
		if f.Submitted() || (st.editing && !focused) {
			st.editing = false
			st.text, st.shown = format(*value), *value
		}
		up, down := f.Element().Shortcut(0, ui.KeyUp), f.Element().Shortcut(0, ui.KeyDown)
		if up || down {
			d := step
			if down {
				d = -step
			}
			*value = min(max(*value+d, lo), hi)
			st.editing = false
			w.changed = true
		}
		w.arrows = Stepper(c, value, lo, hi, step)
		if w.arrows.Changed() {
			st.editing = false
			w.changed = true
		}
	})
	return w
}

// Label names the field for assistive technology and tests: the group,
// the field, and its little arrows "Change" it.
func (w *NumberFieldWidget) Label(l string) *NumberFieldWidget {
	w.el.Role(ui.RoleGroup).Label(l)
	w.field.Label(l)
	w.arrows.Label("Change " + l)
	return w
}

// Changed reports whether the user changed the number since the last frame.
func (w *NumberFieldWidget) Changed() bool { return w.changed }

// Element returns the field's element, its text field and arrows in a row.
func (w *NumberFieldWidget) Element() ui.Element { return w.el }
