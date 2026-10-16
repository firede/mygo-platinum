package main

import (
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
	"github.com/firede/mygo-platinum/examples/desktop/locales"
)

// cabinet is a window browsing a folder of the disk, as icons or as a
// list.
type cabinet struct {
	table *pt.TableState
	dir   *node
	sel   int
	list  bool
	// columns shows the folder and those inside it in columns, path the
	// items chosen in them.
	columns bool
	// buttons shows the items as buttons that a single click opens.
	buttons bool
	path    []*pt.TreeNode
	scroll  float64
	trash   bool
	active  bool // its window is in front, as of this build
	rect    ui.Rect
	typed   string
	desk    *desk
}

const (
	cellW = 84
	cellH = 64
)

func (cb *cabinet) selected() *node {
	if cb.sel < 0 || cb.sel >= len(cb.dir.children) {
		return nil
	}
	return cb.dir.children[cb.sel]
}

func (cb *cabinet) newFolder() {
	m := cb.desk.say()
	name := m.UntitledFolder()
	for i := 2; cb.dir.has(name); i++ {
		name = m.UntitledFolderN(i)
	}
	n := cb.dir.add(&node{name: name, kind: kindFolder})
	for i, c := range cb.dir.children {
		if c == n {
			cb.sel = i
		}
	}
}

func (n *node) has(name string) bool {
	for _, c := range n.children {
		if c.name == name {
			return true
		}
	}
	return false
}

// openFolder brings the window of a folder to the front, opening it if
// need be.
func (d *desk) openFolder(n *node) *win {
	for _, w := range d.windows {
		if w.folder != nil && w.folder.dir == n {
			d.raise(w)
			return w
		}
	}
	w := &win{app: appCabinet, folder: &cabinet{dir: n, sel: -1, desk: d}}
	w.frame = pt.WindowState{W: 340, H: 230}
	w.content = func(c *ui.Context, w *win) { d.cabinetView(c, w) }
	return d.open(w)
}

func humanSize(m locales.Messages, b int) string {
	switch {
	case b >= 1<<20:
		return m.SizeMB(float64(b) / (1 << 20))
	case b >= 1<<10:
		return m.SizeK((b + 1023) >> 10)
	}
	return m.Bytes(b)
}

func (d *desk) cabinetView(c *ui.Context, w *win) {
	cb := w.folder
	cb.active = w.active
	th := pt.ThemeOf(c)
	items := cb.dir.children
	n := len(items)
	if cb.sel >= n {
		cb.sel = n - 1
	}
	ui.Column(c).Fill().AlignItems(ui.Stretch).Children(func() {
		// The header: how many items, and how much room is left.
		m := d.say()
		count := m.Items(n)
		free := m.Available(humanSize(m, 48<<20-d.disk.size()))
		pt.WindowHeader(c, count, free)
		if cb.list {
			d.cabinetTable(c, cb)
			return
		}
		if cb.columns {
			d.cabinetColumns(c, cb)
			return
		}
		if cb.buttons {
			d.cabinetButtons(c, w)
			return
		}
		ui.Row(c).Grow(1).AlignItems(ui.Stretch).Children(func() {
			// The columns of icons, from the window's width as it is now
			// (the body's last painting is a frame behind a resize): the
			// frame's sides and the scroll bar off it.
			ww := w.frame.W
			if w.popup {
				ww = w.dock.W
			}
			cols := max((ww-2*pt.FrameWidth-pt.ScrollBarWidth)/cellW, 1)
			contentH := ((n+cols-1)/cols)*cellH + 8
			visible := int(cb.rect.H)
			body := ui.Box(c).Grow(1).Clip().Focusable().FocusRing(false).Role(ui.RoleList).Label(cb.dir.name)
			d.cabinetKeys(c, body, cb, cols)
			if !cb.trash {
				pt.ContextMenu(c, body, []pt.MenuItem{
					{Title: d.say().Open(), Disabled: cb.selected() == nil, Action: func() { d.openNode(cb.selected()) }},
					{Title: d.say().MoveToWastebasket(), Disabled: cb.selected() == nil || cb.selected().kind == kindApp || cb.selected().kind == kindPanel, Action: func() {
						if it := cb.selected(); it != nil {
							d.toTrash(it)
						}
					}},
					pt.SeparatorItem(),
					{Title: d.say().NewFolder(), Action: cb.newFolder},
				})
			}
			// Items dropped on the window's empty space move into its folder.
			if n, ok := ui.Drop[*node](body); ok && !cb.trash {
				copying := d.copying(c)
				d.later = append(d.later, func() { d.drop(n, cb.dir, copying) })
				c.After(0)
			}
			body.HandleInput(func(ev ui.InputEvent) bool {
				switch ev.Kind {
				case ui.InputPointerDown:
					// A press on no item chooses none.
					cb.sel = cb.itemAt(ev.X, ev.Y+float32(cb.scroll), cols)
					return false // and take the focus
				case ui.InputScroll:
					cb.scroll = min(max(cb.scroll+float64(ev.DY), 0), float64(max(contentH-visible, 0)))
					return true
				case ui.InputText:
					cb.typeSelect(ev.Text)
					return true
				}
				return false
			})
			body.TextCaret(ui.Rect{})
			// The ring and the inverted names show in the window in front
			// alone.
			focused := body.Focused() && cb.active
			body.Draw(func(p *ui.Painter, r ui.Rect) {
				if r.W != cb.rect.W || r.H != cb.rect.H {
					c.Invalidate() // the scroll bar goes by the size
				}
				cb.rect = r
				pn := pt.NewPen(p, r)
				pn.Rect(0, 0, pn.W, pn.H, pt.White)
			})
			body.DrawOver(func(p *ui.Painter, r ui.Rect) {
				if focused {
					pn := pt.NewPen(p, r)
					pn.Frame(0, 0, pn.W, pn.H, th.Accent.Dark)
					pn.Frame(1, 1, pn.W-2, pn.H-2, th.Accent.Dark)
				}
			})
			body.Children(func() {
				for i, it := range items {
					d.cabinetItem(c, cb, i, it, cols, focused, th.Accent)
				}
			})
			// The scroll bar over the lines above it and at the window's right,
			// as a window's own, down to the resize corner's edge.
			ui.Column(c).Width(pt.ScrollBarWidth).Margin(-1, -1, 0, 0).Children(func() {
				pt.ScrollBar(c, &cb.scroll, float64(visible), float64(contentH)).Step(cellH / 4).Element().Grow(1)
				ui.Box(c).Height(pt.GrowBoxSize - 1)
			})
		})
	})
}

// cabinetKeys moves the selection with the arrows and opens it with
// Return or Command-Down.
func (d *desk) cabinetKeys(c *ui.Context, body ui.Element, cb *cabinet, cols int) {
	n := len(cb.dir.children)
	if n == 0 {
		return
	}
	step := func(dx int) {
		if cb.sel < 0 {
			cb.sel = 0
			return
		}
		cb.sel = min(max(cb.sel+dx, 0), n-1)
	}
	across, down := 1, cols
	if body.Shortcut(0, ui.KeyRight) {
		step(across)
	}
	if body.Shortcut(0, ui.KeyLeft) {
		step(-across)
	}
	if body.Shortcut(0, ui.KeyDown) {
		step(down)
	}
	if body.Shortcut(0, ui.KeyUp) {
		step(-down)
	}
	if body.Shortcut(0, ui.KeyEnter) || body.Shortcut(ui.Cmd, ui.KeyDown) {
		if it := cb.selected(); it != nil {
			d.openNode(it)
		}
	}
	if body.Shortcut(ui.Cmd, ui.KeyUp) && cb.dir.parent != nil {
		d.openFolder(cb.dir.parent)
	}
}

func (cb *cabinet) typeSelect(s string) {
	cb.typed = strings.ToLower(s)
	for i, it := range cb.dir.children {
		if strings.HasPrefix(strings.ToLower(it.name), cb.typed) {
			cb.sel = i
			return
		}
	}
}

// itemBox returns where item i is, from the top of the content.
func (cb *cabinet) itemBox(i, cols int) (x, y int) {
	return (i % cols) * cellW, 4 + (i/cols)*cellH
}

func (cb *cabinet) itemAt(x, y float32, cols int) int {
	for i := range cb.dir.children {
		ix, iy := cb.itemBox(i, cols)
		w, h := cellW, cellH
		if x >= float32(ix) && x < float32(ix+w) && y >= float32(iy) && y < float32(iy+h) {
			return i
		}
	}
	return -1
}

// cabinetItem builds an item of a folder window: its icon and name, which a
// click chooses, a double click opens, and the pointer drags — to the
// Wastebasket, another folder's window or a folder's icon.
func (d *desk) cabinetItem(c *ui.Context, cb *cabinet, i int, it *node, cols int, focused bool, a pt.Accent) {
	x, y := cb.itemBox(i, cols)
	e := ui.Box(c.Key(it.id())).Absolute().Left(float32(x)).Top(float32(y)-float32(cb.scroll)).Size(cellW, cellH).
		Role(ui.RoleListItem).Label(it.name)
	if !cb.trash && it.kind != kindApp && it.kind != kindPanel {
		e.Drag(it)
		if e.Dragging() {
			d.dragSeen = true
		}
	}
	if e.Pressed() || e.RightClicked() {
		cb.sel = i
	}
	if e.DoubleClicked() {
		d.zoomFrom = ui.Rect{X: cb.rect.X + float32(x+(cellW-32)/2), Y: cb.rect.Y + float32(y+4) - float32(cb.scroll), W: 32, H: 32}
		d.openNode(it)
	}
	over := false
	if it.kind == kindFolder && !cb.trash {
		if n, ok := ui.Drop[*node](e); ok && n != it {
			copying := d.copying(c)
			d.later = append(d.later, func() { d.drop(n, it, copying) })
			c.After(0)
		}
		_, over = ui.DragOver[*node](e)
		if over {
			d.springAt(c, it)
		}
	}
	sel := cb.sel == i || over
	e.Draw(func(p *ui.Painter, r ui.Rect) {
		pn := pt.NewPen(p, r)
		iconFor(it).draw(p, float32(pn.X+(cellW-32)/2), float32(pn.Y+4), sel)
		drawIconName(pn, it.name, 38, sel && (cb.active || over), a)
	})
}

// move puts n in folder to, from wherever it is.
func (d *desk) move(n, to *node) {
	for p := to; p != nil; p = p.parent {
		if p == n {
			return // not into itself
		}
	}
	if n.parent == to {
		return
	}
	n.remove()
	to.add(n)
}

// cabinetTable shows the folder as a table of names, sizes and kinds,
// which sorts by the column clicked.
func (d *desk) cabinetTable(c *ui.Context, cb *cabinet) {
	if cb.table == nil {
		cb.table = pt.NewTableState()
		cb.table.Widths = []int{150, 70, 100}
		cb.table.SortColumn = 0
		cb.table.ScrollBarAlways = true
		cb.table.NoFocusRing = true // as the classic list views showed none
	}
	items := cb.dir.children
	clear(cb.table.Selected)
	if cb.sel >= 0 && cb.sel < len(items) {
		cb.table.Selected[cb.sel] = true
	}
	m := d.say()
	cols := []pt.TableColumn{{Title: m.ColumnName(), Icon: func(r int) []string { return smallIcon(items[r]) }}, {Title: m.ColumnSize(), Right: true, Less: func(a, b int) bool {
		// Folders, which show no size, after documents; documents by bytes.
		sa, sb := items[a].size(), items[b].size()
		if fa, fb := items[a].kind == kindFolder, items[b].kind == kindFolder; fa != fb {
			return fb
		}
		return sa < sb
	}}, {Title: m.ColumnKind()}}
	tw := pt.Table(c, cb.table, cols, len(items), func(r, col int) string {
		it := items[r]
		switch col {
		case 0:
			return it.name
		case 1:
			if it.kind == kindFolder {
				return "—"
			}
			return humanSize(m, it.size())
		}
		return it.kindName(m)
	})
	tw.Element().Grow(1).FillWidth().Margin(0, 0, pt.GrowBoxSize, 0).Label(cb.dir.name)
	if tw.Changed() {
		cb.sel = -1
		for r := range cb.table.Selected {
			cb.sel = r
		}
	}
	if tw.Submitted() {
		if it := cb.selected(); it != nil {
			d.openNode(it)
		}
	}
}

// treeOf returns the folders and items inside n as tree nodes, each
// holding its node, for a column view.
func treeOf(n *node, seen map[*node]*pt.TreeNode) []*pt.TreeNode {
	var out []*pt.TreeNode
	for _, ch := range n.children {
		t := seen[ch]
		if t == nil {
			t = &pt.TreeNode{Title: ch.name, Data: ch}
			seen[ch] = t
		}
		t.Children = nil
		if ch.kind == kindFolder {
			t.Children = treeOf(ch, seen)
			if t.Children == nil {
				t.Children = []*pt.TreeNode{} // a folder, if empty
			}
		}
		out = append(out, t)
	}
	return out
}

// treeNodes keeps the tree nodes of the disk's nodes, so that a column
// view's choice survives the frames.
var treeNodes = map[*node]*pt.TreeNode{}

// cabinetColumns shows the folder in columns: its items, and right of
// each folder chosen, the items inside it. A double click opens.
func (d *desk) cabinetColumns(c *ui.Context, cb *cabinet) {
	roots := treeOf(cb.dir, treeNodes)
	cv := pt.ColumnView(c, roots, &cb.path).ScrollBarRoom(pt.GrowBoxSize - 1)
	cv.Element().Grow(1).FillWidth().Label(cb.dir.name)
	if cv.Submitted() && len(cb.path) > 0 {
		if n, ok := cb.path[len(cb.path)-1].Data.(*node); ok {
			d.openNode(n)
		}
	}
}

// springDelay is how long a drag rests on a folder before it springs open.
const springDelay = 500 * time.Millisecond

// springAt notes that a drag rests on folder n: once it has rested there
// springDelay, the folder's window springs open, for the drag to go on
// into it, as the classic desktop's spring-loaded folders did.
func (d *desk) springAt(c *ui.Context, n *node) {
	d.springSeen = true
	if d.springNode != n {
		d.springNode, d.springSince = n, d.now()
	}
	if wait := springDelay - d.now().Sub(d.springSince); wait > 0 {
		c.After(wait)
		return
	}
	d.springNode = nil
	d.later = append(d.later, func() { d.openFolder(n) })
	c.After(0)
}

// cabinetButtons shows a folder's items as buttons, as the classic desktop
// could: each a raised square with the item's icon, its name under it,
// which a single click (or Space or Return) opens.
func (d *desk) cabinetButtons(c *ui.Context, w *win) {
	cb := w.folder
	const bw, bh = 56, 56
	ww := w.frame.W
	if w.popup {
		ww = w.dock.W
	}
	cols := max((ww-2*pt.FrameWidth-16)/cellW, 1)
	// Room under it for the resize corner, as the list view leaves.
	box := ui.Box(c).Grow(1).AlignSelf(ui.Stretch).Clip().Role(ui.RoleGroup).Label(cb.dir.name).
		Margin(0, 0, pt.GrowBoxSize, 0)
	box.Draw(func(p *ui.Painter, r ui.Rect) {
		pn := pt.NewPen(p, r)
		pn.Rect(0, 0, pn.W, pn.H, pt.White)
	})
	box.Children(func() {
		for i, it := range cb.dir.children {
			x := 8 + (i%cols)*cellW + 2
			y := 8 + (i/cols)*(bh+22)
			b := ui.ButtonBase(c.Key(it.id())).Absolute().Left(float32(x)).Top(float32(y)).Size(cellW-4, bh+18).
				FocusRing(false).Label(it.name)
			if b.Clicked() {
				n := it
				d.later = append(d.later, func() { d.openNode(n) })
				c.After(0)
			}
			s := pt.StateNormal
			if b.Pressed() {
				s = pt.StatePressed
			}
			focused := b.FocusVisible()
			ic, name := iconFor(it), it.name
			th := pt.ThemeOf(c)
			b.Draw(func(p *ui.Painter, r ui.Rect) {
				pn := pt.NewPen(p, r)
				face := pn.Sub((pn.W-bw)/2, 0, bw, bh)
				pt.DrawBevelButton(face, s, false)
				dy := 0
				if s == pt.StatePressed {
					dy = 1
				}
				ic.draw(p, float32(face.X+(bw-32)/2+dy), float32(face.Y+(bh-32)/2+dy), false)
				if focused {
					pt.DrawFocusRing(face, 0, th.Accent.Dark)
				}
				name = pt.AppFont.Truncate(name, cellW-4)
				pn.TextCentered(0, bh+2, pn.W, 14, name, pt.AppFont, pt.Black)
			})
		}
	})
}
