package platinum

import (
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

// TreeNode is an item of a tree view, which may hold others.
type TreeNode struct {
	Title string
	// Icon is a small picture for DrawIcon, at most 12 rows tall.
	Icon []string
	// Children are the items inside it; it shows a disclosure triangle when
	// it has any.
	Children []*TreeNode
	// Open shows its children under it.
	Open bool
	// Data is anything the app keeps with the item.
	Data any
}

// Tree view metrics, from docs/spec.md.
const (
	TreeIndent  = 16 // how far each level of a tree view is indented
	treeTriLeft = 3
)

type treeRow struct {
	n     *TreeNode
	level int
}

func flatten(nodes []*TreeNode, level int, out []treeRow) []treeRow {
	for _, n := range nodes {
		out = append(out, treeRow{n, level})
		if n.Open && len(n.Children) > 0 {
			out = flatten(n.Children, level+1, out)
		}
	}
	return out
}

// TreeWidget is a tree view being built.
type TreeWidget struct {
	el        ui.Element
	changed   bool
	activated bool
}

type treeState struct {
	top                float64
	rect               ui.Rect
	changed, activated bool
	picked             *TreeNode // chosen between builds, for the next one
	typed              string
	typedAt            time.Time
}

// TreeView builds a list of roots and the items inside them, as an outline:
// a disclosure triangle opens and closes an item, its children indented
// under it. It chooses *selected. A click chooses a row, a click on its
// triangle opens or closes it, and a double click activates it; with the
// keyboard focus, Up and Down move the choice, Right opens the item chosen
// or goes into it, Left closes it or goes to its parent, typing the start
// of a title chooses it, and Return activates. Give it a size.
func TreeView(c *ui.Context, roots []*TreeNode, selected **TreeNode) *TreeWidget {
	th := ThemeOf(c)
	w := &TreeWidget{}
	w.el = ui.Row(c).Focusable().FocusRing(false).Role(ui.RoleTree).Size(200, 160).AlignItems(ui.Stretch).Padding(1)
	st := ui.Local(w.el, "tree", func() treeState { return treeState{} })
	if st.changed {
		*selected = st.picked
	}
	w.changed, w.activated = st.changed, st.activated
	st.changed, st.activated = false, false
	rows := flatten(roots, 0, nil)
	n := len(rows)
	visible := max(int((st.rect.H-4)/ListRowHeight), 1)
	if st.rect.H == 0 {
		visible = 8
	}
	index := func() int {
		for i, r := range rows {
			if r.n == *selected {
				return i
			}
		}
		return -1
	}
	reveal := func(i int) {
		if float64(i) < st.top {
			st.top = float64(i)
		} else if float64(i) >= st.top+float64(visible) {
			st.top = float64(i - visible + 1)
		}
	}
	choose := func(i int) {
		if n == 0 {
			return
		}
		i = min(max(i, 0), n-1)
		if rows[i].n != *selected {
			*selected = rows[i].n
			st.picked = rows[i].n
			st.changed, w.changed = true, true
		}
		reveal(i)
	}
	parentOf := func(i int) int {
		for j := i - 1; j >= 0; j-- {
			if rows[j].level < rows[i].level {
				return j
			}
		}
		return -1
	}
	if i := index(); n > 0 {
		if w.el.Shortcut(0, ui.KeyDown) {
			choose(i + 1)
		}
		if w.el.Shortcut(0, ui.KeyUp) {
			choose(max(i-1, 0))
		}
		if w.el.Shortcut(0, ui.KeyHome) {
			choose(0)
		}
		if w.el.Shortcut(0, ui.KeyEnd) {
			choose(n - 1)
		}
		if w.el.Shortcut(0, ui.KeyRight) && i >= 0 {
			r := rows[i].n
			if len(r.Children) > 0 && !r.Open {
				r.Open = true
			} else if len(r.Children) > 0 {
				choose(i + 1)
			}
		}
		if w.el.Shortcut(0, ui.KeyLeft) && i >= 0 {
			r := rows[i].n
			if r.Open && len(r.Children) > 0 {
				r.Open = false
			} else if p := parentOf(i); p >= 0 {
				choose(p)
			}
		}
		if w.el.Shortcut(0, ui.KeyEnter) && i >= 0 {
			w.activated = true
		}
	}
	listW := func() float32 {
		if n > visible {
			return st.rect.W - ScrollBarWidth
		}
		return st.rect.W
	}
	w.el.HandleInput(func(ev ui.InputEvent) bool {
		switch ev.Kind {
		case ui.InputPointerDown:
			if ev.Button != 0 || ev.X >= listW() {
				return false
			}
			i := int(st.top) + int((ev.Y-2)/ListRowHeight)
			if i < 0 || i >= n {
				return false
			}
			r := rows[i]
			tx := float32(2 + treeTriLeft + r.level*TreeIndent)
			if len(r.n.Children) > 0 && ev.X >= tx-2 && ev.X < tx+DisclosureSize+2 {
				r.n.Open = !r.n.Open
				return true
			}
			if r.n != *selected {
				*selected = r.n
				st.picked = r.n
				st.changed = true
			}
			return false // and take the focus, and count clicks
		case ui.InputText:
			now := wallNow()
			if now.Sub(st.typedAt) > typeSelectGap {
				st.typed = ""
			}
			st.typed += strings.ToLower(ev.Text)
			st.typedAt = now
			for i, r := range rows {
				if strings.HasPrefix(strings.ToLower(r.n.Title), st.typed) {
					*selected = r.n
					st.picked = r.n
					st.changed = true
					reveal(i)
					break
				}
			}
			return true
		case ui.InputScroll:
			st.top = min(max(st.top+float64(ev.DY)/ListRowHeight, 0), float64(max(n-visible, 0)))
			return true
		}
		return false
	})
	w.el.TextCaret(ui.Rect{})
	if w.el.DoubleClicked() && *selected != nil {
		w.activated = true
	}
	focused := w.el.Focused()
	sel := *selected
	top := int(st.top)
	w.el.DrawOver(func(p *ui.Painter, r ui.Rect) {
		resized(c, st.rect, r)
		st.rect = r
		pn := NewPen(p, r)
		DrawContentRing(pn)
		in := pn.Sub(1, 1, pn.W-2, pn.H-2)
		lw := in.W
		if n > visible {
			lw = in.W - ScrollBarWidth + 1
		}
		drawTreeRows(in.Sub(0, 0, lw, in.H), rows, top, sel, focused, th.Accent)
		if focused {
			DrawFocusRing(pn, 0, th.Accent.Dark)
		}
	})
	w.el.Children(func() {
		ui.Column(c).Grow(1).PassThrough().Children(func() {
			for i := top; i < min(top+visible, n); i++ {
				r := rows[i]
				e := ui.Box(c).Height(ListRowHeight).Role(ui.RoleTreeItem).Label(r.n.Title).Level(r.level + 1).PassThrough()
				if len(r.n.Children) > 0 {
					e.Expanded(r.n.Open)
				}
			}
		})
		if n > visible {
			ui.Column(c).Width(ScrollBarWidth).Children(func() {
				ScrollBar(c, &st.top, float64(visible), float64(n)).Step(1).Element().Grow(1)
			})
		}
	})
	report(w.el, w.changed || w.activated)
	return w
}

// Changed reports whether the user chose another item since the last
// frame.
func (w *TreeWidget) Changed() bool { return report(w.el, w.changed) }

// Activated reports whether the user double-clicked an item, or pressed
// Return, since the last frame.
func (w *TreeWidget) Activated() bool { return report(w.el, w.activated) }

// Element returns the tree view's element, to size it.
func (w *TreeWidget) Element() ui.Element { return w.el }

// Label names the tree view.
func (w *TreeWidget) Label(s string) *TreeWidget { w.el.Label(s); return w }

// DrawTree draws the rows of a tree view of roots — those open showing their
// children — from row top in pn's box, sel highlighted (in the accent while
// focused).
func DrawTree(pn Pen, roots []*TreeNode, top int, sel *TreeNode, focused bool, a Accent) {
	drawTreeRows(pn, flatten(roots, 0, nil), top, sel, focused, a)
}

func drawTreeRows(pn Pen, rows []treeRow, top int, sel *TreeNode, focused bool, a Accent) {
	w, h := pn.W, pn.H
	pn.Rect(1, 1, w-2, h-2, White)
	pn.Frame(0, 0, w, h, Black)
	pn.Clip(1, 1, w-2, h-2, func() {
		for i := top; i < len(rows); i++ {
			y := 1 + (i-top)*ListRowHeight
			if y >= h-1 {
				break
			}
			r := rows[i]
			x := 1 + treeTriLeft + r.level*TreeIndent
			ink := Black
			if r.n == sel {
				bg := Gray4
				if focused {
					bg, ink = a.Dark, White
				}
				pn.Rect(x+DisclosureSize+2, y, w-x-DisclosureSize-3, ListRowHeight, bg)
			}
			if len(r.n.Children) > 0 {
				DrawDisclosure(pn.Sub(x, y+(ListRowHeight-DisclosureSize)/2, DisclosureSize, DisclosureSize), r.n.Open, StateNormal)
			}
			tx := x + DisclosureSize + 5
			if r.n.Icon != nil {
				ih := len(r.n.Icon)
				DrawIcon(pn.Sub(tx, y+(ListRowHeight-ih)/2, iconWidth(r.n.Icon), ih), r.n.Icon, a, r.n == sel && focused)
				tx += iconWidth(r.n.Icon) + 4
			}
			pn.TextMiddle(tx, y, ListRowHeight, AppFont.Truncate(r.n.Title, w-tx-4), AppFont, ink)
		}
	})
}
