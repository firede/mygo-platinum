package main

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
)

// BenchmarkDesktopFrame builds and paints a frame of the desk with five
// windows open, without a window.
func BenchmarkDesktopFrame(b *testing.B) {
	defer pt.SetClock(pt.SetClock(nil))
	d := scene("")
	pt.SetClock(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	for b.Loop() {
		d.find(appCalculator).calc.press("C") // something changes every frame
		tt.Frame()
	}
}
