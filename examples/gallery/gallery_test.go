package main

import (
	"bytes"
	"image"
	"image/draw"
	"image/png"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
)

// TestGalleryImage renders the first page of the gallery and compares it,
// pixel by pixel, with testdata/golden/page-1.png, its golden image.
func TestGalleryImage(t *testing.T) { compareGallery(t, 0) }

// TestGalleryPage2Image does the same for the second page.
func TestGalleryPage2Image(t *testing.T) { compareGallery(t, 1) }

// TestGalleryPage3Image does the same for the third page.
func TestGalleryPage3Image(t *testing.T) { compareGallery(t, 2) }

// TestGalleryPage4Image does the same for the fourth page.
func TestGalleryPage4Image(t *testing.T) { compareGallery(t, 3) }

// TestGalleryPage5Image does the same for the fifth page.
func TestGalleryPage5Image(t *testing.T) { compareGallery(t, 4) }

func compareGallery(t *testing.T, page int) {
	path := "testdata/golden/page-" + strconv.Itoa(page+1) + ".png"
	defer pt.SetClock(pt.SetClock(func() time.Time { return time.Unix(0, 0) }))
	g := newGallery()
	g.page = page
	got := ui.Render(g.view, width, height, 1)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	m, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	want := image.NewRGBA(m.Bounds())
	draw.Draw(want, want.Bounds(), m, m.Bounds().Min, draw.Src)
	if want.Bounds() != got.Bounds() {
		t.Fatalf("gallery is %v, %s %v", got.Bounds(), path, want.Bounds())
	}
	n := 0
	for i := range want.Pix {
		if want.Pix[i] != got.Pix[i] {
			n++
		}
	}
	if n > 0 {
		out := t.TempDir() + "/gallery.png"
		f, _ := os.Create(out)
		png.Encode(f, got)
		f.Close()
		t.Fatalf("%d bytes differ from %s; got %s", n, path, out)
	}
}

// TestGalleryInteracts opens the alert from the default button and an
// accent from the Appearance menu.
func TestGalleryInteracts(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return time.Unix(0, 0) }))
	g := newGallery()
	tt := ui.NewTester(g.view, width, height)
	tt.Key(0, ui.KeyEnter) // the default button, OK
	tt.Frame()
	if !g.alertOpen || !tt.HasText("This is the gallery of the Platinum controls.") {
		t.Fatal("Return did not open the alert")
	}
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if g.alertOpen {
		t.Fatal("Return did not close the alert")
	}
	r, _ := tt.Find("Appearance")
	tt.ClickAt(r.X+5, r.Y+5)
	tt.Frame()
	h, ok := tt.Find("Harbor")
	if !ok {
		t.Fatal("Appearance menu did not open")
	}
	tt.ClickAt(h.X+5, h.Y+5)
	tt.Frame()
	tt.Frame()
	if g.theme.Accent.Name != "Harbor" {
		t.Fatalf("accent is %s", g.theme.Accent.Name)
	}
}

// TestGalleryScalesByWholePixels renders every page at scales 1 and 2 and
// checks that the second is the first with each pixel a 2×2 block: nothing
// is blended or moved by half a pixel at the higher scale.
func TestGalleryScalesByWholePixels(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return time.Unix(0, 0) }))
	for page := 0; page < 5; page++ {
		g := newGallery()
		g.page = page
		one := ui.Render(g.view, width, height, 1)
		g = newGallery()
		g.page = page
		two := ui.Render(g.view, width, height, 2)
		bad, first := 0, image.Point{}
		for y := 0; y < 2*height; y++ {
			for x := 0; x < 2*width; x++ {
				if two.RGBAAt(x, y) != one.RGBAAt(x/2, y/2) {
					if bad == 0 {
						first = image.Pt(x, y)
					}
					bad++
				}
			}
		}
		if bad > 0 {
			t.Errorf("page %d: %d pixels at scale 2 are not blocks of scale 1's, the first at %v", page+1, bad, first)
		}
	}
}

// TestGalleryFocusNamed tabs round every page of the gallery: each stop
// the keyboard focus makes must have a name for assistive technology (a
// label of its own or of what holds it), and Tab must come back to the
// first, no control keeping it.
func TestGalleryFocusNamed(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return time.Unix(0, 0) }))
	for page := 0; page < 5; page++ {
		g := newGallery()
		g.page = page
		tt := ui.NewTester(g.view, width, height)
		tt.Frame()
		// Round the page, back to the first stop: no control traps Tab.
		first, back := "", false
		for i := 0; i < 80 && !back; i++ {
			tt.Key(0, ui.KeyTab)
			tt.Frame()
			named := ""
			for _, s := range tt.Texts() {
				if tt.Focused(s) {
					named = s
					break
				}
			}
			if named == "" {
				t.Errorf("page %d: focus stop %d has no name", page+1, i+1)
			}
			if i == 0 {
				first = named
			} else if named == first {
				back = true
			}
		}
		if !back {
			t.Errorf("page %d: 80 tabs did not come back to the first stop, %q", page+1, first)
		}
		// Backward, from the end: past a control that keeps Tab for itself,
		// as a sheet moving between its cells.
		g = newGallery()
		g.page = page
		tt = ui.NewTester(g.view, width, height)
		tt.Frame()
		for i := 0; i < 30; i++ {
			tt.Key(ui.Shift, ui.KeyTab)
			tt.Frame()
			named := false
			for _, s := range tt.Texts() {
				named = named || tt.Focused(s)
			}
			if !named {
				t.Errorf("page %d: focus stop %d from the end has no name", page+1, i+1)
			}
		}
	}
}

// TestGalleryRandomInput clicks, drags, scrolls and presses keys at random
// on every page of the gallery: nothing may panic. PLATINUM_SEED picks
// another run of it.
func TestGalleryRandomInput(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return time.Unix(0, 0) }))
	seed, _ := strconv.Atoi(os.Getenv("PLATINUM_SEED"))
	rng := rand.New(rand.NewSource(int64(1997 + seed)))
	keys := []ui.Key{ui.KeyTab, ui.KeyEnter, ui.KeyEscape, ui.KeySpace, ui.KeyUp, ui.KeyDown, ui.KeyLeft, ui.KeyRight, ui.KeyBackspace, ui.KeyHome, ui.KeyEnd, ui.KeyA, ui.KeyZ, ui.KeyO, ui.KeyS}
	for page := 0; page < 5; page++ {
		g := newGallery()
		g.page = page
		tt := ui.NewTester(g.view, width, height)
		tt.Frame()
		for i := 0; i < 120; i++ {
			x, y := float32(rng.Intn(width)), float32(rng.Intn(height))
			switch rng.Intn(6) {
			case 0, 1:
				tt.ClickAtWith([]ui.Modifiers{0, ui.Shift, ui.Cmd}[rng.Intn(3)], x, y)
			case 2:
				tt.Press(x, y)
				tt.Move(x+float32(rng.Intn(60)-30), y+float32(rng.Intn(60)-30))
				tt.Release(x+float32(rng.Intn(60)-30), y+float32(rng.Intn(60)-30))
			case 3:
				tt.Key([]ui.Modifiers{0, ui.Shift, ui.Cmd}[rng.Intn(3)], keys[rng.Intn(len(keys))])
			case 4:
				tt.Type("7a")
			case 5:
				tt.Scroll(x, y, 0, float32(rng.Intn(60)-30))
			}
			tt.Frame()
		}
	}
}

// TestGalleryFileDialogs shows the open and save dialogs from the File
// menu's keyboard equivalents, and closes them.
func TestGalleryFileDialogs(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return time.Unix(0, 0) }))
	g := newGallery()
	tt := ui.NewTester(g.view, width, height)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyO)
	tt.Frame()
	if !g.opening || !tt.HasText("Eject") {
		t.Fatal("Command-O did not show the open dialog")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	tt.Key(ui.Cmd|ui.Shift, ui.KeyS)
	tt.Frame()
	if !g.saving || !tt.HasText("New Folder") || g.opening {
		t.Fatal("Command-Shift-S did not show the save dialog, or the open one stayed")
	}
	tt.Key(0, ui.KeyEscape) // its Cancel button
	tt.Frame()
	if g.saving {
		t.Fatal("Escape did not close the save dialog")
	}
}
