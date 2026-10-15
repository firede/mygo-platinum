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

// TestDeskIconsFollowEdge checks the desktop's icons as the window grows:
// those the user has not moved stay down its right edge, one moved stays
// where it was put, until Clean Up Desktop.
func TestDeskIconsFollowEdge(t *testing.T) {
	d := newDeskIn(func() time.Time { return sceneTime }, "en")
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	disk, trash := d.icons[0], d.icons[1]
	tt.SetSize(width+200, height)
	tt.Frame()
	if disk.x+deskIconW != width+200-12 || trash.x != disk.x {
		t.Fatalf("after growing 200, the icons stand at %d and %d", disk.x, trash.x)
	}
	x, y := center(t, tt, "Wastebasket")
	tt.Press(x, y-20)
	tt.Move(x-60, y)
	tt.Move(x-100, y+40)
	tt.Release(x-100, y+40)
	tt.Frame()
	moved := trash.x
	tt.SetSize(width+300, height)
	tt.Frame()
	if trash.x != moved || disk.x+deskIconW != width+300-12 {
		t.Fatalf("after growing again, the moved icon at %d (was %d), the disk at %d", trash.x, moved, disk.x)
	}
	for _, di := range d.icons {
		di.placed = false // as Clean Up Desktop does
	}
	tt.Frame()
	if trash.x != disk.x {
		t.Fatal("cleaning up did not put the moved icon back on the edge")
	}
}
