package main

import (
	"strings"

	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
)

// finder is the Find File window: a search of the disk by name or kind,
// as "Find items on local disks whose name contains …", run by its Find
// button.
type finder struct {
	where, what, how int // the choices of its pop-up menus
	query            string
	found            []*node // what the last search found
	searched         bool
	table            *pt.TableState
	said             string // what screen readers were last told was found
}

var (
	findWhere = []string{"on local disks", "on “Platinum Disk”"}
	findWhat  = []string{"name", "kind"}
	findHow   = []string{"contains", "starts with", "ends with", "is"}
)

func newFindWindow(d *desk) *win {
	f := &finder{table: pt.NewTableState()}
	f.table.Widths = []int{150, 100, 182}
	f.table.SortColumn = 0
	f.table.ScrollBarAlways = true
	w := &win{app: appFind, title: "Find File"}
	w.frame = pt.WindowState{W: 470, H: 270}
	w.content = func(c *ui.Context, w *win) { findView(c, d, f) }
	return w
}

// matches returns the nodes of the disk whose name or kind holds the
// query as the menus say.
func (f *finder) matches(root *node) []*node {
	q := strings.ToLower(strings.TrimSpace(f.query))
	if q == "" {
		return nil
	}
	var out []*node
	var walk func(n *node)
	walk = func(n *node) {
		for _, ch := range n.children {
			s := strings.ToLower(ch.name)
			if f.what == 1 {
				s = ch.kindName()
			}
			ok := false
			switch f.how {
			case 0:
				ok = strings.Contains(s, q)
			case 1:
				ok = strings.HasPrefix(s, q)
			case 2:
				ok = strings.HasSuffix(s, q)
			case 3:
				ok = s == q
			}
			if ok {
				out = append(out, ch)
			}
			walk(ch)
		}
	}
	walk(root)
	return out
}

func findView(c *ui.Context, d *desk, f *finder) {
	found := f.found
	// What the search found is news the focus does not bring: screen
	// readers read it out.
	news := ""
	if f.searched {
		news = plural(len(found), "item") + " found"
	}
	if news != f.said {
		if news != "" {
			c.Announce(news)
		}
		f.said = news
	}
	ui.Column(c).Fill().AlignItems(ui.Stretch).Children(func() {
		ui.Column(c).Padding(10, 10, 8, 10).Gap(6).AlignItems(ui.Stretch).Children(func() {
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				pt.Label(c, "Find items")
				pt.PopupMenu(c, &f.where, "", findWhere).Element().Label("Where")
			})
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				pt.Label(c, "whose")
				pt.PopupMenu(c, &f.what, "", findWhat).Element().Label("What")
				pt.PopupMenu(c, &f.how, "", findHow).Element().Label("How")
				field := pt.TextField(c, &f.query).Width(110).Label("Search for")
				field.Element().AutoFocus()
				ui.Box(c).Grow(1)
				if pt.Button(c, "Find").Width(60).Default().Disabled(strings.TrimSpace(f.query) == "").Clicked() {
					f.found, f.searched = f.matches(d.disk), true
				}
			})
		})
		if !f.searched {
			ui.Column(c).Grow(1).Padding(0, 10, 0, 10).Children(func() {
				ui.Box(c).Grow(1).AlignSelf(ui.Stretch).Draw(func(p *ui.Painter, r ui.Rect) {
					pn := pt.NewPen(p, r)
					pt.DrawContentRing(pn)
					pn.Frame(1, 1, pn.W-2, pn.H-2, pt.Black)
					pn.Rect(2, 2, pn.W-4, pn.H-4, pt.White)
				}).Children(func() {
					ui.Column(c).Fill().Children(func() { pt.EmptyState(c, "Choose what to find, and click Find.") })
				})
			})
		} else {
			ui.Column(c).Grow(1).Padding(0, 10, 0, 10).AlignItems(ui.Stretch).Children(func() {
				cols := []pt.Column{{Title: "Name", Icon: func(r int) []string { return smallIcon(found[r]) }}, {Title: "Kind"}, {Title: "Where"}}
				t := pt.Table(c, f.table, cols, len(found), func(r, col int) string {
					n := found[r]
					switch col {
					case 0:
						return n.name
					case 1:
						return n.kindName()
					}
					if n.parent != nil {
						return n.parent.path()
					}
					return ""
				}).Label("Items found")
				t.Element().Grow(1).FillWidth()
				if t.Activated() {
					if r := t.Cursor(); r >= 0 && r < len(found) {
						d.openNode(found[r])
					}
				}
			})
		}
		// The count, its right end against the resize corner.
		ui.Row(c).Height(pt.GrowBoxSize).Margin(8, 0, 0, 0).AlignItems(ui.Stretch).Children(func() {
			count := ""
			if f.searched {
				count = plural(len(found), "item") + " found"
			}
			pt.StatusBar(c, count).Grow(1)
			ui.Box(c).Width(pt.GrowBoxSize)
		})
	})
}

// smallIcon is a node's picture at 16 pixels, for lists of them.
func smallIcon(n *node) []string {
	switch n.kind {
	case kindFolder:
		return smallFolder
	case kindDocument:
		return smallDocument
	case kindApp, kindPanel:
		_, pic := appName(&win{app: n.app})
		if n.kind == kindPanel {
			pic = miniPanel
		}
		return pic
	}
	return nil
}

var (
	smallFolder = []string{
		"................",
		"................",
		".####...........",
		"#wwww#..........",
		"###############.",
		"#wwwwwwwwwwwww#.",
		"#w++++++++++++#.",
		"#w++++++++++++#.",
		"#w++++++++++++#.",
		"#w++++++++++++#.",
		"#w++++++++++++#.",
		"#w++++++++++++#.",
		"#+++++++++++++#.",
		"###############.",
	}
	smallDocument = []string{
		"..########......",
		"..#wwwwww##.....",
		"..#wwwwww#w#....",
		"..#wwwwww####...",
		"..#wwwwwwwww#...",
		"..#w::::::ww#...",
		"..#wwwwwwwww#...",
		"..#w:::::::w#...",
		"..#wwwwwwwww#...",
		"..#w::::::ww#...",
		"..#wwwwwwwww#...",
		"..#w:::::::w#...",
		"..#wwwwwwwww#...",
		"..###########...",
	}
)

func plural(n int, what string) string {
	if n == 1 {
		return "1 " + what
	}
	return strings.Join([]string{itoa(n), what + "s"}, " ")
}

func itoa(n int) string {
	s := ""
	if n == 0 {
		return "0"
	}
	for ; n > 0; n /= 10 {
		s = string(rune('0'+n%10)) + s
	}
	return s
}
