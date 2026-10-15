package platinum

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/egoist/mygo/ui"
)

// Sheet metrics, from docs/spec.md.
const (
	SheetCellW    = 72 // the width of a column of a sheet
	SheetCellH    = 18 // the height of a row
	SheetRowHeadW = 32 // the width of the row numbers
)

// Cell names a cell of a sheet, from 0, 0 at A1.
type Cell struct{ Col, Row int }

// String returns the cell's name, as "B3".
func (c Cell) String() string { return colName(c.Col) + strconv.Itoa(c.Row+1) }

func colName(i int) string {
	s := ""
	for i++; i > 0; i = (i - 1) / 26 {
		s = string(rune('A'+(i-1)%26)) + s
	}
	return s
}

// ParseCell reads a cell's name, as "B3".
func ParseCell(s string) (Cell, bool) {
	s = strings.ToUpper(strings.TrimSpace(s))
	i := 0
	col := 0
	for i < len(s) && s[i] >= 'A' && s[i] <= 'Z' {
		col = col*26 + int(s[i]-'A'+1)
		i++
	}
	if i == 0 || i == len(s) {
		return Cell{}, false
	}
	row, err := strconv.Atoi(s[i:])
	if err != nil || row < 1 {
		return Cell{}, false
	}
	return Cell{col - 1, row - 1}, true
}

// SheetState is a sheet's cells and what the user does to them, kept by
// the app: the text of each cell as typed (numbers, text, or formulas
// beginning with "="), the cell chosen and the range selected, the cell
// being edited, and how far the sheet is scrolled.
type SheetState struct {
	Cells map[Cell]string
	// Cursor is the cell chosen; Anchor the other corner of the range
	// selected with Shift.
	Cursor, Anchor Cell
	// Editing is the text being typed into the cursor's cell, while
	// IsEditing.
	Editing   string
	IsEditing bool

	top, left float64
	rect      ui.Rect
	changed   bool
}

// NewSheetState returns an empty sheet.
func NewSheetState() *SheetState { return &SheetState{Cells: map[Cell]string{}} }

// Value returns what a cell shows: its number or text, or the result of
// its formula, or an error such as "#REF!".
func (s *SheetState) Value(c Cell) string {
	v, err := s.eval(c, map[Cell]bool{})
	if err != "" {
		return err
	}
	return v
}

func (s *SheetState) eval(c Cell, seen map[Cell]bool) (string, string) {
	raw := s.Cells[c]
	if !strings.HasPrefix(raw, "=") {
		return raw, ""
	}
	if seen[c] {
		return "", "#CYCLE!"
	}
	seen[c] = true
	defer delete(seen, c)
	p := &formula{src: strings.ToUpper(raw[1:]), sheet: s, seen: seen}
	v := p.expr()
	p.skip()
	if p.err == "" && p.pos < len(p.src) {
		p.err = "#SYNTAX!"
	}
	if p.err != "" {
		return "", p.err
	}
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return "", "#DIV/0!"
	}
	return strconv.FormatFloat(v, 'g', 10, 64), ""
}

// number returns a cell's value as a number: 0 for empty or text cells.
func (s *SheetState) number(c Cell, seen map[Cell]bool) (float64, string) {
	v, err := s.eval(c, seen)
	if err != "" {
		return 0, err
	}
	f, perr := strconv.ParseFloat(strings.TrimSpace(v), 64)
	if perr != nil {
		return 0, ""
	}
	return f, ""
}

// formula parses and evaluates a formula: numbers, cells, + - * /,
// parentheses, and SUM, AVERAGE, MIN and MAX of ranges such as A1:B3.
type formula struct {
	src   string
	pos   int
	sheet *SheetState
	seen  map[Cell]bool
	err   string
}

func (p *formula) skip() {
	for p.pos < len(p.src) && p.src[p.pos] == ' ' {
		p.pos++
	}
}

func (p *formula) peek() byte {
	p.skip()
	if p.pos < len(p.src) {
		return p.src[p.pos]
	}
	return 0
}

func (p *formula) expr() float64 {
	v := p.term()
	for p.err == "" {
		switch p.peek() {
		case '+':
			p.pos++
			v += p.term()
		case '-':
			p.pos++
			v -= p.term()
		default:
			return v
		}
	}
	return v
}

func (p *formula) term() float64 {
	v := p.factor()
	for p.err == "" {
		switch p.peek() {
		case '*':
			p.pos++
			v *= p.factor()
		case '/':
			p.pos++
			v /= p.factor()
		default:
			return v
		}
	}
	return v
}

func (p *formula) ident() string {
	p.skip()
	start := p.pos
	for p.pos < len(p.src) && (unicode.IsLetter(rune(p.src[p.pos])) || unicode.IsDigit(rune(p.src[p.pos]))) {
		p.pos++
	}
	return p.src[start:p.pos]
}

func (p *formula) factor() float64 {
	switch b := p.peek(); {
	case b == '-':
		p.pos++
		return -p.factor()
	case b == '(':
		p.pos++
		v := p.expr()
		if p.peek() != ')' {
			p.err = "#SYNTAX!"
			return 0
		}
		p.pos++
		return v
	case b >= '0' && b <= '9' || b == '.':
		start := p.pos
		for p.pos < len(p.src) && (p.src[p.pos] >= '0' && p.src[p.pos] <= '9' || p.src[p.pos] == '.') {
			p.pos++
		}
		v, err := strconv.ParseFloat(p.src[start:p.pos], 64)
		if err != nil {
			p.err = "#SYNTAX!"
		}
		return v
	case b >= 'A' && b <= 'Z':
		name := p.ident()
		if p.peek() == '(' {
			return p.call(name)
		}
		c, ok := ParseCell(name)
		if !ok {
			p.err = "#NAME?"
			return 0
		}
		v, err := p.sheet.number(c, p.seen)
		if err != "" {
			p.err = err
		}
		return v
	}
	p.err = "#SYNTAX!"
	return 0
}

func (p *formula) call(name string) float64 {
	p.pos++ // (
	var vals []float64
	for p.err == "" {
		if p.peek() == ')' {
			p.pos++
			break
		}
		start := p.pos
		a := p.ident()
		if ca, ok := ParseCell(a); ok && p.peek() == ':' {
			p.pos++
			cb, ok := ParseCell(p.ident())
			if !ok {
				p.err = "#REF!"
				return 0
			}
			for r := min(ca.Row, cb.Row); r <= max(ca.Row, cb.Row); r++ {
				for col := min(ca.Col, cb.Col); col <= max(ca.Col, cb.Col); col++ {
					v, err := p.sheet.number(Cell{col, r}, p.seen)
					if err != "" {
						p.err = err
					}
					vals = append(vals, v)
				}
			}
		} else {
			p.pos = start
			vals = append(vals, p.expr())
		}
		if p.peek() == ',' {
			p.pos++
		}
	}
	if len(vals) == 0 {
		return 0
	}
	switch name {
	case "SUM", "AVERAGE":
		t := 0.0
		for _, v := range vals {
			t += v
		}
		if name == "AVERAGE" {
			t /= float64(len(vals))
		}
		return t
	case "MIN", "MAX":
		m := vals[0]
		for _, v := range vals {
			if name == "MIN" {
				m = min(m, v)
			} else {
				m = max(m, v)
			}
		}
		return m
	}
	p.err = "#NAME?"
	return 0
}

// SheetWidget is a sheet being built.
type SheetWidget struct {
	el      ui.Element
	changed bool
}

// Sheet builds a grid of cells of rows rows and cols columns, as a
// spreadsheet: lettered column headers, numbered rows, and the cells' values
// right-aligned when they are numbers. A click chooses a cell (Shift
// extends the range); with the keyboard focus, the arrows move the cursor
// (Shift extends), typing starts editing the cell, Return commits and moves
// down, Tab commits and moves right (from the last column, on to the next
// control; Shift-Tab back from the first), Escape cancels, Delete clears
// the range.
// Pair it with a FormulaBar to show and edit the cursor's text.
func Sheet(c *ui.Context, st *SheetState, rows, cols int) *SheetWidget {
	th := ThemeOf(c)
	if st.Cells == nil {
		st.Cells = map[Cell]string{}
	}
	w := &SheetWidget{}
	w.el = ui.Row(c).Focusable().FocusRing(false).Role(ui.RoleTable).Label(say(c, "sheet")).Size(320, 200).AlignItems(ui.Stretch).Padding(1)
	w.changed, st.changed = st.changed, false
	visRows := max(int((st.rect.H-4-SheetCellH)/SheetCellH), 1)
	visCols := max(int((st.rect.W-4-SheetRowHeadW-ScrollBarWidth)/SheetCellW), 1)
	clampCell := func(c Cell) Cell {
		return Cell{min(max(c.Col, 0), cols-1), min(max(c.Row, 0), rows-1)}
	}
	reveal := func() {
		cu := st.Cursor
		if float64(cu.Row) < st.top {
			st.top = float64(cu.Row)
		} else if float64(cu.Row) >= st.top+float64(visRows) {
			st.top = float64(cu.Row - visRows + 1)
		}
		if float64(cu.Col) < st.left {
			st.left = float64(cu.Col)
		} else if float64(cu.Col) >= st.left+float64(visCols) {
			st.left = float64(cu.Col - visCols + 1)
		}
	}
	commit := func() {
		if st.IsEditing {
			if st.Editing == "" {
				delete(st.Cells, st.Cursor)
			} else {
				st.Cells[st.Cursor] = st.Editing
			}
			st.IsEditing = false
			st.changed, w.changed = true, true
		}
	}
	move := func(dc, dr int, extend bool) {
		commit()
		st.Cursor = clampCell(Cell{st.Cursor.Col + dc, st.Cursor.Row + dr})
		if !extend {
			st.Anchor = st.Cursor
		}
		reveal()
	}
	for _, k := range []struct {
		key    ui.Key
		dc, dr int
	}{{ui.KeyLeft, -1, 0}, {ui.KeyRight, 1, 0}, {ui.KeyUp, 0, -1}, {ui.KeyDown, 0, 1}} {
		if !st.IsEditing && w.el.Shortcut(0, k.key) {
			move(k.dc, k.dr, false)
		}
		if !st.IsEditing && w.el.Shortcut(ui.Shift, k.key) {
			move(k.dc, k.dr, true)
		}
	}
	w.el.HandleInput(func(ev ui.InputEvent) bool {
		switch ev.Kind {
		case ui.InputKeyDown:
			switch ev.Key {
			case ui.KeyEnter:
				move(0, 1, false)
				return true
			case ui.KeyTab:
				// From the last column, or back from the first, Tab leaves
				// the sheet for the next control, so that it is no trap.
				back := ev.Mods&ui.Shift != 0
				if back && st.Cursor.Col == 0 || !back && st.Cursor.Col == cols-1 {
					commit()
					return false
				}
				if back {
					move(-1, 0, false)
				} else {
					move(1, 0, false)
				}
				return true
			case ui.KeyEscape:
				if st.IsEditing {
					st.IsEditing = false
					return true
				}
			case ui.KeyBackspace, ui.KeyDelete:
				if st.IsEditing {
					r := []rune(st.Editing)
					if len(r) > 0 {
						st.Editing = string(r[:len(r)-1])
					}
					return true
				}
				for r := min(st.Cursor.Row, st.Anchor.Row); r <= max(st.Cursor.Row, st.Anchor.Row); r++ {
					for col := min(st.Cursor.Col, st.Anchor.Col); col <= max(st.Cursor.Col, st.Anchor.Col); col++ {
						delete(st.Cells, Cell{col, r})
					}
				}
				st.changed = true
				return true
			case ui.KeyLeft, ui.KeyRight, ui.KeyUp, ui.KeyDown:
				if st.IsEditing {
					return true // the caret stays at the end of what is typed
				}
			}
		case ui.InputText:
			if ev.Text == "\r" || ev.Text == "\n" || ev.Text == "\t" {
				return true
			}
			if !st.IsEditing {
				st.IsEditing = true
				st.Editing = ""
			}
			st.Editing += ev.Text
			return true
		case ui.InputPointerDown:
			if ev.Button != 0 {
				return false
			}
			col := int(st.left) + int((ev.X-2-SheetRowHeadW)/SheetCellW)
			row := int(st.top) + int((ev.Y-2-SheetCellH)/SheetCellH)
			if ev.X < 2+SheetRowHeadW || ev.Y < 2+SheetCellH {
				return false
			}
			commit()
			st.Cursor = clampCell(Cell{col, row})
			if ev.Mods&ui.Shift == 0 {
				st.Anchor = st.Cursor
			}
			return false // and take the focus
		case ui.InputScroll:
			st.top = min(max(st.top+float64(ev.DY)/SheetCellH, 0), float64(max(rows-visRows, 0)))
			st.left = min(max(st.left+float64(ev.DX)/SheetCellW, 0), float64(max(cols-visCols, 0)))
			return true
		}
		return false
	})
	w.el.TextCaret(ui.Rect{})
	focused := w.el.Focused()
	top, left := int(st.top), int(st.left)
	bg := inactive(c)
	w.el.DrawOver(func(p *ui.Painter, r ui.Rect) {
		st.rect = r
		pn := NewPen(p, r)
		DrawContentRing(pn)
		in := pn.Sub(1, 1, pn.W-2, pn.H-2)
		sp := in.Sub(0, 0, in.W-ScrollBarWidth+1, in.H)
		DrawSheet(sp, st, top, left, rows, cols, focused, th.Accent)
		if bg {
			// In the background, its headers gray, the cells as they are.
			dp := sp.Dimmed(true)
			dp.Clip(1, 1, dp.W-2, SheetCellH, func() { DrawSheet(dp, st, top, left, rows, cols, focused, th.Accent) })
			dp.Clip(1, 1, SheetRowHeadW, dp.H-2, func() { DrawSheet(dp, st, top, left, rows, cols, focused, th.Accent) })
		}
		if focused {
			DrawFocusRing(pn, 0, th.Accent.Dark)
		}
	})
	w.el.Children(func() {
		ui.Box(c).Grow(1).PassThrough().Role(ui.RoleNone)
		ui.Column(c).Width(ScrollBarWidth).Children(func() {
			ScrollBar(c, &st.top, float64(visRows), float64(rows)).Step(1).Element().Grow(1)
		})
	})
	w.el.Value(st.Cursor.String() + " " + st.Value(st.Cursor))
	return w
}

// Changed reports whether the user changed a cell since the last frame.
func (w *SheetWidget) Changed() bool { return w.changed }

// Element returns the sheet's element, to size it.
func (w *SheetWidget) Element() ui.Element { return w.el }

// Label names the sheet.
func (w *SheetWidget) Label(s string) *SheetWidget { w.el.Label(s); return w }

// DrawSheet draws the cells of a sheet from row top and column left in
// pn's box, which holds its outline.
func DrawSheet(pn Pen, st *SheetState, top, left, rows, cols int, focused bool, a Accent) {
	w, h := pn.W, pn.H
	pn.Rect(1, 1, w-2, h-2, White)
	pn.Frame(0, 0, w, h, Black)
	r0, r1 := min(st.Cursor.Row, st.Anchor.Row), max(st.Cursor.Row, st.Anchor.Row)
	c0, c1 := min(st.Cursor.Col, st.Anchor.Col), max(st.Cursor.Col, st.Anchor.Col)
	pn.Clip(1, 1, w-2, h-2, func() {
		header := func(x, y, cw int, label string, on bool) {
			hp := pn.Sub(x, y, cw, SheetCellH)
			face := Gray2
			if on {
				face = Gray4
			}
			hp.Rect(0, 0, hp.W, hp.H, face)
			hp.Bevel(0, 0, hp.W-1, hp.H-1, White, Gray7)
			hp.HLine(0, hp.H-1, hp.W, Gray9)
			hp.VLine(hp.W-1, 0, hp.H, Gray9)
			hp.TextCentered(0, 0, hp.W-1, hp.H-1, label, AppFont, Black)
		}
		for row := top; row < rows; row++ {
			y := 1 + SheetCellH + (row-top)*SheetCellH
			if y >= h {
				break
			}
			for col := left; col < cols; col++ {
				x := 1 + SheetRowHeadW + (col-left)*SheetCellW
				if x >= w {
					break
				}
				cell := Cell{col, row}
				inRange := row >= r0 && row <= r1 && col >= c0 && col <= c1
				if inRange && cell != st.Cursor {
					bg := Gray3
					if focused {
						bg = a.Light
					}
					pn.Rect(x, y, SheetCellW, SheetCellH, bg)
				}
				pn.HLine(x, y+SheetCellH-1, SheetCellW, Gray3)
				pn.VLine(x+SheetCellW-1, y, SheetCellH, Gray3)
				text := st.Value(cell)
				if st.IsEditing && cell == st.Cursor {
					text = st.Editing
				}
				if text == "" {
					continue
				}
				fit := AppFont.Truncate(text, SheetCellW-8)
				tx := x + 4
				if _, err := strconv.ParseFloat(text, 64); err == nil && !(st.IsEditing && cell == st.Cursor) {
					tx = x + SheetCellW - 4 - AppFont.Measure(fit)
				}
				pn.TextMiddle(tx, y, SheetCellH-1, fit, AppFont, Black)
			}
			header(1, y, SheetRowHeadW, strconv.Itoa(row+1), row >= r0 && row <= r1)
		}
		for col := left; col < cols; col++ {
			x := 1 + SheetRowHeadW + (col-left)*SheetCellW
			if x >= w {
				break
			}
			header(x, 1, SheetCellW, colName(col), col >= c0 && col <= c1)
		}
		header(1, 1, SheetRowHeadW, "", false)
		// The cursor: a frame two pixels wide.
		cx := 1 + SheetRowHeadW + (st.Cursor.Col-left)*SheetCellW
		cy := 1 + SheetCellH + (st.Cursor.Row-top)*SheetCellH
		if st.Cursor.Row >= top && st.Cursor.Col >= left {
			ink := Black
			if focused {
				ink = a.Dark
			}
			pn.Frame(cx-1, cy-1, SheetCellW+1, SheetCellH+1, ink)
			pn.Frame(cx, cy, SheetCellW-1, SheetCellH-1, ink)
			if st.IsEditing {
				tw := AppFont.Measure(AppFont.Truncate(st.Editing, SheetCellW-8))
				pn.VLine(cx+4+tw, cy+3, SheetCellH-7, Black)
			}
		}
	})
}

// FormulaBar builds the bar over a sheet that names its cursor's cell and
// shows the text typed in it — a formula where it has one — in a field
// that edits it: Return puts the text in the cell.
func FormulaBar(c *ui.Context, st *SheetState) ui.Element {
	text := st.Cells[st.Cursor]
	if st.IsEditing {
		text = st.Editing
	}
	bar := ui.Row(c).Gap(6).AlignItems(ui.Center).AlignSelf(ui.Stretch).Padding(4, 6)
	bar.Draw(func(p *ui.Painter, r ui.Rect) {
		pn := NewPen(p, r)
		pn.Rect(0, 0, pn.W, pn.H, Gray2)
		pn.HLine(0, pn.H-2, pn.W, Gray7)
		pn.HLine(0, pn.H-1, pn.W, White)
	})
	bar.Children(func() {
		ref := st.Cursor.String()
		ui.Box(c).Size(44, TextFieldHeight).Role(ui.RoleStatus).Label(say(c, "cell")).Value(ref).Draw(func(p *ui.Painter, r ui.Rect) {
			pn := NewPen(p, r)
			DrawPlacard(pn)
			pn.TextCentered(0, 0, pn.W, pn.H, ref, SystemFont, Black)
		})
		edit := text
		f := TextField(c.Key(fmt.Sprint("formula", ref)), &edit).Label(say(c, "formula"))
		f.Element().Grow(1)
		if f.Changed() {
			st.Editing, st.IsEditing = edit, true
		}
		if f.Submitted() {
			if edit == "" {
				delete(st.Cells, st.Cursor)
			} else {
				st.Cells[st.Cursor] = edit
			}
			st.IsEditing = false
			st.changed = true
		}
	})
	return bar
}
