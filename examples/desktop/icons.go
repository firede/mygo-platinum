package main

import (
	"image"
	"image/color"

	"github.com/egoist/mygo/ui"
)

// icon is a 32×32 picture of the desk, drawn by the code below rather than
// from any image file, with a darkened copy for when it is selected.
type icon struct {
	normal, selected *ui.Bitmap
	img              *image.RGBA
}

// canvas draws an icon on whole pixels.
type canvas struct{ img *image.RGBA }

var (
	black  = color.RGBA{0, 0, 0, 255}
	white  = color.RGBA{255, 255, 255, 255}
	gray1  = color.RGBA{0xEE, 0xEE, 0xEE, 255}
	gray2  = color.RGBA{0xDD, 0xDD, 0xDD, 255}
	gray4  = color.RGBA{0xBB, 0xBB, 0xBB, 255}
	gray5  = color.RGBA{0xAA, 0xAA, 0xAA, 255}
	gray7  = color.RGBA{0x88, 0x88, 0x88, 255}
	gray9  = color.RGBA{0x66, 0x66, 0x66, 255}
	yellow = color.RGBA{0xFF, 0xE0, 0x66, 255}
	amber  = color.RGBA{0xCC, 0x99, 0x33, 255}
	blue   = color.RGBA{0x66, 0x88, 0xCC, 255}
	navy   = color.RGBA{0x33, 0x44, 0x88, 255}
	red    = color.RGBA{0xCC, 0x33, 0x33, 255}
	green  = color.RGBA{0x55, 0x99, 0x55, 255}
	violet = color.RGBA{0x99, 0x88, 0xDD, 255}
)

func (k canvas) rect(x, y, w, h int, c color.RGBA) {
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			k.img.SetRGBA(i, j, c)
		}
	}
}

func (k canvas) frame(x, y, w, h int, c color.RGBA) {
	k.rect(x, y, w, 1, c)
	k.rect(x, y+h-1, w, 1, c)
	k.rect(x, y, 1, h, c)
	k.rect(x+w-1, y, 1, h, c)
}

// box is a filled rectangle with a black outline.
func (k canvas) box(x, y, w, h int, face color.RGBA) {
	k.rect(x, y, w, h, face)
	k.frame(x, y, w, h, black)
}

// disc fills the circle of radius r around cx, cy (in half pixels, for
// circles of even widths), outlined in black.
func (k canvas) disc(cx2, cy2, r2 int, face color.RGBA) {
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			dx, dy := 2*x+1-cx2, 2*y+1-cy2
			d := dx*dx + dy*dy
			if d <= r2*r2 {
				c := face
				if d > (r2-3)*(r2-3) {
					c = black
				}
				k.img.SetRGBA(x, y, c)
			}
		}
	}
}

func (k canvas) line(x0, y0, x1, y1 int, c color.RGBA) {
	dx, dy := abs(x1-x0), -abs(y1-y0)
	sx, sy := sign(x1-x0), sign(y1-y0)
	e := dx + dy
	for {
		k.img.SetRGBA(x0, y0, c)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * e
		if e2 >= dy {
			e += dy
			x0 += sx
		}
		if e2 <= dx {
			e += dx
			y0 += sy
		}
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func sign(v int) int {
	switch {
	case v < 0:
		return -1
	case v > 0:
		return 1
	}
	return 0
}

func newIcon(draw func(k canvas)) *icon {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	draw(canvas{img})
	dark := image.NewRGBA(img.Rect)
	for i := 0; i < len(img.Pix); i += 4 {
		// Selected icons darken by half, as if under a dark glass.
		dark.Pix[i] = img.Pix[i] / 2
		dark.Pix[i+1] = img.Pix[i+1] / 2
		dark.Pix[i+2] = img.Pix[i+2] / 2
		dark.Pix[i+3] = img.Pix[i+3]
	}
	return &icon{normal: ui.NewBitmap(img).Pixelated(), selected: ui.NewBitmap(dark).Pixelated(), img: img}
}

// draw draws the icon with its corner at x, y of the window.
func (ic *icon) draw(p *ui.Painter, x, y float32, selected bool) {
	b := ic.normal
	if selected {
		b = ic.selected
	}
	p.ImageRegion(b, ui.Rect{W: 32, H: 32}, ui.Rect{X: x, Y: y, W: 32, H: 32}, ui.ImageOptions{})
}

// The icons of the desk.
var (
	iconDisk = newIcon(func(k canvas) {
		k.box(2, 9, 28, 16, gray2)
		k.rect(3, 10, 26, 1, white)
		k.rect(3, 10, 1, 14, white)
		k.rect(3, 23, 26, 1, gray7)
		k.rect(28, 11, 1, 13, gray7)
		k.rect(5, 18, 14, 2, black) // the drive's slot
		k.rect(24, 18, 2, 2, green) // its light
		k.rect(3, 25, 27, 1, gray7)
	})
	iconFolder = newIcon(func(k canvas) {
		light, dark := color.RGBA{0xCC, 0xC4, 0xF4, 255}, color.RGBA{0x66, 0x5C, 0xAA, 255}
		k.box(2, 7, 12, 4, violet) // the tab
		k.rect(3, 8, 10, 1, light)
		k.box(2, 10, 28, 18, violet)
		// Lit from the top left, shaded at the bottom right, and the fold
		// of the front flap across the top.
		k.rect(3, 11, 26, 1, light)
		k.rect(3, 11, 1, 16, light)
		k.rect(28, 11, 1, 16, dark)
		k.rect(3, 26, 26, 1, dark)
		k.rect(3, 13, 25, 1, dark)
		k.rect(4, 14, 24, 1, light)
		k.rect(3, 28, 28, 1, gray7)
	})
	iconDocument = newIcon(func(k canvas) {
		k.rect(7, 2, 18, 28, white)
		k.frame(7, 2, 18, 28, black)
		for i := 0; i < 6; i++ { // the dog-ear
			k.rect(25-6+i, 2+i, 6-i, 1, color.RGBA{})
		}
		k.line(19, 2, 24, 7, black)
		k.rect(19, 2, 1, 6, black)
		k.rect(19, 7, 6, 1, black)
		for y := 11; y < 27; y += 3 {
			w := 12
			if y%2 == 0 {
				w = 9
			}
			k.rect(10, y, w, 1, gray7)
		}
		k.rect(8, 30, 18, 1, gray7)
	})
	iconTrash = newIcon(func(k canvas) {
		k.box(9, 3, 14, 3, gray4)
		k.box(5, 5, 22, 4, gray4)
		k.box(7, 9, 18, 21, gray2)
		for x := 11; x < 23; x += 4 {
			k.rect(x, 11, 2, 17, gray5)
			k.rect(x, 11, 1, 17, gray7)
		}
	})
	// The Wastebasket holding something: papers heaped in it lift its lid.
	iconTrashFull = newIcon(func(k canvas) {
		k.box(7, 9, 18, 21, gray2)
		for x := 11; x < 23; x += 4 {
			k.rect(x, 11, 2, 17, gray5)
			k.rect(x, 11, 1, 17, gray7)
		}
		k.box(8, 5, 8, 6, white) // the papers
		k.rect(10, 7, 4, 1, gray5)
		k.box(15, 6, 9, 5, white)
		k.rect(17, 8, 5, 1, gray5)
		k.box(9, 0, 14, 3, gray4) // the lid, raised
		k.box(5, 2, 22, 4, gray4)
	})
	iconCalculator = newIcon(func(k canvas) {
		k.box(5, 2, 22, 28, gray4)
		k.box(8, 5, 16, 6, color.RGBA{0xBB, 0xCC, 0xAA, 255})
		k.rect(17, 7, 5, 2, black)
		for r := 0; r < 4; r++ {
			for col := 0; col < 3; col++ {
				k.box(8+col*6, 13+r*4, 4, 3, gray1)
			}
		}
		k.box(8, 13+4*4, 16, 2, red)
	})
	iconJotter = newIcon(func(k canvas) {
		k.box(5, 4, 22, 26, yellow)
		for x := 8; x < 26; x += 4 {
			k.rect(x, 2, 2, 5, gray7)
			k.rect(x, 2, 1, 5, black)
		}
		for y := 11; y < 28; y += 3 {
			k.rect(7, y, 18, 1, amber)
		}
		k.line(24, 26, 26, 28, black)
	})
	iconScribe = newIcon(func(k canvas) {
		k.rect(4, 6, 18, 24, white)
		k.frame(4, 6, 18, 24, black)
		for y := 10; y < 26; y += 3 {
			k.rect(7, y, 11, 1, gray7)
		}
		// A pen across the page.
		for i := 0; i < 16; i++ {
			k.rect(12+i, 26-i-1, 3, 2, navy)
		}
		k.line(12, 27, 14, 27, black)
		k.rect(27, 8, 3, 2, red)
	})
	iconPuzzle = newIcon(func(k canvas) {
		colors := []color.RGBA{blue, yellow, green, red}
		for i := 0; i < 4; i++ {
			x, y := 3+(i%2)*14, 3+(i/2)*14
			if i == 3 {
				k.frame(x, y, 13, 13, gray7)
				continue
			}
			k.box(x, y, 13, 13, colors[i])
			k.rect(x+1, y+1, 11, 1, white)
			k.rect(x+1, y+1, 1, 11, white)
		}
	})
	iconClock = newIcon(func(k canvas) {
		k.rect(5, 27, 4, 3, black)
		k.rect(23, 27, 4, 3, black)
		k.disc(32, 33, 26, white)
		k.disc(9, 7, 7, red)
		k.disc(55, 7, 7, red)
		k.line(16, 16, 16, 9, black)
		k.line(16, 16, 21, 19, black)
		k.rect(15, 15, 2, 2, black)
	})
	iconAppearance = newIcon(func(k canvas) {
		k.box(3, 4, 26, 20, gray2)
		k.box(6, 7, 20, 14, navy)
		k.rect(7, 8, 18, 4, violet)
		k.rect(7, 12, 9, 8, blue)
		k.rect(16, 12, 9, 8, green)
		k.box(11, 24, 10, 3, gray4)
		k.box(7, 27, 18, 3, gray4)
	})
	iconLedger = newIcon(func(k canvas) {
		k.box(4, 3, 24, 26, white)
		k.rect(5, 4, 22, 4, color.RGBA{0x99, 0xBB, 0x99, 255})
		k.rect(5, 8, 22, 1, black)
		for y := 12; y < 28; y += 4 {
			k.rect(5, y, 22, 1, gray4)
		}
		k.rect(12, 4, 1, 24, gray7)
		k.rect(20, 4, 1, 24, gray7)
		k.rect(22, 21, 4, 2, navy)
		k.rect(14, 13, 4, 2, navy)
	})
	// A window of controls: a check box, a radio button, a slider and a
	// button.
	iconCatalog = newIcon(func(k canvas) {
		k.box(3, 4, 26, 24, gray2)
		k.rect(4, 5, 24, 4, gray4)
		k.rect(3, 9, 26, 1, black)
		k.box(7, 12, 6, 6, white)
		k.rect(9, 14, 2, 2, black)
		k.box(16, 12, 6, 6, white)
		k.rect(18, 14, 2, 2, navy)
		k.rect(7, 21, 18, 1, gray7)
		k.box(15, 19, 3, 5, violet)
		k.box(9, 25, 14, 1, black)
	})
	iconAbout = newIcon(func(k canvas) {
		k.disc(32, 32, 28, violet)
		k.rect(14, 8, 4, 4, white)
		k.rect(12, 14, 6, 3, white)
		k.rect(14, 17, 4, 7, white)
		k.rect(12, 24, 8, 2, white)
	})
)

// iconFor returns the icon of a node.
func iconFor(n *node) *icon {
	switch n.kind {
	case kindDisk:
		return iconDisk
	case kindFolder:
		return iconFolder
	case kindDocument:
		return iconDocument
	case kindPanel:
		return appIcon(n.app)
	}
	return appIcon(n.app)
}
