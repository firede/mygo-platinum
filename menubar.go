package platinum

import (
	"time"

	"github.com/egoist/mygo/ui"
)

const menuBarInset = 10 // from the bar's left edge to its first title

type menuBarState struct {
	t      *menuTracker
	open   int // the menu showing, or -1
	sticky bool
	rect   ui.Rect // the bar on the window, as last painted
	// The title blinking as a keyboard equivalent chose an item of it.
	flash      int
	flashUntil time.Time
	run        func()
}

// MenuBar builds a menu bar of menus along the top of its parent, as wide
// as it: its menus open as the pointer presses their titles, and choose the
// item it is released over, or stay open after a click until an item is
// clicked. Ctrl-F2 opens the first menu from the keyboard; the arrows move
// through the menus and their items, Return chooses and Escape closes.
// Keyboard equivalents of the items work wherever the focus is.
func MenuBar(c *ui.Context, menus []Menu) ui.Element {
	th := ThemeOf(c)
	bar := ui.Box(c).Height(MenuBarHeight).AlignSelf(ui.Stretch).Role(ui.RoleMenuBar).Label(say(c, "menuBar"))
	st := ui.Local(bar, "menubar", func() menuBarState { return menuBarState{open: -1, flash: -1} })
	now := frameNow(c)
	if st.run != nil {
		run := st.run
		st.run = nil
		run()
		report(bar, true)
	}

	barW, _ := c.Size()
	if st.rect.W > 0 {
		barW = st.rect.W
	}
	xs, ws := menuTitles(menus, round(barW))
	titleAt := func(x, y float32) int {
		if y < st.rect.Y || y >= st.rect.Y+MenuBarHeight {
			return -1
		}
		for i := range menus {
			if x >= st.rect.X+float32(xs[i]) && x < st.rect.X+float32(xs[i]+ws[i]) {
				return i
			}
		}
		return -1
	}
	// The window's size, read while building: input handlers run between
	// builds, where the context has none.
	cw, ch := c.Size()
	openMenu := func(i int, keyboard bool) {
		if i < 0 || i >= len(menus) || len(menus[i].Items) == 0 {
			return
		}
		g := layoutMenu(menus[i].Items, round(st.rect.X)+xs[i], round(st.rect.Y)+MenuBarHeight-1, 0)
		g = fitMenu(g, round(cw), round(ch))
		st.open = i
		st.t = newTracker(menus[i].Items, g)
		if keyboard {
			st.t.keyboard = true
			st.t.key(now, ui.KeyDown)
		}
	}
	closeMenu := func() {
		st.open, st.t, st.sticky = -1, nil, false
	}

	// Keyboard equivalents.
	for i, m := range menus {
		for _, it := range m.Items {
			for _, sub := range append([]MenuItem{it}, it.Submenu...) {
				if sub.Key == 0 || sub.Disabled || sub.Separator || sub.Action == nil {
					continue
				}
				if c.Shortcut(sub.mods(), sub.Key) {
					st.flash, st.flashUntil = i, now.Add(MenuFlash)
					c.After(MenuFlash)
					sub.Action()
				}
			}
		}
	}
	if c.Shortcut(ui.Ctrl, ui.KeyF2) && st.open < 0 && len(menus) > 0 {
		openMenu(0, true)
	}

	bar.HandleInput(func(ev ui.InputEvent) bool {
		wx, wy := st.rect.X+ev.X, st.rect.Y+ev.Y
		switch ev.Kind {
		case ui.InputPointerDown:
			if ev.Button != 0 {
				return false
			}
			i := titleAt(wx, wy)
			if i < 0 {
				return true
			}
			openMenu(i, false)
			if st.t == nil {
				// A title with no items, as a clock: its action, if any.
				if len(menus[i].Items) == 0 && menus[i].Action != nil {
					menus[i].Action()
				}
				return true
			}
			st.t.press(wallNow(), wx, wy)
			return true
		case ui.InputPointerMove:
			if st.t == nil || ev.Button != 0 {
				return false
			}
			if i := titleAt(wx, wy); i >= 0 && i != st.open {
				openMenu(i, false)
				st.t.pressed = true
			}
			st.t.move(wx, wy)
			return true
		case ui.InputPointerUp:
			if st.t == nil {
				return true
			}
			st.t.pressed = false
			if st.t.release(wallNow(), wx, wy) {
				return true
			}
			if titleAt(wx, wy) == st.open {
				st.sticky = true
				return true
			}
			closeMenu()
			return true
		}
		return false
	})

	if st.t != nil {
		if it, _, ok := st.t.done(now); ok {
			closeMenu()
			if it.Action != nil {
				st.run = it.Action
				c.After(0)
			}
		} else if st.t.closed {
			closeMenu()
		}
	}
	if st.t != nil {
		t := st.t
		trackerOverlay(c, t, th.Accent, func(ev ui.InputEvent) bool {
			switch ev.Kind {
			case ui.InputPointerDown:
				if i := titleAt(ev.X, ev.Y); i >= 0 {
					if i == st.open {
						closeMenu()
					} else {
						openMenu(i, false)
						st.t.pressed = true
						st.sticky = true
					}
					return true
				}
			case ui.InputPointerMove:
				if i := titleAt(ev.X, ev.Y); i >= 0 && i != st.open {
					pressed := t.pressed
					openMenu(i, false)
					st.t.pressed = pressed
					st.sticky = true
					return true
				}
			}
			return false
		}, func(k ui.Key) bool {
			switch k {
			case ui.KeyLeft, ui.KeyRight:
				if t.key(now, k) {
					return true
				}
				d := 1
				if k == ui.KeyLeft {
					d = -1
				}
				openMenu((st.open+d+len(menus))%len(menus), true)
				st.sticky = true
				return true
			}
			return false
		})
	}

	open, flash := st.open, -1
	if now.Before(st.flashUntil) && !frozen() {
		flash = st.flash
	}
	bar.Draw(func(p *ui.Painter, r ui.Rect) {
		resized(c, st.rect, r)
		st.rect = r
		pn := NewPen(p, r)
		hl := open
		if hl < 0 {
			hl = flash
		}
		DrawMenuBar(pn, menus, hl, th.Accent)
	})
	bar.Children(func() {
		for i, m := range menus {
			title := ui.Box(c.Key([2]any{i, m.Title})).Absolute().Left(float32(xs[i])).Top(0).Size(float32(ws[i]), MenuBarHeight).
				Role(ui.RoleMenuItem).Label(m.Title).Expanded(i == open).PassThrough()
			// Assistive technology presses the title: its menu opens as
			// from the keyboard. The pointer opens menus through the bar.
			if title.Clicked() && st.t == nil {
				if len(m.Items) == 0 && m.Action != nil {
					m.Action()
				}
				openMenu(i, true)
				st.sticky = st.t != nil
			}
		}
	})
	return bar
}

// menuTitles returns where each menu's title starts on a bar width wide,
// and how wide it is: from the left, and the Right ones from the right.
func menuTitles(menus []Menu, width int) (xs, ws []int) {
	xs, ws = make([]int, len(menus)), make([]int, len(menus))
	x := menuBarInset
	right := 0
	for i, m := range menus {
		w := titleWidth(m)
		ws[i] = w
		if m.Right {
			right += w
			continue
		}
		xs[i] = x
		x += w
	}
	x = width - menuBarInset - right
	for i, m := range menus {
		if m.Right {
			xs[i] = x
			x += ws[i]
		}
	}
	return xs, ws
}

func titleWidth(m Menu) int {
	w := SystemFont.Measure(m.Title)
	switch {
	case m.Icon != nil && m.IconTitle:
		w += iconWidth(m.Icon) + 4
	case m.Icon != nil:
		w = iconWidth(m.Icon)
	}
	return w + 2*MenuTitlePadding
}

func iconWidth(rows []string) int {
	w := 0
	for _, r := range rows {
		w = max(w, len(r))
	}
	return w
}

// DrawMenuBar draws a menu bar filling pn's box, with the title of menu hl
// highlighted (-1 for none).
func DrawMenuBar(pn Pen, menus []Menu, hl int, a Accent) {
	w, h := pn.W, pn.H
	pn.Rect(0, 0, w, h, Gray2)
	pn.HLine(0, 0, w, White)
	pn.HLine(0, h-1, w, Black)
	xs, ws := menuTitles(menus, w)
	for i, m := range menus {
		text := Black
		if i == hl && len(m.Items) > 0 {
			pn.Rect(xs[i], 0, ws[i], h-1, a.Dark)
			text = White
		}
		x := xs[i] + MenuTitlePadding
		if m.Icon != nil {
			ih := len(m.Icon)
			DrawIcon(pn.Sub(x, (h-1-ih)/2, iconWidth(m.Icon), ih), m.Icon, a, i == hl)
			if !m.IconTitle {
				continue
			}
			x += iconWidth(m.Icon) + 4
		}
		pn.TextMiddle(x, 0, h-1, m.Title, SystemFont, text)
	}
}

// DrawIcon draws a small picture given as rows of '#' (black), 'o' (the
// accent's dark color), '+' (the accent's mid color), ':' (gray), 'w'
// (white) and other characters (nothing). Inverted, black and white swap,
// for an icon on a highlight.
func DrawIcon(pn Pen, rows []string, a Accent, inverted bool) {
	for y, row := range rows {
		for x := 0; x < len(row); x++ {
			var c ui.Color
			switch row[x] {
			case '#':
				c = Black
				if inverted {
					c = White
				}
			case 'w':
				c = White
				if inverted {
					c = Black
				}
			case 'o':
				c = a.Dark
				if inverted {
					c = a.Light
				}
			case '+':
				c = a.Mid
			case ':':
				c = Gray7
			default:
				continue
			}
			pn.Dot(x, y, c)
		}
	}
}
