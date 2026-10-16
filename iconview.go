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
	scroll             float64 // the offset in whole DIPs when painted
	changed, activated bool
	picked             int // chosen between builds, for the next one
}

type iconViewLayout struct {
	width, height, cols, total int
	scrolls                    bool
}

func layoutIconView(r ui.Rect, n int) iconViewLayout {
	g := iconViewLayout{width: max(round(r.W)-4, 0), height: max(round(r.H)-4, 0)}
	g.cols = max(g.width/IconCellW, 1)
	contentHeight := func() int {
		if n == 0 {
			return 0
		}
		return 8 + (n+g.cols-1)/g.cols*IconCellH
	}
	g.total = contentHeight()
	if g.total > g.height {
		g.scrolls = true
		// The bar shares the content's right outline.
		g.width = max(g.width-ScrollBarWidth+1, 0)
		g.cols = max(g.width/IconCellW, 1)
		g.total = contentHeight()
	}
	return g
}

// IconView builds a grid of icons with their names under them, choosing
// *selected (-1 for none): a click chooses an icon, darkening it and
// inverting its name, and a double click activates it; with the keyboard
// focus, the arrows, Page Up and Down, Home and End move the choice into
// view and Return activates. A scroll bar at the right and the wheel scroll
// the icons when they do not all fit. Give it a size.
//
// The counterpart among MyGo's own controls is ui.GridView.
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
	n := len(items)
	if *selected >= n {
		*selected = n - 1
	}
	r := st.rect
	if r.W == 0 {
		r = ui.Rect{W: 260, H: 140}
	}
	grid := layoutIconView(r, n)
	clamp := func() {
		st.scroll = min(max(st.scroll, 0), float64(max(grid.total-grid.height, 0)))
	}
	clamp()
	reveal := func(i int) {
		top := float64(4 + i/grid.cols*IconCellH)
		bottom := top + IconCellH
		if top < st.scroll {
			st.scroll = top
		} else if bottom > st.scroll+float64(grid.height) {
			st.scroll = min(top, bottom-float64(grid.height))
		}
		clamp()
	}
	move := func(to int) {
		if n == 0 {
			return
		}
		to = min(max(to, 0), n-1)
		if to != *selected {
			*selected = to
			w.changed = true
		}
		reveal(to)
	}
	for _, k := range []struct {
		key ui.Key
		d   int
	}{{ui.KeyRight, 1}, {ui.KeyLeft, -1}, {ui.KeyDown, grid.cols}, {ui.KeyUp, -grid.cols},
		{ui.KeyPageDown, max(grid.height/IconCellH-1, 1) * grid.cols},
		{ui.KeyPageUp, -max(grid.height/IconCellH-1, 1) * grid.cols}} {
		if w.el.Shortcut(0, k.key) {
			if *selected < 0 {
				move(0)
			} else {
				move(*selected + k.d)
			}
		}
	}
	if w.el.Shortcut(0, ui.KeyHome) {
		move(0)
	}
	if w.el.Shortcut(0, ui.KeyEnd) {
		move(n - 1)
	}
	if (w.el.Shortcut(0, ui.KeyEnter) || w.el.DoubleClicked()) && *selected >= 0 {
		w.activated = true
	}
	w.el.HandleInput(func(ev ui.InputEvent) bool {
		switch ev.Kind {
		case ui.InputPointerDown:
			x, y := ev.X-in, ev.Y-in
			if ev.Button != 0 || x < 0 || y < 0 || x >= float32(grid.width) || y >= float32(grid.height) {
				return false
			}
			y += float32(int(st.scroll)) - 4
			i := int(y)/IconCellH*grid.cols + int(x)/IconCellW
			if int(x)/IconCellW >= grid.cols || i >= n || y < 0 {
				i = -1
			}
			if i != *selected {
				*selected = i
				st.picked = i
				st.changed = true
			}
		case ui.InputScroll:
			st.scroll += float64(ev.DY)
			clamp()
			return true
		}
		return false
	})
	focused := w.el.Focused()
	sel := *selected
	offset := int(st.scroll)
	w.el.Draw(func(p *ui.Painter, r ui.Rect) {
		resized(c, st.rect, r)
		st.rect = r
		pn := NewPen(p, r)
		DrawContentRing(pn)
		pn.Frame(1, 1, pn.W-2, pn.H-2, Black)
		// A resize can paint before the next build has used the new size.
		pn.Clip(in, in, max(pn.W-2*in, 0), max(pn.H-2*in, 0), func() {
			drawIconView(pn.Sub(in, in, grid.width, grid.height), items, sel, focused, th.Accent, offset)
		})
	})
	w.el.DrawOver(func(p *ui.Painter, r ui.Rect) {
		if focused {
			DrawFocusRing(NewPen(p, r), 0, th.Accent.Dark)
		}
	})
	w.el.Children(func() {
		ui.Box(c.Key("viewport")).Absolute().Left(in).Top(in).Size(float32(grid.width), float32(grid.height)).Clip().PassThrough().Children(func() {
			for i, it := range items {
				y := 4 + i/grid.cols*IconCellH - offset
				if y+IconCellH <= 0 || y >= grid.height {
					continue
				}
				ui.Box(c.Key(i)).Absolute().Left(float32(i%grid.cols*IconCellW)).Top(float32(y)).
					Size(IconCellW, IconCellH).Role(ui.RoleListItem).Label(it.Name).
					Selected(i == sel).Background(ui.Transparent).PassThrough()
			}
		})
		if grid.scrolls {
			ScrollBar(c.Key("scrollbar"), &st.scroll, float64(grid.height), float64(grid.total)).Step(IconCellH / 4).
				Element().Absolute().Left(float32(in + grid.width)).Top(1).Height(float32(grid.height + 2))
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
	drawIconView(pn, items, sel, focused, a, 0)
	if focused {
		DrawFocusRing(pn.Sub(-2, -2, pn.W+4, pn.H+4), 0, a.Dark)
	}
}

func drawIconView(pn Pen, items []IconItem, sel int, focused bool, a Accent, offset int) {
	if pn.W <= 0 || pn.H <= 0 {
		return
	}
	pn.Rect(0, 0, pn.W, pn.H, White)
	cols := max(pn.W/IconCellW, 1)
	pn.Clip(0, 0, pn.W, pn.H, func() {
		for i, it := range items {
			y := 4 + i/cols*IconCellH - offset
			if y >= pn.H {
				break
			}
			if y+IconCellH <= 0 {
				continue
			}
			cell := pn.Sub(i%cols*IconCellW, y, IconCellW, IconCellH)
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
	})
}
