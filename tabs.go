package platinum

import (
	"github.com/egoist/mygo/ui"
)

// Tab control metrics, from docs/spec.md.
const (
	TabHeight  = 22 // the height of an unselected tab
	TabRaise   = 2  // how much taller the selected tab is
	TabPadding = 12 // the room on each side of a tab's label
	tabInset   = 8  // from the pane's left edge to the first tab
)

// TabsWidget is a tab control being built.
type TabsWidget struct {
	parts ui.TabsParts
	el    ui.Element
}

// TabControl builds a tab control: a row of tabs labeled labels over a
// raised pane that shows the page of the tab chosen, *selected. fn builds
// that page, given its index. A click chooses a tab, and the arrows, Home
// and End move the choice while a tab has the keyboard focus.
func TabControl(c *ui.Context, selected *int, labels []string, fn func(i int)) *TabsWidget {
	th := ThemeOf(c)
	t := &TabsWidget{}
	tabState := StateNormal
	if inactive(c) {
		tabState = StateInactive
	}
	n := len(labels)
	*selected = min(max(*selected, 0), max(n-1, 0))
	widths := make([]int, n)
	for i, l := range labels {
		widths[i] = SystemFont.Measure(l) + 2*TabPadding
	}
	t.el = ui.Column(c).AlignItems(ui.Stretch)
	t.el.Children(func() {
		t.parts = ui.TabsBase(c, selected, n)
		t.parts.List.Height(TabHeight+TabRaise).Padding(0, 0, 0, tabInset).AlignItems(ui.End)
		t.parts.List.Children(func() {
			for i, l := range labels {
				tab := t.parts.Tab(i).FocusRing(false).Label(l).Width(float32(widths[i]))
				sel := i == *selected
				h := TabHeight
				if sel {
					h += TabRaise
				}
				tab.Height(float32(h))
				focused := tab.FocusVisible()
				tab.Draw(func(p *ui.Painter, r ui.Rect) {
					pn := NewPen(p, r)
					DrawTab(pn, l, sel, tabState)
					if focused {
						pn.Rect(5, pn.H-5, pn.W-10, 2, th.Accent.Dark)
					}
				})
			}
		})
		x := tabInset
		for i := 0; i < *selected; i++ {
			x += widths[i]
		}
		var gap [2]int // no tabs, no opening in the pane
		if n > 0 {
			gap = [2]int{x, widths[*selected]}
		}
		ui.Column(c).Grow(1).Padding(GroupPadding).Gap(6).Draw(func(p *ui.Painter, r ui.Rect) {
			DrawTabPane(NewPen(p, r), gap[0], gap[1])
		}).Children(func() {
			if n > 0 {
				fn(*selected)
			}
		})
	})
	return t
}

// Changed reports whether the user chose another tab since the last frame.
func (t *TabsWidget) Changed() bool { return t.parts.List.Changed() }

// Element returns the element holding the tabs and their pane.
func (t *TabsWidget) Element() ui.Element { return t.el }

// DrawTab draws a tab labeled label filling pn's box: its top corners
// rounded, its bottom open onto the pane when selected.
func DrawTab(pn Pen, label string, selected bool, s State) {
	w, h := pn.W, pn.H
	face := Gray4
	if selected {
		face = Gray2
	}
	if s == StatePressed {
		face = Gray5
	}
	// The outline: top corners cut, no bottom edge.
	for i := 0; i < h; i++ {
		d := Inset(i, 2*h, 3)
		if i == 0 {
			pn.HLine(d, 0, w-2*d, Black)
			continue
		}
		pn.Dot(d, i, Black)
		pn.Dot(w-1-d, i, Black)
		pn.HLine(d+1, i, w-2*d-2, face)
	}
	pn.Dot(2, 1, Black)
	pn.Dot(w-3, 1, Black)
	// Its bevel.
	pn.HLine(3, 1, w-6, White)
	pn.Dot(2, 2, White)
	pn.VLine(1, 3, h-3, White)
	pn.VLine(w-2, 3, h-3, Gray7)
	pn.Dot(w-3, 2, Gray7)
	if !selected {
		pn.HLine(0, h-1, w, Black)
	}
	text := Black
	if s == StateDisabled || s == StateInactive {
		text = Gray8
	}
	ty := 0
	if selected {
		ty = TabRaise
	}
	pn.TextCentered(0, ty, w, h-1-ty, label, SystemFont, text)
}

// DrawTabPane draws the raised pane of a tab control filling pn's box,
// open under the selected tab, gapW wide from gapX.
func DrawTabPane(pn Pen, gapX, gapW int) {
	w, h := pn.W, pn.H
	pn.Rect(0, 0, w, h, Gray2)
	pn.Frame(0, 0, w, h, Black)
	pn.HLine(1, 1, w-3, White)
	pn.VLine(1, 1, h-3, White)
	pn.HLine(2, h-2, w-3, Gray7)
	pn.VLine(w-2, 2, h-3, Gray7)
	if gapW > 0 {
		pn.HLine(gapX+1, 0, gapW-2, Gray2)
		pn.HLine(gapX+2, 1, gapW-4, Gray2)
		pn.Dot(gapX+1, 0, White)
		pn.Dot(gapX+gapW-2, 0, Gray7)
		pn.Dot(gapX+gapW-2, 1, Gray7)
	}
}

// DocumentTabsWidget is a row of document tabs being built.
type DocumentTabsWidget struct {
	el     ui.Element
	closed int
}

// DocumentTabs builds the tabs of the documents of a window, choosing
// *selected, over the pane fn builds for the document chosen. Command-W,
// while the tabs have the keyboard focus, and the Close item of a tab's
// context menu close a document: DocumentTabs reports the index of the one
// closed, or -1, for the app to remove it. The arrows move between the
// tabs.
func DocumentTabs(c *ui.Context, selected *int, titles []string, fn func(i int)) *DocumentTabsWidget {
	th := ThemeOf(c)
	w := &DocumentTabsWidget{closed: -1}
	n := len(titles)
	*selected = min(max(*selected, 0), max(n-1, 0))
	st := ui.Local(c.Root(), "doctabs", func() int { return -1 }) // a close chosen from a menu
	if *st >= 0 {
		w.closed, *st = *st, -1
	}
	widths := make([]int, n)
	for i, t := range titles {
		widths[i] = SystemFont.Measure(t) + 2*TabPadding
	}
	w.el = ui.Column(c).AlignItems(ui.Stretch)
	w.el.Children(func() {
		parts := ui.TabsBase(c, selected, n)
		parts.List.Height(TabHeight+TabRaise).Padding(0, 0, 0, 4).AlignItems(ui.End)
		if parts.List.Shortcut(ui.Cmd, ui.KeyW) && n > 0 {
			w.closed = *selected
		}
		parts.List.Children(func() {
			for i, title := range titles {
				tab := parts.Tab(i).FocusRing(false).Label(title).Width(float32(widths[i]))
				sel := i == *selected
				h := TabHeight
				if sel {
					h += TabRaise
				}
				tab.Height(float32(h))
				focused := tab.FocusVisible()
				tab.Draw(func(p *ui.Painter, r ui.Rect) {
					pn := NewPen(p, r)
					DrawTab(pn, title, sel, StateNormal)
					if focused {
						pn.Rect(5, pn.H-5, pn.W-10, 2, th.Accent.Dark)
					}
				})
				ContextMenu(c, tab, []MenuItem{{Title: "Close", Action: func() { *st = i }}})
			}
		})
		x := 4
		for i := 0; i < *selected && i < n; i++ {
			x += widths[i]
		}
		var gap [2]int
		if n > 0 {
			gap = [2]int{x, widths[*selected]}
		}
		ui.Column(c).Grow(1).AlignItems(ui.Stretch).Padding(10).Draw(func(p *ui.Painter, r ui.Rect) {
			DrawTabPane(NewPen(p, r), gap[0], gap[1])
		}).Children(func() {
			if n > 0 {
				fn(*selected)
			} else {
				EmptyState(c, "No documents")
			}
		})
	})
	return w
}

// Closed returns the index of the tab whose close box the user clicked
// since the last frame, or -1.
func (w *DocumentTabsWidget) Closed() int { return w.closed }

// Element returns the element of the tabs and their pane.
func (w *DocumentTabsWidget) Element() ui.Element { return w.el }
