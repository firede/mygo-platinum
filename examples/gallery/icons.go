package main

import (
	"image"
	"image/color"

	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
)

// galleryIcons draws a few 32×32 icons for the icon view, as simple shapes.
func galleryIcons() []pt.IconItem {
	shape := func(face color.RGBA, inset func(x, y int) bool) *ui.Bitmap {
		img := image.NewRGBA(image.Rect(0, 0, 32, 32))
		for y := 0; y < 32; y++ {
			for x := 0; x < 32; x++ {
				if !inset(x, y) {
					continue
				}
				c := face
				if !inset(x-1, y) || !inset(x+1, y) || !inset(x, y-1) || !inset(x, y+1) {
					c = color.RGBA{0, 0, 0, 255}
				}
				img.SetRGBA(x, y, c)
			}
		}
		return ui.NewBitmap(img).Pixelated()
	}
	square := func(x, y int) bool { return x >= 5 && x < 27 && y >= 5 && y < 27 }
	circle := func(x, y int) bool { dx, dy := x-16, y-16; return dx*dx+dy*dy < 12*12 }
	diamond := func(x, y int) bool { return abs(x-16)+abs(y-16) < 13 }
	return []pt.IconItem{
		{Name: "Square", Icon: shape(color.RGBA{0x99, 0x88, 0xDD, 255}, square)},
		{Name: "Circle", Icon: shape(color.RGBA{0xFF, 0xCC, 0x44, 255}, circle)},
		{Name: "Diamond", Icon: shape(color.RGBA{0x66, 0xAA, 0x77, 255}, diamond)},
		{Name: "Another Square", Icon: shape(color.RGBA{0xDD, 0x88, 0x88, 255}, square)},
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
