package platinum

import (
	"sync"

	"github.com/egoist/mygo/ui"
)

// Window frame metrics, from docs/spec.md.
const (
	TitleBarHeight = 19 // the height of a window's title bar, its outline included
	FrameWidth     = 6  // the frame at a window's sides and bottom, outline included
	TitleBoxSize   = 11 // the close, zoom and collapse boxes
	GrowBoxSize    = 15 // the resize corner
	titleBoxInset  = 8  // from the window's side to the first box
	titleBoxGap    = 4  // between the zoom and collapse boxes
	stripeTop      = 4  // the first title bar stripe
	stripeCount    = 6  // pairs of rows of stripes
)

// WindowState is where a window is and how it shows, which the app keeps:
// the window moves and resizes it as the user drags it.
type WindowState struct {
	// X and Y place the window's outline in its parent, and W and H size
	// it, title bar and frame included. An H of 0 makes it as tall as its
	// content.
	X, Y, W, H int
	// MinW and MinH are as small as the grow box makes it (120×80 if
	// zero).
	MinW, MinH int
	// Collapsed shows the title bar alone, as a window shade rolled up.
	Collapsed bool
	// Zoomed shows the window at its zoomed size; the zoom box toggles it
	// with the size the user gave it, which Unzoomed keeps.
	Zoomed   bool
	Unzoomed [4]int
}

// WindowPart is a part of a window's frame.
type WindowPart int

// The parts of a window frame, as WindowPartAt finds them.
const (
	PartNone WindowPart = iota
	PartTitle
	PartClose
	PartZoom
	PartCollapse
	PartGrow
	PartContent
)

// WindowOptions says which parts a window's frame has.
type WindowOptions struct {
	// NoClose, NoZoom and NoCollapse leave out the boxes of the title bar,
	// and NoGrow the resize corner.
	NoClose, NoZoom, NoCollapse, NoGrow bool
	// Active draws the window as the one in front: its title bar striped
	// and its boxes shown. A window in the background shows a plain title
	// bar and dimmed title.
	Active bool
	// ZoomTo is the box the zoom box zooms the window to, in its parent;
	// zero zooms to the window's own size.
	ZoomTo [4]int
	// Modal draws the frame of a movable modal dialog: no close box, no
	// zoom or collapse box, no resize corner, and a thicker border.
	Modal bool
	// Interactive lets the content of a window in the background take
	// clicks, as a floating palette's does; otherwise its first click only
	// brings it to the front.
	Interactive bool
	// Dragging says a drag is under way, which a window in the background
	// lets drop in its content while it keeps its look; the app knows its
	// drags (ui.Element.Dragging).
	Dragging bool
	// Utility draws the frame of a floating utility window, as an
	// inspector or a palette: a narrow title bar with a small close box and
	// its title in the regular face, and a thin frame.
	Utility bool
	// Path is the folders a folder window's folder is in, the outermost
	// first: a Command-press on the title shows the window's title and
	// them as a menu, to go to one, which PathChosen reports.
	Path []string
}

// Utility window metrics, from docs/spec.md.
const (
	UtilityBarHeight = 13 // the height of a utility window's title bar
	UtilityBoxSize   = 9  // its close box
	utilityFrame     = 3  // its frame at the sides and bottom
)

// bar returns the height of the title bar of a window with opts.
func (o WindowOptions) bar() int {
	if o.Utility {
		return UtilityBarHeight
	}
	return TitleBarHeight
}

// side returns the width of the frame at the sides and bottom.
func (o WindowOptions) side() int {
	switch {
	case o.Utility:
		return utilityFrame
	case o.Modal:
		return FrameWidth + 2
	}
	return FrameWidth
}

// boxSize returns the size of the boxes of its title bar, and their top.
func (o WindowOptions) boxSize() (int, int) {
	if o.Utility {
		return UtilityBoxSize, 2
	}
	return TitleBoxSize, stripeTop
}

type windowDrag struct {
	part     WindowPart
	inside   bool
	grabX    float32
	grabY    float32
	origin   WindowState
	rect     ui.Rect // the window on the window, as last painted
	parentX  float32
	parentY  float32
	clicked  WindowPart
	activate bool
	// The path menu of a Command-press on the title, the folder chosen
	// from it, and the window's title and size, as the last build saw them.
	path       *menuTracker
	pathChosen int
	title      string
	cw, ch     float32
	mods       ui.Modifiers // the modifier keys held
}

// WindowWidget is a window being built.
type WindowWidget struct {
	el         ui.Element
	drag       *windowDrag
	closed     bool
	activate   bool
	pathChosen int
}

// Window builds a window of the desktop inside its parent, placed where st
// says: an outlined frame with a striped title bar showing title, a close
// box at its left, zoom and collapse boxes at its right and a resize corner
// at the bottom right, around the content fn builds. Dragging the title bar
// moves the window, dragging the corner resizes it, double-clicking the
// title bar rolls it up, and the boxes act when released over. Put the
// window in a parent that positions absolutely, as a full-window Box, and
// build the windows back to front.
func Window(c *ui.Context, st *WindowState, title string, opts WindowOptions, fn func()) *WindowWidget {
	w := &WindowWidget{}
	if st.MinW == 0 {
		st.MinW = 120
	}
	if st.MinH == 0 {
		st.MinH = 80
	}
	if opts.Utility {
		opts.NoZoom, opts.NoGrow = true, true
	}
	w.el = ui.Column(c).Absolute().Role(ui.RoleDialog).Label(title)
	if opts.Modal {
		opts.NoClose, opts.NoZoom, opts.NoCollapse, opts.NoGrow = true, true, true, true
	}
	d := ui.Local(w.el, "window", func() windowDrag { return windowDrag{pathChosen: -1} })
	w.drag = d
	d.title = title
	d.cw, d.ch = c.Size()
	d.mods = c.Modifiers()
	w.pathChosen, d.pathChosen = d.pathChosen, -1
	if d.path != nil {
		if it, _, ok := d.path.done(frameNow(c)); ok {
			d.path = nil
			if it.Action != nil {
				it.Action()
				w.pathChosen, d.pathChosen = d.pathChosen, -1
				report(w.el, true)
			}
		} else if d.path.closed {
			d.path = nil
		}
	}
	if d.path != nil {
		trackerOverlay(c, d.path, ThemeOf(c).Accent, nil, nil)
	}
	if d.clicked != PartNone {
		switch d.clicked {
		case PartClose:
			w.closed = true
		case PartZoom:
			toggleZoom(st, opts.ZoomTo)
		case PartCollapse:
			st.Collapsed = !st.Collapsed
		}
		d.clicked = PartNone
	}
	if d.activate {
		w.activate = true
		d.activate = false
	}
	// Apply title-box actions before sizing the window so its frame and
	// content use the same state in the frame that handles the click.
	h := st.H
	if st.Collapsed {
		h = opts.bar()
	}
	w.el.Left(float32(st.X)).Top(float32(st.Y)).Width(float32(st.W))
	if h > 0 {
		w.el.Height(float32(h)) // else as tall as its content
	}
	if w.el.Shortcut(ui.Cmd, ui.KeyW) && !opts.NoClose {
		w.closed = true
	}
	side := opts.side()
	w.el.HandleInput(func(ev ui.InputEvent) bool {
		return w.input(ev, st, opts)
	})
	w.el.Padding(float32(opts.bar()), float32(side), float32(side), float32(side))
	pressed := PartNone
	if d.inside {
		pressed = d.part
	}
	w.el.Draw(func(p *ui.Painter, r ui.Rect) {
		resized(c, d.rect, r)
		d.rect = r
		d.parentX, d.parentY = r.X-float32(st.X), r.Y-float32(st.Y)
		DrawWindowFrame(NewPen(p, r), title, opts, st.Collapsed, pressed)
	})
	if !st.Collapsed {
		w.el.Children(func() {
			inner := ui.Column(c).Clip()
			if st.H > 0 {
				inner.Grow(1)
			}
			background := !opts.Active && !opts.Interactive
			if background {
				enterInactive(c)
			}
			inner.Children(func() {
				fn()
				if !opts.Active && !opts.Interactive && !opts.Dragging {
					// A window in the background takes no click in its
					// content: the first only brings it to the front, as on
					// the classic desktop. A transparent shield over the
					// content, and not the title bar, which still drags it,
					// takes it.
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
			if background {
				leaveInactive(c)
			}
		})
	}
	report(w.el, w.closed || w.activate || w.pathChosen >= 0)
	return w
}

func toggleZoom(st *WindowState, to [4]int) {
	if to == [4]int{} {
		return
	}
	if st.Zoomed {
		st.X, st.Y, st.W, st.H = st.Unzoomed[0], st.Unzoomed[1], st.Unzoomed[2], st.Unzoomed[3]
		st.Zoomed = false
		return
	}
	st.Unzoomed = [4]int{st.X, st.Y, st.W, st.H}
	st.X, st.Y, st.W, st.H = to[0], to[1], to[2], to[3]
	st.Zoomed = true
}

// Closed reports whether the user clicked the close box, or pressed
// Command-W in the window, since the last frame.
func (w *WindowWidget) Closed() bool { return report(w.el, w.closed) }

// PathChosen returns the index in WindowOptions.Path of the folder the user
// chose from the title's path menu since the last frame, or -1.
func (w *WindowWidget) PathChosen() int {
	report(w.el, w.pathChosen >= 0)
	return w.pathChosen
}

// Activated reports whether the user pressed the window since the last
// frame, which brings a window in the background to the front.
func (w *WindowWidget) Activated() bool { return report(w.el, w.activate) }

// Element returns the window's element.
func (w *WindowWidget) Element() ui.Element { return w.el }

// WindowPartAt returns the part of a window w×h (title bar included, or
// collapsed) at x, y from its corner.
func WindowPartAt(x, y float32, w, h int, opts WindowOptions, collapsed bool) WindowPart {
	if x < 0 || y < 0 || x >= float32(w) || y >= float32(h) {
		return PartNone
	}
	box, top := opts.boxSize()
	inset := titleBoxInset
	if opts.Utility {
		inset = 4
	}
	in := func(bx int) bool {
		return x >= float32(bx) && x < float32(bx+box) && y >= float32(top) && y < float32(top+box)
	}
	if y < float32(opts.bar()) {
		if !opts.NoClose && in(inset) {
			return PartClose
		}
		x0 := w - inset - box
		if !opts.NoCollapse {
			if in(x0) {
				return PartCollapse
			}
			x0 -= box + titleBoxGap
		}
		if !opts.NoZoom && in(x0) {
			return PartZoom
		}
		return PartTitle
	}
	if !collapsed && !opts.NoGrow && x >= float32(w-GrowBoxSize-1) && y >= float32(h-GrowBoxSize-1) {
		return PartGrow
	}
	return PartContent
}

func (w *WindowWidget) input(ev ui.InputEvent, st *WindowState, opts WindowOptions) bool {
	d := w.drag
	// The pointer in the parent's coordinates: the window may have moved
	// since the event's coordinates were taken, but not its last painting.
	px := d.rect.X - d.parentX + ev.X
	py := d.rect.Y - d.parentY + ev.Y
	ex, ey := px-float32(st.X), py-float32(st.Y) // and in the window's
	h := st.H
	if st.Collapsed {
		h = opts.bar()
	} else if h == 0 {
		h = round(d.rect.H)
	}
	switch ev.Kind {
	case ui.InputPointerDown:
		if ev.Button != 0 {
			return false
		}
		d.activate = true
		part := WindowPartAt(ex, ey, st.W, h, opts, st.Collapsed)
		if part == PartContent || part == PartNone {
			return false
		}
		if part == PartTitle && ev.Clicks == 2 && !opts.NoCollapse {
			st.Collapsed = !st.Collapsed
			return true
		}
		if part == PartTitle && (ev.Mods|d.mods)&ui.Cmd != 0 && len(opts.Path) > 0 {
			// The path menu: the window's folder, then those it is in.
			items := []MenuItem{{Title: d.title, Checked: true}}
			for i := len(opts.Path) - 1; i >= 0; i-- {
				items = append(items, MenuItem{Title: opts.Path[i], Action: func() { d.pathChosen = i }})
			}
			x := d.rect.X + float32((st.W-SystemFont.Measure(d.title))/2-16)
			g := layoutMenu(items, round(x), round(d.rect.Y)+opts.bar(), 0)
			d.path = newTracker(items, fitMenu(g, round(d.cw), round(d.ch)))
			d.path.press(eventNow(ev), d.rect.X+ev.X, d.rect.Y+ev.Y)
			return true
		}
		d.part, d.inside = part, true
		d.grabX, d.grabY = px, py
		d.origin = *st
		return true
	case ui.InputPointerMove:
		if d.path != nil && d.path.pressed {
			d.path.move(d.rect.X+ev.X, d.rect.Y+ev.Y)
			return true
		}
		if d.part == PartNone || ev.Button != 0 {
			return false
		}
		switch d.part {
		case PartTitle:
			st.X = d.origin.X + round(px-d.grabX)
			st.Y = max(d.origin.Y+round(py-d.grabY), 0)
		case PartGrow:
			st.W = max(d.origin.W+round(px-d.grabX), st.MinW)
			st.H = max(d.origin.H+round(py-d.grabY), st.MinH)
			st.Zoomed = false
		default:
			d.inside = WindowPartAt(ex, ey, st.W, h, opts, st.Collapsed) == d.part
		}
		return true
	case ui.InputPointerUp:
		if d.path != nil && d.path.pressed {
			// Released on a folder, it is chosen; a quick click leaves the
			// menu open to choose from; elsewhere, it closes.
			d.path.pressed = false
			if !d.path.release(eventNow(ev), d.rect.X+ev.X, d.rect.Y+ev.Y) && !d.path.quickClick(eventNow(ev)) {
				d.path.closed = true
			}
			return true
		}
		if d.part == PartNone {
			return false
		}
		if d.inside && (d.part == PartClose || d.part == PartZoom || d.part == PartCollapse) {
			d.clicked = d.part
		}
		d.part, d.inside = PartNone, false
		return true
	}
	return false
}

// DrawWindowFrame draws a window's frame filling pn's box: its outline,
// title bar with title, boxes and stripes, side frame and resize corner,
// with the box pressed darkened. The content goes inside, FrameWidth from
// the sides and bottom and TitleBarHeight from the top.
func DrawWindowFrame(pn Pen, title string, opts WindowOptions, collapsed bool, pressed WindowPart) {
	w, h := pn.W, pn.H
	side := opts.side()
	bar := opts.bar()
	box, boxTop := opts.boxSize()
	inset := titleBoxInset
	if opts.Utility {
		inset = 4
	}
	active := opts.Active
	pn.Rect(0, 0, w, h, Gray2)
	pn.Frame(0, 0, w, h, Black)
	if active {
		pn.Bevel(1, 1, w-2, h-2, White, Gray7)
	} else {
		pn.Bevel(1, 1, w-2, h-2, Gray1, Gray5)
	}
	if !collapsed {
		// The sunken edge around the content.
		cx, cy, cw, ch := side-2, bar-2, w-2*side+4, h-bar-side+4
		if opts.Utility {
			// A utility window's content meets its frame in a single line.
			pn.Frame(side-1, bar-1, w-2*side+2, h-bar-side+2, Black)
		} else {
			if active {
				pn.Bevel(cx, cy, cw, ch, Gray7, White)
			} else {
				pn.Bevel(cx, cy, cw, ch, Gray5, Gray1)
			}
			pn.Frame(cx+1, cy+1, cw-2, ch-2, Black)
		}
		if opts.Modal {
			// A modal dialog's thicker border: a second ring.
			pn.Bevel(3, bar-1, w-6, h-bar-2, Gray5, White)
		}
	}
	// The title, centered, and the stripes on either side of it.
	font := SystemFont
	if opts.Utility {
		font = AppFont
	}
	tw := font.Measure(title)
	tx := (w - tw) / 2
	ink := Black
	if !active {
		ink = Gray8
	}
	if active {
		left := inset
		if !opts.NoClose {
			left += box + 4
		}
		right := w - inset
		if !opts.NoCollapse {
			right -= box + 4
		}
		if !opts.NoZoom {
			right -= box + titleBoxGap
		}
		if opts.NoClose {
			left = 4
		}
		if opts.NoCollapse && opts.NoZoom {
			right = w - 4
		}
		textL, textR := tx-6, tx+tw+6
		if title == "" {
			textL, textR = left, left
		}
		stripes, top := stripeCount, stripeTop
		if opts.Utility {
			stripes, top = 4, 2
		}
		for i := 0; i < stripes; i++ {
			y := top + 2*i
			for _, seg := range [][2]int{{left, textL}, {textR, right}} {
				if seg[1] > seg[0] {
					pn.HLine(seg[0], y, seg[1]-seg[0], Gray7)
					pn.HLine(seg[0]+1, y+1, seg[1]-seg[0], White)
				}
			}
		}
		if !opts.NoClose {
			DrawTitleBox(pn.Sub(inset, boxTop, box, box), PartClose, pressed == PartClose)
		}
		x0 := w - inset - box
		if !opts.NoCollapse {
			DrawTitleBox(pn.Sub(x0, boxTop, box, box), PartCollapse, pressed == PartCollapse)
			x0 -= box + titleBoxGap
		}
		if !opts.NoZoom {
			DrawTitleBox(pn.Sub(x0, boxTop, box, box), PartZoom, pressed == PartZoom)
		}
	}
	pn.Clip(0, 0, w, bar, func() {
		if opts.Utility {
			pn.TextMiddle(tx, 0, bar, title, font, ink)
		} else {
			pn.TextMiddle(tx, 1, bar-2, title, font, ink)
		}
	})
	if !collapsed && !opts.NoGrow {
		DrawGrowBox(pn.Sub(w-side-GrowBoxSize, h-side-GrowBoxSize, GrowBoxSize, GrowBoxSize), active)
	}
}

// DrawTitleBox draws the close, zoom or collapse box of a title bar
// filling pn's square box: 11×11, or 9×9 in a utility window.
func DrawTitleBox(pn Pen, part WindowPart, pressed bool) {
	n := pn.W
	// A sunken well with a raised button in it.
	pn.Bevel(0, 0, n, n, Gray7, White)
	pn.Frame(1, 1, n-2, n-2, Black)
	face, tl, br := Gray2, White, Gray7
	if pressed {
		face, tl, br = Gray6, Gray9, Gray4
	}
	pn.Rect(2, 2, n-4, n-4, face)
	pn.Bevel(2, 2, n-4, n-4, tl, br)
	ink := Black
	switch part {
	case PartZoom:
		pn.Frame(2, 2, 5, 5, ink)
	case PartCollapse:
		pn.HLine(2, 4, n-4, ink)
		pn.HLine(2, 6, n-4, ink)
	}
}

// DrawGrowBox draws the resize corner of a window in pn's 15×15 corner.
func DrawGrowBox(pn Pen, active bool) {
	n := GrowBoxSize
	pn.Rect(0, 0, n, n, Gray2)
	pn.HLine(0, 0, n, Black)
	pn.VLine(0, 0, n, Black)
	if !active {
		return
	}
	// Three ridges across the corner, from the bottom left up to the right.
	for k := 0; k < 3; k++ {
		o := 4 + 3*k
		for i := 0; i < o; i++ {
			pn.Dot(n-o+i, n-2-i, Gray7)
			pn.Dot(n-o+i+1, n-2-i, White)
		}
	}
}

// The windows in the background being built, by window: the controls built
// inside one draw inactive.
var (
	inactiveMu sync.Mutex
	inactives  = map[*ui.Context]int{}
)

func enterInactive(c *ui.Context) {
	inactiveMu.Lock()
	inactives[c]++
	inactiveMu.Unlock()
}

func leaveInactive(c *ui.Context) {
	inactiveMu.Lock()
	inactives[c]--
	inactiveMu.Unlock()
}

// activeWithin builds fn as in front, whatever window it is built in: a
// dialog or alert that a window in the background shows is in front of
// every window.
func activeWithin(c *ui.Context, fn func()) {
	inactiveMu.Lock()
	n := inactives[c]
	inactives[c] = 0
	inactiveMu.Unlock()
	fn()
	inactiveMu.Lock()
	inactives[c] = n
	inactiveMu.Unlock()
}

// inactive reports whether the control being built is in a window in the
// background.
func inactive(c *ui.Context) bool {
	inactiveMu.Lock()
	defer inactiveMu.Unlock()
	return inactives[c] > 0
}

// WindowHeaderHeight is the height of a window header.
const WindowHeaderHeight = 18

// WindowHeader builds the band across the top of a window's content that
// tells about it, as a folder window's count of items and the room left:
// left at its left and right at its right, in the regular face.
func WindowHeader(c *ui.Context, left, right string) ui.Element {
	return ui.Box(c).Height(WindowHeaderHeight).AlignSelf(ui.Stretch).Role(ui.RoleStatus).Label(left + ", " + right).
		Draw(func(p *ui.Painter, r ui.Rect) { DrawWindowHeader(NewPen(p, r), left, right) })
}

// DrawWindowHeader draws a window header filling pn's box: Gray2, an
// etched line and a Black one at its bottom, left 8 from its left and
// right 8 from its right.
func DrawWindowHeader(pn Pen, left, right string) {
	pn.Rect(0, 0, pn.W, pn.H, Gray2)
	pn.HLine(0, pn.H-2, pn.W, Gray7)
	pn.HLine(0, pn.H-1, pn.W, Black)
	pn.TextMiddle(8, 0, pn.H-2, left, AppFont, Black)
	pn.TextMiddle(pn.W-AppFont.Measure(right)-8, 0, pn.H-2, right, AppFont, Black)
}
