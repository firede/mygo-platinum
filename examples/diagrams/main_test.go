package main

import (
	"bytes"
	"flag"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/firede/limu"
	"github.com/firede/limu/scene"
	platinum "github.com/firede/mygo-platinum"
	"github.com/firede/mygo-platinum/diagram"
)

var update = flag.Bool("update", false, "write the golden images anew")

// Every example of every type lays out and draws in the look, and the
// window shows each type.
func TestSamples(t *testing.T) {
	a := newApp()
	// limu's gallery has every type limu reads but info, which shows only
	// limu's version.
	if want := len(limu.Types()) - 1; len(a.types) != want {
		t.Errorf("%d types of samples, want %d", len(a.types), want)
	}
	for i, typ := range a.types {
		if len(typ.examples) == 0 {
			t.Errorf("%s has no examples", typ.title)
		}
		for j, e := range typ.examples {
			sc, err := limu.Layout(e.src, diagram.Layout())
			if err != nil {
				t.Errorf("%s, %q: %v", typ.title, e.title, err)
				continue
			}
			if len(sc.Items) == 0 || sc.Width <= 0 || sc.Height <= 0 {
				t.Errorf("%s, %q: an empty scene", typ.title, e.title)
			}
			draw(sc)
			if j == 0 {
				a.choose(i, 0)
				ui.Render(a.view, width, height, 1)
				if a.err != "" || a.sc == nil {
					t.Errorf("%s in the window: %q", typ.title, a.err)
				}
			}
		}
	}
}

// The first example of each type, drawn at scale 1, compared pixel by pixel
// with testdata/golden.
func TestSnapshots(t *testing.T) {
	for _, typ := range samples() {
		name := strings.ToLower(strings.ReplaceAll(typ.title, " ", "-"))
		t.Run(name, func(t *testing.T) {
			sc, err := limu.Layout(typ.examples[0].src, diagram.Layout())
			if err != nil {
				t.Fatal(err)
			}
			golden(t, name, draw(sc))
		})
	}
}

// draw renders a scene in the look, on white paper with a margin of 8.
func draw(sc *scene.Scene) *image.RGBA {
	return ui.Render(func(c *ui.Context) {
		platinum.Use(c, nil)
		ui.Box(c).Fill().Background(platinum.White).Padding(8).Children(func() { diagram.View(c, sc) })
	}, int(sc.Width)+16, int(sc.Height)+16, 1)
}

func golden(t *testing.T, name string, img *image.RGBA) {
	t.Helper()
	path := filepath.Join("testdata", "golden", name+".png")
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("golden %s: %v (run go test -update after checking the image)", path, err)
	}
	want, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if n, at := differ(want, img); n > 0 {
		out := filepath.Join(os.TempDir(), "platinum-diagrams-"+name+".png")
		if err := os.WriteFile(out, buf.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Fatalf("%s differs from its golden: %d pixels, the first at %v; got %s", name, n, at, out)
	}
}

// differ counts the pixels that differ between two images, and gives the
// first; images of other sizes differ everywhere.
func differ(a, b image.Image) (int, image.Point) {
	if a.Bounds() != b.Bounds() {
		return a.Bounds().Dx() * a.Bounds().Dy(), a.Bounds().Min
	}
	n, first := 0, image.Point{}
	r := a.Bounds()
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if a.At(x, y) != b.At(x, y) {
				if n == 0 {
					first = image.Point{x, y}
				}
				n++
			}
		}
	}
	return n, first
}
