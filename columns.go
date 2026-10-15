package platinum

import (
	"github.com/egoist/mygo/ui"
)

// Column view metrics, from docs/spec.md.
const ColumnViewWidth = 140 // the width of each column of a column view

type columnViewState struct {
	height  float32 // the columns' height, as last painted
	width   float32 // and their width
	focus   int     // the column to give the keyboard focus to, or -1
	changed bool
	left    float64 // how far the columns are scrolled left
	toEnd   bool    // scroll to the last column once the width is known
	room    int     // the room left at the scroll bar's right
	columns int     // how many columns there were
}

// ColumnViewWidget is a column view being built.
type ColumnViewWidget struct {
	el        ui.Element
	st        *columnViewState
	c         *ui.Context
	changed   bool
	submitted bool
}

// ColumnView builds a browser of a tree in columns, as a list of roots and,
// right of it, a list of the children of the item chosen in each column:
// *path holds the item chosen in each column, from the first. A click
// chooses an item; with the keyboard focus, Up and Down move in a column,
// Right goes into the item chosen, Left back to its column. A horizontal
// scroll bar under the columns scrolls them, the newest in view when the
// path changes; its arrows dim while they all fit. Give it a size.
func ColumnView(c *ui.Context, roots []*TreeNode, path *[]*TreeNode) *ColumnViewWidget {
	th := ThemeOf(c)
	w := &ColumnViewWidget{c: c}
	w.el = ui.Column(c).Role(ui.RoleGroup).Size(3*ColumnViewWidth, 140+ScrollBarWidth).AlignItems(ui.Stretch)
	st := ui.Local(w.el, "columns", func() columnViewState { return columnViewState{focus: -1, toEnd: true} })
	w.st = st
	w.changed, st.changed = st.changed, false
	// The columns: the roots, then the children of each item chosen that
	// has any.
	lists := [][]*TreeNode{roots}
	for i := 0; i < len(*path); i++ {
		n := (*path)[i]
		found := false
		for _, it := range lists[i] {
			found = found || it == n
		}
		if !found {
			*path = (*path)[:i]
			break
		}
		if len(n.Children) > 0 {
			lists = append(lists, n.Children)
		}
	}
	focus := st.focus
	st.focus = -1
	h := st.height
	if h == 0 {
		h = 140
		c.After(0) // lay out again once the size is known
	}
	total := float64(len(lists)*(ColumnViewWidth-1) + 1)
	if len(lists) != st.columns {
		st.columns = len(lists)
		st.toEnd = true
	}
	if st.toEnd && st.width > 0 {
		st.left = max(total-float64(st.width), 0)
		st.toEnd = false
	}
	st.left = min(max(st.left, 0), max(total-float64(st.width), 0))
	focused := w.el.FocusWithin()
	w.el.Children(func() {
		cols := ui.Row(c).Grow(1).AlignItems(ui.Stretch).Clip()
		cols.Draw(func(p *ui.Painter, r ui.Rect) {
			resized(c, ui.Rect{W: st.width, H: st.height}, r)
			st.height, st.width = r.H, r.W
			pn := NewPen(p, r)
			pn.Rect(0, 0, pn.W, pn.H, White)
		})
		cols.DrawOver(func(p *ui.Painter, r ui.Rect) {
			if focused {
				// One ring around all the columns, inside them.
				pn := NewPen(p, r)
				pn.Frame(0, 0, pn.W, pn.H, th.Accent.Dark)
				pn.Frame(1, 1, pn.W-2, pn.H-2, th.Accent.Dark)
			}
		})
		cols.Children(func() {
			for col, items := range lists {
				sel := -1
				if col < len(*path) {
					for i, it := range items {
						if it == (*path)[col] {
							sel = i
						}
					}
				}
				names := make([]string, len(items))
				for i, it := range items {
					names[i] = it.Title
				}
				lb := ListBox(c.Key(col), &sel, names).FocusRing(false)
				lb.Element().Size(ColumnViewWidth, h).Margin(0, -1, 0, 0)
				if col == 0 {
					lb.Element().Margin(0, -1, 0, -float32(st.left))
				}
				if col == focus {
					lb.Element().Focus()
				}
				if lb.Submitted() && sel >= 0 {
					*path = append((*path)[:col], items[sel])
					w.submitted = true
				}
				if lb.Changed() && sel >= 0 {
					*path = append((*path)[:col], items[sel])
					w.changed, st.changed = true, true
				}
				if lb.Element().Shortcut(0, ui.KeyRight) && sel >= 0 && len(items[sel].Children) > 0 {
					child := items[sel].Children[0]
					*path = append((*path)[:col+1], child)
					st.focus = col + 1
					w.changed, st.changed = true, true
					c.After(0)
				}
				if lb.Element().Shortcut(0, ui.KeyLeft) && col > 0 {
					*path = (*path)[:col]
					st.focus = col - 1
					w.changed, st.changed = true, true
					c.After(0)
				}
				// An arrow on the rows that hold others.
				arrows := make([]bool, len(items))
				for i, it := range items {
					arrows[i] = len(it.Children) > 0
				}
				lb.Arrows(arrows)
			}
		})
		ui.Row(c).Height(ScrollBarWidth).Margin(-1, 0, 0, 0).AlignItems(ui.Stretch).Children(func() {
			sb := ScrollBar(c, &st.left, float64(st.width), total).Horizontal().Step(ColumnViewWidth / 2)
			sb.Element().Grow(1)
			if st.room > 0 {
				ui.Box(c).Width(float32(st.room))
			}
		})
	})
	report(w.el, w.changed || w.submitted)
	return w
}

// ScrollBarRoom leaves px pixels at the right of the horizontal scroll bar,
// as for the resize corner of the window it fills.
func (w *ColumnViewWidget) ScrollBarRoom(px int) *ColumnViewWidget {
	if w.st.room != px {
		w.st.room = px
		w.c.After(0) // built from the next frame
	}
	return w
}

// Changed reports whether the user chose another item since the last
// frame.
func (w *ColumnViewWidget) Changed() bool { return report(w.el, w.changed) }

// Element returns the column view's element, to size it.
func (w *ColumnViewWidget) Element() ui.Element { return w.el }

// Label names the column view.
func (w *ColumnViewWidget) Label(s string) *ColumnViewWidget { w.el.Label(s); return w }

// Submitted reports whether the user double-clicked an item, or chose one
// and pressed Return, since the last frame: the last of the path.
func (w *ColumnViewWidget) Submitted() bool { return report(w.el, w.submitted) }
