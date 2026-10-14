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
