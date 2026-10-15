package main

import (
	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
	"github.com/firede/mygo-platinum/examples/desktop/locales"
)

// getInfo shows a window telling about n, as the classic Get Info: its
// icon and name, its kind, size, where it is and when it changed; or
// brings the one already open to the front.
func (d *desk) getInfo(n *node) {
	for _, w := range d.windows {
		if w.info == n {
			d.raise(w)
			return
		}
	}
	w := &win{app: appInfo, info: n, opts: pt.WindowOptions{NoZoom: true, NoGrow: true}}
	w.frame = pt.WindowState{W: 250}
	w.content = func(c *ui.Context, w *win) { infoView(c, d, w.info) }
	d.open(w)
}

// infoTarget is what Get Info tells about: the item chosen in the folder
// window in front, or else the disk or the Wastebasket chosen on the
// desktop.
func (d *desk) infoTarget(selected *node) *node {
	if selected != nil {
		return selected
	}
	if d.selIcon < 0 || d.selIcon >= len(d.icons) {
		return nil
	}
	switch d.icons[d.selIcon].icon {
	case iconDisk:
		return d.disk
	case iconTrash:
		if d.trashNode == nil {
			d.trashNode = &node{key: "wastebasket", name: d.say().Wastebasket(), kind: kindFolder, modified: d.now()}
		}
		d.trashNode.children = append(d.trashNode.children[:0], d.trash...)
		return d.trashNode
	}
	return nil
}

// onDisk is a size as Get Info tells it, in the language of m: the room
// it takes in blocks of 4K, and the bytes it uses.
func onDisk(m locales.Messages, bytes int) string {
	k := (bytes + 4095) / 4096 * 4
	used := m.Bytes(bytes)
	if k == 0 {
		return m.ZeroKOnDisk(used)
	}
	return m.OnDiskK(k, used)
}

func infoView(c *ui.Context, d *desk, n *node) {
	m := d.say()
	where := m.TheDesktop()
	if n.parent != nil {
		where = n.parent.path()
	}
	size := onDisk(m, n.size())
	ui.Column(c).FillWidth().Padding(12).Gap(8).AlignItems(ui.Stretch).Children(func() {
		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			ic := iconFor(n)
			ui.Box(c).Size(32, 32).Role(ui.RoleImage).Label(n.kindName(m)).Draw(func(p *ui.Painter, r ui.Rect) {
				ic.draw(p, r.X, r.Y, false)
			})
			pt.Label(c, n.name)
		})
		pt.Separator(c)
		for _, row := range [][2]string{
			{m.KindLabel(), n.kindName(m)},
			{m.SizeLabel(), size},
			{m.WhereLabel(), where},
			{m.ModifiedLabel(), i18n.FormatDate(d.formats, n.modified, i18n.DateOptions{Skeleton: "yMMMEdjm"})},
		} {
			ui.Row(c).Gap(8).AlignItems(ui.Start).Children(func() {
				ui.Box(c).Width(62).Children(func() { pt.Label(c, row[0]) })
				pt.Paragraph(c, row[1], pt.AppFont, 150)
			})
		}
		pt.Separator(c)
		pt.Label(c, m.CommentsLabel())
		pt.TextArea(c, &n.comment).Label(m.Comments()).Element().Height(3*pt.TextLineHeight + 6).AlignSelf(ui.Stretch)
	})
}

// duplicate puts a copy of n beside it, named as classic desktops named copies:
// "name copy", then "name copy 2" and on.
func (d *desk) duplicate(n *node) {
	dir := n.parent
	if dir == nil {
		return
	}
	m := d.say()
	name := m.CopyName(n.name)
	for i := 2; dir.has(name); i++ {
		name = m.CopyNameN(n.name, i)
	}
	c := n.clone()
	c.name, c.modified = name, d.now()
	dir.add(c)
}

// drop puts n, dropped on folder to, in it: moved there, or with Option
// held, a copy of it, as classic desktops copied on an Option-drag.
func (d *desk) drop(n, to *node, copying bool) {
	if !copying {
		d.move(n, to)
		return
	}
	for p := to; p != nil; p = p.parent {
		if p == n {
			return // not into itself
		}
	}
	c := n.clone()
	c.name = n.name
	if to.has(c.name) {
		m := d.say()
		c.name = m.CopyName(n.name)
		for i := 2; to.has(c.name); i++ {
			c.name = m.CopyNameN(n.name, i)
		}
	}
	to.add(c)
}
