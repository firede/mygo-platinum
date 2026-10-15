package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

var boldIcon = []string{
	"######..",
	".##..##.",
	".##..##.",
	".#####..",
	".##..##.",
	".##..##.",
	"######..",
}

var pageIcon = []string{
	"#####...",
	"#...##..",
	"#...###.",
	"#.....#.",
	"#.###.#.",
	"#.....#.",
	"#.###.#.",
	"#.....#.",
	"#######.",
}

func TestSnapshotBevelAndSegmented(t *testing.T) {
	freeze(t)
	snap(t, "bevel-states", 460, 110, func(c *ui.Context) {
		on, off := true, false
		seg := 1
		ui.Column(c).Padding(8).Gap(8).Children(func() {
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				BevelButton(c, "Normal")
				BevelButton(c, "On").Toggle(&on)
				BevelButton(c, "Off").Toggle(&off)
				BevelButton(c, "Disabled").Disabled(true)
				BevelButton(c, "Page").Icon(pageIcon)
				IconButton(c, boldIcon, "Bold").Toggle(&on)
				BevelButton(c, "Menu").Menu([]MenuItem{{Title: "One"}})
			})
			ui.Row(c).Gap(6).Children(func() {
				ui.Box(c).Size(60, 22).Draw(func(p *ui.Painter, r ui.Rect) {
					pn := NewPen(p, r)
					DrawBevelButton(pn, StatePressed, false)
					pn.TextCentered(1, 1, pn.W, pn.H, "Pressed", SystemFont, Black)
				})
				Segmented(c, &seg, []string{"Left", "Center", "Right"})
			})
			Toolbar(c, []ToolItem{
				{Label: "New", Icon: pageIcon, IconOnly: true},
				{Separator: true},
				{Label: "Bold", Icon: boldIcon, IconOnly: true, On: &on},
				{Label: "Plain"},
			})
		})
	})
}

func TestBevelToggleAndMenu(t *testing.T) {
	freeze(t)
	on := false
	chosen := 0
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Row(c).Padding(10).Gap(8).Children(func() {
			BevelButton(c, "Bold").Toggle(&on)
			BevelButton(c, "Style").Menu([]MenuItem{{Title: "Plain"}, {Title: "Italic", Action: func() { chosen++ }}})
		})
	}
	tt := ui.NewTester(view, 300, 160)
	if err := tt.Click("Bold"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !on {
		t.Fatal("a click did not toggle the bevel button on")
	}
	if err := tt.Click("Style"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	x, y := findCenter(t, tt, "Italic")
	tt.ClickAt(x, y)
	tt.Frame()
	tt.Frame()
	if chosen != 1 {
		t.Fatalf("the menu's item ran %d times", chosen)
	}
}

func TestSegmentedKeys(t *testing.T) {
	sel := 0
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Row(c).Padding(10).Children(func() {
			Segmented(c, &sel, []string{"One", "Two", "Three"})
		})
	}
	tt := ui.NewTester(view, 300, 60)
	if err := tt.Click("Three"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if sel != 2 {
		t.Fatalf("click chose %d", sel)
	}
	tt.Key(0, ui.KeyLeft)
	tt.Frame()
	if sel != 1 {
		t.Fatalf("Left chose %d", sel)
	}
}

// TestBevelMenuHighlight checks the first item of a bevel button's menu:
// none highlighted when a click opens it, for the pointer to choose, and
// the first when the keyboard does, for the arrows to go on from.
func TestBevelMenuHighlight(t *testing.T) {
	freeze(t)
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Row(c).Padding(10).Children(func() {
			BevelButton(c, "Style").Menu([]MenuItem{{Title: "Plain"}, {Title: "Italic"}})
		})
	}
	highlighted := func(tt *ui.Tester) bool {
		r, ok := tt.Find("Plain")
		if !ok {
			t.Fatal("the menu did not open")
		}
		c := tt.Image().RGBAAt(int(r.X)+2, int(r.Y+r.H/2))
		return ui.RGB(c.R, c.G, c.B) == Lavender.Dark
	}
	tt := ui.NewTester(view, 200, 120)
	if err := tt.Click("Style"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if highlighted(tt) {
		t.Error("a click opened the menu with its first item highlighted")
	}
	tt = ui.NewTester(view, 200, 120)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyDown)
	tt.Frame()
	if !highlighted(tt) {
		t.Error("Down opened the menu with no item highlighted")
	}
}
