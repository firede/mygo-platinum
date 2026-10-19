// Package diagram draws Mermaid diagrams in the Platinum look: laid out by
// limu on whole pixels, with the library's bitmap faces, and drawn with its
// Pen — boxes as its controls are drawn, lines and arrowheads a pixel at a
// time.
//
//	sc, err := limu.Layout(src, diagram.Layout())
//	...
//	diagram.View(c, sc)
//
// It is the third level of limu's styles at work: the look's theme and its
// overrides set the colors, as for any backend, and a drawing of its own
// replaces how each kind of part is drawn.
package diagram

import (
	"github.com/egoist/mygo/ui"
	"github.com/firede/limu/layout"
	"github.com/firede/limu/mygoui"
	"github.com/firede/limu/scene"
	"github.com/firede/limu/style"
	"github.com/firede/limu/text"
	platinum "github.com/firede/mygo-platinum"
)

// FontSize is the size layout is told the text is: the bitmap faces have one
// size, and this one gives their lines room.
const FontSize = 12

// Measurer measures text in AppFont, as the drawing sets it: bold and italic
// as TextStyled draws them, and code in the same face, as it has no
// monospaced one.
var Measurer text.Measurer = measurer{}

type measurer struct{}

func (measurer) Width(s string, size float64, f text.Face) float64 {
	return float64(platinum.AppFont.MeasureStyled(s, textStyle(f)))
}

func textStyle(f text.Face) platinum.TextStyle {
	var st platinum.TextStyle
	if f&text.Bold != 0 {
		st |= platinum.Bold
	}
	if f&text.Italic != 0 {
		st |= platinum.Italic
	}
	return st
}

// Layout returns the layout options of the look: text measured in the
// bitmap faces, edges in runs across and down, and everything on whole
// pixels.
func Layout() *layout.Options {
	return &layout.Options{Measurer: Measurer, FontSize: FontSize, Snap: 1, Routing: layout.Orthogonal}
}

// Theme is the limu theme of a Platinum theme: white paper, black lines and
// text, faces of the gray ramp, and the accent's colors.
func Theme(th *platinum.Theme) *style.Theme {
	if th == nil {
		th = platinum.DefaultTheme()
	}
	c := func(u ui.Color) style.Color { return style.RGB(u.R, u.G, u.B) }
	accents := platinum.Accents()
	var series [8]style.Color
	for i := range series {
		if i < len(accents) {
			series[i] = c(accents[i].Mid)
		} else {
			series[i] = c([]ui.Color{platinum.Gray5, platinum.Gray8}[i-len(accents)])
		}
	}
	return &style.Theme{
		Bg:          c(platinum.White),
		Fg:          c(platinum.Black),
		Muted:       c(platinum.Gray8),
		Line:        c(platinum.Black),
		Surface:     c(platinum.Gray2),
		SurfaceAlt:  c(platinum.Gray3),
		Border:      c(platinum.Black),
		Accent:      c(th.Accent.Dark),
		Critical:    c(platinum.Ember.Dark),
		Series:      series,
		Radius:      2,
		StrokeWidth: 1,
	}
}

// Options returns mygoui's options for the look, in a Platinum theme: its
// colors, and the drawing of its own. Overrides of roles may be added to
// its Look.
func Options(th *platinum.Theme) *mygoui.Options {
	return &mygoui.Options{Look: style.Look{Theme: Theme(th)}, Scale: 1, Draw: Draw}
}

// Paint draws a scene laid out with Layout on p, its top left corner at x,
// y, in the look of th.
func Paint(p *ui.Painter, sc *scene.Scene, x, y float32, th *platinum.Theme) {
	mygoui.Paint(p, sc, x, y, Options(th))
}

// View shows a scene laid out with Layout, in the theme of the window, and
// reports what the pointer is over and what was clicked.
func View(c *ui.Context, sc *scene.Scene) *mygoui.Widget {
	return mygoui.View(c, sc, Options(platinum.ThemeOf(c)))
}
