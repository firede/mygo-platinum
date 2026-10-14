package platinum

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	"github.com/egoist/mygo/ui"
)

func findCenter(t *testing.T, tt *ui.Tester, label string) (float32, float32) {
	t.Helper()
	r, ok := tt.Find(label)
	if !ok {
		t.Fatalf("nothing shows %q; texts %q", label, tt.Texts())
	}
	return r.X + r.W/2, r.Y + r.H/2
}

var folderMini = []string{
	"###.......",
	"#..######.",
	"#.......#.",
	"#.......#.",
	"#.......#.",
	"#########.",
}

func sampleTree() []*TreeNode {
	return []*TreeNode{
		{Title: "Documents", Icon: folderMini, Open: true, Children: []*TreeNode{
			{Title: "Letters", Icon: folderMini, Children: []*TreeNode{{Title: "To Ada"}, {Title: "To Grace"}}},
			{Title: "Recipes"},
		}},
		{Title: "Pictures", Icon: folderMini, Children: []*TreeNode{{Title: "Beach"}}},
		{Title: "Read Me"},
	}
}

type fruit struct {
	name  string
	color string
	price float64
}

var produce = []fruit{
	{"Kiwi", "green", 0.5}, {"Apricot", "orange", 1.25}, {"Banana", "yellow", 0.25},
	{"Cherry", "red", 4}, {"Fig", "purple", 2.5}, {"Lemon", "yellow", 0.75},
	{"Grape", "green", 3}, {"Date", "brown", 6}, {"Mango", "orange", 1.5},
}

func produceCell(r, col int) string {
	f := produce[r]
	switch col {
	case 0:
		return f.name
	case 1:
		return f.color
	}
	return fmt.Sprintf("%.2f", f.price)
}

var produceCols = []Column{{Title: "Name", Width: 80}, {Title: "Color", Width: 70}, {Title: "Price", Width: 60, Right: true}}

func testIcon() *ui.Bitmap {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 4; y < 28; y++ {
		for x := 6; x < 26; x++ {
			c := color.RGBA{0x99, 0x88, 0xDD, 255}
			if x == 6 || x == 25 || y == 4 || y == 27 {
				c = color.RGBA{0, 0, 0, 255}
			}
			img.SetRGBA(x, y, c)
		}
	}
	return ui.NewBitmap(img).Pixelated()
}
