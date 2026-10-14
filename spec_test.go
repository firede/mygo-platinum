package platinum

import (
	"bufio"
	"image"
	"image/color"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// These tests check the controls against docs/spec.md: its color tables
// against the palette, and the sizes and colors of rendered controls
// against its numbers.

func rgba(c ui.Color) color.RGBA { return color.RGBA{c.R, c.G, c.B, 255} }

func hexColor(t *testing.T, s string) color.RGBA {
	t.Helper()
	c := ui.Hex(s)
	return color.RGBA{c.R, c.G, c.B, 255}
}

func specLines(t *testing.T) []string {
	t.Helper()
	f, err := os.Open("docs/spec.md")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines
}

func TestSpecGrayRamp(t *testing.T) {
	ramp := map[string]ui.Color{
		"White": White, "Gray1": Gray1, "Gray2": Gray2, "Gray3": Gray3, "Gray4": Gray4,
		"Gray5": Gray5, "Gray6": Gray6, "Gray7": Gray7, "Gray8": Gray8, "Gray9": Gray9,
		"Gray10": Gray10, "Gray11": Gray11, "Gray12": Gray12, "Black": Black,
	}
	row := regexp.MustCompile("^\\| (\\w+) +\\| `(#[0-9A-F]{6})` \\|")
	seen := 0
	for _, l := range specLines(t) {
		m := row.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		c, ok := ramp[m[1]]
		if !ok {
			continue
		}
		seen++
		if rgba(c) != hexColor(t, m[2]) {
			t.Errorf("%s is %v, the spec says %s", m[1], rgba(c), m[2])
		}
	}
	if seen != len(ramp) {
		t.Errorf("the spec lists %d of the %d grays", seen, len(ramp))
	}
}

func TestSpecAccents(t *testing.T) {
	row := regexp.MustCompile("^\\| (\\w+)[^|]*\\| `(#[0-9A-F]{6})` \\| `(#[0-9A-F]{6})` \\| `(#[0-9A-F]{6})` \\|")
	byName := map[string]Accent{}
	for _, a := range Accents() {
		byName[a.Name] = a
	}
	seen := 0
	for _, l := range specLines(t) {
		m := row.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		a, ok := byName[m[1]]
		if !ok {
			t.Errorf("the spec lists an accent %q the library lacks", m[1])
			continue
		}
		seen++
		for i, c := range []ui.Color{a.Dark, a.Mid, a.Light} {
			if rgba(c) != hexColor(t, m[2+i]) {
				t.Errorf("%s's color %d is %v, the spec says %s", a.Name, i, rgba(c), m[2+i])
			}
		}
	}
	if seen != len(byName) {
		t.Errorf("the spec lists %d of the %d accents", seen, len(byName))
	}
}

// probe renders view and returns the frame and a function finding an
// element's box by its label.
func probe(t *testing.T, w, h int, view func(c *ui.Context)) (*image.RGBA, func(string) image.Rectangle) {
	t.Helper()
	tt := ui.NewTester(func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Gap(10).AlignItems(ui.Start).Children(func() { view(c) })
	}, w, h)
	img := tt.Image()
	return img, func(label string) image.Rectangle {
		r, ok := tt.Find(label)
		if !ok {
			t.Fatalf("no element %q", label)
		}
		return image.Rect(int(r.X), int(r.Y), int(r.X+r.W), int(r.Y+r.H))
	}
}

func expectSize(t *testing.T, what string, r image.Rectangle, w, h int) {
	t.Helper()
	if w > 0 && r.Dx() != w || h > 0 && r.Dy() != h {
		t.Errorf("%s is %dx%d, the spec says %dx%d", what, r.Dx(), r.Dy(), w, h)
	}
}

func expectPixel(t *testing.T, img *image.RGBA, what string, x, y int, want ui.Color) {
	t.Helper()
	if got := img.RGBAAt(x, y); got != rgba(want) {
		t.Errorf("%s at %d,%d is %v, want %v", what, x, y, got, rgba(want))
	}
}

func TestSpecPushButton(t *testing.T) {
	img, find := probe(t, 300, 120, func(c *ui.Context) {
		Button(c, "OK")
		Button(c, "A longer label")
		Button(c, "Go").Default()
	})
	ok := find("OK")
	expectSize(t, "a push button", ok, ButtonMinWidth, ButtonHeight)
	expectSize(t, "the minimum width", ok, 58, 20)
	long := find("A longer label")
	if want := SystemFont.Measure("A longer label") + 24; long.Dx() != want {
		t.Errorf("a long button is %d wide, want its label plus 12 on each side, %d", long.Dx(), want)
	}
	expectSize(t, "a default button", find("Go"), 58+8, 28)
	x, y := ok.Min.X, ok.Min.Y
	expectPixel(t, img, "the corner, cut", x, y, Gray2)
	expectPixel(t, img, "the outline", x+3, y, Black)
	expectPixel(t, img, "the outer highlight", x+5, y+1, White)
	expectPixel(t, img, "the inner highlight", x+5, y+2, Gray1)
	expectPixel(t, img, "the outer shadow", x+5, y+18, Gray7)
	expectPixel(t, img, "the inner shadow", x+5, y+17, Gray5)
	expectPixel(t, img, "the face", x+8, y+10, Gray2)
	g := find("Go")
	expectPixel(t, img, "the default ring's outline", g.Min.X+5, g.Min.Y, Black)
	expectPixel(t, img, "the default ring's face", g.Min.X+5, g.Min.Y+2, Gray4)
}

func TestSpecSmallControls(t *testing.T) {
	on, off := true, false
	choice := 1
	v := 0.5
	img, find := probe(t, 300, 200, func(c *ui.Context) {
		Checkbox(c, &on, "Check")
		Radio(c, &choice, 1, "Radio")
		Slider(c, &v, 0, 1).Label("Slider")
		Disclosure(c, &off, "More")
	})
	cb := find("Check")
	expectSize(t, "a check box row", cb, CheckSize+CheckGap+SystemFont.Measure("Check"), 16)
	bx, by := cb.Min.X, cb.Min.Y+2
	expectPixel(t, img, "the check box outline", bx, by, Black)
	expectPixel(t, img, "the check box outline", bx+11, by+11, Black)
	expectPixel(t, img, "the check box highlight", bx+1, by+5, White)
	expectPixel(t, img, "the check box cross", bx+2, by+2, Black)
	r := find("Radio")
	expectPixel(t, img, "the radio circle's cut corner", r.Min.X, r.Min.Y+2, Gray2)
	expectPixel(t, img, "the radio circle's top", r.Min.X+4, r.Min.Y+2, Black)
	expectPixel(t, img, "the radio dot", r.Min.X+5, r.Min.Y+2+5, Black)
	s := find("Slider")
	expectSize(t, "a slider", s, 0, SliderHeight)
	expectPixel(t, img, "the slider track's shadow", s.Min.X+3, s.Min.Y+sliderTrackTop, Gray7)
	expectPixel(t, img, "the track, unfilled left of the thumb", s.Min.X+8, s.Min.Y+sliderTrackTop+2, Gray4)
	expectSize(t, "a disclosure triangle row", find("More"), DisclosureSize+CheckGap+SystemFont.Measure("More"), 16)
}

func TestSpecFieldsAndBars(t *testing.T) {
	name := "x"
	sel := 0
	v := 0.0
	img, find := probe(t, 300, 260, func(c *ui.Context) {
		TextField(c, &name).Label("Field")
		ProgressBar(c, 0.5).Label("Progress")
		ScrollBar(c, &v, 10, 40).Element().Height(100).Label("Scroll")
		PopupMenu(c, &sel, "", []string{"One", "Two"})
	})
	f := find("Field")
	expectSize(t, "a text field", f, 0, TextFieldHeight)
	expectSize(t, "a text field", f, 0, 22)
	expectPixel(t, img, "the field's sunken edge", f.Min.X+3, f.Min.Y, Gray7)
	expectPixel(t, img, "the field's outline", f.Min.X+3, f.Min.Y+1, Black)
	expectPixel(t, img, "the field's face", f.Min.X+3, f.Min.Y+3, White)
	p := find("Progress")
	expectSize(t, "a progress bar", p, 0, 14)
	expectPixel(t, img, "the progress outline", p.Min.X+4, p.Min.Y, Black)
	expectPixel(t, img, "the progress fill", p.Min.X+4, p.Min.Y+6, Lavender.Mid)
	expectPixel(t, img, "the empty channel", p.Max.X-6, p.Min.Y+6, Gray4)
	s := find("Scroll")
	expectSize(t, "a scroll bar", s, 16, 0)
	expectPixel(t, img, "the scroll outline", s.Min.X, s.Min.Y+5, Black)
	expectPixel(t, img, "the arrow face", s.Min.X+5, s.Min.Y+5, Gray2)
	expectPixel(t, img, "the line under the arrow", s.Min.X+5, s.Min.Y+15, Black)
	pp := find("One")
	expectSize(t, "a pop-up menu", pp, 0, PopupHeight)
}

func TestSpecWindowAndMenus(t *testing.T) {
	a := &menuApp{}
	tt := ui.NewTester(func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Fill().Children(func() {
			MenuBar(c, a.menus())
			ui.Box(c).Grow(1).Children(func() {
				st := WindowState{X: 10, Y: 10, W: 200, H: 120}
				Window(c, &st, "Specs", WindowOptions{Active: true}, func() {})
			})
		})
	}, 300, 200)
	img := tt.Image()
	expectPixel(t, img, "the menu bar's top", 5, 0, White)
	expectPixel(t, img, "the menu bar's face", 5, 10, Gray2)
	expectPixel(t, img, "the menu bar's face above its bottom", 5, 18, Gray2)
	expectPixel(t, img, "the menu bar's bottom", 5, 19, Black)
	// The window, 20 down.
	wx, wy := 10, 30
	expectPixel(t, img, "the window outline", wx+40, wy, Black)
	expectPixel(t, img, "the window highlight", wx+40, wy+1, White)
	expectPixel(t, img, "a stripe's dark row", wx+30, wy+stripeTop, Gray7)
	expectPixel(t, img, "a stripe's light row", wx+31, wy+stripeTop+1, White)
	expectPixel(t, img, "the last stripe", wx+30, wy+stripeTop+10, Gray7)
	expectPixel(t, img, "the face below the stripes", wx+30, wy+stripeTop+12, Gray2)
	expectPixel(t, img, "the close box outline", wx+titleBoxInset+1, wy+stripeTop+1, Black)
	expectPixel(t, img, "the content's sunken edge", wx+40, wy+TitleBarHeight-2, Gray7)
	expectPixel(t, img, "the content's outline", wx+40, wy+TitleBarHeight-1, Black)
	expectPixel(t, img, "the side frame", wx+3, wy+60, Gray2)
	expectPixel(t, img, "the side outline", wx, wy+60, Black)
	// An open menu: its highlight.
	r, _ := tt.Find("File")
	tt.Press(r.X+5, r.Y+5)
	tt.Frame()
	s, _ := tt.Find("Save")
	tt.Move(s.X+40, s.Y+5)
	tt.Frame()
	img = tt.Image()
	expectPixel(t, img, "the open title", int(r.X)+2, 10, Lavender.Dark)
	expectPixel(t, img, "the highlighted item", int(s.X)+4, int(s.Y)+1, Lavender.Dark)
	if s.H != MenuItemHeight {
		t.Errorf("a menu item is %v tall, the spec says %d", s.H, MenuItemHeight)
	}
}

func TestSpecFaces(t *testing.T) {
	for _, tc := range []struct {
		f                       Font
		height, ascent, xh, cap int
	}{
		{SystemFont, 12, 9, 6, 9},
		{AppFont, 11, 8, 5, 8},
	} {
		face := tc.f.face()
		if face.Height != tc.height || face.Ascent != tc.ascent || face.XHeight != tc.xh || face.CapHeight != tc.cap {
			t.Errorf("%s face: height %d ascent %d x %d cap %d", face.Name, face.Height, face.Ascent, face.XHeight, face.CapHeight)
		}
		for r := rune(0x21); r < 0x7F; r++ {
			if !face.Has(r) {
				t.Errorf("%s face lacks %q", face.Name, r)
			}
		}
		// Capitals are as tall as the cap height and sit on the baseline.
		for _, r := range "EHT" {
			rows := face.Rows(r)
			if len(rows) != tc.cap || !strings.Contains(rows[0], "#") {
				t.Errorf("%s face: %q is %d rows, the cap height is %d", face.Name, r, len(rows), tc.cap)
			}
		}
	}
}

func TestSpecBevelAndToolbar(t *testing.T) {
	on := true
	img, find := probe(t, 400, 160, func(c *ui.Context) {
		BevelButton(c, "Bevel")
		BevelButton(c, "On").Toggle(&on)
		Toolbar(c, []ToolItem{{Label: "Tool"}}).Width(300).Label("Bar")
		Stepper(c, new(float64), 0, 9, 1).Label("Arrows")
	})
	b := find("Bevel")
	expectSize(t, "a bevel button", b, SystemFont.Measure("Bevel")+2*bevelPadding, BevelHeight)
	expectPixel(t, img, "a bevel button's square corner", b.Min.X, b.Min.Y, Gray10)
	expectPixel(t, img, "its outline", b.Min.X+5, b.Min.Y, Gray10)
	expectPixel(t, img, "its outer highlight", b.Min.X+5, b.Min.Y+1, White)
	expectPixel(t, img, "its face", b.Min.X+3, b.Min.Y+10, Gray2)
	o := find("On")
	expectPixel(t, img, "an on bevel button's face", o.Min.X+3, o.Min.Y+10, Gray4)
	tb := find("Bar")
	expectSize(t, "a toolbar", tb, 0, ToolbarHeight)
	expectPixel(t, img, "the toolbar's etched line", tb.Min.X+5, tb.Max.Y-2, Gray7)
	expectSize(t, "the little arrows", find("Arrows"), 13, 23)
}

func TestSpecStatusControls(t *testing.T) {
	img, find := probe(t, 300, 160, func(c *ui.Context) {
		Placard(c, "Placard")
		Badge(c, "9")
		Tag(c, "tag")
		ChasingArrows(c).Label("Busy")
	})
	p := find("Placard")
	expectSize(t, "a placard", p, 0, PlacardHeight)
	expectPixel(t, img, "a placard's top, a highlight and no outline", p.Min.X+3, p.Min.Y, White)
	expectPixel(t, img, "its bottom, a shadow", p.Min.X+3, p.Max.Y-1, Gray7)
	bd := find("9")
	expectSize(t, "a badge", bd, BadgeHeight, BadgeHeight)
	expectPixel(t, img, "a badge's square corner", bd.Min.X, bd.Min.Y, Black)
	expectPixel(t, img, "its face", bd.Min.X+1, bd.Min.Y+1, Black)
	tg := find("tag")
	expectPixel(t, img, "a tag's outline", tg.Min.X, tg.Min.Y, Black)
	expectPixel(t, img, "its face", tg.Min.X+1, tg.Min.Y+1, Gray1)
	expectSize(t, "the chasing arrows", find("Busy"), 16, 16)
}

func TestSpecLists(t *testing.T) {
	sel := 0
	img, find := probe(t, 300, 200, func(c *ui.Context) {
		ListBox(c, &sel, []string{"One", "Two"}).Element().Size(120, 60).Label("List")
	})
	l := find("List")
	expectPixel(t, img, "a list's sunken ring, top", l.Min.X+5, l.Min.Y, Gray7)
	expectPixel(t, img, "its sunken ring, bottom", l.Min.X+5, l.Max.Y-1, White)
	expectPixel(t, img, "its outline", l.Min.X+5, l.Min.Y+1, Black)
	expectPixel(t, img, "its face", l.Min.X+5, l.Min.Y+2+ListRowHeight+3, White)
	expectPixel(t, img, "the unfocused selection", l.Min.X+5, l.Min.Y+2+3, Gray4)
}

func TestSpecCalendarSheetRuler(t *testing.T) {
	day := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	st := NewSheetState()
	rs := &RulerState{Left: 16, Right: 200}
	img, find := probe(t, 420, 420, func(c *ui.Context) {
		Calendar(c, &day)
		Sheet(c, st, 10, 4).Element().Size(330, 100)
		Ruler(c, rs).Width(300)
	})
	expectSize(t, "a calendar", find("Calendar"), CalendarWidth, CalendarHeight)
	expectSize(t, "a ruler", find("Ruler"), 300, RulerHeight)
	r := find("Ruler")
	expectPixel(t, img, "the ruler's bottom line", r.Min.X+5, r.Max.Y-1, Black)
	expectPixel(t, img, "the scale between the margins", r.Min.X+20, r.Min.Y+4, White)
	expectPixel(t, img, "the scale outside them", r.Min.X+250, r.Min.Y+4, Gray1)
	s := find("Sheet")
	// The cursor, A1, framed two pixels wide in black while unfocused.
	cx := s.Min.X + 2 + SheetRowHeadW
	cy := s.Min.Y + 2 + SheetCellH
	expectPixel(t, img, "the cursor's outer frame", cx+10, cy-1, Black)
	expectPixel(t, img, "the cursor's inner frame", cx+10, cy, Black)
	expectPixel(t, img, "a cell", cx+10, cy+SheetCellH+5, White)
	expectPixel(t, img, "the line between cells", cx+10, cy+2*SheetCellH-1, Gray3)
}

func TestSpecUtilityWindow(t *testing.T) {
	st := WindowState{X: 10, Y: 10, W: 160, H: 80}
	tt := ui.NewTester(func(c *ui.Context) {
		Use(c, nil)
		ui.Box(c).Fill().Children(func() {
			Window(c, &st, "Tools", WindowOptions{Utility: true, Active: true}, func() {})
		})
	}, 200, 120)
	img := tt.Image()
	expectPixel(t, img, "the utility title bar's first stripe", 10+30, 10+2, Gray7)
	expectPixel(t, img, "its close box outline", 10+4+1, 10+2+1, Black)
	expectPixel(t, img, "the line under the title bar", 10+40, 10+UtilityBarHeight-1, Black)
	expectPixel(t, img, "the thin side frame", 10+1, 10+40, White)
	expectPixel(t, img, "the content's line", 10+utilityFrame-1, 10+40, Black)
}

// TestSpecDisabledColors checks disabled controls come out in the spec's
// colors on the window's face, although MyGo paints disabled elements at
// half opacity.
func TestSpecDisabledColors(t *testing.T) {
	on := true
	v := 0.0
	img, find := probe(t, 300, 200, func(c *ui.Context) {
		Button(c, "Off").Disabled(true)
		Checkbox(c, &on, "Box").Disabled(true)
		Slider(c, &v, 0, 1).Disabled(true).Label("Dim")
	})
	b := find("Off")
	expectPixel(t, img, "a disabled button's outline", b.Min.X+5, b.Min.Y, Gray6)
	expectPixel(t, img, "its face", b.Min.X+5, b.Min.Y+3, Gray2)
	expectPixel(t, img, "its inner highlight", b.Min.X+5, b.Min.Y+1, Gray1)
	cb := find("Box")
	expectPixel(t, img, "a disabled check box's outline", cb.Min.X, cb.Min.Y+2, Gray6)
	expectPixel(t, img, "its mark", cb.Min.X+2, cb.Min.Y+4, Gray6)
	s := find("Dim")
	expectPixel(t, img, "a disabled slider's track line", s.Min.X+60, s.Min.Y+sliderTrackTop+1, Gray6)
}
