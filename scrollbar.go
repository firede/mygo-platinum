package platinum

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Scroll bar metrics, from docs/spec.md.
const (
	ScrollBarWidth = 16 // the width of a vertical scroll bar, and the height of a horizontal one
	ScrollMinThumb = 16 // the shortest a thumb gets
	scrollRepeat   = 60 * time.Millisecond
	scrollDelay    = 300 * time.Millisecond
)

// ScrollPart is a part of a scroll bar.
type ScrollPart int

// The parts of a scroll bar.
const (
	ScrollNone ScrollPart = iota
	ScrollUpArrow
	ScrollDownArrow
	ScrollPageUp
	ScrollPageDown
	ScrollThumb
)

type scrollState struct {
	part     ScrollPart // the part pressed
	inside   bool       // the pointer is over the part pressed
	grab     float32    // where the thumb was grabbed, from its top
	next     time.Time  // when the part pressed acts again
	step     float64    // the arrows' step when it was pressed
	length   float32    // the bar's length, as last painted
	rect     ui.Rect
	pointer  float32 // the pointer along the bar, from its start
	dragging bool
	changed  bool // the user scrolled since the last build
}

// ScrollBarWidget is a scroll bar being built.
type ScrollBarWidget struct {
	c              *ui.Context
	inactiveWindow bool
	el             ui.Element
	value          *float64
	visible        float64
	total          float64
	horizontal     bool
	step           float64
	changed        bool
	st             *scrollState
	th             *Theme
	active         bool
}

// ScrollBar builds a vertical scroll bar setting *value, the offset of the
// content in view, from 0 to total-visible: arrows at both ends step it,
// a click in the track pages it, and the thumb, as long as the part of the
// content in view is of the whole, drags it. With nothing to scroll, it
// shows an empty track. Its element is ScrollBarWidth wide; give it a
// height, or let it stretch.
func ScrollBar(c *ui.Context, value *float64, visible, total float64) *ScrollBarWidget {
	s := &ScrollBarWidget{c: c, value: value, visible: visible, total: total, step: 16, th: ThemeOf(c), active: true, inactiveWindow: inactive(c)}
	s.el = ui.Box(c).Width(ScrollBarWidth).Role(ui.RoleScroll)
	s.st = ui.Local(s.el, "scroll", func() scrollState { return scrollState{} })
	s.changed, s.st.changed = s.st.changed, false
	s.clamp()
	st := s.st
	now := frameNow(c)
	// Arrows and the track repeat while held.
	if st.part != ScrollNone && st.part != ScrollThumb && !st.dragging {
		if st.inside && !now.Before(st.next) {
			// Step has not been called yet: repeat by the step pressed with.
			s.step = st.step
			s.act(st.part)
			st.next = now.Add(scrollRepeat)
		}
		c.After(scrollRepeat)
	}
	s.el.HandleInput(func(ev ui.InputEvent) bool { return s.input(ev) })
	s.el.Draw(s.draw)
	report(s.el, s.changed)
	return s
}

// Horizontal makes the bar run left to right, ScrollBarWidth tall.
func (s *ScrollBarWidget) Horizontal() *ScrollBarWidget {
	s.horizontal = true
	s.el.Width(0).Height(ScrollBarWidth).Role(ui.RoleScroll)
	return s
}

// Step sets how far the arrows move the content, 16 by default.
func (s *ScrollBarWidget) Step(step float64) *ScrollBarWidget {
	s.step = step
	return s
}

// Inactive draws the bar as in a window in the background: a flat gray
// track, no thumb and no arrows. Bars in a Window in the background draw so
// of themselves.
func (s *ScrollBarWidget) Inactive() *ScrollBarWidget {
	s.active = false
	return s
}

// Changed reports whether the user scrolled since the last frame.
func (s *ScrollBarWidget) Changed() bool { return report(s.el, s.changed) }

// Element returns the scroll bar's element.
func (s *ScrollBarWidget) Element() ui.Element { return s.el }

// Label names the scroll bar.
func (s *ScrollBarWidget) Label(l string) *ScrollBarWidget { s.el.Label(l); return s }

func (s *ScrollBarWidget) max() float64 { return max(s.total-s.visible, 0) }

func (s *ScrollBarWidget) clamp() {
	*s.value = min(max(*s.value, 0), s.max())
}

func (s *ScrollBarWidget) set(v float64) {
	old := *s.value
	*s.value = v
	s.clamp()
	if *s.value != old {
		s.changed = true
		s.st.changed = true
	}
}

func (s *ScrollBarWidget) act(p ScrollPart) {
	page := max(s.visible-s.step, s.step)
	switch p {
	case ScrollUpArrow:
		s.set(*s.value - s.step)
	case ScrollDownArrow:
		s.set(*s.value + s.step)
	case ScrollPageUp:
		s.set(*s.value - page)
	case ScrollPageDown:
		s.set(*s.value + page)
	}
}

// thumb returns where the thumb starts along the track and how long it is,
// for a bar length long; ok is false where it shows none.
func scrollThumb(length int, value, visible, total float64) (pos, size int, ok bool) {
	track := length - 2*ScrollBarWidth + 2
	if total <= visible || track < ScrollMinThumb {
		return 0, 0, false
	}
	size = max(ScrollMinThumb, int(float64(track)*visible/total))
	size = min(size, track)
	f := value / (total - visible)
	pos = ScrollBarWidth - 1 + int(f*float64(track-size)+0.5)
	return pos, size, true
}

func (s *ScrollBarWidget) partAt(at float32) ScrollPart {
	length := int(s.st.length)
	a := int(at)
	if a < ScrollBarWidth {
		return ScrollUpArrow
	}
	if a >= length-ScrollBarWidth {
		return ScrollDownArrow
	}
	pos, size, ok := scrollThumb(length, *s.value, s.visible, s.total)
	if !ok {
		return ScrollNone
	}
	switch {
	case a < pos:
		return ScrollPageUp
	case a >= pos+size:
		return ScrollPageDown
	}
	return ScrollThumb
}

func (s *ScrollBarWidget) input(ev ui.InputEvent) bool {
	st := s.st
	at := ev.Y
	if s.horizontal {
		at = ev.X
	}
	switch ev.Kind {
	case ui.InputPointerDown:
		if ev.Button != 0 || s.total <= s.visible {
			return true
		}
		st.part = s.partAt(at)
		st.inside = true
		st.pointer = at
		if st.part == ScrollThumb {
			pos, _, _ := scrollThumb(int(st.length), *s.value, s.visible, s.total)
			st.grab = at - float32(pos)
			st.dragging = true
		} else {
			s.act(st.part)
			st.step = s.step
			st.next = eventNow(ev).Add(scrollDelay)
		}
		return true
	case ui.InputPointerMove:
		if st.part == ScrollNone {
			return false
		}
		st.pointer = at
		if st.dragging {
			_, size, ok := scrollThumb(int(st.length), *s.value, s.visible, s.total)
			if ok {
				track := float32(int(st.length) - 2*ScrollBarWidth + 2 - size)
				f := (at - st.grab - float32(ScrollBarWidth-1)) / track
				s.set(float64(f) * s.max())
			}
			return true
		}
		// Arrows and the track act only while the pointer stays over the
		// part pressed, as the pointer leaving a button releases it.
		p := s.partAt(at)
		st.inside = p == st.part || (st.part == ScrollPageUp || st.part == ScrollPageDown) && p == st.part
		return true
	case ui.InputPointerUp:
		st.part, st.dragging, st.inside = ScrollNone, false, false
		return true
	case ui.InputScroll:
		d := ev.DY
		if s.horizontal {
			d = ev.DX
			if d == 0 {
				d = ev.DY
			}
		}
		s.set(*s.value + float64(d))
		return true
	}
	return false
}

func (s *ScrollBarWidget) draw(p *ui.Painter, r ui.Rect) {
	pn := NewPen(p, r)
	st := s.st
	resized(s.c, st.rect, r)
	st.rect = r
	st.length = r.H
	if s.horizontal {
		st.length = r.W
	}
	pressed := ScrollNone
	if st.inside || st.dragging {
		pressed = st.part
	}
	if s.inactiveWindow || !s.active {
		DrawScrollBarInactive(pn, s.horizontal)
		return
	}
	DrawScrollBar(pn, s.horizontal, *s.value, s.visible, s.total, pressed, s.total > s.visible, s.th.Accent)
}

var (
	arrowUp    = []string{"...#...", "..###..", ".#####.", "#######"}
	arrowDown  = []string{"#######", ".#####.", "..###..", "...#..."}
	arrowLeft  = []string{"...#", "..##", ".###", "####", ".###", "..##", "...#"}
	arrowRight = []string{"#...", "##..", "###.", "####", "###.", "##..", "#..."}
)

// DrawScrollBar draws a scroll bar filling pn's box: its arrows, track and
// thumb for value, visible and total as ScrollBar takes them, the part
// pressed darkened. An inactive bar shows its track empty and its arrows
// dimmed.
func DrawScrollBar(pn Pen, horizontal bool, value, visible, total float64, pressed ScrollPart, active bool, a Accent) {
	// Draw a vertical bar; a horizontal one is the same turned, which the
	// pen does by swapping the axes of every call through rot.
	length, breadth := pn.H, pn.W
	if horizontal {
		length, breadth = pn.W, pn.H
	}
	rect := func(x, y, w, h int, c ui.Color) {
		if horizontal {
			pn.Rect(y, x, h, w, c)
		} else {
			pn.Rect(x, y, w, h, c)
		}
	}
	// The frame around it all.
	rect(0, 0, breadth, length, Black)
	// The track.
	track := length - 2*ScrollBarWidth + 2
	if active {
		rect(1, ScrollBarWidth, breadth-2, track-2, Gray4)
		rect(1, ScrollBarWidth, 1, track-2, Gray7)
		rect(2, ScrollBarWidth, breadth-3, 1, Gray7)
		rect(breadth-2, ScrollBarWidth+1, 1, track-3, Gray3)
		if pressed == ScrollPageUp || pressed == ScrollPageDown {
			pos, size, _ := scrollThumb(length, value, visible, total)
			if pressed == ScrollPageUp {
				rect(2, ScrollBarWidth+1, breadth-4, pos-ScrollBarWidth-1, Gray6)
			} else {
				rect(2, pos+size, breadth-4, length-ScrollBarWidth-1-pos-size, Gray6)
			}
		}
	} else {
		rect(1, ScrollBarWidth, breadth-2, track-2, Gray2)
	}
	// The arrows.
	for i, part := range []ScrollPart{ScrollUpArrow, ScrollDownArrow} {
		y := 0
		if i == 1 {
			y = length - ScrollBarWidth
		}
		face, tl, br, ink := Gray2, White, Gray7, Black
		if pressed == part {
			face, tl, br = Gray5, Gray7, Gray3
		}
		if !active {
			ink = Gray6
		}
		rect(1, y+1, breadth-2, ScrollBarWidth-2, face)
		rect(1, y+1, breadth-3, 1, tl)
		rect(1, y+2, 1, ScrollBarWidth-4, tl)
		rect(2, y+ScrollBarWidth-2, breadth-3, 1, br)
		rect(breadth-2, y+2, 1, ScrollBarWidth-4, br)
		var bm []string
		switch {
		case !horizontal && i == 0:
			bm = arrowUp
		case !horizontal:
			bm = arrowDown
		case i == 0:
			bm = arrowLeft
		default:
			bm = arrowRight
		}
		bw, bh := len(bm[0]), len(bm)
		if horizontal {
			pn.Bitmap(y+(ScrollBarWidth-bw)/2, (breadth-bh)/2, bm, ink)
		} else {
			pn.Bitmap((breadth-bw)/2, y+(ScrollBarWidth-bh)/2, bm, ink)
		}
	}
	if !active {
		return
	}
	pos, size, ok := scrollThumb(length, value, visible, total)
	if !ok {
		return
	}
	face, dark := thumbFace(a), a.Dark
	if pressed == ScrollThumb {
		face = a.Mid
	}
	rect(1, pos, breadth-2, size, Black)
	rect(1, pos+1, breadth-2, size-2, face)
	rect(1, pos+1, breadth-3, 1, White)
	rect(1, pos+2, 1, size-4, White)
	rect(2, pos+size-2, breadth-3, 1, dark)
	rect(breadth-2, pos+2, 1, size-4, dark)
	// Grip ridges across the middle of the thumb.
	mid := pos + size/2
	for k := -2; k < 2; k++ {
		rect(4, mid+2*k, breadth-8, 1, dark)
		rect(5, mid+2*k+1, breadth-8, 1, White)
	}
}

// DrawScrollBarInactive draws a scroll bar of a window in the background
// filling pn's box: its outline round a flat gray track, with no arrows or
// thumb.
func DrawScrollBarInactive(pn Pen, horizontal bool) {
	pn.Rect(0, 0, pn.W, pn.H, Gray2)
	pn.Frame(0, 0, pn.W, pn.H, Black)
}
