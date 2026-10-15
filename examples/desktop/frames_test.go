package main

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// settled fails t when the frame drawn after the last input is not the
// one the desk draws when built again: what the input changed would show
// only at the next input.
func settled(t *testing.T, tt *ui.Tester, what string) {
	t.Helper()
	before := bytes.Clone(tt.Image().Pix)
	tt.Frame()
	if !bytes.Equal(before, tt.Image().Pix) {
		t.Errorf("%s shows only in the frame after the next", what)
	}
}

// TestDeskFramesShowInput checks that the frame after an input shows what
// it did: Tab going round the desktop's icons with windows open, which
// keep the focus where it goes; the window of the disk opened by a
// double-click, its scroll bar as long as its folder; and a window in the
// background pressed to the front.
func TestDeskFramesShowInput(t *testing.T) {
	d := scene("")
	d.launch(appCalculator)
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	for i := range 30 {
		tt.KeyDown(ui.KeyTab)
		settled(t, tt, fmt.Sprintf("Tab %d", i+1))
		tt.KeyUp(ui.KeyTab)
	}

	d = newDeskIn(func() time.Time { return sceneTime }, "en")
	tt = ui.NewTester(d.view, width, height)
	tt.Frame()
	x, y := center(t, tt, "Platinum Disk")
	tt.ClickAt(x, y)
	tt.Press(x, y)
	settled(t, tt, "the disk's window opened")
	tt.Release(x, y)

	k := d.launch(appCalculator)
	k.frame.X, k.frame.Y = 300, 40
	tt.Frame()
	tt.Press(60, 120) // the disk's window, behind
	if d.front() == k {
		t.Fatal("the press did not raise the disk's window")
	}
	settled(t, tt, "the window raised")
	tt.Release(60, 120)
}
