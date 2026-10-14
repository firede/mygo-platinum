package platinum

import (
	"github.com/egoist/mygo/ui"
)

// Group box metrics, from docs/spec.md.
const (
	GroupPadding = 12 // the room inside a group box around its content
	groupTitleX  = 8
)

// GroupBox builds a box that groups related controls under title: an
// etched outline whose top edge the title interrupts. fn builds the
// controls inside, in a column. An empty title draws the outline alone.
func GroupBox(c *ui.Context, title string, fn func()) ui.Element {
	top := GroupPadding
	if title != "" {
		top = SystemFont.Height() + 8
	}
	el := ui.Column(c).Padding(float32(top), GroupPadding, GroupPadding, GroupPadding).Gap(6)
	if title != "" {
		el.Role(ui.RoleGroup).Label(title)
	}
	dim := inactive(c) // in a window in the background, its title gray
	el.Draw(func(p *ui.Painter, r ui.Rect) {
		DrawGroupBox(NewPen(p, r), title, dim)
	})
	return el.Children(fn)
}

// DrawGroupBox draws the outline of a group box filling pn's box, its title
// on its top edge, dimmed when disabled.
func DrawGroupBox(pn Pen, title string, disabled bool) {
	y := 0
	if title != "" {
		y = SystemFont.CapHeight() / 2
	}
	w, h := pn.W, pn.H-y
	pn.Frame(0, y, w-1, h-1, Gray7)
	pn.HLine(1, y+1, w-3, White)
	pn.VLine(1, y+2, h-4, White)
	pn.HLine(0, y+h-1, w, White)
	pn.VLine(w-1, y, h, White)
	if title != "" {
		tw := SystemFont.Measure(title)
		pn.Rect(groupTitleX, y, tw+6, 2, Gray2)
		c := Black
		if disabled {
			c = Gray8
		}
		pn.Text(groupTitleX+3, 0, title, SystemFont, c)
	}
}

// Separator builds a horizontal etched line, for dividing a column of
// controls. Give it a width, or let it stretch across its column.
func Separator(c *ui.Context) ui.Element {
	return ui.Box(c).Height(2).AlignSelf(ui.Stretch).Role(ui.RoleNone).Draw(func(p *ui.Painter, r ui.Rect) {
		pn := NewPen(p, r)
		pn.HLine(0, 0, pn.W, Gray7)
		pn.HLine(0, 1, pn.W, White)
	})
}

// Card builds a raised panel that groups what fn builds, in a column: a
// black outline round a raised face, as a panel of a control panel.
func Card(c *ui.Context, fn func()) ui.Element {
	return ui.Column(c).Padding(10).Gap(6).Draw(func(p *ui.Painter, r ui.Rect) {
		pn := NewPen(p, r)
		pn.Rect(0, 0, pn.W, pn.H, Gray2)
		pn.Frame(0, 0, pn.W, pn.H, Black)
		pn.Bevel(1, 1, pn.W-2, pn.H-2, White, Gray7)
	}).Children(fn)
}
