package platinum

import (
	"bytes"
	"testing"

	"github.com/egoist/mygo/ui"
)

// settled fails t when the frame drawn after the last input is not the
// one the view draws when built again: what the input changed would show
// only in a later frame, which a window draws at its next input alone.
func settled(t *testing.T, tt *ui.Tester, what string) {
	t.Helper()
	before := bytes.Clone(tt.Image().Pix)
	tt.Frame()
	if !bytes.Equal(before, tt.Image().Pix) {
		t.Errorf("%s shows only in the frame after the next", what)
	}
}

// TestFramesShowInput checks that the frame after an input shows what it
// did, where what it did reaches what was built before the control that
// took it: a window brought to the front over the windows built before
// it, a label above a list naming the row chosen, the field that had the
// focus once a pop-up menu takes it.
func TestFramesShowInput(t *testing.T) {
	t.Run("window raised", func(t *testing.T) {
		states := []*WindowState{{X: 20, Y: 20, W: 160, H: 100}, {X: 80, Y: 60, W: 160, H: 100}}
		titles := map[*WindowState]string{states[0]: "Back", states[1]: "Front"}
		view := func(c *ui.Context) {
			Use(c, nil)
			ui.Box(c).Fill().Background(Gray5).Children(func() {
				for i, st := range states {
					w := Window(c.Key(titles[st]), st, titles[st], WindowOptions{Active: i == len(states)-1}, func() {})
					if w.Activated() && i < len(states)-1 {
						states = append(append(states[:i:i], states[i+1:]...), st)
					}
				}
			})
		}
		tt := ui.NewTester(view, 300, 200)
		tt.Press(40, 60) // the back window, where the front one leaves it
		if titles[states[1]] != "Back" {
			t.Fatal("the press did not raise the window")
		}
		settled(t, tt, "the window raised")
		tt.Release(40, 60)
	})
	t.Run("list chosen", func(t *testing.T) {
		sel := 0
		items := []string{"Apples", "Pears", "Plums"}
		chosen := items[0]
		view := func(c *ui.Context) {
			Use(c, nil)
			ui.Column(c).Padding(10).Gap(6).Children(func() {
				Label(c, "Chosen: "+chosen)
				l := ListBox(c, &sel, items)
				l.Element().Size(120, 60)
				if l.Changed() {
					chosen = items[sel]
				}
			})
		}
		tt := ui.NewTester(view, 200, 120)
		x, y := findCenter(t, tt, "Plums")
		tt.Press(x, y)
		settled(t, tt, "the row chosen")
		tt.Release(x, y)
		if chosen != "Plums" {
			t.Fatalf("chose %q", chosen)
		}
	})
	t.Run("focus taken by a menu", func(t *testing.T) {
		freeze(t)
		text, sel := "", 0
		view := func(c *ui.Context) {
			Use(c, nil)
			ui.Column(c).Padding(10).Gap(6).Children(func() {
				TextField(c, &text).Label("Name")
				PopupMenu(c, &sel, "Sort by:", []string{"Name", "Date"})
			})
		}
		tt := ui.NewTester(view, 240, 140)
		tt.Key(0, ui.KeyTab)
		tt.Key(0, ui.KeyTab)
		tt.KeyDown(ui.KeySpace)
		if _, ok := tt.Find("Date"); !ok {
			t.Fatal("Space did not open the menu")
		}
		settled(t, tt, "the menu taking the focus")
	})
}
