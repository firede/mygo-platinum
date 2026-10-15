package platinum

import (
	"github.com/egoist/mygo/ui"
)

// ToolItem is an item of a toolbar: a bevel button, or a separator.
type ToolItem struct {
	// Label names the item; it shows beside the icon, or alone, unless
	// IconOnly. Icon is a small picture for DrawIcon.
	Label    string
	Icon     []string
	IconOnly bool
	// On makes the item a toggle, pressed in while *On.
	On       *bool
	Disabled bool
	// Action runs when the user clicks the item.
	Action func()
	// Separator makes the item an etched line between groups.
	Separator bool
}

const (
	ToolbarHeight = 32 // the height of a toolbar
	toolbarGap    = 4
	toolbarSepW   = 6
	overflowW     = 22
)

type toolbarState struct{ width float32 }

// itemWidth returns how wide a toolbar item is.
func (it ToolItem) width() int {
	if it.Separator {
		return toolbarSepW
	}
	label := it.Label
	if it.IconOnly {
		label = ""
	}
	w, _ := bevelContentSize(label, it.Icon, false)
	w += 2 * bevelPadding
	if label == "" && it.Icon != nil {
		w = max(BevelHeight, max(iconWidth(it.Icon), len(it.Icon))+6)
	}
	return w
}

// Toolbar builds a row of bevel buttons along the top of a window, on a
// face with an etched line under it, one stop of Tab whose buttons the
// arrows move between. Items that do not fit go into a menu that a »
// button at its end opens.
func Toolbar(c *ui.Context, items []ToolItem) ui.Element {
	bar := ui.Row(c).Height(ToolbarHeight).Padding(0, 6).Gap(toolbarGap).AlignItems(ui.Center).AlignSelf(ui.Stretch).
		Role(ui.RoleToolbar).FocusGroup(ui.Horizontal).Clip()
	st := ui.Local(bar, "toolbar", func() toolbarState { return toolbarState{} })
	bar.Draw(func(p *ui.Painter, r ui.Rect) {
		resized(c, ui.Rect{W: st.width, H: r.H}, r)
		st.width = r.W
		pn := NewPen(p, r)
		pn.Rect(0, 0, pn.W, pn.H, Gray2)
		pn.HLine(0, 0, pn.W, White)
		pn.HLine(0, pn.H-2, pn.W, Gray7)
		pn.HLine(0, pn.H-1, pn.W, White)
	})
	// How many items fit, leaving room for the » button when some do not.
	room := int(st.width) - 12
	shown := len(items)
	if st.width > 0 {
		total := 0
		for i, it := range items {
			total += it.width() + toolbarGap
			if total > room {
				shown = i
				break
			}
		}
		if shown < len(items) {
			total = 0
			for i, it := range items {
				total += it.width() + toolbarGap
				if total > room-overflowW-toolbarGap {
					shown = i
					break
				}
			}
		}
	}
	bar.Children(func() {
		for i, it := range items[:shown] {
			if it.Separator {
				ToolbarSeparator(c.Key(i))
				continue
			}
			var b *BevelWidget
			if it.IconOnly {
				b = IconButton(c.Key(i), it.Icon, it.Label)
			} else {
				b = BevelButton(c.Key(i), it.Label).Icon(it.Icon)
			}
			if it.On != nil {
				b.Toggle(it.On)
			}
			b.Disabled(it.Disabled)
			if b.Clicked() && it.Action != nil {
				it.Action()
			}
		}
		if shown < len(items) {
			var menu []MenuItem
			for _, it := range items[shown:] {
				if it.Separator {
					menu = append(menu, Divider())
					continue
				}
				mi := MenuItem{Title: it.Label, Disabled: it.Disabled, Action: it.Action}
				if it.On != nil {
					on := it.On
					mi.Checked = *on
					act := it.Action
					mi.Action = func() {
						*on = !*on
						if act != nil {
							act()
						}
					}
				}
				menu = append(menu, mi)
			}
			ui.Box(c).Grow(1)
			more := BevelButton(c.Key("more"), "»")
			more.noArrow = true
			more.Size(overflowW, BevelHeight).Menu(menu).Element().Label(say(c, "more"))
		}
	})
	return bar
}

// ToolbarSeparator builds an etched vertical line between groups of
// controls of a toolbar.
func ToolbarSeparator(c *ui.Context) ui.Element {
	return ui.Box(c).Size(toolbarSepW, 22).Role(ui.RoleNone).Draw(func(p *ui.Painter, r ui.Rect) {
		pn := NewPen(p, r)
		pn.VLine(2, 0, pn.H, Gray7)
		pn.VLine(3, 0, pn.H, White)
	})
}
