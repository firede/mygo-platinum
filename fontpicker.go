package platinum

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

// FontChoice is a font, size and style as a font picker chooses them.
type FontChoice struct {
	Family                  string
	Size                    float64
	Bold, Italic, Underline bool
}

// FontPickerWidget is a font picker being built.
type FontPickerWidget struct {
	el      ui.Element
	changed bool
}

var fontSizes = []float64{9, 10, 12, 14, 18, 24, 36}

// FontPicker builds a panel choosing a font, as a Format menu's dialog: a
// list of families, a list of sizes with a field for any other, the styles
// as check boxes, and a sample of the choice. families are the families
// the app has. The sample is drawn in the library's regular face, in the
// style chosen.
func FontPicker(c *ui.Context, ch *FontChoice, families []string) *FontPickerWidget {
	w := &FontPickerWidget{}
	fam := -1
	for i, f := range families {
		if f == ch.Family {
			fam = i
		}
	}
	size := -1
	names := make([]string, len(fontSizes))
	for i, s := range fontSizes {
		names[i] = fmt.Sprint(s)
		if s == ch.Size {
			size = i
		}
	}
	w.el = ui.Column(c).Gap(8).Role(ui.RoleGroup).Label(say(c, "font"))
	w.el.Children(func() {
		ui.Row(c).Gap(10).AlignItems(ui.Start).Children(func() {
			ui.Column(c).Gap(4).Children(func() {
				Label(c, say(c, "fontLabel"))
				fl := ListBox(c, &fam, families)
				fl.Element().Size(140, 98).Label(say(c, "families"))
				if fl.Changed() && fam >= 0 {
					ch.Family = families[fam]
					w.changed = true
				}
			})
			ui.Column(c).Gap(4).Children(func() {
				Label(c, say(c, "sizeLabel"))
				if NumberField(c, &ch.Size, 4, 288, 1, 0).Label(say(c, "size")).Changed() {
					w.changed = true
				}
				sl := ListBox(c, &size, names)
				sl.Element().Size(76, 70).Label(say(c, "sizes"))
				if sl.Changed() && size >= 0 {
					ch.Size = fontSizes[size]
					w.changed = true
				}
			})
			ui.Column(c).Gap(2).Children(func() {
				Label(c, say(c, "styleLabel"))
				for _, s := range []struct {
					name string
					v    *bool
				}{{say(c, "bold"), &ch.Bold}, {say(c, "italic"), &ch.Italic}, {say(c, "underline"), &ch.Underline}} {
					if Checkbox(c, s.v, s.name).Changed() {
						w.changed = true
					}
				}
			})
		})
		choice := *ch
		sample := say(c, "sampleText")
		ui.Box(c).Height(40).AlignSelf(ui.Stretch).Role(ui.RoleImage).Label(say(c, "sample")).Draw(func(p *ui.Painter, r ui.Rect) {
			pn := NewPen(p, r)
			DrawWell(pn)
			f := AppFont
			k := max(1, int(choice.Size+6)/12)
			var style TextStyle
			if choice.Bold {
				style |= Bold
			}
			if choice.Italic {
				style |= Italic
			}
			if choice.Underline {
				style |= Underline
			}
			text := choice.Family + " " + fmt.Sprint(choice.Size)
			pn.Clip(3, 3, pn.W-6, pn.H-6, func() {
				if k == 1 {
					pn.TextStyled(8, (pn.H-f.Height())/2, sample+" — "+text, f, Black, style)
					return
				}
				x := pn.TextScaled(8, (pn.H-f.Height()*k)/2, "Aa "+text, f, Black, k)
				if choice.Underline {
					pn.Rect(8, (pn.H-f.Height()*k)/2+(f.Ascent()+1)*k, x-8, k, Black)
				}
			})
		})
	})
	return w
}

// Changed reports whether the user changed the choice since the last frame.
func (w *FontPickerWidget) Changed() bool { return w.changed }

// Element returns the picker's element.
func (w *FontPickerWidget) Element() ui.Element { return w.el }
