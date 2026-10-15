package platinum

import (
	"github.com/egoist/mygo/ui"
)

var pathFolder = []string{
	"###.......",
	"#..######.",
	"#.......#.",
	"#.......#.",
	"#.......#.",
	"#########.",
}

// PathPopup builds a bevel button showing the last of path, a folder's
// name, whose menu lists the folders it is in, from it up to the outermost,
// as the path menu of a window's title. It returns the index in path of the
// folder chosen, or -1.
func PathPopup(c *ui.Context, path []string) int {
	if len(path) == 0 {
		return -1
	}
	b := BevelButton(c, path[len(path)-1]).Icon(pathFolder)
	// The menu keeps the items of the frame it opened in: their actions
	// write where every frame reads.
	chosen := ui.Local(b.Element(), "path", func() int { return -1 })
	var items []MenuItem
	for i := len(path) - 1; i >= 0; i-- {
		items = append(items, MenuItem{Title: path[i], Checked: i == len(path)-1, Action: func() { *chosen = i }})
	}
	b.Menu(items)
	got := *chosen
	*chosen = -1
	return got
}

var crumbArrow = []string{"#..", "##.", "###", "##.", "#.."}

// Breadcrumbs builds a row of the folders of path, the outermost first,
// with arrows between them; a click on one, or Space or Return while it has
// the keyboard focus, goes there. It returns the index of the one chosen,
// or -1.
//
// The counterpart among MyGo's own controls is ui.Breadcrumbs.
func Breadcrumbs(c *ui.Context, path []string) int {
	th := ThemeOf(c)
	chosen := -1
	ui.Row(c).Gap(2).AlignItems(ui.Center).Role(ui.RoleGroup).Label(say(c, "path")).FocusGroup(ui.Horizontal).Children(func() {
		for i, name := range path {
			if i > 0 {
				ui.Box(c).Size(7, 12).Role(ui.RoleNone).Draw(func(p *ui.Painter, r ui.Rect) {
					NewPen(p, r).Bitmap(2, 3, crumbArrow, Gray8)
				})
			}
			last := i == len(path)-1
			b := ui.ButtonBase(c).FocusRing(false).Label(name).Size(float32(SystemFont.Measure(name)+8), 18)
			pressed := b.Pressed()
			focused := b.FocusVisible()
			b.Draw(func(p *ui.Painter, r ui.Rect) {
				pn := NewPen(p, r)
				ink := Black
				if pressed {
					pn.Rect(0, 0, pn.W, pn.H, th.Accent.Dark)
					ink = White
				} else if !last {
					ink = Gray9
				}
				pn.TextCentered(0, 0, pn.W, pn.H, name, SystemFont, ink)
				if focused {
					pn.Rect(3, pn.H-2, pn.W-6, 2, th.Accent.Dark)
				}
			})
			if b.Clicked() {
				chosen = i
			}
		}
	})
	return chosen
}

// Pagination builds the arrows and the "Page n of m" between them that move
// *page (from 0) through pages pages; Left and Right do too while either
// arrow has the keyboard focus.
func Pagination(c *ui.Context, page *int, pages int) ui.Element {
	*page = min(max(*page, 0), max(pages-1, 0))
	row := ui.Row(c).Gap(8).AlignItems(ui.Center).Role(ui.RoleGroup).Label(say(c, "pages")).
		Value(sayWith(c, "pageOf", map[string]any{"page": *page + 1, "pages": pages}))
	row.Children(func() {
		back := BevelButton(c, "").Icon(monthBack).Size(BevelHeight, BevelHeight).Disabled(*page == 0)
		back.Element().Label(say(c, "previousPage"))
		if back.Clicked() || row.Shortcut(0, ui.KeyLeft) && *page > 0 {
			*page--
		}
		TextIn(c, sayWith(c, "pageOf", map[string]any{"page": *page + 1, "pages": pages}), SystemFont, Black)
		next := BevelButton(c, "").Icon(monthNext).Size(BevelHeight, BevelHeight).Disabled(*page >= pages-1)
		next.Element().Label(say(c, "nextPage"))
		if next.Clicked() || row.Shortcut(0, ui.KeyRight) && *page < pages-1 {
			*page++
		}
	})
	return row
}
