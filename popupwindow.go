package platinum

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Pop-up window metrics, from docs/spec.md.
const (
	PopupTabHeight = 20 // the tab of a pop-up window, its outline included
	popupTabInset  = 12 // from the window's left side to its tab
	popupTabPad    = 16 // from the tab's sides to its title
	popupTabMinW   = 80
	popupTearOff   = 24 // how far up dragging the tab tears the window off
	popupDragSlop  = 3  // how far the tab moves before a press is a drag
)

// PopupWindowState is where a pop-up window is and whether it is open,
// which the app keeps.
type PopupWindowState struct {
	// X places the window's left side in its parent; W is its width and
	// H its height when open, under its tab.
	X, W, H int
	// Open shows the window over the bottom of its parent, its tab on top;
	// closed, the tab alone sits on the bottom edge.
	Open bool
}

type popupDrag struct {
	down     bool
	moved    bool
	grabX    float32
	grabY    float32
	originX  int
	rect     ui.Rect // the window, as last painted
	parentX  float32
	dragged  bool // the last press dragged the tab: its click does not count
	tornOff  bool
	activate bool
}

// PopupWindowWidget is a pop-up window being built.
type PopupWindowWidget struct {
	el       ui.Element
	tab      ui.Element
	closed   bool
	tornOff  bool
	activate bool
}

// PopupWindow builds a window docked at the bottom of its parent, as the
// classic desktop's pop-up windows were: a tab showing title stands on the window's
// top edge, and the window rises from the bottom of the parent while
// st.Open, or hides, leaving the tab on the bottom edge. A click on the
// tab, or Space or Enter while it has the keyboard focus, opens and closes
// the window; dragging the tab sideways moves it; dragging it up tears the
// window off, which TornOff reports for the app to show it as a Window
// instead. Command-W closes it. Of opts, Active, Interactive and Dragging
// mean what they do for Window. Build it in a parent that positions
// absolutely, as Window.
func PopupWindow(c *ui.Context, st *PopupWindowState, title string, opts WindowOptions, fn func()) *PopupWindowWidget {
	w := &PopupWindowWidget{}
	active := opts.Active
	background := !opts.Active && !opts.Interactive // its content inactive
	st.W = max(st.W, popupTabMinW+2*popupTabInset)
	st.H = max(st.H, 2*FrameWidth+20)
	tabW := min(max(SystemFont.Measure(title)+2*popupTabPad, popupTabMinW), st.W-popupTabInset-4)
	h := PopupTabHeight
	if st.Open {
		h += st.H - 1
	}
	w.el = ui.Column(c).Absolute().Left(float32(st.X)).Bottom(0).Width(float32(st.W)).Height(float32(h)).
		Role(ui.RoleDialog).Label(title)
	d := ui.Local(w.el, "popupwindow", func() popupDrag { return popupDrag{} })
	w.tornOff, d.tornOff = d.tornOff, false
	w.activate, d.activate = d.activate, false
	if w.el.Shortcut(ui.Cmd, ui.KeyW) {
		w.closed = true
	}
	pressed := d.down && !d.moved
	w.el.Draw(func(p *ui.Painter, r ui.Rect) {
		resized(c, d.rect, r)
		d.rect = r
		d.parentX = r.X - float32(st.X)
		DrawPopupWindow(NewPen(p, r), title, tabW, st.Open, active, pressed)
	})
	w.el.AlignItems(ui.Stretch).Children(func() {
		// The tab, a button for the keyboard and assistive technologies;
		// the pointer drags it as well.
		ui.Row(c).Height(PopupTabHeight).Padding(0, 0, 0, popupTabInset).Children(func() {
			tab := ui.ButtonBase(c).FocusRing(false).Role(ui.RoleButton).Label(title).Expanded(st.Open).
				Size(float32(tabW), PopupTabHeight)
			w.tab = tab
			if tab.Clicked() && !d.dragged {
				st.Open = !st.Open
			}
			if !d.down {
				d.dragged = false
			}
			focused := tab.FocusVisible()
			ring := ThemeOf(c).Accent.Dark
			tab.Draw(func(p *ui.Painter, r ui.Rect) {
				if focused {
					pn := NewPen(p, r)
					pn.Rect(popupTabPad-4, pn.H-5, pn.W-2*popupTabPad+8, 2, ring)
				}
			})
			tab.HandleInput(func(ev ui.InputEvent) bool {
				return w.tabInput(ev, d, st)
			})
		})
		if !st.Open {
			return
		}
		inner := ui.Column(c).Grow(1).Padding(FrameWidth-1, FrameWidth, 0, FrameWidth).AlignItems(ui.Stretch).Clip()
		inner.Children(func() {
			content := ui.Column(c).Grow(1).AlignItems(ui.Stretch).Clip()
			if background {
				enterInactive(c)
			}
			content.Children(fn)
			if background {
				leaveInactive(c)
			}
			if background && !opts.Dragging {
				// The first click in a window in the background brings it
				// to the front, as for Window.
				shield := ui.Box(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).Role(ui.RoleNone)
				shield.HandleInput(func(ev ui.InputEvent) bool {
					if ev.Kind == ui.InputPointerDown {
						d.activate = true
						return true
					}
					return ev.Kind == ui.InputPointerUp || ev.Kind == ui.InputPointerMove && ev.Button >= 0
				})
			}
		})
	})
	report(w.el, w.closed || w.tornOff || w.activate)
	return w
}

func (w *PopupWindowWidget) tabInput(ev ui.InputEvent, d *popupDrag, st *PopupWindowState) bool {
	// The pointer in the parent's coordinates, as in Window.input.
	px := d.rect.X - d.parentX + ev.X
	switch ev.Kind {
	case ui.InputPointerDown:
		if ev.Button != 0 {
			return false
		}
		d.activate = true
		d.down, d.moved = true, false
		d.grabX, d.grabY = px, ev.Y
		d.originX = st.X
		return true
	case ui.InputPointerMove:
		if !d.down || ev.Button != 0 {
			return false
		}
		dx := px - d.grabX
		if dx > popupDragSlop || dx < -popupDragSlop || d.grabY-ev.Y > popupDragSlop {
			d.moved = true
		}
		if d.moved {
			st.X = max(d.originX+round(dx), 0)
		}
		return true
	case ui.InputPointerUp:
		if !d.down {
			return false
		}
		d.tornOff = d.grabY-ev.Y > popupTearOff
		d.dragged = d.moved
		d.down, d.moved = false, false
		return false // for the button to see the click
	}
	return false
}

// Closed reports whether the user pressed Command-W in the window since
// the last frame.
func (w *PopupWindowWidget) Closed() bool { return report(w.el, w.closed) }

// SpringDelay is how long a drag rests on a closed pop-up window's tab
// before SpringOpen opens it.
const SpringDelay = 500 * time.Millisecond

// SpringOpen opens a closed pop-up window w when a drag of a T has rested
// on its tab for SpringDelay, as the classic pop-up windows sprang open
// under a drag, for what is dragged to be dropped in the window. Call it
// after PopupWindow, with its state.
func SpringOpen[T any](c *ui.Context, w *PopupWindowWidget, st *PopupWindowState) {
	since := ui.Local(w.tab, "spring", func() time.Time { return time.Time{} })
	if _, over := ui.DragOver[T](w.tab); !over || st.Open {
		*since = time.Time{}
		return
	}
	now := frameNow(c)
	if since.IsZero() {
		*since = now
	}
	if now.Sub(*since) >= SpringDelay {
		st.Open = true
		*since = time.Time{}
		report(w.el, true)
		return
	}
	c.After(SpringDelay - now.Sub(*since))
}

// TornOff reports whether the user dragged the tab up, tearing the window
// off the bottom edge, since the last frame.
func (w *PopupWindowWidget) TornOff() bool { return report(w.el, w.tornOff) }

// Activated reports whether the user pressed the window since the last
// frame, which brings a window in the background to the front.
func (w *PopupWindowWidget) Activated() bool { return report(w.el, w.activate) }

// Element returns the pop-up window's element.
func (w *PopupWindowWidget) Element() ui.Element { return w.el }

// Tab returns the element of the window's tab, for the app to take drops
// on it (ui.Drop), as the classic pop-up windows took items dropped on
// their tabs.
func (w *PopupWindowWidget) Tab() ui.Element { return w.tab }

// DrawPopupWindow draws a pop-up window filling pn's box: its tab, tabW
// wide, with the title, and, while open, the window under it, its frame
// and the sunken edge of its content running off the bottom. The tab is
// pressed in while pressed; in the background, its title is gray.
func DrawPopupWindow(pn Pen, title string, tabW int, open, active, pressed bool) {
	tx := popupTabInset
	th := PopupTabHeight
	face, light, dark := Gray2, White, Gray7
	if !active {
		light, dark = Gray1, Gray5
	}
	if pressed {
		face, light, dark = Gray5, Gray7, Gray3
	}
	// The window, under the tab's bottom row, its bottom off the edge.
	if open {
		top := th - 1
		h := pn.H - top
		pn.Rect(0, top, pn.W, h, Gray2)
		pn.Rect(0, top, pn.W, 1, Black)
		pn.VLine(0, top, h, Black)
		pn.VLine(pn.W-1, top, h, Black)
		pn.HLine(1, top+1, pn.W-2, light)
		pn.VLine(1, top+1, h-1, light)
		pn.VLine(pn.W-2, top+2, h-2, dark)
		// The sunken edge around the content, open at the bottom.
		s := FrameWidth
		cx, cy, cw := s-2, top+s-2, pn.W-2*s+4
		pn.HLine(cx, cy, cw-1, dark)
		pn.VLine(cx, cy, h, dark)
		pn.VLine(cx+cw-1, cy+1, h, light)
		pn.HLine(cx+1, cy+1, cw-2, Black)
		pn.VLine(cx+1, cy+1, h, Black)
		pn.VLine(cx+cw-2, cy+1, h, Black)
	}
	// The tab: its corners cut, its bottom open into the window, or on
	// the parent's edge.
	pn.Rect(tx+1, 1, tabW-2, th-1, face)
	pn.HLine(tx+2, 0, tabW-4, Black)
	pn.Dot(tx+1, 1, Black)
	pn.Dot(tx+tabW-2, 1, Black)
	pn.VLine(tx, 2, th-2, Black)
	pn.VLine(tx+tabW-1, 2, th-2, Black)
	pn.HLine(tx+2, 1, tabW-4, light)
	pn.VLine(tx+1, 2, th-2, light)
	pn.VLine(tx+tabW-2, 2, th-2, dark)
	if open {
		// The tab's face runs into the window's.
		pn.HLine(tx+1, th-1, tabW-2, face)
		pn.Dot(tx+1, th-1, light)
		pn.Dot(tx+tabW-2, th-1, dark)
		if !pressed {
			pn.HLine(tx+1, th, tabW-2, Gray2)
		}
	}
	ink := Black
	if !active {
		ink = Gray8
	}
	pn.Clip(tx+2, 0, tabW-4, th, func() {
		pn.TextMiddle(tx+(tabW-SystemFont.Measure(title))/2, 1, th-2, title, SystemFont, ink)
	})
}
