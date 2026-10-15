package platinum

import "github.com/egoist/mygo/ui"

// RichText is a document of styled text, kept by the app: its text, the
// style of each of its runes, and the alignment of each of its paragraphs.
// A RichTextArea edits it, and a FormatBar and a SizeMenu style it.
type RichText struct {
	// Text is the document's text. Setting it from outside keeps the
	// styles of the runes at the same places.
	Text string
	ed   *editor
}

// NewRichText returns a document of text in the plain style.
func NewRichText(text string) *RichText {
	d := &RichText{Text: text, ed: &editor{font: AppFont, multi: true, rich: true}}
	d.ed.set(text)
	return d
}

func (d *RichText) editor() *editor {
	if d.ed == nil {
		d.ed = &editor{font: AppFont, multi: true, rich: true}
		d.ed.set(d.Text)
	}
	return d.ed
}

// Run is a stretch of a document in one style.
type Run struct {
	Text  string
	Style TextStyle
}

// Runs returns the document as runs of one style each, as for saving it.
func (d *RichText) Runs() []Run {
	ed := d.editor()
	var runs []Run
	for i := 0; i < len(ed.text); {
		j := i
		for j < len(ed.text) && ed.styles[j] == ed.styles[i] {
			j++
		}
		runs = append(runs, Run{string(ed.text[i:j]), ed.styles[i]})
		i = j
	}
	return runs
}

// SetRuns replaces the document with runs.
func (d *RichText) SetRuns(runs []Run) {
	ed := d.editor()
	ed.text, ed.styles = nil, nil
	for _, r := range runs {
		for _, ch := range r.Text {
			ed.text = append(ed.text, ch)
			ed.styles = append(ed.styles, r.Style)
		}
	}
	ed.paras = nil
	ed.caret, ed.anchor = 0, 0
	ed.undo, ed.redo, ed.lines = nil, nil, nil
	d.Text = string(ed.text)
}

// Style returns the style the format bar shows: the styles all of the
// selection has, at the scale Size reports (1 where it is mixed), or, with
// none, the style typing takes — that of the rune before the caret, unless
// the user chose another since the caret moved.
func (d *RichText) Style() TextStyle {
	ed := d.editor()
	a, b := ed.selection()
	if a == b {
		return ed.typing
	}
	st := ed.styles[a]
	for i := a; i < b; i++ {
		st &= ed.styles[i] | scaleBits
	}
	return st.Scaled(d.Size())
}

// Size returns the scale of the selection, where all of it has one, or 0;
// with no selection, the scale typing takes. See TextStyle.Scale.
func (d *RichText) Size() int {
	ed := d.editor()
	a, b := ed.selection()
	if a == b {
		return ed.typing.Scale()
	}
	k := ed.styles[a].Scale()
	for i := a; i < b; i++ {
		if ed.styles[i].Scale() != k {
			return 0
		}
	}
	return k
}

// SetSize scales the selection, or, with none, what is typed next, to k,
// from 1 to MaxScale.
func (d *RichText) SetSize(k int) {
	ed := d.editor()
	a, b := ed.selection()
	if a == b {
		ed.typing = ed.typing.Scaled(k)
		return
	}
	ed.snapshot(editOther)
	ed.styles = append([]TextStyle(nil), ed.styles...)
	for i := a; i < b; i++ {
		ed.styles[i] = ed.styles[i].Scaled(k)
	}
	ed.lines = nil
}

// Toggle turns style s on or off: over the selection — on, unless all of
// it has it already — or, with none, for what is typed next.
func (d *RichText) Toggle(s TextStyle) {
	ed := d.editor()
	a, b := ed.selection()
	if a == b {
		ed.typing ^= s
		return
	}
	on := d.Style()&s == 0
	ed.snapshot(editOther)
	ed.styles = append([]TextStyle(nil), ed.styles...)
	for i := a; i < b; i++ {
		if on {
			ed.styles[i] |= s
		} else {
			ed.styles[i] &^= s
		}
	}
	ed.lines = nil // bold runs are wider
}

// Align returns the alignment the format bar shows: of the paragraph
// holding the caret, or the start of the selection.
func (d *RichText) Align() Align {
	ed := d.editor()
	a, _ := ed.selection()
	return ed.paraAlign(ed.paraOf(a))
}

// SetAlign aligns the paragraphs the selection touches, or the one holding
// the caret, as one edit to undo.
func (d *RichText) SetAlign(align Align) {
	ed := d.editor()
	a, b := ed.selection()
	ed.snapshot(editOther)
	ed.paras = append([]Align(nil), ed.paras...)
	ed.setAlign(ed.paraOf(a), ed.paraOf(b), align)
}

// Aligns returns the alignment of each of the document's paragraphs, the
// lines between newlines, as for saving it with its runs.
func (d *RichText) Aligns() []Align {
	ed := d.editor()
	aligns := make([]Align, ed.paraOf(len(ed.text))+1)
	for p := range aligns {
		aligns[p] = ed.paraAlign(p)
	}
	return aligns
}

// SetAligns aligns the document's paragraphs, from the first; those it
// leaves out are aligned left.
func (d *RichText) SetAligns(aligns []Align) {
	ed := d.editor()
	ed.paras = append([]Align(nil), aligns...)
}

// RichTextArea builds a box editing a document of styled text, as TextArea
// edits plain text: what is typed takes the style the format bar shows,
// and the lines wrap, aligned, at the box's width. Command-B, -I and -U
// toggle bold, italic and underline; Command-Shift-. and -, make the text
// a size larger and smaller. Give it a size.
func RichTextArea(c *ui.Context, d *RichText) *TextFieldWidget {
	ed := d.editor()
	if d.Text != ed.String() {
		ed.set(d.Text)
	}
	w := newTextWidgetWith(c, &d.Text, true, ed)
	for _, k := range []struct {
		key ui.Key
		s   TextStyle
	}{{ui.KeyB, Bold}, {ui.KeyI, Italic}, {ui.KeyU, Underline}} {
		if w.el.Shortcut(ui.Cmd, k.key) {
			d.Toggle(k.s)
		}
	}
	if w.el.Shortcut(ui.Cmd|ui.Shift, ui.KeyPeriod) {
		d.SetSize(max(d.Size(), 1) + 1)
	}
	if w.el.Shortcut(ui.Cmd|ui.Shift, ui.KeyComma) {
		d.SetSize(max(d.Size(), 2) - 1)
	}
	return w
}

// SizeMenu returns a Size menu for a document, to add to the menu bar: an
// item for each scale, named by its size in points (the App face is 11),
// the selection's checked.
func SizeMenu(d *RichText) Menu {
	m := Menu{Title: sayApp("size")}
	size := d.Size()
	for k := 1; k <= MaxScale; k++ {
		m.Items = append(m.Items, MenuItem{
			Title:   sayAppWith("points", map[string]any{"size": 11 * k}),
			Checked: size == k,
			Action:  func() { d.SetSize(k) },
		})
	}
	return m
}

// The 16×16 icons of the format bar, as rows for DrawIcon, for toolbars
// of one's own.
var (
	IconBold = []string{
		"................",
		"................",
		"..#########.....",
		"...###...###....",
		"...###....###...",
		"...###....###...",
		"...###...###....",
		"...########.....",
		"...###....###...",
		"...###.....###..",
		"...###.....###..",
		"...###....###...",
		"..##########....",
		"................",
		"................",
		"................",
	}
	IconItalic = []string{
		"................",
		"................",
		"......#######...",
		".........##.....",
		".........##.....",
		"........##......",
		"........##......",
		".......##.......",
		".......##.......",
		"......##........",
		"......##........",
		".....##.........",
		"...#######......",
		"................",
		"................",
		"................",
	}
	IconUnderline = []string{
		"................",
		"................",
		"..####....####..",
		"...##......##...",
		"...##......##...",
		"...##......##...",
		"...##......##...",
		"...##......##...",
		"...##......##...",
		"....##....##....",
		".....######.....",
		"................",
		"..############..",
		"................",
		"................",
		"................",
	}
	IconAlign = [3][]string{
		{"................", "................", "..############..", "................", "..########......", "................", "..############..", "................", "..########......", "................", "..############..", "................", "..########......", "................", "................", "................"},
		{"................", "................", "..############..", "................", "....########....", "................", "..############..", "................", "....########....", "................", "..############..", "................", "....########....", "................", "................", "................"},
		{"................", "................", "..############..", "................", "......########..", "................", "..############..", "................", "......########..", "................", "..############..", "................", "......########..", "................", "................", "................"},
	}
)

// FormatBar builds a toolbar styling d: bold, italic and underline toggles
// showing the style of the selection, and the alignment of its lines.
func FormatBar(c *ui.Context, d *RichText) ui.Element {
	return Toolbar(c, FormatItems(d))
}

// FormatItems returns the items of a format bar, to put in a toolbar of
// one's own.
func FormatItems(d *RichText) []ToolItem {
	st := d.Style()
	bold, italic, under := st&Bold != 0, st&Italic != 0, st&Underline != 0
	align := d.Align()
	left, center, right := align == AlignLeft, align == AlignCenter, align == AlignRight
	return []ToolItem{
		{Label: sayApp("bold"), Icon: IconBold, IconOnly: true, On: &bold, Action: func() { d.Toggle(Bold) }},
		{Label: sayApp("italic"), Icon: IconItalic, IconOnly: true, On: &italic, Action: func() { d.Toggle(Italic) }},
		{Label: sayApp("underline"), Icon: IconUnderline, IconOnly: true, On: &under, Action: func() { d.Toggle(Underline) }},
		{Separator: true},
		{Label: sayApp("alignLeft"), Icon: IconAlign[0], IconOnly: true, On: &left, Action: func() { d.SetAlign(AlignLeft) }},
		{Label: sayApp("alignCenter"), Icon: IconAlign[1], IconOnly: true, On: &center, Action: func() { d.SetAlign(AlignCenter) }},
		{Label: sayApp("alignRight"), Icon: IconAlign[2], IconOnly: true, On: &right, Action: func() { d.SetAlign(AlignRight) }},
	}
}
