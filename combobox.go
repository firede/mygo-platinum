package platinum

import (
	"strings"

	"github.com/egoist/mygo/ui"
)

// ComboboxWidget is a combo box being built.
type ComboboxWidget struct {
	el      ui.Element
	field   ui.Element
	button  ui.Element
	changed bool
	locale  string
}

type comboState struct {
	t    *menuTracker
	rect ui.Rect
}

// Combobox builds a text field editing *value with an arrow button at its
// right that opens a menu of items: those that start as the text does, or
// all of them while it is empty. Choosing one puts it in the field. Down
// opens the menu from the keyboard.
//
// The counterpart among MyGo's own controls is ui.Combobox.
func Combobox(c *ui.Context, value *string, items []string) *ComboboxWidget {
	th := ThemeOf(c)
	w := &ComboboxWidget{locale: c.Locale()}
	w.el = ui.Row(c).Gap(0).AlignItems(ui.Center).AlignSelf(ui.Start)
	st := ui.Local(w.el, "combo", func() comboState { return comboState{} })
	cw, ch := c.Size()
	w.el.Draw(func(p *ui.Painter, r ui.Rect) {
		resized(c, st.rect, r)
		st.rect = r
	})
	open := func() {
		var menu []MenuItem
		for _, it := range items {
			if *value == "" || strings.HasPrefix(strings.ToLower(it), strings.ToLower(*value)) {
				menu = append(menu, MenuItem{Title: it, Checked: it == *value})
			}
		}
		if len(menu) == 0 {
			return
		}
		g := layoutMenu(menu, round(st.rect.X), round(st.rect.Y+st.rect.H)-1, round(st.rect.W))
		st.t = newTracker(menu, fitMenu(g, round(cw), round(ch)))
		st.t.keyboard = true
		st.t.key(frameNow(c), ui.KeyDown)
	}
	w.el.Children(func() {
		f := TextField(c, value).Width(140)
		w.field = f.Element()
		if f.Changed() {
			w.changed = true
		}
		if f.Element().Shortcut(0, ui.KeyDown) && st.t == nil {
			open()
		}
		btn := ui.ButtonBase(c).Size(PopupArrowBox+2, TextFieldHeight).FocusRing(false).Label(say(c, "choices")).Margin(0, 0, 0, -1)
		w.button = btn
		if btn.Clicked() && st.t == nil {
			open()
		}
		pressed := btn.Pressed() || st.t != nil
		btn.Draw(func(p *ui.Painter, r ui.Rect) {
			pn := NewPen(p, r)
			s := StateNormal
			if pressed {
				s = StatePressed
			}
			DrawBevelButton(pn, s, false)
			ink := Black
			pn.Bitmap((pn.W-7)/2, (pn.H-4)/2, popupArrow, ink)
		})
	})
	if st.t != nil {
		if it, _, ok := st.t.done(frameNow(c)); ok {
			st.t = nil
			if *value != it.Title {
				*value = it.Title
				w.changed = true
			}
		} else if st.t.closed {
			st.t = nil
		}
	}
	if st.t != nil {
		trackerOverlay(c, st.t, th.Accent, nil, nil)
	}
	report(w.el, w.changed)
	return w
}

// Changed reports whether the text changed since the last frame.
func (w *ComboboxWidget) Changed() bool { return report(w.el, w.changed) }

// Element returns the combo box's element.
func (w *ComboboxWidget) Element() ui.Element { return w.el }

// Label names the combo box for assistive technology and tests: the
// group, the field, and the button of its choices, as "<l> choices".
func (w *ComboboxWidget) Label(l string) *ComboboxWidget {
	w.el.Role(ui.RoleGroup).Label(l)
	w.field.Label(l)
	w.button.Label(catalog().Format(w.locale, "choicesOf", map[string]any{"name": l}))
	return w
}
