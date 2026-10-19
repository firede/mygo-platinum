package diagram

import (
	"math"
	"slices"

	"github.com/egoist/mygo/ui"
	"github.com/firede/limu/scene"
	platinum "github.com/firede/mygo-platinum"
)

// A pixel is one DIP of the pen's grid, by its top left corner.
type pixel struct{ x, y int }

// pixels is a set of pixels, gathered and then drawn at once, so that a pixel two lines
// cross is drawn once, and a color that is not opaque comes out even.
type pixels map[pixel]struct{}

func (px pixels) add(x, y int) { px[pixel{x, y}] = struct{}{} }

func (px pixels) has(x, y int) bool {
	_, ok := px[pixel{x, y}]
	return ok
}

// draw fills the pixels in c, in runs along their rows.
func (px pixels) draw(pn platinum.Pen, c ui.Color) {
	if len(px) == 0 || c.A == 0 {
		return
	}
	all := make([]pixel, 0, len(px))
	for p := range px {
		all = append(all, p)
	}
	slices.SortFunc(all, func(a, b pixel) int {
		if a.y != b.y {
			return a.y - b.y
		}
		return a.x - b.x
	})
	for i := 0; i < len(all); {
		j := i + 1
		for j < len(all) && all[j].y == all[i].y && all[j].x == all[j-1].x+1 {
			j++
		}
		pn.HLine(all[i].x, all[i].y, j-i, c)
		i = j
	}
}

// fill is the pixels whose centers are inside polygons, by the nonzero rule, as SVG
// fills.
func fill(polys []scene.Polyline) pixels {
	px := pixels{}
	top, bottom := math.Inf(1), math.Inf(-1)
	for _, l := range polys {
		for _, p := range l.Points {
			top, bottom = min(top, p.Y), max(bottom, p.Y)
		}
	}
	if top > bottom {
		return px
	}
	type crossing struct {
		x   float64
		dir int
	}
	var xs []crossing
	for y := int(math.Floor(top)); float64(y) < bottom; y++ {
		yc := float64(y) + 0.5
		xs = xs[:0]
		for _, l := range polys {
			n := len(l.Points)
			for i := range n {
				a, b := l.Points[i], l.Points[(i+1)%n]
				if (a.Y <= yc) == (b.Y <= yc) {
					continue
				}
				dir := 1
				if b.Y < a.Y {
					dir = -1
				}
				xs = append(xs, crossing{a.X + (yc-a.Y)*(b.X-a.X)/(b.Y-a.Y), dir})
			}
		}
		slices.SortFunc(xs, func(a, b crossing) int { return cmpFloat(a.x, b.x) })
		wind := 0
		var from float64
		for _, c := range xs {
			was := wind
			wind += c.dir
			switch {
			case was == 0 && wind != 0:
				from = c.x
			case was != 0 && wind == 0:
				for x := int(math.Ceil(from - 0.5)); float64(x)+0.5 < c.x; x++ {
					px.add(x, y)
				}
			}
		}
	}
	return px
}

func cmpFloat(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// border is the pixels of a region next to a pixel outside it, across or down: its outline,
// a pixel wide, as QuickDraw frames a region.
func (px pixels) border() pixels {
	out := pixels{}
	for p := range px {
		if !px.has(p.x-1, p.y) || !px.has(p.x+1, p.y) || !px.has(p.x, p.y-1) || !px.has(p.x, p.y+1) {
			out[p] = struct{}{}
		}
	}
	return out
}

// without is the pixels of px that are not in other.
func (px pixels) without(other pixels) pixels {
	out := pixels{}
	for p := range px {
		if !other.has(p.x, p.y) {
			out[p] = struct{}{}
		}
	}
	return out
}

// A stroke is how lines are put on pixels: as wide as a brush, in dashes of a pattern,
// and from a point of the scene to the pixel it falls in, by place.
type stroke struct {
	width int
	dash  []int // lengths on and off, in pixels; none is solid
	at    func(p scene.Point) pixel
}

// lines puts polylines on pixels, one pixel after the other as Bresenham steps, the pattern
// of dashes running on from one line to the next.
func (s stroke) lines(px pixels, lines []scene.Polyline) {
	step := 0
	put := func(x, y int) {
		if len(s.dash) > 0 {
			on, at := true, step%patternLen(s.dash)
			for _, d := range s.dash {
				if at < d {
					break
				}
				at -= d
				on = !on
			}
			step++
			if !on {
				return
			}
		}
		lo := -(s.width - 1) / 2
		for dy := lo; dy < lo+s.width; dy++ {
			for dx := lo; dx < lo+s.width; dx++ {
				px.add(x+dx, y+dy)
			}
		}
	}
	for _, l := range lines {
		pts := l.Points
		if l.Closed && len(pts) > 0 {
			pts = append(pts[:len(pts):len(pts)], pts[0])
		}
		var last pixel
		for i, p := range pts {
			q := s.at(p)
			if i == 0 {
				put(q.x, q.y)
			} else {
				bresenham(last, q, put)
			}
			last = q
		}
	}
}

func patternLen(dash []int) int {
	n := 0
	for _, d := range dash {
		n += d
	}
	return max(n, 1)
}

// bresenham calls put for each pixel of the line from a to b, but a.
func bresenham(a, b pixel, put func(x, y int)) {
	dx, dy := abs(b.x-a.x), -abs(b.y-a.y)
	sx, sy := sign(b.x-a.x), sign(b.y-a.y)
	err := dx + dy
	x, y := a.x, a.y
	for x != b.x || y != b.y {
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x += sx
		}
		if e2 <= dx {
			err += dx
			y += sy
		}
		put(x, y)
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func sign(v int) int {
	switch {
	case v < 0:
		return -1
	case v > 0:
		return 1
	}
	return 0
}

// floorAt is the pixel a point of a line falls in.
func floorAt(p scene.Point) pixel {
	return pixel{int(math.Floor(p.X)), int(math.Floor(p.Y))}
}

// inward is where a point of a line inside a box falls: on the pixel by it toward the
// box's middle, so that a line drawn at the box's sides lands on the pixels of its frame,
// and lines alike from either side are alike from it.
func inward(b scene.Rect) func(p scene.Point) pixel {
	cx, cy := b.X+b.W/2, b.Y+b.H/2
	at := func(v, c float64) int {
		if v <= c {
			return int(math.Floor(v))
		}
		return int(math.Ceil(v)) - 1
	}
	return func(p scene.Point) pixel { return pixel{at(p.X, cx), at(p.Y, cy)} }
}
