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
// keyboard focus. *pos is the first pane's width, kept by the app; with
// less room than that and the second pane's least, the first gives way,
// and *pos stays for when there is room again. Give the split a size, or
// let it grow.
func Split(c *ui.Context, pos *float32, first, second func()) ui.Element {
	return split(c, pos, false, false, first, second)
}

// SplitVertical builds a Split whose panes are stacked, *pos being the
// first one's height.
func SplitVertical(c *ui.Context, pos *float32, first, second func()) ui.Element {
	return split(c, pos, true, false, first, second)
}

// SplitPercent builds a Split with *p the percentage of the room the first
// pane takes, from 0 to 100, so that both panes grow and shrink with the
// split, in proportion.
func SplitPercent(c *ui.Context, p *float32, first, second func()) ui.Element {
	return split(c, p, false, true, first, second)
}

// SplitVerticalPercent builds a SplitPercent whose panes are stacked.
func SplitVerticalPercent(c *ui.Context, p *float32, first, second func()) ui.Element {
	return split(c, p, true, true, first, second)
}

func split(c *ui.Context, pos *float32, vertical, percent bool, first, second func()) ui.Element {
	th := ThemeOf(c)
	el := ui.Row(c).AlignItems(ui.Stretch)
	if vertical {
		el = ui.Column(c).AlignItems(ui.Stretch)
	}
	st := ui.Local(el, "split", func() splitState { return splitState{} })
	length := func(e ui.Element, v float32) {
		if vertical {
			e.Height(v)
		} else {
			e.Width(v)
		}
	}
	least := func(e ui.Element, v float32) {
		if vertical {
			e.MinHeight(v)
		} else {
			e.MinWidth(v)
		}
	}
	// The room the panes share, as last painted; none before.
	room := st.size - SplitterWidth
	fit := func(v float32) float32 {
		if room > 0 {
			v = min(v, room-SplitMin)
		}
		return max(v, SplitMin)
	}
	share := min(max(*pos, 0), 100) / 100
	// shown is the first pane's length as last laid out: the divider
	// moves from there, not from *pos, which a small split does not show.
	shown := fit(*pos)
	if percent {
		shown = fit(share * room)
	}
	set := func(v float32) {
		v = fit(v)
		if percent {
			if room <= 0 {
				return
			}
			v = v * 100 / room
		}
		*pos = v
	}
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
		one := ui.Column(c).AlignItems(ui.Stretch).Clip().Shrink(1)
		if percent {
			one.Basis(0).Grow(share)
			least(one, SplitMin)
		} else {
			length(one, *pos)
			least(one, min(*pos, SplitMin))
		}
		one.Children(first)
		div := ui.Box(c).Focusable().FocusRing(false).Shrink(0).Role(ui.RoleSplitter).Label(say(c, "divider"))
		if vertical {
			div.Height(SplitterWidth).Cursor(ui.CursorResizeRow)
		} else {
			div.Width(SplitterWidth).Cursor(ui.CursorResizeColumn)
		}
		div.Range(SplitMin, float64(max(room-SplitMin, SplitMin)), float64(shown))
		back, fwd := ui.KeyLeft, ui.KeyRight
		if vertical {
			back, fwd = ui.KeyUp, ui.KeyDown
		}
		if div.Shortcut(0, back) {
			set(shown - splitStep)
		}
		if div.Shortcut(0, fwd) {
			set(shown + splitStep)
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
					// at is from where the divider was last laid out, at
					// shown.
					set(shown + at - st.grab)
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
		two := ui.Column(c).Basis(0).Grow(1).Shrink(1).AlignItems(ui.Stretch).Clip()
		if percent {
			two.Grow(1 - share)
		}
		least(two, SplitMin)
		two.Children(second)
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
