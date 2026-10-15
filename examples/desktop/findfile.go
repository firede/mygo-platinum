package main

import (
	"strings"

	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
	"github.com/firede/mygo-platinum/examples/desktop/locales"
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

func newFindWindow(d *desk) *win {
	f := &finder{table: pt.NewTableState()}
	f.table.Widths = []int{150, 100, 182}
	f.table.SortColumn = 0
	f.table.ScrollBarAlways = true
	w := &win{app: appFind}
	w.frame = pt.WindowState{W: 470, H: 270}
	w.content = func(c *ui.Context, w *win) { findView(c, d, f) }
	return w
}

// matches returns the nodes of the disk whose name or kind, in the
// language of m, holds the query as the menus say.
func (f *finder) matches(root *node, m locales.Messages) []*node {
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
				s = ch.kindName(m)
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
	m := d.say()
	found := f.found
	// What the search found is news the focus does not bring: screen
	// readers read it out.
	news := ""
	if f.searched {
		news = m.FoundCount(len(found))
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
				pt.Label(c, m.FindItems())
				pt.PopupMenu(c, &f.where, "", []string{m.OnLocalDisks(), m.OnDisk(d.disk.name)}).Element().Label(m.Where())
			})
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				pt.Label(c, m.Whose())
				pt.PopupMenu(c, &f.what, "", []string{m.FindName(), m.FindKind()}).Element().Label(m.What())
				pt.PopupMenu(c, &f.how, "", []string{m.Contains(), m.StartsWith(), m.EndsWith(), m.IsExactly()}).Element().Label(m.How())
				field := pt.TextField(c, &f.query).Width(110).Label(m.SearchFor())
				field.Element().AutoFocus()
				ui.Box(c).Grow(1)
				if pt.Button(c, m.Find()).Width(60).Default().Disabled(strings.TrimSpace(f.query) == "").Clicked() {
					f.found, f.searched = f.matches(d.disk, m), true
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
					ui.Column(c).Fill().Children(func() { pt.EmptyState(c, m.ChooseWhatToFind()) })
				})
			})
		} else {
			ui.Column(c).Grow(1).Padding(0, 10, 0, 10).AlignItems(ui.Stretch).Children(func() {
				cols := []pt.TableColumn{{Title: m.ColumnName(), Icon: func(r int) []string { return smallIcon(found[r]) }}, {Title: m.ColumnKind()}, {Title: m.ColumnWhere()}}
				t := pt.Table(c, f.table, cols, len(found), func(r, col int) string {
					n := found[r]
					switch col {
					case 0:
						return n.name
					case 1:
						return n.kindName(m)
					}
					if n.parent != nil {
						return n.parent.path()
					}
					return ""
				}).Label(m.ItemsFound())
				t.Element().Grow(1).FillWidth()
				if t.Submitted() {
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
				count = m.FoundCount(len(found))
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
		if n.kind == kindPanel {
			return miniPanel
		}
		return appPic(n.app)
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
