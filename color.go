package platinum

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

// Color well metrics, from docs/spec.md.
const (
	ColorWellW = 44 // the width of a color well
	ColorWellH = 22 // its height
	swatchSize = 14 // a swatch of the picker's palette
)

// pickerPalette is the color picker's palette: eight hues in five shades,
// and a row of grays.
var pickerPalette = func() []ui.Color {
	var p []ui.Color
	hues := [][3]uint8{{0xCC, 0x22, 0x22}, {0xEE, 0x88, 0x22}, {0xEE, 0xCC, 0x22}, {0x44, 0x99, 0x33},
		{0x22, 0x99, 0x99}, {0x33, 0x55, 0xCC}, {0x66, 0x33, 0xAA}, {0xBB, 0x44, 0x99}}
	for _, shade := range []float32{0.6, 0.35, 0, -0.25, -0.5} {
		for _, h := range hues {
			c := ui.RGB(h[0], h[1], h[2])
			if shade > 0 {
				c = c.Mix(White, shade)
			} else if shade < 0 {
				c = c.Mix(Black, -shade)
			}
			p = append(p, ui.RGB(c.R, c.G, c.B))
		}
	}
	for _, g := range []ui.Color{White, Gray2, Gray4, Gray6, Gray8, Gray10, Gray12, Black} {
		p = append(p, g)
	}
	return p
}()

type wellState struct {
	open    bool
	r, g, b float64
	old     ui.Color
	label   string // the well's name, the picker's title
}

// ColorWellWidget is a color well being built.
type ColorWellWidget struct {
	el       ui.Element
	st       *wellState
	changed  bool
	disabled bool
}

// ColorWell builds a color well: a bevel button holding a swatch of *col.
// Clicking it, or Space, opens a color picker dialog — a palette, sliders of
// red, green and blue, and the new color beside the old — whose OK sets
// *col, and Cancel leaves it. Label names it, and titles the picker.
func ColorWell(c *ui.Context, col *ui.Color) *ColorWellWidget {
	th := ThemeOf(c)
	w := &ColorWellWidget{}
	w.el = ui.ButtonBase(c).FocusRing(false).Role(ui.RoleColorWell).
		Value(fmt.Sprintf("#%02X%02X%02X", col.R, col.G, col.B)).Size(ColorWellW, ColorWellH)
	st := ui.Local(w.el, "well", func() wellState { return wellState{} })
	w.st = st
	title := st.label // as the last frame named it
	if title == "" {
		title = say(c, "color")
	}
	if w.el.Clicked() && !st.open {
		st.open = true
		st.old = *col
		st.r, st.g, st.b = float64(col.R), float64(col.G), float64(col.B)
	}
	pressed := w.el.Pressed() || st.open
	focused := w.el.FocusVisible()
	shown := *col
	w.el.Draw(func(p *ui.Painter, r ui.Rect) {
		pn := NewPen(p, r).Faded(w.disabled)
		s := StateNormal
		if pressed {
			s = StatePressed
		}
		if w.disabled {
			s = StateDisabled
		}
		DrawBevelButton(pn, s, false)
		d := 0
		if pressed {
			d = 1
		}
		pn.Frame(5+d, 5+d, pn.W-10, pn.H-10, Black)
		pn.Rect(6+d, 6+d, pn.W-12, pn.H-12, shown)
		if focused && !w.disabled {
			DrawFocusRing(pn, 0, th.Accent.Dark)
		}
	})
	if st.open {
		Dialog(c, &st.open, title, 300, func() {
			now := ui.RGB(uint8(st.r), uint8(st.g), uint8(st.b))
			ui.Row(c).Gap(12).AlignItems(ui.Start).Children(func() {
				// The palette.
				grid := ui.Box(c).Size(8*swatchSize+4, 6*swatchSize+4).Role(ui.RoleGroup).Label(say(c, "palette"))
				grid.HandleInput(func(ev ui.InputEvent) bool {
					if ev.Kind != ui.InputPointerDown {
						return false
					}
					i := int(ev.Y-2)/swatchSize*8 + int(ev.X-2)/swatchSize
					if i >= 0 && i < len(pickerPalette) && ev.X >= 2 && ev.Y >= 2 {
						pc := pickerPalette[i]
						st.r, st.g, st.b = float64(pc.R), float64(pc.G), float64(pc.B)
					}
					return true
				})
				grid.Draw(func(p *ui.Painter, r ui.Rect) {
					DrawPalette(NewPen(p, r), now)
				})
				// The new color over the old.
				old := st.old
				ui.Box(c).Size(60, 50).Role(ui.RoleImage).Label(say(c, "newAndOld")).Draw(func(p *ui.Painter, r ui.Rect) {
					pn := NewPen(p, r)
					DrawWell(pn)
					pn.Rect(3, 3, pn.W-6, (pn.H-6)/2, now)
					pn.Rect(3, 3+(pn.H-6)/2, pn.W-6, pn.H-6-(pn.H-6)/2, old)
				})
			})
			for _, ch := range []struct {
				name string
				v    *float64
			}{{say(c, "red"), &st.r}, {say(c, "green"), &st.g}, {say(c, "blue"), &st.b}} {
				ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
					ui.Box(c).Width(42).Children(func() { Label(c, sayWith(c, "labelled", map[string]any{"name": ch.name})) })
					Slider(c, ch.v, 0, 255).Width(160).Step(1).Label(ch.name)
					TextIn(c, fmt.Sprint(int(*ch.v)), AppFont, Black)
				})
			}
			ui.Row(c).Gap(10).Justify(ui.End).Children(func() {
				if Button(c, say(c, "cancel")).Cancel().Clicked() {
					st.open = false
				}
				if Button(c, say(c, "ok")).Default().Clicked() {
					*col = now
					st.open = false
					w.changed = true
				}
			})
		})
	}
	report(w.el, w.changed)
	return w
}

// Changed reports whether the user chose another color since the last
// frame.
func (w *ColorWellWidget) Changed() bool { return report(w.el, w.changed) }

// Label names the well, as screen readers read it, and titles its color
// picker.
func (w *ColorWellWidget) Label(s string) *ColorWellWidget {
	w.el.Label(s)
	w.st.label = s
	return w
}

// Disabled dims the well and makes it take no input while d is true.
func (w *ColorWellWidget) Disabled(d bool) *ColorWellWidget {
	w.disabled = d
	w.el.Disabled(d)
	return w
}

// Element returns the well's element.
func (w *ColorWellWidget) Element() ui.Element { return w.el }

// DrawPalette draws the picker's palette in pn's box, the swatch of chosen,
// if any, outlined.
func DrawPalette(pn Pen, chosen ui.Color) {
	DrawWell(pn)
	for i, sc := range pickerPalette {
		x, y := 2+i%8*swatchSize, 2+i/8*swatchSize
		pn.Rect(x, y, swatchSize, swatchSize, sc)
		if sc.R == chosen.R && sc.G == chosen.G && sc.B == chosen.B {
			pn.Frame(x, y, swatchSize, swatchSize, Black)
			pn.Frame(x+1, y+1, swatchSize-2, swatchSize-2, White)
		}
	}
}
