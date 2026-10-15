package platinum

import (
	"github.com/egoist/mygo/ui"
)

// Pop-up menu metrics, from docs/spec.md.
const (
	PopupHeight   = 20 // the height of a pop-up menu button
	PopupArrowBox = 18 // the width of the box holding its arrow
	popupTextLeft = 9
	popupLabelGap = 6
)

type popupState struct {
	t      *menuTracker
	rect   ui.Rect // the button on the window, as last painted
	sticky bool
}

// PopupMenuWidget is a pop-up menu being built.
type PopupMenuWidget struct {
	inactive bool
	el       ui.Element
	changed  bool
	disabled bool
}

// PopupMenu builds a pop-up menu choosing *selected, an index of items: a
// beveled button showing the item chosen, with an arrow box at its right.
// Pressing it shows the menu over it, the chosen item over the button;
// releasing over an item chooses it, and a click without a drag leaves the
// menu open for a second click. With the keyboard focus, Space, Return, Up
// and Down open it. An item "-" is a separator. A label, if not empty,
// shows before the button and names it.
//
// The counterpart among MyGo's own controls is ui.Select.
func PopupMenu(c *ui.Context, selected *int, label string, items []string) *PopupMenuWidget {
	th := ThemeOf(c)
	w := &PopupMenuWidget{inactive: inactive(c)}
	menu := make([]MenuItem, len(items))
	textW := 0
	for i, s := range items {
		if s == "-" {
			menu[i] = SeparatorItem()
			continue
		}
		menu[i] = MenuItem{Title: s, Checked: i == *selected}
		textW = max(textW, SystemFont.Measure(s))
	}
	*selected = min(max(*selected, 0), max(len(items)-1, 0))
	cur := ""
	if *selected < len(items) {
		cur = items[*selected]
	}
	labelW := 0
	if label != "" {
		labelW = SystemFont.Measure(label) + popupLabelGap
	}
	btnW := popupTextLeft + textW + 8 + PopupArrowBox

	w.el = ui.Box(c).Focusable().FocusRing(false).Role(ui.RolePopUpButton).
		Label(label).Value(cur).Size(float32(labelW+btnW), PopupHeight)
	if label == "" {
		w.el.Label(cur)
	}
	st := ui.Local(w.el, "popup", func() popupState { return popupState{} })
	w.el.Expanded(st.t != nil)
	now := frameNow(c)
	cw, ch := c.Size()

	open := func(keyboard bool) {
		if len(items) == 0 || w.disabled {
			return
		}
		x := round(st.rect.X) + labelW
		y := round(st.rect.Y)
		g := layoutMenu(menu, 0, 0, btnW)
		// The chosen item over the button.
		g.X, g.Y = x, y-g.tops[*selected]+(PopupHeight-MenuItemHeight)/2
		if g.Y < 0 {
			g.Y = 0
		}
		g = fitMenu(g, round(cw), round(ch))
		st.t = newTracker(menu, g)
		st.t.hl = *selected
		st.t.keyboard = keyboard
		st.sticky = keyboard
	}
	w.el.HandleInput(func(ev ui.InputEvent) bool {
		if w.disabled {
			return false
		}
		wx, wy := st.rect.X+ev.X, st.rect.Y+ev.Y
		switch ev.Kind {
		case ui.InputPointerDown:
			if ev.Button != 0 || ev.X < float32(labelW) {
				return false
			}
			open(false)
			if st.t != nil {
				st.t.press(wallNow(), st.rect.X+ev.X, st.rect.Y+ev.Y)
			}
			return true
		case ui.InputPointerMove:
			if st.t == nil || ev.Button != 0 {
				return false
			}
			st.t.move(wx, wy)
			return true
		case ui.InputPointerUp:
			if st.t == nil {
				return false
			}
			st.t.pressed = false
			if st.t.quickClick(wallNow()) {
				st.sticky = true
				return true
			}
			if st.t.release(wallNow(), wx, wy) {
				return true
			}
			if wx >= st.rect.X+float32(labelW) && wx < st.rect.X+st.rect.W && wy >= st.rect.Y && wy < st.rect.Y+st.rect.H {
				st.sticky = true
				st.t.hl = *selected
				return true
			}
			st.t = nil
			return true
		}
		return false
	})
	for _, k := range []ui.Key{ui.KeySpace, ui.KeyEnter, ui.KeyUp, ui.KeyDown} {
		if st.t == nil && w.el.Shortcut(0, k) {
			open(true)
		}
	}
	if st.t != nil {
		if _, path, ok := st.t.done(now); ok {
			st.t = nil
			if path[0] != *selected {
				*selected = path[0]
				w.changed = true
			}
		} else if st.t.closed {
			st.t = nil
		}
	}
	if st.t != nil {
		trackerOverlay(c, st.t, th.Accent, nil, nil)
	}
	pressed := st.t != nil && st.t.pressed
	focused := w.el.FocusVisible()
	w.el.Draw(func(p *ui.Painter, r ui.Rect) {
		st.rect = ui.Rect{X: r.X, Y: r.Y, W: r.W, H: r.H}
		pn := NewPen(p, r).Faded(w.disabled)
		s := StateNormal
		switch {
		case w.disabled:
			s = StateDisabled
		case w.inactive:
			s = StateInactive
		case pressed:
			s = StatePressed
		}
		if label != "" {
			pn.TextMiddle(0, 0, pn.H, label, SystemFont, labelColor(s))
		}
		btn := pn.Sub(labelW, 0, btnW, PopupHeight)
		DrawPopupButton(btn, cur, s)
		if focused && !w.disabled {
			DrawFocusRing(btn, 3, th.Accent.Dark)
		}
	})
	report(w.el, w.changed)
	return w
}

// Disabled dims the pop-up menu and makes it take no input while d is true.
func (w *PopupMenuWidget) Disabled(d bool) *PopupMenuWidget {
	w.disabled = d
	w.el.Disabled(d)
	return w
}

// Changed reports whether the user chose another item since the last frame.
func (w *PopupMenuWidget) Changed() bool { return report(w.el, w.changed) }

// Element returns the pop-up menu's element.
func (w *PopupMenuWidget) Element() ui.Element { return w.el }

var popupArrow = []string{
	"#######",
	".#####.",
	"..###..",
	"...#...",
}

// DrawPopupButton draws the button of a pop-up menu showing text, filling
// pn's box.
func DrawPopupButton(pn Pen, text string, s State) {
	DrawButton(pn, "", s)
	w, h := pn.W, pn.H
	ink := Black
	switch s {
	case StatePressed:
		ink = White
	case StateDisabled, StateInactive:
		ink = Gray8
	}
	// The arrow box, set off by an etched line.
	ax := w - PopupArrowBox
	line, hi := Gray7, White
	if s == StatePressed {
		line, hi = Gray11, Gray8
	}
	if s == StateDisabled || s == StateInactive {
		line, hi = Gray4, Gray1
	}
	pn.VLine(ax, 3, h-6, line)
	pn.VLine(ax+1, 3, h-6, hi)
	pn.Bitmap(ax+(PopupArrowBox-7)/2+1, (h-4)/2, popupArrow, ink)
	pn.Clip(2, 0, ax-4, h, func() {
		pn.TextMiddle(popupTextLeft, 0, h, SystemFont.Truncate(text, ax-popupTextLeft-4), SystemFont, ink)
	})
}
