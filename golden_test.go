package platinum

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

var update = flag.Bool("update", false, "write the golden images of failing snapshots anew")

// snap renders view in a w×h window at scale 1 and compares it, pixel by
// pixel, with testdata/golden/name.png.
func snap(t *testing.T, name string, w, h int, view func(c *ui.Context)) {
	t.Helper()
	img := ui.Render(func(c *ui.Context) {
		Use(c, nil)
		view(c)
	}, w, h, 1)
	compareGolden(t, name, img)
}

// snapScale is snap at another scale.
func snapScale(t *testing.T, name string, w, h int, scale float32, view func(c *ui.Context)) {
	t.Helper()
	img := ui.Render(func(c *ui.Context) {
		Use(c, nil)
		view(c)
	}, w, h, scale)
	compareGolden(t, name, img)
}

func compareGolden(t *testing.T, name string, img *image.RGBA) {
	t.Helper()
	path := filepath.Join("testdata", "golden", name+".png")
	want, err := readPNG(path)
	if err != nil {
		if *update {
			writePNG(t, path, img)
			return
		}
		t.Fatalf("golden %s: %v (run go test -update after checking the image)", path, err)
	}
	if diff := diffImages(want, img); diff != "" {
		out := filepath.Join(os.TempDir(), "platinum-"+name+".png")
		writePNG(t, out, img)
		if *update {
			writePNG(t, path, img)
			t.Logf("updated %s: %s", path, diff)
			return
		}
		t.Fatalf("%s differs from its golden: %s; got %s", name, diff, out)
	}
}

func readPNG(path string) (*image.RGBA, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	m, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	b := m.Bounds()
	rgba := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			rgba.Set(x, y, m.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return rgba, nil
}

func writePNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

// diffImages describes how got differs from want, or returns "".
func diffImages(want, got *image.RGBA) string {
	if want.Bounds().Size() != got.Bounds().Size() {
		return fmt.Sprintf("size %v, want %v", got.Bounds().Size(), want.Bounds().Size())
	}
	n, first := 0, image.Point{-1, -1}
	b := want.Bounds()
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			if want.RGBAAt(x, y) != got.RGBAAt(x, y) {
				if n == 0 {
					first = image.Pt(x, y)
				}
				n++
			}
		}
	}
	if n == 0 {
		return ""
	}
	return fmt.Sprintf("%d pixels differ, the first at %v: %v, want %v", n, first, got.RGBAAt(first.X, first.Y), want.RGBAAt(first.X, first.Y))
}
