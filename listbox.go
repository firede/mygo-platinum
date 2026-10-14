package platinum

import (
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

// List box metrics, from docs/spec.md.
const (
	ListRowHeight = 16 // the height of a row of a list box
	listTextLeft  = 4
	typeSelectGap = time.Second // how long typed letters wait for the next
)

type listState struct {
	top float64 // the first row in view, in rows
	// What the user did between builds, which the next build reports.
	changed, activated bool
	// picked is the row the pointer or a key chose between builds, which
	// the next build puts in *selected: the pointer the handler holds may
	// be a variable of an earlier build.
	picked   int
	rect     ui.Rect
	typed    string
	typedAt  time.Time
	dragging bool
	always   bool // ScrollBarAlways, from the frame it was asked in
}

// ListBoxWidget is a list box being built.
type ListBoxWidget struct {
	arrows    []bool
	icons     [][]string
	dim       []bool
	always    bool // the scroll bar shows when the rows fit too
	noRing    bool // the list draws no focus ring: what holds it does
	st        *listState
	c         *ui.Context
	el        ui.Element
	changed   bool
	activated bool
	disabled  bool
}

// ListBox builds a list box of items choosing *selected (-1 for none): a
// white, sunken box of rows, the chosen one highlighted, with a scroll bar
// at its right when the rows do not all fit. A click chooses a row, and a
// double click activates it; with the keyboard focus, the arrows, Page Up
// and Down, Home and End move the choice, typing the start of an item
// chooses it, and Return activates it. Give it a size; its height in rows
// is its height divided by ListRowHeight.
func ListBox(c *ui.Context, selected *int, items []string) *ListBoxWidget {
	th := ThemeOf(c)
	w := &ListBoxWidget{c: c}
	w.el = ui.Row(c).Focusable().FocusRing(false).Role(ui.RoleList).Size(160, 96).AlignItems(ui.Stretch).Padding(1)
	st := ui.Local(w.el, "list", func() listState { return listState{} })
	w.st = st
	w.always = st.always
	if st.changed {
		*selected = st.picked
	}
	w.changed, w.activated = st.changed, st.activated
	st.changed, st.activated = false, false
	n := len(items)
	if *selected >= n {
		*selected = n - 1
	}
	rows := max(int((st.rect.H-4)/ListRowHeight), 1)
	if st.rect.H == 0 {
		rows = 6
	}
	scrolls := n > rows
	focused := w.el.Focused() || w.el.FocusWithin()
	focusVisible := w.el.FocusVisible()

	dimmed := func(i int) bool { return i >= 0 && i < len(w.dim) && w.dim[i] }
	choose := func(i int) {
		if n == 0 {
			return
		}
		i = min(max(i, 0), n-1)
		if dimmed(i) {
			// A dimmed row takes no choosing: the nearest one that does,
			// the way the choice was going.
			dir := 1
			if i < *selected {
				dir = -1
			}
			j := i
			for j >= 0 && j < n && dimmed(j) {
				j += dir
			}
			if j < 0 || j >= n {
				return
			}
			i = j
		}
		if i != *selected {
			*selected = i
			st.picked = i
			st.changed = true
			w.changed = true
		}
		// Keep it in view.
		if float64(i) < st.top {
			st.top = float64(i)
		} else if float64(i) >= st.top+float64(rows) {
			st.top = float64(i - rows + 1)
		}
	}
	if !w.disabled {
		keys := map[ui.Key]func(){
			ui.KeyUp:       func() { choose(max(*selected-1, 0)) },
			ui.KeyDown:     func() { choose(*selected + 1) },
			ui.KeyHome:     func() { choose(0) },
			ui.KeyEnd:      func() { choose(n - 1) },
			ui.KeyPageUp:   func() { choose(*selected - rows + 1) },
			ui.KeyPageDown: func() { choose(*selected + rows - 1) },
		}
		for _, k := range []ui.Key{ui.KeyUp, ui.KeyDown, ui.KeyHome, ui.KeyEnd, ui.KeyPageUp, ui.KeyPageDown} {
			if w.el.Shortcut(0, k) {
				keys[k]()
			}
		}
		if w.el.Shortcut(0, ui.KeyEnter) && *selected >= 0 {
			w.activated = true
		}
	}
	rowAt := func(y float32) int {
		return int(st.top) + int((y-2)/ListRowHeight)
	}
	listW := func() float32 {
		if scrolls || w.always {
			return st.rect.W - 1 - ScrollBarWidth + 1
		}
		return st.rect.W
	}
	w.el.HandleInput(func(ev ui.InputEvent) bool {
		if w.disabled {
			return false
		}
		switch ev.Kind {
		case ui.InputPointerDown:
			if ev.Button != 0 || ev.X >= listW() {
				return false
			}
			if i := rowAt(ev.Y); i < n && !dimmed(i) {
				choose(i)
			}
			st.dragging = true
			return false // and take the focus
		case ui.InputPointerMove:
			if st.dragging && ev.Button == 0 {
				choose(rowAt(min(max(ev.Y, 2), st.rect.H-3)))
				return true
			}
		case ui.InputPointerUp:
			st.dragging = false
		case ui.InputText:
			now := wallNow()
			if now.Sub(st.typedAt) > typeSelectGap {
				st.typed = ""
			}
			st.typed += strings.ToLower(ev.Text)
			st.typedAt = now
			for i, it := range items {
				if !dimmed(i) && strings.HasPrefix(strings.ToLower(it), st.typed) {
					choose(i)
					break
				}
			}
			return true
		case ui.InputScroll:
			st.top = min(max(st.top+float64(ev.DY)/ListRowHeight, 0), float64(max(n-rows, 0)))
			return true
		}
		return false
	})
	w.el.TextCaret(ui.Rect{})
	if w.el.DoubleClicked() && *selected >= 0 {
		w.activated = true
	}

	sel := *selected
	top := int(st.top)
	// Drawn over the rows' elements, which are there for assistive
	// technology alone.
	w.el.DrawOver(func(p *ui.Painter, r ui.Rect) {
		st.rect = r
		pn := NewPen(p, r).Faded(w.disabled)
		DrawContentRing(pn)
		in := pn.Sub(1, 1, pn.W-2, pn.H-2)
		lw := in.W
		if scrolls || w.always {
			lw = in.W - ScrollBarWidth + 1
		}
		if w.icons != nil || w.dim != nil {
			drawListRows(in.Sub(0, 0, lw, in.H), items, w.icons, w.dim, top, sel, focused, w.disabled, th.Accent)
		} else {
			DrawListBox(in.Sub(0, 0, lw, in.H), items, top, sel, focused, w.disabled, th.Accent)
		}
		for i := top; i < len(w.arrows); i++ {
			y := 1 + (i-top)*ListRowHeight
			if !w.arrows[i] || y+ListRowHeight > in.H-1 {
				continue
			}
			ink := Black
			if i == sel && focused {
				ink = White
			}
			in.Bitmap(lw-11, y+(ListRowHeight-7)/2, rowArrow, ink)
		}
		if (focusVisible || focused) && !w.noRing {
			DrawFocusRing(pn, 0, th.Accent.Dark)
		}
	})
	w.el.Children(func() {
		// Rows for assistive technology and tests, over where they show.
		ui.Column(c).Grow(1).PassThrough().Children(func() {
			for i := top; i < min(top+rows, n); i++ {
				row := ui.Box(c).Height(ListRowHeight).Role(ui.RoleListItem).Label(items[i]).PassThrough()
				if i == sel {
					row.Selected(true)
				}
			}
		})
		if scrolls || w.always {
			ui.Column(c).Width(ScrollBarWidth).Children(func() {
				sb := ScrollBar(c, &st.top, float64(rows), float64(max(n, rows))).Step(1)
				if !scrolls {
					sb.Inactive() // the rows fit: a flat track
				}
				sb.Element().Grow(1)
			})
		}
	})
	return w
}

// Arrows draws an arrow at the right of the rows marked, as a column view
// marks the items that hold others.
func (w *ListBoxWidget) Arrows(rows []bool) *ListBoxWidget {
	w.arrows = rows
	return w
}

var rowArrow = []string{"#...", "##..", "###.", "####", "###.", "##..", "#..."}

// Icons draws a small picture, at most 16×16 as rows for DrawIcon, at the
// left of each row, its text after it.
func (w *ListBoxWidget) Icons(icons [][]string) *ListBoxWidget {
	w.icons = icons
	return w
}

// Dim draws the rows marked in Gray8 and keeps them from being chosen, as
// a save dialog shows the files it is not saving over.
func (w *ListBoxWidget) Dim(rows []bool) *ListBoxWidget {
	w.dim = rows
	return w
}

// FocusRing(false) leaves the focus ring to what holds the list, as a
// column view draws one around all its columns.
func (w *ListBoxWidget) FocusRing(on bool) *ListBoxWidget {
	w.noRing = !on
	return w
}

// ScrollBarAlways keeps the scroll bar at the list's right when its rows
// all fit too, as a flat track.
func (w *ListBoxWidget) ScrollBarAlways() *ListBoxWidget {
	if !w.st.always {
		w.st.always = true
		w.c.After(0) // built from the next frame
	}
	w.always = true
	return w
}

// drawListRows draws rows as DrawListBox does, with an icon at the left of
// each that has one and the dimmed rows in Gray8.
func drawListRows(pn Pen, items []string, icons [][]string, dim []bool, top, sel int, focused, disabled bool, a Accent) {
	w, h := pn.W, pn.H
	pn.Rect(1, 1, w-2, h-2, White)
	pn.Frame(0, 0, w, h, Black)
	pn.Clip(1, 1, w-2, h-2, func() {
		for i := top; i < len(items); i++ {
			y := 1 + (i-top)*ListRowHeight
			if y >= h-1 {
				break
			}
			c := Black
			if disabled || i < len(dim) && dim[i] {
				c = Gray8
			}
			hl := false
			if i == sel {
				bg := Gray4
				if focused && !disabled {
					bg, c, hl = a.Dark, White, true
				}
				pn.Rect(1, y, w-2, ListRowHeight, bg)
			}
			x := listTextLeft
			if i < len(icons) && icons[i] != nil {
				ic := icons[i]
				ip := pn.Sub(x-1, y+(ListRowHeight-len(ic))/2, iconWidth(ic), len(ic)).Dimmed(i < len(dim) && dim[i])
				DrawIcon(ip, ic, a, hl)
			}
			if icons != nil {
				x += 19
			}
			pn.TextMiddle(x, y, ListRowHeight, AppFont.Truncate(items[i], w-x-listTextLeft), AppFont, c)
		}
	})
}

// Disabled dims the list and makes it take no input while d is true.
func (w *ListBoxWidget) Disabled(d bool) *ListBoxWidget {
	w.disabled = d
	w.el.Disabled(d)
	return w
}

// Changed reports whether the user chose another row since the last frame.
func (w *ListBoxWidget) Changed() bool { return w.changed }

// Activated reports whether the user double-clicked a row, or pressed
// Return, since the last frame.
func (w *ListBoxWidget) Activated() bool { return w.activated }

// Element returns the list box's element, to size it.
func (w *ListBoxWidget) Element() ui.Element { return w.el }

// Label names the list box.
func (w *ListBoxWidget) Label(s string) *ListBoxWidget { w.el.Label(s); return w }

// DrawListBox draws the rows of a list box from row top in pn's box, row
// sel highlighted: in the accent while the list has the keyboard focus,
// in gray otherwise.
func DrawListBox(pn Pen, items []string, top, sel int, focused, disabled bool, a Accent) {
	w, h := pn.W, pn.H
	pn.Rect(1, 1, w-2, h-2, White)
	pn.Frame(0, 0, w, h, Black)
	text := Black
	if disabled {
		text = Gray8
	}
	pn.Clip(1, 1, w-2, h-2, func() {
		for i := top; i < len(items); i++ {
			y := 1 + (i-top)*ListRowHeight
			if y >= h-1 {
				break
			}
			c := text
			if i == sel {
				bg := Gray4
				if focused && !disabled {
					bg, c = a.Dark, White
				}
				pn.Rect(1, y, w-2, ListRowHeight, bg)
			}
			pn.TextMiddle(listTextLeft, y, ListRowHeight, AppFont.Truncate(items[i], w-2*listTextLeft), AppFont, c)
		}
	})
}
