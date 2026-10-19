package diagram

import (
	"math"

	"github.com/egoist/mygo/ui"
	"github.com/firede/limu/mygoui"
	"github.com/firede/limu/scene"
	"github.com/firede/limu/style"
	"github.com/firede/limu/text"
	platinum "github.com/firede/mygo-platinum"
)

// Draw draws an item of a scene in the Platinum look, as mygoui's Options.Draw: on whole
// pixels with a Pen, a node's box as a button and a group's as a group box, every other
// outline, line and arrowhead a pixel at a time, and text in AppFont. It draws as mygoui
// does where the scene is not drawn a DIP to its unit, as the Pen's pixels are DIPs.
func Draw(p *mygoui.Painter, it *scene.Item, s style.Resolved, next func()) {
	if p.Len(1) != 1 {
		next()
		return
	}
	x, y := p.Point(scene.Point{})
	d := drawer{p: p, pn: platinum.NewPen(p.Painter, ui.Rect{X: x, Y: y}), it: it, s: s}
	d.parts()
	d.ends()
	p.Picture(it, s)
	d.label()
}

type drawer struct {
	p  *mygoui.Painter
	pn platinum.Pen
	it *scene.Item
	s  style.Resolved
}

func (d *drawer) color(c style.Color) ui.Color { return mygoui.Color(c, d.s.Opacity) }

// stroke is how the item's lines go on pixels: an edge's fall where they are, the lines of
// a shape toward its middle.
func (d *drawer) stroke() stroke {
	st := stroke{width: max(1, int(math.Round(d.s.StrokeWidth))), at: floorAt}
	for _, v := range d.s.Dash {
		st.dash = append(st.dash, max(1, int(math.Round(v))))
	}
	if d.it.Kind != scene.Edge {
		st.at = inward(d.it.Box)
	}
	return st
}

func (d *drawer) parts() {
	for _, part := range d.it.Parts {
		face, line := d.color(d.s.Fill), d.color(d.s.Stroke)
		switch part.Paint {
		case scene.StrokeOnly:
			face = ui.Color{}
		case scene.FillOnly:
			line = ui.Color{}
		case scene.Solid:
			face = d.color(d.s.Line)
		}
		if part.Path == nil {
			d.box(part, face, line)
			continue
		}
		lines := part.Path.Flatten(0.25)
		if face.A == 0 {
			if line.A > 0 {
				px := pixels{}
				d.stroke().lines(px, lines)
				px.draw(d.pn, line)
			}
			continue
		}
		region := fill(lines)
		if line.A == 0 {
			region.draw(d.pn, face)
			continue
		}
		// The frame, as wide as the stroke, inside the region, and the face within it.
		frame := pixels{}
		inner := region
		for range max(1, int(math.Round(d.s.StrokeWidth))) {
			b := inner.border()
			for q := range b {
				frame[q] = struct{}{}
			}
			inner = inner.without(b)
		}
		inner.draw(d.pn, face)
		frame.draw(d.pn, line)
	}
}

// box draws a box of an outline: a node's as a button, raised in its face; a group's as a
// group box, etched; any other filled and framed.
func (d *drawer) box(part scene.Part, fill, line ui.Color) {
	b := part.Box
	x, y := int(math.Round(b.X)), int(math.Round(b.Y))
	w, h := int(math.Round(b.X+b.W))-x, int(math.Round(b.Y+b.H))-y
	if w <= 0 || h <= 0 {
		return
	}
	r := 0
	if part.Rounded {
		r = min(4, max(0, int(math.Round(d.s.Radius))))
	}
	pn := d.pn.Sub(x, y, w, h)
	switch {
	case d.it.Kind == scene.Node && fill.A == 255 && line.A > 0 && w >= 8 && h >= 8:
		button(pn, r, fill, line)
	case d.it.Kind == scene.Group && line.A > 0 && !d.it.Author.Stroke.Set():
		pn.Rect(0, 0, w, h, fill)
		groupBox(pn)
	default:
		pn.RoundFrame(0, 0, w, h, r, line, fill)
	}
}

// button draws a box as DrawButton draws one: framed, with two rings of bevel inside, lit
// from the top left, in the shades of its face.
func button(pn platinum.Pen, r int, face, frame ui.Color) {
	w, h := pn.W, pn.H
	tl1, tl2 := mix(face, platinum.White, 1), mix(face, platinum.White, 0.5)
	br1, br2 := mix(face, platinum.Black, 5.0/13), mix(face, platinum.Black, 3.0/13)
	if r == 0 {
		pn.Rect(0, 0, w, h, face)
		pn.Frame(0, 0, w, h, frame)
		pn.Bevel(1, 1, w-2, h-2, tl1, br1)
		pn.Bevel(2, 2, w-4, h-4, tl2, br2)
		return
	}
	pn.RoundFrame(0, 0, w, h, 3, frame, face)
	pn.HLine(3, 1, w-6, tl1)
	pn.VLine(1, 3, h-6, tl1)
	pn.Dot(2, 2, tl1)
	pn.HLine(3, h-2, w-6, br1)
	pn.VLine(w-2, 3, h-6, br1)
	pn.Dot(w-3, h-3, br1)
	pn.HLine(3, 2, w-7, tl2)
	pn.VLine(2, 3, h-7, tl2)
	pn.HLine(4, h-3, w-7, br2)
	pn.VLine(w-3, 4, h-7, br2)
}

// groupBox draws the etched frame of DrawGroupBox about a box.
func groupBox(pn platinum.Pen) {
	w, h := pn.W, pn.H
	pn.Frame(0, 0, w-1, h-1, platinum.Gray7)
	pn.HLine(1, 1, w-3, platinum.White)
	pn.VLine(1, 2, h-4, platinum.White)
	pn.HLine(0, h-1, w, platinum.White)
	pn.VLine(w-1, 0, h, platinum.White)
}

// mix is a color t of the way from a to b, keeping a's alpha.
func mix(a, b ui.Color, t float64) ui.Color {
	m := func(u, v uint8) uint8 { return uint8(math.Round(float64(u) + (float64(v)-float64(u))*t)) }
	return ui.Color{R: m(a.R, b.R), G: m(a.G, b.G), B: m(a.B, b.B), A: a.A}
}

// ends draws an edge's markers in the color of its line: an arrowhead along a run across
// or down as the pixel arrowheads of old, growing a pixel each side every other row; any
// other as its outline falls on pixels.
func (d *drawer) ends() {
	c := d.color(d.s.Stroke)
	if c.A == 0 {
		return
	}
	px := pixels{}
	for _, m := range d.it.Ends {
		if m.Type == "arrow_point" && (m.DX == 0) != (m.DY == 0) {
			arrowhead(px, m)
			continue
		}
		path, filled := m.Outline()
		if path == nil {
			continue
		}
		if filled {
			for q := range fill(path.Flatten(0.25)) {
				px[q] = struct{}{}
			}
		} else {
			stroke{width: 1, at: floorAt}.lines(px, path.Flatten(0.25))
		}
	}
	px.draw(d.pn, c)
}

// arrowhead puts an arrowhead of MarkerSize rows on pixels, its tip on the pixel before
// the marker's tip, in line with the edge's pixels.
func arrowhead(px pixels, m scene.Marker) {
	dx, dy := int(math.Round(m.DX)), int(math.Round(m.DY))
	// The tip's pixel: the one the line's last half pixel falls in.
	tip := floorAt(scene.Point{X: m.At.X - m.DX*0.5, Y: m.At.Y - m.DY*0.5})
	if dx == 0 {
		tip.x = int(math.Floor(m.At.X))
	} else {
		tip.y = int(math.Floor(m.At.Y))
	}
	for row := range scene.MarkerSize {
		half := int(math.Round(float64(row) * 0.42))
		for k := -half; k <= half; k++ {
			// Back along the edge by row, across it by k.
			px.add(tip.x-dx*row+dy*k, tip.y-dy*row+dx*k)
		}
	}
}

// label draws an item's label in AppFont, each line centered on the item's At, or starting
// or ending there, its capitals in the middle of the line's height; with a halo, drawn about
// it in the halo's color first.
func (d *drawer) label() {
	l := d.it.Label
	if l == nil || l.Empty() {
		return
	}
	c := d.color(d.s.Text)
	if c.A == 0 {
		return
	}
	if halo := d.color(d.s.Halo); halo.A > 0 && d.s.HaloWidth > 0 {
		r := max(1, int(d.s.HaloWidth/2))
		for oy := -r; oy <= r; oy++ {
			for ox := -r; ox <= r; ox++ {
				if (ox != 0 || oy != 0) && abs(ox)+abs(oy) <= r+r/2 {
					d.text(l, ox, oy, halo)
				}
			}
		}
	}
	d.text(l, 0, 0, c)
}

func (d *drawer) text(l *text.Label, ox, oy int, c ui.Color) {
	f := platinum.AppFont
	lh := int(math.Round(l.LineHeight))
	top := int(math.Round(d.it.At.Y-l.Height/2)) + oy
	for i, line := range l.Lines {
		var x float64
		switch d.it.Align {
		case scene.Start:
			x = d.it.At.X
		case scene.End:
			x = d.it.At.X - l.Widths[i]
		default:
			x = d.it.At.X - l.Widths[i]/2
		}
		px := int(math.Round(x)) + ox
		ty := top + i*lh + (lh-f.CapHeight())/2 - (f.Ascent() - f.CapHeight())
		for _, sp := range line {
			if sp.Text != "" && sp.Icon == "" {
				d.pn.TextStyled(px, ty, sp.Text, f, c, textStyle(sp.Face))
			}
			px += int(math.Round(sp.Width))
		}
	}
}
