package platinum

import (
	"image"
	"image/draw"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSnapshotTabStates(t *testing.T) {
	snap(t, "tabs-states", 420, 44, func(c *ui.Context) {
		board(c, 70, 24,
			func(pn Pen) { DrawTab(pn.Sub(0, 2, 70, 22), "Normal", false, StateNormal) },
			func(pn Pen) { DrawTab(pn, "Selected", true, StateNormal) },
			func(pn Pen) { DrawTab(pn.Sub(0, 2, 70, 22), "Pressed", false, StatePressed) },
			func(pn Pen) { DrawTab(pn.Sub(0, 2, 70, 22), "Disabled", false, StateDisabled) },
			func(pn Pen) {
				DrawTab(pn, "Focused", true, StateNormal)
				pn.Rect(5, pn.H-5, pn.W-10, 2, Lavender.Dark)
			},
		)
	})
}

func TestSnapshotTitleBoxes(t *testing.T) {
	snap(t, "titlebox-states", 220, 30, func(c *ui.Context) {
		var cells []func(Pen)
		for _, part := range []WindowPart{PartClose, PartZoom, PartCollapse} {
			for _, pressed := range []bool{false, true} {
				cells = append(cells, func(pn Pen) { DrawTitleBox(pn.Sub(0, 0, 11, 11), part, pressed) })
			}
		}
		cells = append(cells, func(pn Pen) { DrawTitleBox(pn.Sub(0, 0, 9, 9), PartClose, false) })
		cells = append(cells, func(pn Pen) { DrawGrowBox(pn.Sub(0, 0, 15, 15), true) })
		board(c, 15, 15, cells...)
	})
}

// TestSnapshotKeyboardFocus moves the keyboard focus through one control of
// each kind with Tab, and puts the focused control of each step side by
// side, so that every control's focused look is in one golden image.
func TestSnapshotKeyboardFocus(t *testing.T) {
	freeze(t)
	on, v, sel, tab, seg, list := true, 40.0, 0, 0, 0, 1
	open, bold, num := false, false, 3.0
	name := "Name"
	labels := []string{"Button", "Check", "Radio", "Volume", "Sort", "More", "Field", "List", "Tabs", "Bevel", "Seg", "Count"}
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Gap(10).AlignItems(ui.Start).Children(func() {
			Button(c, "Button")
			Checkbox(c, &on, "Check")
			RadioGroup(c, func() { Radio(c, &sel, 0, "Radio") })
			Slider(c, &v, 0, 100).Width(120).Label("Volume")
			PopupMenu(c, &sel, "Sort", []string{"Name", "Date"})
			Disclosure(c, &open, "More")
			TextField(c, &name).Width(100).Label("Field")
			ListBox(c, &list, []string{"One", "Two"}).Element().Size(90, 38).Label("List")
			TabControl(c, &tab, []string{"Tabs", "Two"}, func(int) {}).Element().Size(130, 50).Label("Tabs")
			BevelButton(c, "Bevel").Toggle(&bold)
			Segmented(c, &seg, "Seg", "Two").Element().Label("Seg")
			Stepper(c, &num, 0, 9, 1).Label("Count")
		})
	}
	tt := ui.NewTester(view, 220, 640)
	const cellW, cellH = 150, 64
	montage := image.NewRGBA(image.Rect(0, 0, 4*cellW, 3*cellH))
	draw.Draw(montage, montage.Bounds(), image.NewUniform(rgba(Gray2)), image.Point{}, draw.Src)
	for i, l := range labels {
		tt.Key(0, ui.KeyTab)
		tt.Frame()
		if !tt.Focused(l) {
			t.Errorf("Tab %d did not reach %q", i+1, l)
		}
		r, _ := tt.Find(l)
		src := image.Rect(int(r.X)-5, int(r.Y)-5, int(r.X+r.W)+5, int(r.Y+r.H)+5)
		dst := image.Rect(i%4*cellW, i/4*cellH, i%4*cellW+src.Dx(), i/4*cellH+src.Dy()).
			Intersect(image.Rect(i%4*cellW, i/4*cellH, i%4*cellW+cellW, i/4*cellH+cellH))
		draw.Draw(montage, dst, tt.Image(), src.Min, draw.Src)
	}
	compareGolden(t, "focus-states", montage)
}

func TestSnapshotZoomRects(t *testing.T) {
	snap(t, "zoomrects", 300, 140, func(c *ui.Context) {
		ui.Box(c).Fill().Draw(func(p *ui.Painter, r ui.Rect) {
			pn := Pen{P: p}
			from := ui.Rect{X: 10, Y: 100, W: 32, H: 32}
			DrawZoomRects(pn, from, ui.Rect{X: 60, Y: 10, W: 90, H: 70}, 0.35)
			DrawZoomRects(pn, from, ui.Rect{X: 170, Y: 10, W: 120, H: 100}, 0.7)
		})
	})
}

// TestSnapshotInactiveWindow puts a control of each kind in a window in
// the background, where they draw inactive.
func TestSnapshotInactiveWindow(t *testing.T) {
	on, v, sel, tab, num := true, 40.0, 1, 0, 3.0
	open, bold := false, true
	name := "Text"
	top := 0.0
	list := 1
	front := WindowState{X: 450, Y: 10, W: 120, H: 60}
	back := WindowState{X: 10, Y: 10, W: 430, H: 220}
	ts := NewTableState()
	ts.SortColumn = 0
	snap(t, "inactive-window", 580, 240, func(c *ui.Context) {
		ui.Box(c).Fill().Children(func() {
			Window(c, &back, "In the background", WindowOptions{}, func() {
				ui.Row(c).Padding(8).Gap(10).AlignItems(ui.Start).Children(func() {
					ui.Column(c).Gap(6).AlignItems(ui.Start).Children(func() {
						Button(c, "Button")
						Button(c, "Default").Default()
						Checkbox(c, &on, "Check")
						Radio(c, &sel, 1, "Radio")
						Disclosure(c, &open, "More")
						Slider(c, &v, 0, 100).Width(120).Label("Slider")
						PopupMenu(c, &sel, "", []string{"One", "Two"})
					})
					ui.Column(c).Gap(6).AlignItems(ui.Start).Children(func() {
						TextField(c, &name).Width(110)
						BevelButton(c, "Bevel").Toggle(&bold)
						Stepper(c, &num, 0, 9, 1).Label("Count")
						ListBox(c, &list, []string{"A", "B", "C", "D", "E", "F"}).Element().Size(110, 52)
						ui.Row(c).Gap(4).Children(func() {
							ScrollBar(c, &top, 10, 40).Element().Height(44)
						})
					})
					ui.Column(c).Gap(6).AlignItems(ui.Start).Children(func() {
						GroupBox(c, "Group", func() { Label(c, "Inside") })
						Table(c, ts, []TableColumn{{Title: "Name", Width: 60}, {Title: "Size", Width: 40}}, 2, func(r, col int) string {
							return [][]string{{"Letter", "2 K"}, {"Recipes", "4 K"}}[r][col]
						}).Element().Size(110, 60)
					})
				})
				_ = tab
			})
			Window(c, &front, "In front", WindowOptions{Active: true}, func() {})
		})
	})
}
