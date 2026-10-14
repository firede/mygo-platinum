package platinum

import (
	"github.com/egoist/mygo/ui"
)

// ImageWell builds an image well: a sunken white square showing bmp
// centered, as a picture a document or a setting holds. Label it for
// assistive technology; size it with Size (48×48 by default).
func ImageWell(c *ui.Context, bmp *ui.Bitmap) ui.Element {
	el := ui.Box(c).Size(48, 48).Role(ui.RoleImage)
	el.Draw(func(p *ui.Painter, r ui.Rect) {
		pn := NewPen(p, r)
		DrawWell(pn)
		if bmp != nil {
			bw, bh := bmp.Size()
			x := pn.X + (pn.W-bw)/2
			y := pn.Y + (pn.H-bh)/2
			pn.Clip(3, 3, pn.W-6, pn.H-6, func() {
				p.Image(bmp, ui.Rect{X: float32(x), Y: float32(y), W: float32(bw), H: float32(bh)}, ui.FillBox)
			})
		}
	})
	return el
}

// DrawWell draws a sunken white well filling pn's box, as an image well's.
func DrawWell(pn Pen) {
	pn.Bevel(0, 0, pn.W, pn.H, Gray7, White)
	pn.Frame(1, 1, pn.W-2, pn.H-2, Black)
	pn.Rect(2, 2, pn.W-4, pn.H-4, White)
}
