package platinum

import (
	"sort"
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"
)

// Table metrics, from docs/spec.md.
const (
	TableHeaderHeight = 18 // the height of a table's column headers
	TableMinColumn    = 30 // the narrowest a column gets
	tableCellPad      = 5
	resizeGrip        = 3 // how near a column's edge the pointer resizes it
)

// Column is a column of a table.
type Column struct {
	Title string
	// Width is the column's width when the table first shows; the user
	// resizes it from there.
	Width int
	// Right aligns the column's cells at its right, as numbers.
	Right bool
	// Fixed keeps the user from sorting by the column.
	Fixed bool
	// Less, if set, orders rows a and b of the model by the column, for
	// cells whose text does not sort as their values do (as "2 K" and
	// "300 bytes"), or text to order as a language does (a collator of
	// golang.org/x/text/collate); otherwise cells sort as numbers where
	// both are, and as text otherwise, by code point.
	Less func(a, b int) bool
	// Icon, if set, gives a small picture (at most 16×16, as rows for
	// DrawIcon) for row r of the model, drawn at the left of its cell.
	Icon func(r int) []string
}

// TableState is what a table keeps between frames, held by the app: the
// widths of its columns, how its rows are sorted, which are selected, and
// how far it is scrolled.
type TableState struct {
	// Widths are the columns' widths; the table fills it from Column.Width.
	Widths []int
	// SortColumn sorts the rows by that column's cells (-1, or a Fixed
	// column, for the model's order); Descending turns the order over.
	SortColumn int
	Descending bool
	// Selected holds the rows selected, by their index in the model.
	Selected map[int]bool
	// ScrollBarAlways keeps the scroll bar when the rows fit too, as a
	// flat track.
	ScrollBarAlways bool
	// NoFocusRing draws no ring around the table while it has the keyboard
	// focus, as a list view filling a window draws none.
	NoFocusRing bool

	top      float64
	anchor   int // the row a Shift-click or Shift-arrow extends from
	cursor   int // the row the arrows move from, in sorted order
	rect     ui.Rect
	resizing int // the column being resized, or -1
	grabX    float32
	grabW    int
	changed  bool
	sorted   bool
	order    []int
}

// NewTableState returns the state of a table sorted by none of its columns.
func NewTableState() *TableState {
	return &TableState{SortColumn: -1, Selected: map[int]bool{}, resizing: -1}
}

// TableWidget is a table being built.
type TableWidget struct {
	el        ui.Element
	st        *TableState
	changed   bool
	sorted    bool
	activated bool
}

// lessCells compares two cells as numbers where both are, and as text
// otherwise.
func lessCells(a, b string) bool {
	fa, ea := strconv.ParseFloat(strings.TrimSpace(a), 64)
	fb, eb := strconv.ParseFloat(strings.TrimSpace(b), 64)
	if ea == nil && eb == nil {
		return fa < fb
	}
	return strings.ToLower(a) < strings.ToLower(b)
}

// Table builds a table of rows rows in columns, whose cells cell returns:
// column headers along its top, which sort the rows by their column as
// clicked (again to turn the order over) and resize it as their right
// edges are dragged, over rows that the pointer and the keyboard select —
// a click selects one, Shift extends the selection, Command adds or takes
// away a row. With the keyboard focus, Up and Down move the selection
// (Shift extends it), Command-A selects every row, Home and End go to the
// ends. Give it a size.
func Table(c *ui.Context, st *TableState, cols []Column, rows int, cell func(row, col int) string) *TableWidget {
	th := ThemeOf(c)
	if st.Selected == nil {
		st.Selected = map[int]bool{}
	}
	for len(st.Widths) < len(cols) {
		w := cols[len(st.Widths)].Width
		if w <= 0 {
			w = 100
		}
		st.Widths = append(st.Widths, w)
	}
	w := &TableWidget{st: st}
	w.el = ui.Row(c).Focusable().FocusRing(false).Role(ui.RoleTable).Size(300, 160).AlignItems(ui.Stretch).Padding(1)
	w.changed, st.changed = st.changed, false
	w.sorted, st.sorted = st.sorted, false

	// The order the rows show in.
	order := make([]int, rows)
	for i := range order {
		order[i] = i
	}
	if sc := st.SortColumn; sc >= 0 && sc < len(cols) && !cols[sc].Fixed {
		less := cols[sc].Less
		if less == nil {
			less = func(a, b int) bool { return lessCells(cell(a, sc), cell(b, sc)) }
		}
		sort.SliceStable(order, func(i, j int) bool {
			if st.Descending {
				return less(order[j], order[i])
			}
			return less(order[i], order[j])
		})
	}
	st.order = order
	// The model may have shrunk since the last frame: keep what points into
	// it inside it.
	st.anchor = min(max(st.anchor, 0), max(rows-1, 0))
	st.cursor = min(max(st.cursor, 0), max(rows-1, 0))
	for r := range st.Selected {
		if r < 0 || r >= rows {
			delete(st.Selected, r)
		}
	}
	st.top = min(st.top, float64(max(rows-1, 0)))
	visible := max(int((st.rect.H-TableHeaderHeight-3)/ListRowHeight), 1)
	if st.rect.H == 0 {
		visible = 8
	}
	fits := rows <= visible
	scrolls := !fits || st.ScrollBarAlways
	reveal := func(i int) {
		if float64(i) < st.top {
			st.top = float64(i)
		} else if float64(i) >= st.top+float64(visible) {
			st.top = float64(i - visible + 1)
		}
	}
	selectRange := func(a, b int) {
		clear(st.Selected)
		if rows == 0 {
			return
		}
		a, b = min(max(a, 0), rows-1), min(max(b, 0), rows-1)
		for i := min(a, b); i <= max(a, b); i++ {
			st.Selected[order[i]] = true
		}
	}
	move := func(to int, extend bool) {
		if rows == 0 {
			return
		}
		to = min(max(to, 0), rows-1)
		if extend {
			selectRange(st.anchor, to)
		} else {
			clear(st.Selected)
			st.Selected[order[to]] = true
			st.anchor = to
		}
		st.cursor = to
		st.changed, w.changed = true, true
		reveal(to)
	}
	for _, k := range []struct {
		key ui.Key
		to  func() int
	}{
		{ui.KeyDown, func() int { return st.cursor + 1 }},
		{ui.KeyUp, func() int { return st.cursor - 1 }},
		{ui.KeyHome, func() int { return 0 }},
		{ui.KeyEnd, func() int { return rows - 1 }},
		{ui.KeyPageDown, func() int { return st.cursor + visible - 1 }},
		{ui.KeyPageUp, func() int { return st.cursor - visible + 1 }},
	} {
		if w.el.Shortcut(0, k.key) {
			move(k.to(), false)
		}
		if w.el.Shortcut(ui.Shift, k.key) {
			move(k.to(), true)
		}
	}
	if (w.el.Shortcut(0, ui.KeyEnter) || w.el.DoubleClicked()) && len(st.Selected) > 0 {
		w.activated = true
	}
	if w.el.Shortcut(ui.Cmd, ui.KeyA) && rows > 0 {
		selectRange(0, rows-1)
		st.changed, w.changed = true, true
	}

	listW := func() float32 {
		if scrolls {
			return st.rect.W - ScrollBarWidth
		}
		return st.rect.W
	}
	// edgeAt returns the column whose right edge is at x, or -1.
	edgeAt := func(x float32) int {
		cx := 2
		for i, cw := range st.Widths {
			cx += cw
			if x >= float32(cx-resizeGrip) && x <= float32(cx+resizeGrip) {
				return i
			}
		}
		return -1
	}
	columnAt := func(x float32) int {
		cx := 2
		for i, cw := range st.Widths {
			if x >= float32(cx) && x < float32(cx+cw) {
				return i
			}
			cx += cw
		}
		return -1
	}
	w.el.HandleInput(func(ev ui.InputEvent) bool {
		switch ev.Kind {
		case ui.InputPointerDown:
			if ev.Button != 0 || ev.X >= listW() {
				return false
			}
			if ev.Y < TableHeaderHeight+1 {
				if e := edgeAt(ev.X); e >= 0 {
					st.resizing, st.grabX, st.grabW = e, ev.X, st.Widths[e]
					return true
				}
				if col := columnAt(ev.X); col >= 0 && !cols[col].Fixed {
					if st.SortColumn == col {
						st.Descending = !st.Descending
					} else {
						st.SortColumn, st.Descending = col, false
					}
					st.sorted = true
				}
				return true
			}
			i := int(st.top) + int((ev.Y-TableHeaderHeight-1)/ListRowHeight)
			if i < 0 || i >= rows {
				return false
			}
			switch {
			case ev.Mods&ui.Shift != 0:
				selectRange(st.anchor, i)
			case ev.Mods&ui.Cmd != 0:
				r := st.order[i]
				if st.Selected[r] {
					delete(st.Selected, r)
				} else {
					st.Selected[r] = true
				}
				st.anchor = i
			default:
				clear(st.Selected)
				st.Selected[st.order[i]] = true
				st.anchor = i
			}
			st.cursor = i
			st.changed = true
			return false // and take the focus
		case ui.InputPointerMove:
			if st.resizing >= 0 && ev.Button == 0 {
				st.Widths[st.resizing] = max(st.grabW+round(ev.X-st.grabX), TableMinColumn)
				return true
			}
		case ui.InputPointerUp:
			if st.resizing >= 0 {
				st.resizing = -1
				return true
			}
		case ui.InputScroll:
			st.top = min(max(st.top+float64(ev.DY)/ListRowHeight, 0), float64(max(rows-visible, 0)))
			return true
		}
		return false
	})
	focused := w.el.Focused()
	top := int(st.top)
	bg := inactive(c)
	w.el.DrawOver(func(p *ui.Painter, r ui.Rect) {
		resized(c, st.rect, r)
		st.rect = r
		pn := NewPen(p, r)
		DrawContentRing(pn)
		in := pn.Sub(1, 1, pn.W-2, pn.H-2)
		lw := in.W
		if scrolls {
			lw = in.W - ScrollBarWidth + 1
		}
		DrawTable(in.Sub(0, 0, lw, in.H), st, cols, order, top, cell, focused, th.Accent)
		if bg {
			// In the background, its headers gray, the cells as they are.
			tp := in.Sub(0, 0, lw, in.H).Dimmed(true)
			tp.Clip(1, 1, tp.W-2, TableHeaderHeight-1, func() {
				DrawTable(tp, st, cols, order, top, cell, focused, th.Accent)
			})
		}
		if scrolls {
			// The corner over the scroll bar, a header of its own.
			corner := in.Sub(lw-1, 0, in.W-lw+1, TableHeaderHeight)
			corner.Rect(0, 0, corner.W, corner.H, Gray2)
			corner.Frame(0, 0, corner.W, corner.H, Black)
			corner.Bevel(1, 1, corner.W-2, corner.H-2, White, Gray7)
		}
		if focused && !st.NoFocusRing {
			DrawFocusRing(pn, 0, th.Accent.Dark)
		}
	})
	w.el.Children(func() {
		// The cells for assistive technology, which may be wider than the
		// table: they clip rather than push the scroll bar out.
		ui.Column(c).Grow(1).MinWidth(0).Clip().PassThrough().Children(func() {
			ui.Row(c).Height(TableHeaderHeight).PassThrough().Children(func() {
				for i, col := range cols {
					h := ui.Box(c).Size(float32(st.Widths[i]), TableHeaderHeight).Role(ui.RoleColumnHeader).Label(col.Title).PassThrough()
					if i == st.SortColumn {
						if st.Descending {
							h.Value("descending")
						} else {
							h.Value("ascending")
						}
					}
				}
			})
			for i := top; i < min(top+visible, rows); i++ {
				row := ui.Row(c).Height(ListRowHeight).Role(ui.RoleRow).PassThrough()
				if st.Selected[order[i]] {
					row.Selected(true)
				}
				row.Children(func() {
					for j := range cols {
						ui.Box(c).Size(float32(st.Widths[j]), ListRowHeight).Role(ui.RoleCell).Label(cell(order[i], j)).PassThrough()
					}
				})
			}
		})
		if scrolls {
			ui.Column(c).Width(ScrollBarWidth).Children(func() {
				ui.Box(c).Height(TableHeaderHeight - 1).PassThrough()
				sb := ScrollBar(c, &st.top, float64(visible), float64(max(rows, visible))).Step(1)
				if fits {
					sb.Inactive() // the rows fit: a flat track
				}
				sb.Element().Grow(1)
			})
		}
	})
	report(w.el, w.changed || w.activated || w.sorted)
	return w
}

// Changed reports whether the user changed the selection since the last
// frame.
func (w *TableWidget) Changed() bool { return report(w.el, w.changed) }

// Activated reports whether the user double-clicked a row, or pressed
// Return, with rows selected, since the last frame.
func (w *TableWidget) Activated() bool { return report(w.el, w.activated) }

// Cursor returns the model index of the row the arrows last moved to or a
// click chose, or -1 with no rows.
func (w *TableWidget) Cursor() int {
	if len(w.st.order) == 0 {
		return -1
	}
	return w.st.order[min(max(w.st.cursor, 0), len(w.st.order)-1)]
}

// Sorted reports whether the user sorted the rows anew since the last
// frame.
func (w *TableWidget) Sorted() bool { return report(w.el, w.sorted) }

// Order returns the rows of the model in the order they show.
func (w *TableWidget) Order() []int { return w.st.order }

// Element returns the table's element, to size it.
func (w *TableWidget) Element() ui.Element { return w.el }

// Label names the table.
func (w *TableWidget) Label(s string) *TableWidget { w.el.Label(s); return w }

var (
	sortUp   = []string{"..#..", ".###.", "#####"}
	sortDown = []string{"#####", ".###.", "..#.."}
)

// DrawTable draws a table in pn's box: its headers and the rows of order
// from row top.
func DrawTable(pn Pen, st *TableState, cols []Column, order []int, top int, cell func(row, col int) string, focused bool, a Accent) {
	w, h := pn.W, pn.H
	pn.Rect(1, 1, w-2, h-2, White)
	pn.Frame(0, 0, w, h, Black)
	pn.Clip(1, 1, w-2, h-2, func() {
		// The column the rows are sorted by, shaded all the way down.
		x := 1
		for j, col := range cols {
			if j == st.SortColumn && !col.Fixed {
				pn.Rect(x, TableHeaderHeight, st.Widths[j], h-1-TableHeaderHeight, Gray1)
			}
			x += st.Widths[j]
		}
		for i := top; i < len(order); i++ {
			y := TableHeaderHeight + (i-top)*ListRowHeight
			if y >= h-1 {
				break
			}
			sel := st.Selected[order[i]]
			ink := Black
			switch {
			case sel && focused:
				pn.Rect(1, y, w-2, ListRowHeight, a.Dark)
				ink = White
			case sel:
				pn.Rect(1, y, w-2, ListRowHeight, Gray4)
			}
			x := 1
			for j, col := range cols {
				cw := st.Widths[j]
				if !sel && j == st.SortColumn && !col.Fixed {
					// The column the rows are sorted by, shaded.
					pn.Rect(x, y, cw, ListRowHeight, Gray1)
				}
				tx := x + tableCellPad
				if col.Icon != nil {
					if ic := col.Icon(order[i]); ic != nil {
						DrawIcon(pn.Sub(x+4, y+(ListRowHeight-len(ic))/2, iconWidth(ic), len(ic)), ic, a, sel && focused)
					}
					tx = x + 4 + 16 + 4
				}
				text := AppFont.Truncate(cell(order[i], j), cw-(tx-x)-tableCellPad)
				if col.Right {
					tx = x + cw - tableCellPad - AppFont.Measure(text)
				}
				pn.Clip(x, y, cw, ListRowHeight, func() {
					pn.TextMiddle(tx, y, ListRowHeight, text, AppFont, ink)
				})
				x += cw
				if !sel {
					pn.VLine(x-1, y, ListRowHeight, Gray3)
				}
			}
		}
		// The headers.
		x = 1
		for j, col := range cols {
			cw := st.Widths[j]
			hp := pn.Sub(x, 1, cw, TableHeaderHeight-1)
			face := Gray2
			if j == st.SortColumn && !col.Fixed {
				face = Gray4
			}
			hp.Rect(0, 0, hp.W, hp.H, face)
			hp.Bevel(0, 0, hp.W, hp.H-1, White, Gray7)
			hp.HLine(0, hp.H-1, hp.W, Black)
			hp.VLine(hp.W-1, 0, hp.H, Gray7)
			title := AppFont.Truncate(col.Title, cw-2*tableCellPad-8)
			hp.TextMiddle(tableCellPad, 0, hp.H-1, title, AppFont, Black)
			if j == st.SortColumn && !col.Fixed {
				arrow := sortUp
				if st.Descending {
					arrow = sortDown
				}
				hp.Bitmap(hp.W-tableCellPad-5, (hp.H-4)/2, arrow, Black)
			}
			x += cw
		}
		if x < w-1 {
			pn.Rect(x, 1, w-1-x, TableHeaderHeight-1, Gray2)
			pn.HLine(x, TableHeaderHeight-1, w-1-x, Black)
		}
	})
}
