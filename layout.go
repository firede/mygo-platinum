package platinum

import (
	"github.com/egoist/mygo/ui"
)

// Splitter metrics, from docs/spec.md.
const (
	SplitterWidth = 7  // the width of a splitter's divider
	SplitMin      = 40 // the least either pane keeps
	splitStep     = 8  // how far the arrows move the divider
)

type splitState struct {
	size     float32 // the split's length along its axis, as last painted
	dragging bool
	grab     float32
}

// Split builds two panes side by side, first and second, with a divider
// between them that the pointer drags, or the arrows move while it has the
// keyboard focus. *pos is the first pane's width, kept by the app. Give the
// split a size, or let it grow.
func Split(c *ui.Context, pos *float32, first, second func()) ui.Element {
	return split(c, pos, false, first, second)
}

// SplitVertical builds a Split whose panes are stacked, *pos being the
// first one's height.
func SplitVertical(c *ui.Context, pos *float32, first, second func()) ui.Element {
	return split(c, pos, true, first, second)
}

func split(c *ui.Context, pos *float32, vertical bool, first, second func()) ui.Element {
	th := ThemeOf(c)
	el := ui.Row(c).AlignItems(ui.Stretch)
	if vertical {
		el = ui.Column(c).AlignItems(ui.Stretch)
	}
	st := ui.Local(el, "split", func() splitState { return splitState{} })
	clamp := func(v float32) float32 {
		if st.size <= 0 {
			return max(v, SplitMin)
		}
		return min(max(v, SplitMin), st.size-SplitterWidth-SplitMin)
	}
	*pos = clamp(*pos)
	el.Draw(func(p *ui.Painter, r ui.Rect) {
		size := r.W
		if vertical {
			size = r.H
		}
		if size != st.size {
			c.Invalidate()
		}
		st.size = size
	})
	el.Children(func() {
		one := ui.Column(c).AlignItems(ui.Stretch).Clip()
		if vertical {
			one.Height(*pos)
		} else {
			one.Width(*pos)
		}
		one.Children(first)
		div := ui.Box(c).Focusable().FocusRing(false).Role(ui.RoleSplitter).Label(say(c, "divider"))
		if vertical {
			div.Height(SplitterWidth).Cursor(ui.CursorResizeRow)
		} else {
			div.Width(SplitterWidth).Cursor(ui.CursorResizeColumn)
		}
		div.Range(SplitMin, float64(st.size-SplitterWidth-SplitMin), float64(*pos))
		back, fwd := ui.KeyLeft, ui.KeyRight
		if vertical {
			back, fwd = ui.KeyUp, ui.KeyDown
		}
		if div.Shortcut(0, back) {
			*pos = clamp(*pos - splitStep)
		}
		if div.Shortcut(0, fwd) {
			*pos = clamp(*pos + splitStep)
		}
		div.HandleInput(func(ev ui.InputEvent) bool {
			at := ev.X
			if vertical {
				at = ev.Y
			}
			switch ev.Kind {
			case ui.InputPointerDown:
				st.dragging, st.grab = true, at
				return false // and take the focus
			case ui.InputPointerMove:
				if st.dragging && ev.Button == 0 {
					*pos = clamp(*pos + at - st.grab)
					return true
				}
			case ui.InputPointerUp:
				st.dragging = false
			}
			return false
		})
		focused := div.FocusVisible()
		div.Draw(func(p *ui.Painter, r ui.Rect) {
			pn := NewPen(p, r)
			DrawSplitter(pn, vertical)
			if focused {
				DrawFocusRing(pn, 0, th.Accent.Dark)
			}
		})
		ui.Column(c).Grow(1).AlignItems(ui.Stretch).Clip().Children(second)
	})
	return el
}

// DrawSplitter draws the divider of a split filling pn's box: a raised bar
// with a grip of dots in its middle, across a vertical split if vertical.
func DrawSplitter(pn Pen, vertical bool) {
	w, h := pn.W, pn.H
	pn.Rect(0, 0, w, h, Gray2)
	if vertical {
		pn.HLine(0, 0, w, Black)
		pn.HLine(0, 1, w, White)
		pn.HLine(0, h-2, w, Gray7)
		pn.HLine(0, h-1, w, Black)
		for i := -2; i <= 2; i++ {
			x := w/2 + 4*i
			pn.Dot(x, h/2-1, Gray7)
			pn.Dot(x+1, h/2, White)
		}
		return
	}
	pn.VLine(0, 0, h, Black)
	pn.VLine(1, 0, h, White)
	pn.VLine(w-2, 0, h, Gray7)
	pn.VLine(w-1, 0, h, Black)
	for i := -2; i <= 2; i++ {
		y := h/2 + 4*i
		pn.Dot(w/2-1, y, Gray7)
		pn.Dot(w/2, y+1, White)
	}
}

// ScrollView builds a box that scrolls what fn builds up and down, with a
// Platinum scroll bar at its right that follows it, kept in *st by the
// app. Give it a size, or let it grow.
//
// The counterpart among MyGo's own controls is ui.Scroll.
func ScrollView(c *ui.Context, st *ui.ScrollState, fn func()) ui.Element {
	el := ui.Row(c).AlignItems(ui.Stretch).Padding(1)
	view := ui.Local(el, "scrollview", func() splitState { return splitState{} })
	el.Draw(func(p *ui.Painter, r ui.Rect) {
		pn := NewPen(p, r)
		DrawContentRing(pn)
		pn.Rect(2, 2, pn.W-4, pn.H-4, White)
		pn.Frame(1, 1, pn.W-2, pn.H-2, Black)
		if view.size != r.H-4 {
			c.Invalidate()
		}
		view.size = r.H - 4
	})
	el.Children(func() {
		ui.Scroll(c).Grow(1).Margin(1, 0, 1, 1).TrackScroll(st).Children(func() {
			ui.Column(c).Padding(4, 6).Gap(2).Children(fn)
		})
		visible := float64(view.size)
		total := float64(st.MaxY) + visible
		v := float64(st.Y)
		ui.Column(c).Width(ScrollBarWidth).Children(func() {
			bar := ScrollBar(c, &v, visible, total)
			bar.Element().Grow(1)
			if bar.Changed() {
				st.Y = float32(v)
			}
		})
	})
	return el
}
