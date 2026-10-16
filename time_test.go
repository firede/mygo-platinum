package platinum

import (
	"bytes"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// These tests check what the controls do over time — flashes, blinks,
// repeats and how long a click or typing waits — by a Tester's clock,
// which only the test moves. The package's other tests freeze the time,
// which stills all of it.

// testersClock gives the controls the frames' time, a Tester's clock, in
// place of the package's frozen moment, for the rest of the test.
func testersClock(t *testing.T) {
	prev := SetClock(nil)
	t.Cleanup(func() { SetClock(prev) })
}

// A default button chosen with Return shows pressed for FlashDuration.
func TestButtonFlashes(t *testing.T) {
	testersClock(t)
	clicks := 0
	tt := ui.NewTester(func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() {
			if Button(c, "OK").Default().Clicked() {
				clicks++
			}
		})
	}, 100, 50)
	tt.Frame()
	normal := bytes.Clone(tt.Image().Pix)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if clicks != 1 {
		t.Fatalf("Return clicked %d times", clicks)
	}
	if bytes.Equal(normal, tt.Image().Pix) {
		t.Fatal("the button did not flash")
	}
	tt.Advance(FlashDuration - time.Millisecond)
	if bytes.Equal(normal, tt.Image().Pix) {
		t.Fatal("the flash ended early")
	}
	tt.Advance(time.Millisecond)
	if !bytes.Equal(normal, tt.Image().Pix) {
		t.Fatal("the button stayed pressed after the flash")
	}
}

// The caret shows for caretBlink and hides for as long, and shows at once
// again when the user types.
func TestCaretBlinks(t *testing.T) {
	testersClock(t)
	s := ""
	tt := ui.NewTester(func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() { TextField(c, &s).Width(100) })
	}, 130, 50)
	tt.Key(0, ui.KeyTab)
	tt.Type("a")
	tt.Frame()
	shown := bytes.Clone(tt.Image().Pix)
	tt.Advance(caretBlink)
	hidden := bytes.Clone(tt.Image().Pix)
	if bytes.Equal(shown, hidden) {
		t.Fatal("the caret did not hide")
	}
	tt.Advance(caretBlink)
	if !bytes.Equal(shown, tt.Image().Pix) {
		t.Fatal("the caret did not show again")
	}
	tt.Advance(caretBlink + caretBlink/2) // hidden, halfway
	if !bytes.Equal(hidden, tt.Image().Pix) {
		t.Fatal("the caret did not hide again")
	}
	tt.Key(0, ui.KeyLeft)
	tt.Frame()
	tt.Key(0, ui.KeyRight)
	tt.Frame()
	if !bytes.Equal(shown, tt.Image().Pix) {
		t.Fatal("a key did not show the caret at once")
	}
}

// A chosen menu item blinks for MenuFlash, the menu open, before its
// action runs and the menu closes.
func TestMenuFlashesBeforeItsAction(t *testing.T) {
	testersClock(t)
	a := &menuApp{}
	tt := ui.NewTester(a.view, 300, 200)
	r, _ := tt.Find("File")
	tt.ClickAt(r.X+5, r.Y+5)
	tt.Frame()
	n, _ := tt.Find("New")
	tt.ClickAt(n.X+5, n.Y+5)
	tt.Frame()
	before := bytes.Clone(tt.Image().Pix)
	if a.opened != 0 {
		t.Fatal("the action ran before the item blinked")
	}
	tt.Advance(MenuFlash / 4)
	if bytes.Equal(before, tt.Image().Pix) {
		t.Fatal("the item did not blink")
	}
	if _, ok := tt.Find("Save"); !ok || a.opened != 0 {
		t.Fatal("the menu closed while the item blinked")
	}
	tt.Advance(MenuFlash)
	if a.opened != 1 {
		t.Fatalf("New ran %d times after the blink", a.opened)
	}
	if _, ok := tt.Find("Save"); ok {
		t.Fatal("the menu stayed open after the blink")
	}
}

// A press on a pop-up menu let go within clickHold leaves its menu open to
// choose from; one held longer chooses the item under it, and closes it.
func TestPopupMenuClickHold(t *testing.T) {
	testersClock(t)
	for _, held := range []time.Duration{clickHold - 50*time.Millisecond, clickHold + 50*time.Millisecond} {
		sel := 0
		tt := ui.NewTester(func(c *ui.Context) {
			Use(c, nil)
			ui.Column(c).Padding(10).Children(func() { PopupMenu(c, &sel, "", []string{"One", "Two", "Three"}) })
		}, 200, 120)
		r, _ := tt.Find("One")
		tt.Press(r.X+10, r.Y+r.H/2)
		tt.Frame()
		tt.Advance(held)
		tt.Release(r.X+10, r.Y+r.H/2)
		tt.Advance(MenuFlash)
		_, open := tt.Find("Two")
		if open != (held < clickHold) || sel != 0 {
			t.Fatalf("held %v: the menu open %v, chose %d", held, open, sel)
		}
	}
}

// Letters typed within typeSelectGap of each other choose the item they
// begin; after it, they start anew.
func TestListTypeSelectGap(t *testing.T) {
	testersClock(t)
	sel := -1
	items := []string{"Apple", "Banana", "Blueberry", "Cherry"}
	tt := ui.NewTester(func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() { ListBox(c, &sel, items).Element().Size(120, 80) })
	}, 150, 100)
	tt.Key(0, ui.KeyTab)
	for _, step := range []struct {
		wait  time.Duration
		typed string
		want  int
	}{
		{0, "b", 1},
		{typeSelectGap, "l", 2}, // "bl"
		{typeSelectGap + time.Millisecond, "b", 1}, // anew
		{0, "c", 1}, // "bc", which nothing begins
		{typeSelectGap + time.Millisecond, "c", 3},
	} {
		tt.Advance(step.wait)
		tt.Type(step.typed)
		tt.Frame()
		if sel != step.want {
			t.Fatalf("%q after %v chose %d, want %d", step.typed, step.wait, sel, step.want)
		}
	}
}

// Chasing arrows and an indeterminate progress bar move on with the time.
func TestStatusAnimates(t *testing.T) {
	testersClock(t)
	tt := ui.NewTester(func(c *ui.Context) {
		Use(c, nil)
		ui.Row(c).Padding(10).Gap(10).Children(func() {
			ChasingArrows(c).Label("Arrows")
			ProgressBar(c, -1).Label("Bar")
		})
	}, 220, 40)
	tt.Frame()
	region := func(name string) []byte {
		r, _ := tt.Find(name)
		img := tt.Image()
		var b []byte
		for y := int(r.Y); y < int(r.Y+r.H); y++ {
			for x := int(r.X); x < int(r.X+r.W); x++ {
				c := img.RGBAAt(x, y)
				b = append(b, c.R, c.G, c.B)
			}
		}
		return b
	}
	arrows, bar := region("Arrows"), region("Bar")
	tt.Advance(time.Second / chasingRate)
	if bytes.Equal(arrows, region("Arrows")) {
		t.Error("the chasing arrows stood still")
	}
	tt.Advance(time.Second / stripeSpeed)
	if bytes.Equal(bar, region("Bar")) {
		t.Error("the progress bar's stripes stood still")
	}
}

// A drag resting on a closed pop-up window's tab opens it after
// SpringDelay; one passing over does not.
func TestSpringOpen(t *testing.T) {
	testersClock(t)
	tools := PopupWindowState{X: 150, W: 140, H: 80}
	tt := ui.NewTester(func(c *ui.Context) {
		Use(c, nil)
		ui.Box(c).Fill().Background(Gray5).Children(func() {
			ui.Box(c).Absolute().Left(10).Top(10).Size(60, 30).Label("Source").Drag("note")
			w := PopupWindow(c, &tools, "Tools", WindowOptions{Active: true}, func() {})
			SpringOpen[string](c, w, &tools)
		})
	}, 300, 140)
	tt.Frame()
	x0, y0 := findCenter(t, tt, "Source")
	r, _ := tt.Find("Tools")
	x1, y1 := r.X+popupTabInset+20, r.Y+PopupTabHeight/2
	tt.Press(x0, y0)
	for i := 1; i <= 4; i++ {
		f := float32(i) / 4
		tt.Move(x0+(x1-x0)*f, y0+(y1-y0)*f)
		tt.Frame()
	}
	tt.Advance(SpringDelay - time.Millisecond)
	if tools.Open {
		t.Fatal("the window sprang open early")
	}
	tt.Move(x1+80, y0) // off the tab, and back: the rest starts anew
	tt.Frame()
	tt.Move(x1, y1)
	tt.Frame()
	tt.Advance(SpringDelay - time.Millisecond)
	if tools.Open {
		t.Fatal("the rest went on while the drag was off the tab")
	}
	tt.Advance(time.Millisecond)
	if !tools.Open {
		t.Fatal("the window did not spring open")
	}
	tt.Release(x1, y1)
}

// A stepper's arrow held down repeats after scrollDelay, every
// scrollRepeat.
func TestStepperRepeats(t *testing.T) {
	testersClock(t)
	v := 0.0
	tt := ui.NewTester(func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() { Stepper(c, &v, 0, 100, 1).Label("Count") })
	}, 60, 60)
	r, _ := tt.Find("Count")
	tt.Press(r.X+r.W/2, r.Y+4) // the up arrow
	tt.Frame()
	for _, step := range []struct {
		d    time.Duration
		want float64
	}{{0, 1}, {scrollDelay - time.Millisecond, 1}, {time.Millisecond, 2}, {3 * scrollRepeat, 5}} {
		tt.Advance(step.d)
		if v != step.want {
			t.Fatalf("held %v more: %g, want %g", step.d, v, step.want)
		}
	}
	tt.Release(r.X+r.W/2, r.Y+4)
	tt.Advance(time.Second)
	if v != 5 {
		t.Fatalf("the arrow repeated after it was let go: %g", v)
	}
}
