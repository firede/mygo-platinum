package platinum

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Menu metrics, from docs/spec.md.
const (
	MenuBarHeight       = 20 // the height of the menu bar
	MenuTitlePadding    = 9  // the room on each side of a menu's title
	MenuItemHeight      = 18 // the height of a menu item
	MenuSeparatorHeight = 8  // the height of a separator between items
	menuLeft            = 18 // from a menu's edge to its items' titles
	menuRight           = 12 // from the end of the longest title to the edge
	menuShortcutGap     = 20 // between an item's title and its shortcut
	menuBorder          = 2  // the frame and bevel around the items
)

// MenuFlash is how long a chosen menu item blinks before its action runs,
// as the classic menus confirm a choice.
const MenuFlash = 150 * time.Millisecond

// MenuItem is an item of a menu, or a separator.
type MenuItem struct {
	// Title is what the item says.
	Title string
	// Key, with Mods, is the item's keyboard equivalent, which chooses it
	// wherever the focus is; Mods of 0 is ui.Cmd. The menu shows it at the
	// item's right.
	Key  ui.Key
	Mods ui.Modifiers
	// Disabled dims the item, which the user cannot choose.
	Disabled bool
	// Checked shows a check mark before the title.
	Checked bool
	// Separator makes the item an etched line between groups of items.
	Separator bool
	// Submenu makes the item open a menu of its own, to its right.
	Submenu []MenuItem
	// Action runs when the user chooses the item, in a frame after the
	// one that built the menu: it should change the app's state, not
	// variables of the view that built it.
	Action func()
}

// SeparatorItem returns a separator, for a menu's list of items, as MyGo's
// menus have Separator; Separator draws one in a window.
func SeparatorItem() MenuItem { return MenuItem{Separator: true} }

// Menu is a menu of a menu bar.
type Menu struct {
	// Title is the menu's title on the bar.
	Title string
	// Icon shows a small picture in place of the title, given as rows of
	// '#' (black), 'o' (the accent's dark color), ':' (gray) and '.' or ' '
	// (nothing), at most 16 rows tall. Title still names it for assistive
	// technology.
	Icon []string
	// IconTitle shows the title beside the icon, as an application menu
	// shows the application's name.
	IconTitle bool
	// Right places the menu at the right end of the bar, after those
	// before it there, as the clock and the application menu; the last is
	// rightmost.
	Right bool
	// Items are the menu's items. A menu without items shows its title
	// alone and does not open, as a clock.
	Items []MenuItem
	// Action runs when the user clicks the title of a menu without items,
	// as a clock clicked to show the date.
	Action func()
}

func (it MenuItem) mods() ui.Modifiers {
	if it.Mods == 0 {
		return ui.Cmd
	}
	return it.Mods
}

// MacKeys is whether menus show the keys of shortcuts as a Mac's keyboard
// marks them, as ⌘N, or by their names, as Ctrl+N, as other systems do.
// It starts as the system's, in a browser the system the browser runs on;
// tests set it, to draw the same everywhere.
var MacKeys = macKeyboard()

// shortcutText returns how a menu shows the keyboard equivalent.
func (it MenuItem) shortcutText() string {
	if it.Key == 0 {
		return ""
	}
	m := it.mods()
	if !MacKeys {
		// Cmd is Control where keys go by name.
		if m&ui.Cmd != 0 {
			m = m&^ui.Cmd | ui.Ctrl
		}
		s := ""
		for _, mod := range []struct {
			mod  ui.Modifiers
			name string
		}{{ui.Ctrl, "Ctrl+"}, {ui.Super, "Super+"}, {ui.Alt, "Alt+"}, {ui.Shift, "Shift+"}} {
			if m&mod.mod != 0 {
				s += mod.name
			}
		}
		return s + keyName(it.Key)
	}
	s := ""
	if m&ui.Ctrl != 0 && m&ui.Cmd == 0 {
		s += "^"
	}
	if m&ui.Alt != 0 {
		s += "~"
	}
	if m&ui.Shift != 0 {
		s += "↑"
	}
	if m&ui.Cmd != 0 {
		s += "⌘"
	}
	return s + keyName(it.Key)
}

// keyName returns how a menu writes k: by a symbol on a Mac, else by
// name.
func keyName(k ui.Key) string {
	switch {
	case k >= ui.KeyA && k <= ui.KeyZ:
		return string(rune('A' + int(k-ui.KeyA)))
	case k >= ui.Key0 && k <= ui.Key9:
		return string(rune('0' + int(k-ui.Key0)))
	}
	switch k {
	case ui.KeyComma:
		return ","
	case ui.KeyPeriod:
		return "."
	case ui.KeySlash:
		return "/"
	case ui.KeyMinus:
		return "-"
	case ui.KeyEqual:
		return "="
	case ui.KeyBracketLeft:
		return "["
	case ui.KeyBracketRight:
		return "]"
	case ui.KeyEnter:
		if !MacKeys {
			return "Enter"
		}
		return "↵"
	case ui.KeyBackspace:
		if !MacKeys {
			return "Backspace"
		}
		return "⌫"
	case ui.KeyDelete:
		if !MacKeys {
			return "Del"
		}
		return "⌫"
	}
	return "?"
}

// menuGeom is where the items of a menu panel are, in the window.
type menuGeom struct {
	X, Y, W, H int
	tops       []int // the top of each item, from Y
}

func layoutMenu(items []MenuItem, x, y, minW int) menuGeom {
	g := menuGeom{X: x, Y: y}
	titleW, keyW := 0, 0
	sub := false
	cy := menuBorder
	for _, it := range items {
		g.tops = append(g.tops, cy)
		if it.Separator {
			cy += MenuSeparatorHeight
			continue
		}
		cy += MenuItemHeight
		titleW = max(titleW, SystemFont.Measure(it.Title))
		if s := it.shortcutText(); s != "" {
			keyW = max(keyW, SystemFont.Measure(s))
		}
		if it.Submenu != nil {
			sub = true
		}
	}
	w := menuLeft + titleW + menuRight
	if keyW > 0 {
		w += menuShortcutGap + keyW
	} else if sub {
		w += menuShortcutGap
	}
	g.W = max(w+2*menuBorder, minW)
	g.H = cy + menuBorder
	return g
}

// itemAt returns the item under the window point x, y, or -1.
func (g menuGeom) itemAt(items []MenuItem, x, y float32) int {
	if x < float32(g.X) || x >= float32(g.X+g.W) {
		return -1
	}
	for i, it := range items {
		top := float32(g.Y + g.tops[i])
		h := float32(MenuItemHeight)
		if it.Separator {
			h = MenuSeparatorHeight
		}
		if y >= top && y < top+h {
			return i
		}
	}
	return -1
}

func (g menuGeom) contains(x, y float32) bool {
	return x >= float32(g.X) && x < float32(g.X+g.W) && y >= float32(g.Y) && y < float32(g.Y+g.H)
}

// fitMenu moves a panel inside a window of w×h.
func fitMenu(g menuGeom, w, h int) menuGeom {
	if g.X+g.W+1 > w {
		g.X = max(0, w-g.W-1)
	}
	if g.Y+g.H+1 > h {
		g.Y = max(0, h-g.H-1)
	}
	return g
}

// DrawMenuPanel draws a menu of items at the panel's corner of pn, with
// item hl highlighted (-1 for none), as wide as w.
func DrawMenuPanel(pn Pen, items []MenuItem, hl, w int, a Accent) {
	g := layoutMenu(items, 0, 0, w)
	w, h := g.W, g.H
	// A shadow a pixel deep along the right and bottom.
	pn.Rect(2, h, w-1, 1, Gray9.Alpha(0.6))
	pn.Rect(w, 2, 1, h-1, Gray9.Alpha(0.6))
	pn.Rect(0, 0, w, h, Gray2)
	pn.Frame(0, 0, w, h, Black)
	pn.Bevel(1, 1, w-2, h-2, White, Gray7)
	for i, it := range items {
		top := g.tops[i]
		if it.Separator {
			pn.HLine(menuBorder, top+3, w-2*menuBorder, Gray7)
			pn.HLine(menuBorder, top+4, w-2*menuBorder, White)
			continue
		}
		DrawMenuItem(pn.Sub(menuBorder, top, w-2*menuBorder, MenuItemHeight), it, i == hl, a)
	}
}

var submenuArrow = []string{
	"#...",
	"##..",
	"###.",
	"####",
	"###.",
	"##..",
	"#...",
}

// DrawMenuItem draws an item of a menu filling pn's box.
func DrawMenuItem(pn Pen, it MenuItem, highlighted bool, a Accent) {
	text := Black
	if it.Disabled {
		text = Gray8
	}
	if highlighted && !it.Disabled {
		pn.Rect(0, 0, pn.W, pn.H, a.Dark)
		text = White
	}
	if it.Checked {
		pn.TextMiddle(4, 0, pn.H, "✓", SystemFont, text)
	}
	pn.TextMiddle(menuLeft-menuBorder, 0, pn.H, it.Title, SystemFont, text)
	if s := it.shortcutText(); s != "" {
		sw := SystemFont.Measure(s)
		pn.TextMiddle(pn.W-menuRight+menuBorder-sw, 0, pn.H, s, SystemFont, text)
	} else if it.Submenu != nil {
		pn.Bitmap(pn.W-menuRight, (pn.H-7)/2, submenuArrow, text)
	}
}

// menuTracker follows the pointer and the keys through an open menu and
// its submenu, which a menu bar, a pop-up menu and a context menu share.
type menuTracker struct {
	items   []MenuItem
	geom    menuGeom
	hl      int
	sub     int // the item whose submenu shows, or -1
	subGeom menuGeom
	subHl   int
	pressed bool // the pointer went down and has not come up
	// Where and when the press that opened the menu went down, and whether
	// the pointer moved away from there since: a quick click leaves the
	// menu open.
	downX, downY float32
	downAt       time.Time
	moved        bool
	keyboard     bool // the keys move the highlight
	builds       int  // how many builds the menu has shown for
	// The choice blinking before it is made.
	chosen   bool
	chosenAt time.Time
	path     [2]int
	closed   bool
}

func newTracker(items []MenuItem, g menuGeom) *menuTracker {
	return &menuTracker{items: items, geom: g, hl: -1, sub: -1, subHl: -1}
}

// subItems returns the submenu showing.
func (t *menuTracker) subItems() []MenuItem {
	if t.sub < 0 {
		return nil
	}
	return t.items[t.sub].Submenu
}

func (t *menuTracker) openSub(i int) {
	if t.sub == i {
		return
	}
	t.sub, t.subHl = -1, -1
	if i >= 0 && t.items[i].Submenu != nil && !t.items[i].Disabled {
		t.sub = i
		t.subGeom = layoutMenu(t.items[i].Submenu, t.geom.X+t.geom.W-3, t.geom.Y+t.geom.tops[i]-menuBorder, 0)
	}
}

// move follows the pointer at x, y of the window.
func (t *menuTracker) move(x, y float32) {
	if t.chosen {
		return
	}
	if dx, dy := x-t.downX, y-t.downY; dx*dx+dy*dy > 9 {
		t.moved = true
	}
	if sub := t.subItems(); sub != nil && t.subGeom.contains(x, y) {
		t.subHl = t.subGeom.itemAt(sub, x, y)
		return
	}
	t.hl = t.geom.itemAt(t.items, x, y)
	if t.hl >= 0 {
		t.openSub(t.hl)
		t.subHl = -1
	} else if t.sub >= 0 {
		t.hl = t.sub
	}
}

// press records where the press that opened the menu went down.
func (t *menuTracker) press(now time.Time, x, y float32) {
	t.pressed, t.downX, t.downY, t.downAt, t.moved = true, x, y, now, false
}

// clickHold is how quickly a click must come up for the menu it opened to
// stay open.
const clickHold = 400 * time.Millisecond

// quickClick reports whether a release at now ends a click that should
// leave the menu open.
func (t *menuTracker) quickClick(now time.Time) bool {
	return !t.moved && now.Sub(t.downAt) < clickHold
}

// release chooses the item under x, y, if any, and reports whether it did.
func (t *menuTracker) release(now time.Time, x, y float32) bool {
	t.move(x, y)
	if sub := t.subItems(); sub != nil && t.subHl >= 0 {
		return t.choose(now, t.sub, t.subHl)
	}
	if t.hl >= 0 && t.items[t.hl].Submenu == nil {
		return t.choose(now, t.hl, -1)
	}
	return false
}

func (t *menuTracker) choose(now time.Time, i, j int) bool {
	it := t.items[i]
	if j >= 0 {
		it = it.Submenu[j]
	}
	if it.Separator || it.Disabled || it.Submenu != nil {
		return false
	}
	t.chosen, t.chosenAt, t.path = true, now, [2]int{i, j}
	return true
}

// key handles a key of the keyboard while the menu is open, and reports
// whether it took it.
func (t *menuTracker) key(now time.Time, k ui.Key) bool {
	if t.chosen {
		return true
	}
	t.keyboard = true
	step := func(items []MenuItem, i, d int) int {
		for n := 0; n < len(items); n++ {
			i = (i + d + len(items)) % len(items)
			if !items[i].Separator && !items[i].Disabled {
				return i
			}
		}
		return -1
	}
	switch k {
	case ui.KeyDown, ui.KeyUp:
		d := 1
		if k == ui.KeyUp {
			d = -1
		}
		if sub := t.subItems(); sub != nil && t.subHl >= 0 {
			t.subHl = step(sub, t.subHl, d)
			return true
		}
		start := t.hl
		if start < 0 && d < 0 {
			start = 0
		} else if start < 0 {
			start = len(t.items) - 1
		}
		t.hl = step(t.items, start, d)
		if t.sub >= 0 && t.sub != t.hl {
			t.openSub(-1)
		}
		return true
	case ui.KeyRight:
		if t.hl >= 0 && t.items[t.hl].Submenu != nil && t.subHl < 0 {
			t.openSub(t.hl)
			t.subHl = step(t.subItems(), -1, 1)
			return true
		}
		return false
	case ui.KeyLeft:
		if t.subHl >= 0 {
			t.subHl = -1
			t.openSub(-1)
			return true
		}
		return false
	case ui.KeyEnter, ui.KeySpace:
		if t.subHl >= 0 {
			t.choose(now, t.sub, t.subHl)
		} else if t.hl >= 0 {
			if t.items[t.hl].Submenu != nil {
				return t.key(now, ui.KeyRight)
			}
			t.choose(now, t.hl, -1)
		}
		return true
	}
	return false
}

// done reports whether the blink of the choice is over, and returns the
// item chosen.
func (t *menuTracker) done(now time.Time) (MenuItem, [2]int, bool) {
	if !t.chosen || (now.Sub(t.chosenAt) < MenuFlash && !frozen()) {
		return MenuItem{}, [2]int{}, false
	}
	it := t.items[t.path[0]]
	if t.path[1] >= 0 {
		it = it.Submenu[t.path[1]]
	}
	return it, t.path, true
}

// blinkOff reports whether the chosen item shows unhighlighted now, in
// the middle of its blink.
func (t *menuTracker) blinkOff(now time.Time) bool {
	if !t.chosen {
		return false
	}
	phase := now.Sub(t.chosenAt) * 4 / MenuFlash
	return phase%2 == 0
}

// draw draws the open panels.
func (t *menuTracker) draw(p *ui.Painter, a Accent) {
	now := paintNow(p)
	hl, subHl := t.hl, t.subHl
	if t.chosen && t.blinkOff(now) {
		if t.path[1] >= 0 {
			subHl = -1
		} else {
			hl = -1
		}
	}
	pn := Pen{P: p}
	DrawMenuPanel(pn.Sub(t.geom.X, t.geom.Y, t.geom.W, t.geom.H), t.items, hl, t.geom.W, a)
	if sub := t.subItems(); sub != nil {
		DrawMenuPanel(pn.Sub(t.subGeom.X, t.subGeom.Y, t.subGeom.W, t.subGeom.H), sub, subHl, t.subGeom.W, a)
	}
}

// accessibleItems builds an element for each item over where it shows,
// which assistive technology reads and tests find; the pointer passes
// through them to the tracker.
func accessibleItems(c *ui.Context, t *menuTracker, items []MenuItem, g menuGeom, hl int, path [1]int) {
	ui.Column(c).Absolute().Left(float32(g.X)).Top(float32(g.Y)).Size(float32(g.W), float32(g.H)).
		Role(ui.RoleMenu).PassThrough().Children(func() {
		for i, it := range items {
			if it.Separator {
				ui.Box(c).Absolute().Top(float32(g.tops[i])).Height(MenuSeparatorHeight).Role(ui.RoleNone).PassThrough()
				continue
			}
			e := ui.Box(c.Key(it.Title)).Absolute().Left(menuBorder).Top(float32(g.tops[i])).
				Size(float32(g.W-2*menuBorder), MenuItemHeight).
				Role(ui.RoleMenuItem).Label(it.Title).PassThrough().Disabled(it.Disabled)
			if it.Checked {
				e.Role(ui.RoleMenuItemCheckBox).Checked(true)
			}
			if it.Submenu != nil {
				e.Expanded(i == hl)
			}
			// Assistive technology presses the item. The pointer never
			// reaches it, passing through to the menu's layer; and a menu
			// just opened takes no press, which is the pointer's that
			// opened it.
			if e.Clicked() && t.builds > 1 && !t.pressed && !t.chosen {
				j := i
				if path[0] < 0 {
					if it.Submenu != nil {
						t.hl = j
						t.openSub(j)
					} else {
						t.choose(frameNow(c), j, -1)
					}
				} else {
					t.choose(frameNow(c), path[0], j)
				}
			}
		}
	})
}

// trackerOverlay builds the layer that shows an open menu over the
// window and takes the pointer and the keys while it is open. onPointer
// may take pointer input first, as the menu bar does for its titles; it
// returns whether it did.
func trackerOverlay(c *ui.Context, t *menuTracker, a Accent, onPointer func(ev ui.InputEvent) bool, onKey func(k ui.Key) bool) {
	t.builds++
	ui.Overlay(c, func() {
		// The layer keeps the keyboard while the menu is open, as the
		// classic menus did: shortcuts of the window wait.
		layer := ui.Box(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).Focusable().FocusRing(false).Modal()
		shown(c, layer)
		layer.HandleInput(func(ev ui.InputEvent) bool {
			if onPointer != nil && onPointer(ev) {
				return true
			}
			now := eventNow(ev)
			switch ev.Kind {
			case ui.InputPointerMove:
				t.keyboard = false
				t.move(ev.X, ev.Y)
				return true
			case ui.InputPointerDown:
				in := t.geom.contains(ev.X, ev.Y) || t.subItems() != nil && t.subGeom.contains(ev.X, ev.Y)
				if !in {
					t.closed = true
					return true
				}
				t.pressed = true
				t.move(ev.X, ev.Y)
				return true
			case ui.InputPointerUp:
				t.pressed = false
				t.release(now, ev.X, ev.Y)
				return true
			case ui.InputScroll:
				return true
			}
			return false
		})
		for _, k := range []ui.Key{ui.KeyUp, ui.KeyDown, ui.KeyLeft, ui.KeyRight, ui.KeyEnter, ui.KeySpace} {
			if layer.Shortcut(0, k) || layer.OverlayShortcut(0, k) {
				if onKey == nil || !onKey(k) {
					t.key(frameNow(c), k)
				}
			}
		}
		if layer.Shortcut(0, ui.KeyEscape) || layer.OverlayShortcut(0, ui.KeyEscape) || layer.Shortcut(ui.Cmd, ui.KeyPeriod) {
			t.closed = true
		}
		layer.Draw(func(p *ui.Painter, r ui.Rect) { t.draw(p, a) })
		layer.Children(func() {
			accessibleItems(c, t, t.items, t.geom, t.hl, [1]int{-1})
			if sub := t.subItems(); sub != nil {
				accessibleItems(c, t, sub, t.subGeom, t.subHl, [1]int{t.sub})
			}
		})
		if t.chosen {
			c.After(MenuFlash / 4)
		}
	})
}

// ContextMenu gives target a context menu of items, which opens at the
// pointer on a click of the secondary button, or a Control-click, and
// chooses as the menus of a menu bar do. Shift-F10 opens it from the
// keyboard while target has the focus.
//
// The counterpart among MyGo's own controls is Element.ContextMenu.
func ContextMenu(c *ui.Context, target ui.Element, items []MenuItem) {
	th := ThemeOf(c)
	st := ui.Local(target, "contextmenu", func() bevelState { return bevelState{} })
	cw, ch := c.Size()
	open := func(x, y float32, keyboard bool) {
		g := layoutMenu(items, round(x), round(y), 0)
		st.t = newTracker(items, fitMenu(g, round(cw), round(ch)))
		st.t.keyboard = keyboard
		if keyboard {
			st.t.key(frameNow(c), ui.KeyDown)
		}
	}
	px, py, _ := target.PointerPosition()
	// A Control-press opens it too, without asking target for clicks, which
	// would make Return press it.
	if target.RightClicked() || st.t == nil && target.Pressed() && c.Modifiers()&ui.Ctrl != 0 {
		wx, wy := target.LocalToWindow(px, py)
		open(wx, wy, false)
	}
	if target.Shortcut(ui.Shift, ui.KeyF10) {
		wx, wy := target.LocalToWindow(4, 4)
		open(wx, wy, true)
	}
	if st.t != nil {
		if it, _, ok := st.t.done(frameNow(c)); ok {
			st.t = nil
			if it.Action != nil {
				it.Action()
				report(target, true)
			}
		} else if st.t.closed {
			st.t = nil
		}
	}
	if st.t != nil {
		trackerOverlay(c, st.t, th.Accent, nil, nil)
	}
}
