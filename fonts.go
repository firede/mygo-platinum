package platinum

import (
	"image"
	"image/color"
	"sync"

	"github.com/egoist/mygo/ui"
	"github.com/firede/platinum-fonts/bitmap"
	"github.com/firede/platinum-fonts/sans"
	"github.com/firede/platinum-fonts/sanssmall"
)

// face is a Latin face of the library, packed in an atlas: its glyphs side
// by side, the top of each cell Above rows down from the atlas's top.
type face struct {
	Name                                         string
	Height, Ascent, CapHeight, XHeight, Tracking int
	// Above is how many rows over the cell a mark over a capital takes.
	Above  int
	Glyphs map[rune]atlasGlyph
	Atlas  *image.RGBA
	src    *bitmap.Face
}

// atlasGlyph is a glyph of a face: Width columns of the atlas from X.
type atlasGlyph struct{ X, Width, Advance int }

// tracking is the blank column after every glyph of the faces.
const tracking = 1

func newFace(name string, src *bitmap.Face, capHeight, xHeight int) *face {
	f := &face{
		Name: name, Height: src.Ascent() + src.Descent(), Ascent: src.Ascent(),
		CapHeight: capHeight, XHeight: xHeight, Tracking: tracking,
		Glyphs: map[rune]atlasGlyph{}, src: src,
	}
	var runes []rune
	src.Runes(func(r rune) bool { runes = append(runes, r); return true })
	x := 0
	for _, r := range runes {
		g, _ := src.Glyph(r)
		f.Above = max(f.Above, -g.Bounds.Min.Y-f.Ascent)
		f.Glyphs[r] = atlasGlyph{X: x, Width: g.Bounds.Dx(), Advance: g.Advance}
		x += g.Bounds.Dx() + 1
	}
	f.Atlas = image.NewRGBA(image.Rect(0, 0, max(x, 1), f.Above+f.Height))
	for _, r := range runes {
		g, _ := src.Glyph(r)
		at := f.Glyphs[r].X - g.Bounds.Min.X
		for y := g.Bounds.Min.Y; y < g.Bounds.Max.Y; y++ {
			for gx := g.Bounds.Min.X; gx < g.Bounds.Max.X; gx++ {
				if g.At(gx, y) {
					f.Atlas.SetRGBA(at+gx, f.Above+f.Ascent+y, color.RGBA{255, 255, 255, 255})
				}
			}
		}
	}
	return f
}

// Glyph returns the glyph of r, or of the replacement character where the
// face lacks r.
func (f *face) Glyph(r rune) atlasGlyph {
	if g, ok := f.Glyphs[r]; ok {
		return g
	}
	if g, ok := f.Glyphs[0xFFFD]; ok {
		return g
	}
	return f.Glyphs['?']
}

// Has reports whether the face draws r.
func (f *face) Has(r rune) bool {
	_, ok := f.Glyphs[r]
	return ok
}

// The faces: SystemFont is Platinum Sans Bold, AppFont Platinum Sans Small
// Regular; bold, for styled text, Platinum Sans Small Bold and Platinum
// Sans Bold emboldened.
var faces = sync.OnceValue(func() (fs struct{ regular, bold [2]*face }) {
	systemBold, err := bitmap.Embolden(sans.Bold())
	if err != nil {
		panic(err)
	}
	fs.regular[SystemFont] = newFace("System", sans.Bold(), 9, 6)
	fs.regular[AppFont] = newFace("App", sanssmall.Regular(), 8, 5)
	fs.bold[SystemFont] = newFace("System bold", systemBold, 9, 6)
	fs.bold[AppFont] = newFace("App bold", sanssmall.Bold(), 8, 5)
	return fs
})

var (
	atlasMu sync.Mutex
	atlases = map[*face]*ui.Bitmap{}
)

func atlasOf(f *face) *ui.Bitmap {
	atlasMu.Lock()
	defer atlasMu.Unlock()
	b, ok := atlases[f]
	if !ok {
		b = ui.NewBitmap(f.Atlas).Pixelated()
		atlases[f] = b
	}
	return b
}

// The East Asian faces SetEastAsian gives, and the glyphs drawn of them.
var eastAsian struct {
	sync.Mutex
	regular, bold func() *bitmap.Face
	glyphs        map[eastAsianKey]glyph
}

type eastAsianKey struct {
	bold bool
	r    rune
}

// SetEastAsian makes the controls draw the characters of East Asian
// scripts, which the library's faces lack, with the 12px faces regular
// and bold return: those of a package of platinum-fonts' cjk12, whose
// forms are of the language the app shows. Bold draws them in the bold
// face, SystemFont, and in bold styled text. Characters neither face draws
// are drawn in the system's font, as they all are until SetEastAsian is
// called; nil functions return to it.
//
//	import "github.com/firede/platinum-fonts/cjk12/zhhans"
//
//	pt.SetEastAsian(zhhans.Regular, zhhans.Bold)
func SetEastAsian(regular, bold func() *bitmap.Face) {
	eastAsian.Lock()
	defer eastAsian.Unlock()
	eastAsian.regular, eastAsian.bold = regular, bold
	eastAsian.glyphs = map[eastAsianKey]glyph{}
}

// glyph is what draws a character: a region of a bitmap, w columns from x
// and rows tall, up of them over the baseline, left of the pen from left;
// eastAsian for a glyph of the faces SetEastAsian gives.
type glyph struct {
	bmp           *ui.Bitmap
	x, w          int
	up, rows      int
	left, advance int
	eastAsian     bool
}

// glyph returns the glyph that draws r in f, bold or not: the face's own,
// or the East Asian face's. It reports false for a character neither has.
func (f Font) glyph(r rune, bold bool) (glyph, bool) {
	fs := faces()
	fc := fs.regular[f]
	if bold {
		fc = fs.bold[f]
	}
	if g, ok := fc.Glyphs[r]; ok {
		return glyph{
			bmp: atlasOf(fc), x: g.X, w: g.Width,
			up: fc.Ascent + fc.Above, rows: fc.Above + fc.Height, advance: g.Advance,
		}, true
	}
	return eastAsianGlyph(r, bold || f == SystemFont)
}

func eastAsianGlyph(r rune, bold bool) (glyph, bool) {
	eastAsian.Lock()
	defer eastAsian.Unlock()
	key := eastAsianKey{bold, r}
	if g, ok := eastAsian.glyphs[key]; ok {
		return g, g.bmp != nil
	}
	src := eastAsian.regular
	if bold {
		src = eastAsian.bold
	}
	if src == nil {
		return glyph{}, false
	}
	g, ok := src().Glyph(r)
	if !ok || g.Bounds.Empty() && g.Advance == 0 {
		eastAsian.glyphs[key] = glyph{}
		return glyph{}, false
	}
	b := g.Bounds
	img := image.NewRGBA(image.Rect(0, 0, max(b.Dx(), 1), max(b.Dy(), 1)))
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if g.At(x, y) {
				img.SetRGBA(x-b.Min.X, y-b.Min.Y, color.RGBA{255, 255, 255, 255})
			}
		}
	}
	gl := glyph{
		bmp: ui.NewBitmap(img).Pixelated(), w: b.Dx(),
		up: -b.Min.Y, rows: b.Dy(), left: b.Min.X, advance: g.Advance,
		eastAsian: true,
	}
	eastAsian.glyphs[key] = gl
	return gl, true
}

// Fallback returns the character drawn for one the face lacks: the
// replacement character, or a question mark.
func (f *face) Fallback() rune {
	if f.Has(0xFFFD) {
		return 0xFFFD
	}
	return '?'
}
