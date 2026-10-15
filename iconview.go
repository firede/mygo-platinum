package platinum

import (
	"github.com/egoist/mygo/ui"
)

// IconItem is an item of an icon view: a name under a 32×32 picture.
type IconItem struct {
	Name string
	Icon *ui.Bitmap
}

// Icon view metrics, from docs/spec.md.
const (
	IconCellW = 84 // the width of an icon view's cell
	IconCellH = 64 // its height
)

// IconViewWidget is an icon view being built.
type IconViewWidget struct {
	el        ui.Element
	changed   bool
	activated bool
}

type iconViewState struct {
	rect               ui.Rect
	changed, activated bool
	picked             int // chosen between builds, for the next one
}

// IconView builds a grid of icons with their names under them, choosing
// *selected (-1 for none): a click chooses an icon, darkening it and
// inverting its name, and a double click activates it; with the keyboard
// focus, the arrows move the choice and Return activates. Give it a size.
func IconView(c *ui.Context, items []IconItem, selected *int) *IconViewWidget {
	th := ThemeOf(c)
	w := &IconViewWidget{}
	w.el = ui.Box(c).Focusable().FocusRing(false).Role(ui.RoleList).Size(260, 140).Clip()
	const in = 2 // the ring and outline around the icons
	st := ui.Local(w.el, "iconview", func() iconViewState { return iconViewState{} })
	if st.rect.W == 0 {
		c.After(0) // lay out again once the width is known
	}
	if st.changed {
		*selected = st.picked
	}
	w.changed, w.activated = st.changed, st.activated
	st.changed, st.activated = false, false
	cols := max(int(st.rect.W-2*in)/IconCellW, 1)
	n := len(items)
	move := func(to int) {
		if n == 0 {
			return
		}
		to = min(max(to, 0), n-1)
		if to != *selected {
			*selected = to
			w.changed = true
		}
	}
	for _, k := range []struct {
		key ui.Key
		d   int
	}{{ui.KeyRight, 1}, {ui.KeyLeft, -1}, {ui.KeyDown, cols}, {ui.KeyUp, -cols}} {
		if w.el.Shortcut(0, k.key) {
			if *selected < 0 {
				move(0)
			} else {
				move(*selected + k.d)
			}
		}
	}
	if (w.el.Shortcut(0, ui.KeyEnter) || w.el.DoubleClicked()) && *selected >= 0 {
		w.activated = true
	}
	w.el.HandleInput(func(ev ui.InputEvent) bool {
		if ev.Kind == ui.InputPointerDown && ev.Button == 0 {
			x, y := ev.X-in, ev.Y-in
			i := int(y-4)/IconCellH*cols + int(x)/IconCellW
			if x < 0 || int(x)/IconCellW >= cols || i >= n || y < 4 {
				i = -1
			}
			if i != *selected {
				*selected = i
				st.picked = i
				st.changed = true
			}
		}
		return false
	})
	focused := w.el.Focused()
	sel := *selected
	w.el.DrawOver(func(p *ui.Painter, r ui.Rect) {
		resized(c, st.rect, r)
		st.rect = r
		pn := NewPen(p, r)
		DrawContentRing(pn)
		pn.Frame(1, 1, pn.W-2, pn.H-2, Black)
		DrawIconView(pn.Sub(in, in, pn.W-2*in, pn.H-2*in), items, sel, focused, th.Accent)
	})
	w.el.Children(func() {
		for i, it := range items {
			ui.Box(c).Absolute().Left(float32(in+i%cols*IconCellW)).Top(float32(in+4+i/cols*IconCellH)).
				Size(IconCellW, IconCellH).Role(ui.RoleListItem).Label(it.Name).PassThrough()
		}
	})
	report(w.el, w.changed || w.activated)
	return w
}

// Changed reports whether the user chose another icon since the last frame.
func (w *IconViewWidget) Changed() bool { return report(w.el, w.changed) }

// Submitted reports whether the user double-clicked an icon, or pressed
// Return, since the last frame, as MyGo's lists report it.
func (w *IconViewWidget) Submitted() bool { return report(w.el, w.activated) }

// Element returns the icon view's element, to size it.
func (w *IconViewWidget) Element() ui.Element { return w.el }

// Label names the icon view.
func (w *IconViewWidget) Label(s string) *IconViewWidget { w.el.Label(s); return w }

// DrawIconView draws the icons of an icon view in pn's box, icon sel
// selected.
func DrawIconView(pn Pen, items []IconItem, sel int, focused bool, a Accent) {
	pn.Rect(0, 0, pn.W, pn.H, White)
	cols := max(pn.W/IconCellW, 1)
	for i, it := range items {
		cell := pn.Sub(i%cols*IconCellW, 4+i/cols*IconCellH, IconCellW, IconCellH)
		if it.Icon != nil {
			o := ui.ImageOptions{}
			if i == sel {
				o.Tint = ui.RGB(0x80, 0x80, 0x80) // darkened by half
			}
			pn.P.ImageRegion(it.Icon, ui.Rect{W: 32, H: 32},
				ui.Rect{X: float32(cell.X + (IconCellW-32)/2), Y: float32(cell.Y + 4), W: 32, H: 32}, o)
		}
		name := AppFont.Truncate(it.Name, IconCellW-4)
		tw := AppFont.Measure(name)
		x := (IconCellW - tw) / 2
		bg, ink := White, Black
		if i == sel {
			bg, ink = Black, White
			if focused {
				bg = a.Dark
			}
		}
		cell.Rect(x-2, 38, tw+4, AppFont.Height()+1, bg)
		cell.Text(x, 39, name, AppFont, ink)
	}
	if focused {
		DrawFocusRing(pn.Sub(-2, -2, pn.W+4, pn.H+4), 0, a.Dark)
	}
}
