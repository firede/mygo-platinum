package platinum

import (
	"strings"
	"time"
	"unicode"

	"github.com/egoist/mygo/ui"
)

// editor is the text, caret, selection and undo history that text fields
// and text areas edit, with the layout of its lines in a bitmap font.
type editor struct {
	text          []rune
	caret, anchor int
	compose       string // the composition of an input method, at the caret
	composeCaret  int    // the rune offset inside the composition
	undo, redo    []editSnap
	lastEdit      editKind
	// Layout.
	font        Font
	wrap        int // the width lines wrap at, or 0 for none
	multi       bool
	lines       []lineSpan
	laidFor     string
	goalX       int  // the x the caret keeps going up and down
	verticalRun bool // the last move was up or down, keeping goalX
	dragging    bool
	clicks      int
	maxLen      int
	readOnly    bool
	noCopy      bool // a password: its text never leaves it
	// A rich text's style of each rune, the style typing gives, and the
	// alignment of each of its paragraphs (left where it has none).
	rich   bool
	styles []TextStyle
	typing TextStyle
	paras  []Align
	indent int // how far the first line of each paragraph starts in
}

type editSnap struct {
	text          []rune
	styles        []TextStyle
	paras         []Align
	caret, anchor int
}

type editKind int

const (
	editNone editKind = iota
	editTyping
	editDeleting
	editOther
)

type lineSpan struct{ start, end, para int } // runes [start, end) of a visual line, in paragraph para

func (e *editor) String() string { return string(e.text) }

// set replaces the text from outside, as when the bound string changes.
func (e *editor) set(s string) {
	e.text = []rune(s)
	if e.rich {
		// Text set from outside keeps the styles of the runes it keeps.
		st := make([]TextStyle, len(e.text))
		copy(st, e.styles)
		e.styles = st
		e.paras = append([]Align(nil), e.paras[:min(len(e.paras), e.paraOf(len(e.text))+1)]...)
	}
	e.lines = nil
	e.caret = min(e.caret, len(e.text))
	e.anchor = min(e.anchor, len(e.text))
	e.undo, e.redo = nil, nil
	e.lastEdit = editNone
}

func (e *editor) selection() (int, int) {
	return min(e.caret, e.anchor), max(e.caret, e.anchor)
}

func (e *editor) hasSelection() bool { return e.caret != e.anchor }

func (e *editor) selectedText() string {
	a, b := e.selection()
	return string(e.text[a:b])
}

// displayEditor lays out preedit text in place of the selection without
// changing the document or its undo history.
func (e *editor) displayEditor() *editor {
	if e.compose == "" {
		return e
	}
	a, b := e.selection()
	preedit := []rune(e.compose)
	text := make([]rune, 0, len(e.text)-(b-a)+len(preedit))
	text = append(text, e.text[:a]...)
	text = append(text, preedit...)
	text = append(text, e.text[b:]...)
	view := &editor{
		text: text, font: e.font, wrap: e.wrap, multi: e.multi,
		rich: true, indent: e.indent,
		styles: make([]TextStyle, len(text)),
		paras:  append([]Align(nil), e.paras...),
	}
	if e.rich {
		copy(view.styles[:a], e.styles[:a])
		copy(view.styles[a+len(preedit):], e.styles[b:])
	}
	for i := a; i < a+len(preedit); i++ {
		view.styles[i] = e.typing | Underline
	}
	// Replacing a selection can join paragraphs; preedit can split them.
	p := e.paraOf(a)
	gone := strings.Count(string(e.text[a:b]), "\n")
	added := make([]Align, strings.Count(e.compose, "\n"))
	for i := range added {
		added[i] = e.paraAlign(p)
	}
	if len(view.paras) > p {
		view.paras = append(view.paras[:p+1:p+1], append(added, view.paras[min(p+1+gone, len(view.paras)):]...)...)
	}
	view.caret = a + min(max(e.composeCaret, 0), len(preedit))
	view.anchor = view.caret
	return view
}

func (e *editor) snapshot(kind editKind) {
	if kind != editOther && kind == e.lastEdit {
		return // typing and deleting group into one undo
	}
	e.undo = append(e.undo, editSnap{append([]rune(nil), e.text...), append([]TextStyle(nil), e.styles...), append([]Align(nil), e.paras...), e.caret, e.anchor})
	if len(e.undo) > 200 {
		e.undo = e.undo[1:]
	}
	e.redo = nil
	e.lastEdit = kind
}

// replace puts s in place of the selection.
func (e *editor) replace(s string, kind editKind) bool {
	if e.readOnly {
		return false
	}
	if !e.multi {
		s = strings.Map(func(r rune) rune {
			if r == '\n' || r == '\r' {
				return ' '
			}
			return r
		}, s)
	}
	rs := []rune(s)
	a, b := e.selection()
	if e.maxLen > 0 && len(e.text)-(b-a)+len(rs) > e.maxLen {
		rs = rs[:max(0, e.maxLen-(len(e.text)-(b-a)))]
	}
	if a == b && len(rs) == 0 {
		return false
	}
	e.snapshot(kind)
	if e.rich {
		st := make([]TextStyle, len(rs))
		for i := range st {
			st[i] = e.typing
		}
		e.styles = append(e.styles[:a:a], append(st, e.styles[b:]...)...)
		// The paragraphs the selection joined go; those typed take the
		// alignment of the paragraph they were typed in.
		p := e.paraOf(a)
		gone := strings.Count(string(e.text[a:b]), "\n")
		added := make([]Align, strings.Count(string(rs), "\n"))
		for i := range added {
			added[i] = e.paraAlign(p)
		}
		if len(e.paras) > p {
			e.paras = append(e.paras[:p+1:p+1], append(added, e.paras[min(p+1+gone, len(e.paras)):]...)...)
		}
	}
	e.text = append(e.text[:a:a], append(rs, e.text[b:]...)...)
	e.caret = a + len(rs)
	e.anchor = e.caret
	return true
}

func (e *editor) undoEdit() bool {
	if len(e.undo) == 0 {
		return false
	}
	e.redo = append(e.redo, editSnap{e.text, e.styles, e.paras, e.caret, e.anchor})
	s := e.undo[len(e.undo)-1]
	e.undo = e.undo[:len(e.undo)-1]
	e.text, e.styles, e.paras, e.caret, e.anchor = s.text, s.styles, s.paras, s.caret, s.anchor
	e.lines = nil
	e.lastEdit = editNone
	return true
}

func (e *editor) redoEdit() bool {
	if len(e.redo) == 0 {
		return false
	}
	e.undo = append(e.undo, editSnap{e.text, e.styles, e.paras, e.caret, e.anchor})
	s := e.redo[len(e.redo)-1]
	e.redo = e.redo[:len(e.redo)-1]
	e.text, e.styles, e.paras, e.caret, e.anchor = s.text, s.styles, s.paras, s.caret, s.anchor
	e.lines = nil
	e.lastEdit = editNone
	return true
}

func isWord(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' }

func (e *editor) wordLeft(i int) int {
	for i > 0 && !isWord(e.text[i-1]) {
		i--
	}
	for i > 0 && isWord(e.text[i-1]) {
		i--
	}
	return i
}

func (e *editor) wordRight(i int) int {
	n := len(e.text)
	for i < n && !isWord(e.text[i]) {
		i++
	}
	for i < n && isWord(e.text[i]) {
		i++
	}
	return i
}

// layout breaks the text into visual lines.
func (e *editor) layout() {
	key := string(e.text)
	if e.lines != nil && e.laidFor == key {
		return
	}
	e.laidFor = key
	e.lines = e.lines[:0]
	start, para := 0, 0
	for i := 0; i <= len(e.text); i++ {
		if i < len(e.text) && e.text[i] != '\n' {
			continue
		}
		e.wrapParagraph(start, i, para)
		start = i + 1
		para++
	}
}

func (e *editor) wrapParagraph(start, end, para int) {
	if !e.multi || e.wrap <= 0 {
		e.lines = append(e.lines, lineSpan{start, end, para})
		return
	}
	face := e.font.face()
	first := true
	for {
		w, lastSpace, i := 0, -1, start
		wrap := e.wrap
		if first {
			wrap -= e.indent
		}
		first = false
		for ; i < end; i++ {
			a := e.adv(i)
			if w+a-face.Tracking > wrap && i > start {
				break
			}
			if e.text[i] == ' ' {
				lastSpace = i
			}
			w += a
		}
		if i >= end {
			e.lines = append(e.lines, lineSpan{start, end, para})
			return
		}
		brk := i
		if lastSpace >= start {
			brk = lastSpace + 1 // the space stays at the end of the line
		}
		e.lines = append(e.lines, lineSpan{start, brk, para})
		start = brk
	}
}

// paraOf returns the paragraph holding rune offset i.
func (e *editor) paraOf(i int) int {
	n := 0
	for _, r := range e.text[:min(i, len(e.text))] {
		if r == '\n' {
			n++
		}
	}
	return n
}

// paraAlign returns the alignment of paragraph p.
func (e *editor) paraAlign(p int) Align {
	if p < len(e.paras) {
		return e.paras[p]
	}
	return AlignLeft
}

// setAlign aligns the paragraphs from a to b.
func (e *editor) setAlign(a, b int, align Align) {
	for len(e.paras) <= b {
		e.paras = append(e.paras, AlignLeft)
	}
	for p := a; p <= b; p++ {
		e.paras[p] = align
	}
}

// lineOf returns the visual line holding rune offset i.
func (e *editor) lineOf(i int) int {
	e.layout()
	for n, l := range e.lines {
		if i < l.end || i == l.end && (n == len(e.lines)-1 || e.lines[n+1].start > l.end) {
			return n
		}
	}
	return len(e.lines) - 1
}

// adv returns the advance of rune i: a pixel more in a bold run, and the
// slant's room after the last rune of an italic run.
func (e *editor) adv(i int) int {
	a := e.font.advance(e.text[i])
	if !e.rich || i >= len(e.styles) {
		return a
	}
	st := e.styles[i]
	if st&Bold != 0 {
		a++
	}
	if st&Italic != 0 && (i+1 >= len(e.text) || e.styles[i+1]&(Italic|scaleBits) != st&(Italic|scaleBits) || e.text[i+1] == '\n') {
		a += ItalicSlant(e.font)
	}
	return a * st.Scale()
}

// lineScale returns the largest scale of the runes of line n.
func (e *editor) lineScale(n int) int {
	k := 1
	if e.rich {
		l := e.lines[n]
		for i := l.start; i < l.end && i < len(e.styles); i++ {
			k = max(k, e.styles[i].Scale())
		}
	}
	return k
}

// lineHeight returns the height of line n, where a line of text at scale 1
// is lineH tall: a line of larger text is taller by what its text is.
func (e *editor) lineHeight(n, lineH int) int {
	return lineH + (e.lineScale(n)-1)*e.font.Height()
}

// linesHeight returns the height of the lines from a to b.
func (e *editor) linesHeight(a, b, lineH int) int {
	h := 0
	for n := a; n < b && n < len(e.lines); n++ {
		h += e.lineHeight(n, lineH)
	}
	return h
}

// lineAtY returns the line y below the top of line top.
func (e *editor) lineAtY(top, y, lineH int) int {
	n := min(top, len(e.lines)-1)
	for y < 0 && n > 0 { // above it, as a drag goes
		n--
		y += e.lineHeight(n, lineH)
	}
	for ; n < len(e.lines)-1; n++ {
		h := e.lineHeight(n, lineH)
		if y < h {
			break
		}
		y -= h
	}
	return n
}

// measure returns the width of the runes from a to b.
func (e *editor) measure(a, b int) int {
	if !e.rich {
		w := 0
		for _, r := range e.text[a:b] {
			w += e.font.advance(r)
		}
		return w
	}
	w := 0
	for i := a; i < b; i++ {
		w += e.adv(i)
	}
	return w
}

// lineX returns how far line n starts in from the left, by the alignment.
func (e *editor) lineX(n int) int {
	l := e.lines[n]
	in := 0
	if e.indent > 0 && e.multi && e.wrap > 0 && (l.start == 0 || e.text[l.start-1] == '\n') {
		in = e.indent // the first line of a paragraph
	}
	align := e.paraAlign(l.para)
	if align == AlignLeft || e.wrap <= 0 {
		return in
	}
	end := l.end
	for end > l.start && (e.text[end-1] == ' ' || e.text[end-1] == '\n') {
		end--
	}
	room := e.wrap - in - e.measure(l.start, end)
	if align == AlignCenter {
		return in + max(room/2, 0)
	}
	return in + max(room, 0)
}

// xOf returns the x of rune offset i on its line.
func (e *editor) xOf(i int) int {
	n := e.lineOf(i)
	l := e.lines[n]
	return e.lineX(n) + e.measure(l.start, i)
}

// caretRect is shared by painting and the input method's candidate anchor.
func (e *editor) caretRect(x0, y0, lineH, top, scrollX int) ui.Rect {
	n := e.lineOf(e.caret)
	return ui.Rect{
		X: float32(x0 - scrollX + e.xOf(e.caret) - 1),
		Y: float32(y0 + e.linesHeight(top, n, lineH) + 1),
		W: 1,
		H: float32(e.lineHeight(n, lineH) - 2),
	}
}

// offsetAt returns the rune offset nearest x on line n.
func (e *editor) offsetAt(n, x int) int {
	e.layout()
	n = min(max(n, 0), len(e.lines)-1)
	l := e.lines[n]
	cx := e.lineX(n)
	end := l.end
	if n < len(e.lines)-1 && e.lines[n+1].start == l.end && end > l.start {
		end-- // the caret goes before the space a line wrapped at
	}
	for i := l.start; i < end; i++ {
		a := e.adv(i)
		if x < cx+a/2 {
			return i
		}
		cx += a
	}
	return end
}

func (e *editor) moveTo(i int, extend bool) {
	e.caret = min(max(i, 0), len(e.text))
	if !extend {
		e.anchor = e.caret
		if e.rich && len(e.styles) > 0 {
			// Typing goes on in the style of the rune before the caret.
			e.typing = e.styles[max(e.caret-1, 0)]
		}
	}
	e.lastEdit = editNone
}

// key handles a key, reporting whether it took it and whether the text
// changed.
func (e *editor) key(ev ui.InputEvent, svc clipboard) (took, changed bool) {
	shift := ev.Mods&ui.Shift != 0
	cmd := ev.Mods&ui.Cmd != 0
	alt := ev.Mods&ui.Alt != 0
	e.layout()
	vertical := false
	switch ev.Key {
	case ui.KeyLeft:
		switch {
		case e.hasSelection() && !shift && !cmd && !alt:
			a, _ := e.selection()
			e.moveTo(a, false)
		case cmd:
			e.moveTo(e.lines[e.lineOf(e.caret)].start, shift)
		case alt:
			e.moveTo(e.wordLeft(e.caret), shift)
		default:
			e.moveTo(e.caret-1, shift)
		}
	case ui.KeyRight:
		switch {
		case e.hasSelection() && !shift && !cmd && !alt:
			_, b := e.selection()
			e.moveTo(b, false)
		case cmd:
			l := e.lines[e.lineOf(e.caret)]
			e.moveTo(e.offsetAt(e.lineOf(e.caret), 1<<20), shift)
			_ = l
		case alt:
			e.moveTo(e.wordRight(e.caret), shift)
		default:
			e.moveTo(e.caret+1, shift)
		}
	case ui.KeyUp, ui.KeyDown:
		if !e.multi {
			if ev.Key == ui.KeyUp {
				e.moveTo(0, shift)
			} else {
				e.moveTo(len(e.text), shift)
			}
			break
		}
		if cmd {
			if ev.Key == ui.KeyUp {
				e.moveTo(0, shift)
			} else {
				e.moveTo(len(e.text), shift)
			}
			break
		}
		n := e.lineOf(e.caret)
		if e.lastEdit != editNone || !e.verticalRun {
			e.goalX = e.xOf(e.caret)
		}
		d := 1
		if ev.Key == ui.KeyUp {
			d = -1
		}
		switch {
		case n+d < 0:
			e.moveTo(0, shift)
		case n+d >= len(e.lines):
			e.moveTo(len(e.text), shift)
		default:
			e.moveTo(e.offsetAt(n+d, e.goalX), shift)
		}
		vertical = true
	case ui.KeyHome:
		e.moveTo(0, shift)
	case ui.KeyEnd:
		e.moveTo(len(e.text), shift)
	case ui.KeyBackspace, ui.KeyDelete:
		if e.readOnly {
			return true, false
		}
		if !e.hasSelection() {
			back := ev.Key == ui.KeyBackspace
			switch {
			case back && cmd:
				e.anchor = e.lines[e.lineOf(e.caret)].start
			case back && alt:
				e.anchor = e.wordLeft(e.caret)
			case back:
				e.anchor = max(e.caret-1, 0)
			case alt:
				e.anchor = e.wordRight(e.caret)
			default:
				e.anchor = min(e.caret+1, len(e.text))
			}
		}
		changed = e.replace("", editDeleting)
	case ui.KeyEnter:
		if !e.multi || e.readOnly {
			return false, false
		}
		changed = e.replace("\n", editOther)
	case ui.KeyTab:
		return false, false
	case ui.KeyA, ui.KeyC, ui.KeyX, ui.KeyV, ui.KeyZ:
		if !cmd {
			return false, false
		}
		name := map[ui.Key]string{ui.KeyA: "selectAll", ui.KeyC: "copy", ui.KeyX: "cut", ui.KeyV: "paste", ui.KeyZ: "undo"}[ev.Key]
		if ev.Key == ui.KeyZ && shift {
			name = "redo"
		}
		return e.command(name, svc)
	default:
		return false, false
	}
	e.verticalRun = vertical
	return true, changed
}

// command performs an edit command of the menus.
func (e *editor) command(name string, svc clipboard) (took, changed bool) {
	switch name {
	case "selectAll":
		e.anchor, e.caret = 0, len(e.text)
	case "copy":
		if e.noCopy {
			return true, false
		}
		if e.hasSelection() && svc != nil {
			svc.WriteClipboard(e.selectedText())
		}
	case "cut":
		if e.noCopy {
			return true, false
		}
		if e.hasSelection() && !e.readOnly {
			if svc != nil {
				svc.WriteClipboard(e.selectedText())
			}
			changed = e.replace("", editOther)
		}
	case "paste":
		if svc != nil && !e.readOnly {
			changed = e.replace(svc.ReadClipboard(), editOther)
		}
	case "undo":
		changed = !e.readOnly && e.undoEdit()
	case "redo":
		changed = !e.readOnly && e.redoEdit()
	case "delete":
		changed = e.replace("", editOther)
	default:
		return false, false
	}
	return true, changed
}

// clipboard is what the editor copies to and pastes from.
type clipboard interface {
	ReadClipboard() string
	WriteClipboard(string)
}

// selectWord selects the word around offset i.
func (e *editor) selectWord(i int) {
	a, b := i, i
	for a > 0 && isWord(e.text[a-1]) {
		a--
	}
	for b < len(e.text) && isWord(e.text[b]) {
		b++
	}
	e.anchor, e.caret = a, b
}

// drawRuns draws line l of a rich text from x, y, a run of one style at
// a time, all on the baseline of the line's largest text, at scale k.
func (e *editor) drawRuns(pn Pen, x, y int, l lineSpan, k int, ink ui.Color) {
	for i := l.start; i < l.end; {
		st := e.styles[i]
		j := i
		for j < l.end && e.styles[j] == st && e.text[j] != '\n' {
			j++
		}
		if j == i {
			i++
			continue
		}
		// Each run from where the caret measures it to start.
		pn.TextStyled(x+e.measure(l.start, i), y+(k-st.Scale())*e.font.Ascent(), string(e.text[i:j]), e.font, ink, st)
		i = j
	}
}

// caretBlink is how long the caret shows, and then hides.
const caretBlink = 530 * time.Millisecond

// drawLines draws the lines of the editor from line top, each lineH tall
// or taller for larger text, at x0 (scrolled left by scrollX) in pn, with
// its selection and caret.
func (e *editor) drawLines(pn Pen, x0, y0, lineH0, top, scrollX int, ink, selBG ui.Color, caret bool) {
	e.layout()
	a, b := e.selection()
	ty := (lineH0 - e.font.Height()) / 2
	y := y0
	for n := top; n < len(e.lines); n++ {
		if y >= pn.H {
			break
		}
		lineH := e.lineHeight(n, lineH0)
		y += lineH
		y := y - lineH
		l := e.lines[n]
		lx := e.lineX(n)
		// The selection behind the text.
		if a != b && a <= l.end && b >= l.start {
			sa, sb := max(a, l.start), min(b, l.end)
			xa := x0 - scrollX + lx + e.measure(l.start, sa)
			xb := x0 - scrollX + lx + e.measure(l.start, sb)
			if b > l.end && n < len(e.lines)-1 {
				xb = pn.W // the selection runs on past the line's end
			}
			pn.Rect(xa, y, xb-xa, lineH, selBG)
		}
		x := x0 - scrollX + lx
		seg := string(e.text[l.start:l.end])
		if e.rich {
			e.drawRuns(pn, x, y+ty, l, e.lineScale(n), ink)
			continue
		}
		pn.Text(x, y+ty, strings.TrimRight(seg, "\n"), e.font, ink)
	}
	if caret && !e.hasSelection() {
		n := e.lineOf(e.caret)
		if n >= top {
			r := e.caretRect(x0, y0, lineH0, top, scrollX)
			pn.VLine(int(r.X), int(r.Y), int(r.H), ink)
		}
	}
}

// TextStyle is the style of a run of rich text.
type TextStyle uint8

// The styles of rich text, combined as bits, with a scale (see Scaled).
const (
	Bold TextStyle = 1 << iota
	Italic
	Underline
)

// MaxScale is the largest scale of rich text, as Scaled sets it.
const MaxScale = 4

const scaleBits TextStyle = 3 << 3 // the scale less one

// Scale returns the size of text in style s, as a multiple of the face's:
// text at scale k draws each pixel of its glyphs as k by k, as a bitmap
// font was drawn at a size it had no bitmaps for.
func (s TextStyle) Scale() int { return int(s&scaleBits>>3) + 1 }

// Scaled returns s at scale k, from 1 to MaxScale.
func (s TextStyle) Scaled(k int) TextStyle {
	k = min(max(k, 1), MaxScale)
	return s&^scaleBits | TextStyle(k-1)<<3
}

// Align is the alignment of the lines of rich text.
type Align uint8

// The alignments of the lines of rich text.
const (
	AlignLeft Align = iota
	AlignCenter
	AlignRight
)
