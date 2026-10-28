package diagram

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
	"github.com/firede/limu"
	"github.com/firede/limu/scene"
	"github.com/firede/limu/text"
	platinum "github.com/firede/mygo-platinum"
)

var update = flag.Bool("update", false, "write the golden images of failing snapshots anew")

// The flowcharts in testdata, laid out and drawn at scale 1, compared pixel by pixel with
// testdata/golden.
func TestSnapshots(t *testing.T) {
	for _, name := range []string{"order", "desk", "fills"} {
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(filepath.Join("testdata", name+".mmd"))
			if err != nil {
				t.Fatal(err)
			}
			sc, err := limu.Layout(string(src), Layout())
			if err != nil {
				t.Fatal(err)
			}
			img := ui.Render(func(c *ui.Context) {
				platinum.Use(c, nil)
				ui.Box(c).Fill().Background(platinum.White).Padding(8).Children(func() { View(c, sc) })
			}, int(sc.Width)+16, int(sc.Height)+16, 1)
			compareGolden(t, name, img)
		})
	}
}

// Text is measured as AppFont draws it, in its styles.
func TestMeasurer(t *testing.T) {
	for _, f := range []text.Face{0, text.Bold, text.Italic, text.Bold | text.Italic, text.Mono} {
		for _, s := range []string{"", "Notes", "Quality check", "a b"} {
			if got, want := Measurer.Width(s, 99, f), float64(platinum.AppFont.MeasureStyled(s, textStyle(f))); got != want {
				t.Errorf("Width(%q, %v) = %v, want %v", s, f, got, want)
			}
		}
	}
}

// A box filled is the pixels within it, and its border the ring of them at its edge.
func TestFillBox(t *testing.T) {
	sq := new(scene.Path).Polygon(2, 3, 7, 3, 7, 6, 2, 6)
	px := fill(sq.Flatten(0.25))
	if len(px) != 15 || !px.has(2, 3) || !px.has(6, 5) || px.has(7, 3) || px.has(2, 6) {
		t.Errorf("fill: %d pixels %v", len(px), px)
	}
	if b := px.border(); len(b) != 12 || !b.has(2, 4) || b.has(3, 4) {
		t.Errorf("border: %d pixels %v", len(b), b)
	}
}

// A line covers each pixel from its start to its end once, eight-connected.
func TestLines(t *testing.T) {
	px := pixels{}
	stroke{width: 1, at: floorAt}.lines(px, []scene.Polyline{{Points: []scene.Point{{X: 0, Y: 0}, {X: 6, Y: 3}}}})
	if len(px) != 7 || !px.has(0, 0) || !px.has(6, 3) {
		t.Errorf("line: %v", px)
	}
	px = pixels{}
	stroke{width: 1, dash: []int{2, 1}, at: floorAt}.lines(px, []scene.Polyline{{Points: []scene.Point{{X: 0, Y: 0}, {X: 8, Y: 0}}}})
	for x := range 9 {
		if px.has(x, 0) != (x%3 != 2) {
			t.Errorf("dashes: %v", px)
			break
		}
	}
}

// An arrowhead down a column ends on the pixel above its tip, centered on the column, and
// one up it on the pixel at its tip.
func TestArrowhead(t *testing.T) {
	for _, m := range []scene.Marker{
		{Type: "arrow_point", At: scene.Point{X: 10, Y: 20}, DY: 1},
		{Type: "arrow_point", At: scene.Point{X: 10, Y: 20}, DY: -1},
		{Type: "arrow_point", At: scene.Point{X: 20, Y: 10}, DX: 1},
	} {
		px := pixels{}
		arrowhead(px, m)
		tip := pixel{int(m.At.X - max(m.DX, 0)), int(m.At.Y - max(m.DY, 0))}
		if len(px) != 1+1+3+3+5+5+7+7 || !px.has(tip.x, tip.y) {
			t.Errorf("%+v: %d pixels, tip %v: %v", m, len(px), tip, px)
		}
		back := pixel{tip.x - int(m.DX)*7, tip.y - int(m.DY)*7}
		across := pixel{int(m.DY) * 3, int(m.DX) * 3}
		if !px.has(back.x+across.x, back.y+across.y) || !px.has(back.x-across.x, back.y-across.y) {
			t.Errorf("%+v: base not 7 wide about %v: %v", m, back, px)
		}
	}
}

func compareGolden(t *testing.T, name string, img *image.RGBA) {
	t.Helper()
	path := filepath.Join("testdata", "golden", name+".png")
	data, err := os.ReadFile(path)
	if err != nil {
		if *update {
			writePNG(t, path, img)
			return
		}
		t.Fatalf("golden %s: %v (run go test -update after checking the image)", path, err)
	}
	want, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if diff := diffImages(want, img); diff != "" {
		out := filepath.Join(os.TempDir(), "platinum-diagram-"+name+".png")
		writePNG(t, out, img)
		if *update {
			writePNG(t, path, img)
			t.Logf("updated %s: %s", path, diff)
			return
		}
		t.Fatalf("%s differs from its golden: %s; got %s", name, diff, out)
	}
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
func diffImages(want image.Image, got *image.RGBA) string {
	b := want.Bounds()
	if b.Size() != got.Bounds().Size() {
		return fmt.Sprintf("size %v, want %v", got.Bounds().Size(), b.Size())
	}
	n, first := 0, image.Point{-1, -1}
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			r1, g1, b1, a1 := want.At(b.Min.X+x, b.Min.Y+y).RGBA()
			r2, g2, b2, a2 := got.At(x, y).RGBA()
			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
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
	return fmt.Sprintf("%d pixels differ, the first at %v", n, first)
}
