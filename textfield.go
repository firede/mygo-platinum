package platinum

import (
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

// Text field metrics, from docs/spec.md.
const (
	TextFieldHeight = 22 // the height of a text field
	textInset       = 4  // from the field's frame to its text
	TextLineHeight  = 14 // the height of a line of a text area
)

type textState struct {
	ed        *editor
	scrollX   int
	top       float64 // the first line in view of a text area
	rect      ui.Rect
	changed   bool
	submitted bool
	activity  time.Time // when the caret last moved, which restarts its blink
}

// TextFieldWidget is a text field or text area being built.
type TextFieldWidget struct {
	c           *ui.Context
	svc         ui.Services
	el          ui.Element
	st          *textState
	value       *string
	placeholder string
	password    bool
	disabled    bool
	multi       bool
	search      bool
	// doc makes a text area a document's content: it fills its window,
	// its text starts docLeft in and wraps at docWrap (0: at its width).
	doc              bool
	docLeft, docWrap int
	changed          bool
	submitted        bool
	th               *Theme
	inactive         bool
}

// TextField builds a field editing *value on one line: a white, sunken
// box with a caret, in the regular face. It takes the keys of editing:
// the arrows (with Shift to select, Option by words, Command to the ends),
// Delete, the clipboard's Command-X, C and V, Command-Z to undo and
// Command-Shift-Z to redo, and text from input methods. Return submits it.
//
// The counterpart among MyGo's own controls is ui.TextInput.
func TextField(c *ui.Context, value *string) *TextFieldWidget {
	return newTextWidget(c, value, false)
}

// TextArea builds a box editing *value on many lines, which wrap at its
// width, with a scroll bar at its right. It edits as TextField does, and
// Return starts a new line. Give it a size.
func TextArea(c *ui.Context, value *string) *TextFieldWidget {
	return newTextWidget(c, value, true)
}

func newTextWidget(c *ui.Context, value *string, multi bool) *TextFieldWidget {
	return newTextWidgetWith(c, value, multi, nil)
}

// newTextWidgetWith builds a text widget editing with ed, which the app
// keeps (as a rich text's), or with an editor of its own if ed is nil.
func newTextWidgetWith(c *ui.Context, value *string, multi bool, ed *editor) *TextFieldWidget {
	w := &TextFieldWidget{c: c, svc: c.Services(), value: value, multi: multi, th: ThemeOf(c), inactive: inactive(c)}
	w.el = ui.Row(c).Focusable().FocusRing(false).Role(ui.RoleTextField).AlignItems(ui.Stretch)
	if multi {
		w.el.Size(200, 120).Padding(1)
	} else {
		w.el.Size(160, TextFieldHeight)
	}
	w.st = ui.Local(w.el, "text", func() textState {
		return textState{ed: &editor{font: AppFont, multi: multi}}
	})
	st := w.st
	if ed != nil {
		st.ed = ed
	}
	st.ed.noCopy = false // until Password says otherwise
	if st.changed {
		// An edit made between builds, for a value of the view bound anew.
		*value = st.ed.String()
	}
	if st.ed.String() != *value {
		st.ed.set(*value)
	}
	w.changed, st.changed = st.changed, false
	w.submitted, st.submitted = st.submitted, false
	w.el.HandleInput(w.input)
	w.el.Value(*value)
	if w.el.Focused() {
		registerEditTarget(c, func(cmd string) {
			if w.disabled {
				return
			}
			if w.password && (cmd == "copy" || cmd == "cut") {
				return
			}
			if _, changed := st.ed.command(cmd, w.svc); changed {
				*w.value = st.ed.String()
				st.changed = true
				st.ed.lines = nil
				st.ed.laidFor = ""
			}
			st.activity = wallNow()
			w.keepCaretVisible()
		})
	}
	return w.finish()
}

// finish builds the field's drawing and scroll bar. The drawing reads the
// options set after it, which come before painting.
func (w *TextFieldWidget) finish() *TextFieldWidget {
	st := w.st
	ed := st.ed
	focused := w.el.Focused()
	th := w.th
	now := frameNow(w.c)
	// The caret, for input methods.
	ed.layout()
	show := ed.displayEditor()
	show.layout()
	if !w.multi && st.rect.W > 0 {
		w.keepLineCaretVisible(show, w.textWidth(int(st.rect.W)))
	}
	lines := w.linesInView(show, int(st.top))
	w.el.TextCaret(show.caretRect(w.textLeft(), w.lineTop(TextFieldHeight), TextLineHeight, int(st.top), st.scrollX))
	w.el.Draw(func(p *ui.Painter, r ui.Rect) {
		disabled, placeholder, password := w.disabled, w.placeholder, w.password
		resized(w.c, st.rect, r)
		st.rect = r
		pn := NewPen(p, r).Faded(w.disabled)
		width := pn.W
		if w.doc {
			// A document's content: white to the window's edges, its scroll
			// bar against the window's right side.
			width = pn.W - ScrollBarWidth
			pn.Rect(0, 0, width, pn.H, White)
		} else if w.multi {
			// The ring goes round the text and its scroll bar, which shares
			// the text's outline.
			width = pn.W - 1 - ScrollBarWidth + 1
			DrawContentRing(pn)
			face, frame := White, Black
			if disabled {
				face, frame = Gray2, Gray6
			} else if w.inactive {
				frame = Gray8
			}
			pn.Frame(1, 1, width-1, pn.H-2, frame)
			pn.Rect(2, 2, width-3, pn.H-4, face)
		} else {
			DrawTextFrame(pn, disabled)
			if w.inactive && !disabled {
				pn.Frame(1, 1, pn.W-2, pn.H-2, Gray8)
			}
		}
		inner := w.textWidth(width)
		if w.search {
			DrawSearchDecor(pn.Sub(0, 0, width, pn.H), len(ed.text) > 0, disabled)
		}
		if w.multi && ed.wrap != inner-2 {
			ed.wrap = inner - 2
			ed.lines = nil
		}
		show := ed.displayEditor()
		if password {
			show = &editor{font: AppFont, text: []rune(strings.Repeat("•", len(ed.text))), caret: ed.caret, anchor: ed.anchor}
		}
		ink := Black
		if disabled {
			ink = Gray8
		}
		sel := th.Accent.Light
		if !focused {
			sel = Gray3
		}
		blinkOn := frozen() || (now.Sub(st.activity)/caretBlink)%2 == 0
		if focused && !frozen() {
			p.After(caretBlink - now.Sub(st.activity)%caretBlink)
		}
		if !w.multi {
			// Keep the caret in view, scrolling the line sideways.
			show.layout()
			w.keepLineCaretVisible(show, inner)
		}
		pn.Clip(2, 2, width-4, pn.H-4, func() {
			if len(ed.text) == 0 && placeholder != "" && ed.compose == "" {
				pn.Text(w.textLeft(), w.lineTop(pn.H)+textTop(AppFont, TextLineHeight), placeholder, AppFont, Gray7)
			}
			top := 0
			if w.multi {
				top = int(st.top)
			}
			show.drawLines(pn, w.textLeft(), w.lineTop(pn.H), TextLineHeight, top, st.scrollX, ink, sel, focused && blinkOn && !disabled)
		})
		if focused && !disabled && !w.doc {
			DrawFocusRing(pn, 0, th.Accent.Dark)
		}
	})
	if w.multi {
		total := float64(len(show.lines))
		w.el.Children(func() {
			ui.Box(w.c).Grow(1).PassThrough()
			ui.Column(w.c).Width(ScrollBarWidth).Children(func() {
				ScrollBar(w.c, &st.top, float64(lines), total).Step(1).Element().Grow(1)
			})
		})
	}
	report(w.el, w.changed || w.submitted)
	return w
}

func (w *TextFieldWidget) textWidth(width int) int {
	inner := width - 2*textInset - 2
	if w.doc {
		inner = width - w.docLeft - 4
		if w.docWrap > 0 {
			inner = min(inner, w.docWrap+2)
		}
	}
	if w.search {
		inner -= searchLeft + searchRight
	}
	return inner
}

func (w *TextFieldWidget) keepLineCaretVisible(ed *editor, width int) {
	x := ed.xOf(ed.caret)
	if x-w.st.scrollX > width-2 {
		w.st.scrollX = x - width + 2
	} else if x < w.st.scrollX {
		w.st.scrollX = x
	}
	w.st.scrollX = max(0, min(w.st.scrollX, max(0, ed.font.Measure(string(ed.text))-width+2)))
}

// textLeft returns where the text starts in the field.
func (w *TextFieldWidget) textLeft() int {
	if w.doc {
		return w.docLeft
	}
	if w.search {
		return textInset + 1 + searchLeft
	}
	return textInset + 1
}

// lineTop returns the top of the first line in a box h tall: a text
// field centers its capitals.
func (w *TextFieldWidget) lineTop(h int) int {
	if w.multi {
		return textInset
	}
	return (h-AppFont.CapHeight())/2 - textTop(AppFont, TextLineHeight) - (AppFont.Ascent() - AppFont.CapHeight())
}

// linesInView returns how many lines from line top a text area shows
// whole, at least one.
func (w *TextFieldWidget) linesInView(ed *editor, top int) int {
	st := w.st
	if !w.multi || st.rect.H == 0 {
		return 1
	}
	room := int(st.rect.H) - 2*textInset
	n := 0
	for top+n < len(ed.lines) {
		room -= ed.lineHeight(top+n, TextLineHeight)
		if room < 0 {
			break
		}
		n++
	}
	// Past the last line, lines of the usual height.
	return max(n+max(room, 0)/TextLineHeight, 1)
}

// keepCaretVisible scrolls a text area to its caret.
func (w *TextFieldWidget) keepCaretVisible() {
	st := w.st
	if !w.multi || st.rect.H == 0 {
		return
	}
	ed := st.ed.displayEditor()
	ed.layout()
	n := ed.lineOf(ed.caret)
	if float64(n) < st.top {
		st.top = float64(n)
	}
	for n >= int(st.top)+w.linesInView(ed, int(st.top)) && int(st.top) < n {
		st.top++
	}
}

func (w *TextFieldWidget) input(ev ui.InputEvent) bool {
	st := w.st
	ed := st.ed
	if w.disabled {
		return false
	}
	took, changed := false, false
	switch ev.Kind {
	case ui.InputKeyDown:
		if ev.Key == ui.KeyEnter && !w.multi {
			st.submitted = true
			return false // a default button may take it too
		}
		if ev.Key == ui.KeyEscape && w.search && len(ed.text) > 0 {
			ed.anchor, ed.caret = 0, len(ed.text)
			changed = ed.replace("", editOther)
			took = true
			break
		}
		took, changed = ed.key(ev, w.svc)
	case ui.InputText:
		ed.compose = ""
		ed.composeCaret = 0
		changed = ed.replace(ev.Text, editTyping)
		took = true
	case ui.InputCompose:
		ed.compose = ev.Text
		ed.composeCaret = min(max(ev.Caret, 0), len([]rune(ev.Text)))
		took = true
	case ui.InputCommand:
		if w.password && (ev.Text == "copy" || ev.Text == "cut") {
			return true
		}
		took, changed = ed.command(ev.Text, w.svc)
	case ui.InputPointerDown:
		if ev.Button != 0 {
			return false
		}
		if w.search && len(ed.text) > 0 && ev.X >= st.rect.W-searchRight-4 {
			// The clear button.
			ed.anchor, ed.caret = 0, len(ed.text)
			changed = ed.replace("", editOther)
			*w.value = ed.String()
			st.changed = true
			ed.lines, ed.laidFor = nil, ""
			return false
		}
		i := w.offsetAt(ev.X, ev.Y)
		switch {
		case ev.Clicks >= 3:
			ed.anchor, ed.caret = 0, len(ed.text)
		case ev.Clicks == 2:
			ed.selectWord(i)
		default:
			ed.moveTo(i, ev.Mods&ui.Shift != 0)
		}
		ed.dragging = true
		took = false // and take the focus
	case ui.InputPointerMove:
		if ed.dragging && ev.Button == 0 {
			ed.moveTo(w.offsetAt(ev.X, ev.Y), true)
			took = true
		}
	case ui.InputPointerUp:
		ed.dragging = false
	case ui.InputScroll:
		if w.multi {
			st.top = min(max(st.top+float64(ev.DY)/TextLineHeight, 0), float64(max(len(ed.lines)-1, 0)))
			return true
		}
	}
	st.activity = eventNow(ev)
	if changed {
		*w.value = ed.String()
		st.changed = true
		ed.lines = nil
		ed.laidFor = ""
	}
	if ev.Kind != ui.InputScroll {
		w.keepCaretVisible()
	}
	return took
}

func (w *TextFieldWidget) offsetAt(x, y float32) int {
	ed := w.st.ed
	ed.layout()
	n := 0
	if w.multi {
		n = ed.lineAtY(int(w.st.top), int(y)-textInset, TextLineHeight)
	}
	return ed.offsetAt(n, int(x)-w.textLeft()+w.st.scrollX)
}

// Document makes a text area the content of a document window: white to
// the window's edges with its scroll bar against the window's right side,
// no ring or outline of its own, and no focus ring — the window shows where
// typing goes. Its text starts left pixels in and wraps at width (0: at the
// area's width), as a ruler's margins set them.
func (w *TextFieldWidget) Document(left, width int) *TextFieldWidget {
	w.doc, w.docLeft, w.docWrap = true, left, width
	w.el.Padding(0)
	return w
}

// Indent starts the first line of each paragraph of a text area first
// pixels further in, as a ruler's first-line indent sets it.
func (w *TextFieldWidget) Indent(first int) *TextFieldWidget {
	if ed := w.st.ed; ed.indent != max(first, 0) {
		ed.indent = max(first, 0)
		ed.lines = nil
	}
	return w
}

// Width sets the field's width.
func (w *TextFieldWidget) Width(px int) *TextFieldWidget {
	w.el.Width(float32(px))
	return w
}

// Placeholder shows text in gray while the field is empty.
func (w *TextFieldWidget) Placeholder(s string) *TextFieldWidget {
	w.placeholder = s
	return w
}

// Password shows the field's text as bullets, keeps it from being copied
// or cut by any means, and gives assistive technology bullets in place of
// the text.
func (w *TextFieldWidget) Password() *TextFieldWidget {
	w.password = true
	w.st.ed.noCopy = true
	w.el.Value(strings.Repeat("\u2022", len(w.st.ed.text)))
	return w
}

// MaxLength keeps the text to n characters at most.
func (w *TextFieldWidget) MaxLength(n int) *TextFieldWidget {
	w.st.ed.maxLen = n
	return w
}

// Label names the field for assistive technology and tests.
func (w *TextFieldWidget) Label(s string) *TextFieldWidget {
	w.el.Label(s)
	return w
}

// Disabled dims the field and makes it take no input while d is true.
func (w *TextFieldWidget) Disabled(d bool) *TextFieldWidget {
	w.disabled = d
	w.st.ed.readOnly = d
	w.el.Disabled(d)
	return w
}

// Changed reports whether the user changed the text since the last frame.
func (w *TextFieldWidget) Changed() bool { return report(w.el, w.changed) }

// Submitted reports whether the user pressed Return in a text field since
// the last frame.
func (w *TextFieldWidget) Submitted() bool { return report(w.el, w.submitted) }

// Element returns the field's element.
func (w *TextFieldWidget) Element() ui.Element { return w.el }

// DrawContentRing draws the sunken ring around a content box — a text
// field, list, tree, table or scroll view — just inside pn's box: Gray7
// along the top and left, White along the bottom and right. The content's
// black outline goes inside it.
func DrawContentRing(pn Pen) {
	pn.Bevel(0, 0, pn.W, pn.H, Gray7, White)
}

// DrawTextFrame draws the sunken white box of a text field or text area
// filling pn's box.
func DrawTextFrame(pn Pen, disabled bool) {
	w, h := pn.W, pn.H
	face, frame := White, Black
	if disabled {
		face, frame = Gray2, Gray6
	}
	pn.Bevel(0, 0, w, h, Gray7, White)
	pn.Frame(1, 1, w-2, h-2, frame)
	pn.Rect(2, 2, w-4, h-4, face)
}

const (
	searchLeft  = 14 // the room of a search field's magnifier
	searchRight = 14 // and of its clear button
)

// SearchField builds a text field for searching: a magnifier at its left,
// and a button at its right, once there is text, that clears it, as Escape
// does.
func SearchField(c *ui.Context, value *string) *TextFieldWidget {
	w := newTextWidget(c, value, false)
	w.search = true
	w.el.Role(ui.RoleTextField)
	if w.placeholder == "" {
		w.placeholder = say(c, "search")
	}
	return w
}

var (
	magnifier = []string{
		".###..",
		"#...#.",
		"#...#.",
		"#...#.",
		".####.",
		"....##",
		".....#",
	}
	clearMark = []string{
		"..###..",
		".#####.",
		"##.#.##",
		"###.###",
		"##.#.##",
		".#####.",
		"..###..",
	}
)

// DrawSearchDecor draws the magnifier and, when clear, the clear button of
// a search field filling pn's box.
func DrawSearchDecor(pn Pen, clear, disabled bool) {
	ink := Gray9
	if disabled {
		ink = Gray6
	}
	pn.Bitmap(textInset+1, (pn.H-7)/2, magnifier, ink)
	if clear && !disabled {
		pn.Bitmap(pn.W-textInset-2-7, (pn.H-7)/2, clearMark, Gray7)
	}
}

// Select selects the runes from a to b of the text field or area, as a
// match found, and scrolls them into view.
func (w *TextFieldWidget) Select(a, b int) *TextFieldWidget {
	ed := w.st.ed
	n := len(ed.text)
	ed.anchor, ed.caret = min(max(a, 0), n), min(max(b, 0), n)
	w.keepCaretVisible()
	return w
}

// Selection returns the runes selected in the text field or area, from a
// to b; a equals b at the caret.
func (w *TextFieldWidget) Selection() (int, int) { return w.st.ed.selection() }
