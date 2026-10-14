package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// board draws cells side by side, each w×h with a gap, for a snapshot of a
// control's states.
func board(c *ui.Context, w, h int, cells ...func(pn Pen)) {
	ui.Row(c).Padding(6).Gap(6).Children(func() {
		for _, cell := range cells {
			ui.Box(c).Size(float32(w), float32(h)).Draw(func(p *ui.Painter, r ui.Rect) {
				cell(NewPen(p, r))
			})
		}
	})
}

func TestSnapshotButton(t *testing.T) {
	snap(t, "button-states", 400, 40, func(c *ui.Context) {
		board(c, 70, 28,
			func(pn Pen) { DrawButton(pn.Sub(4, 4, 62, 20), "Normal", StateNormal) },
			func(pn Pen) { DrawButton(pn.Sub(4, 4, 62, 20), "Pressed", StatePressed) },
			func(pn Pen) { DrawButton(pn.Sub(4, 4, 62, 20), "Disabled", StateDisabled) },
			func(pn Pen) {
				DrawDefaultRing(pn, 0, 0, 70, 28, StateNormal)
				DrawButton(pn.Sub(4, 4, 62, 20), "Default", StateNormal)
			},
			func(pn Pen) {
				DrawButton(pn.Sub(4, 4, 62, 20), "Focus", StateNormal)
				DrawFocusRing(pn.Sub(4, 4, 62, 20), 3, Lavender.Dark)
			},
		)
	})
}

func TestSnapshotCheckbox(t *testing.T) {
	snap(t, "checkbox-states", 260, 30, func(c *ui.Context) {
		var cells []func(Pen)
		for _, s := range []State{StateNormal, StatePressed, StateDisabled} {
			for _, m := range []CheckMark{CheckNone, CheckOn, CheckMixed} {
				cells = append(cells, func(pn Pen) { DrawCheckbox(pn.Sub(3, 3, 12, 12), m, s) })
			}
		}
		cells = append(cells, func(pn Pen) {
			DrawCheckbox(pn.Sub(3, 3, 12, 12), CheckOn, StateNormal)
			DrawFocusRing(pn.Sub(3, 3, 12, 12), 0, Lavender.Dark)
		})
		board(c, 18, 18, cells...)
	})
}

func TestSnapshotRadio(t *testing.T) {
	snap(t, "radio-states", 200, 30, func(c *ui.Context) {
		var cells []func(Pen)
		for _, s := range []State{StateNormal, StatePressed, StateDisabled} {
			for _, on := range []bool{false, true} {
				cells = append(cells, func(pn Pen) { DrawRadio(pn.Sub(3, 3, 12, 12), on, s) })
			}
		}
		cells = append(cells, func(pn Pen) {
			DrawRadio(pn.Sub(3, 3, 12, 12), true, StateNormal)
			DrawFocusCircle(pn.Sub(3, 3, 12, 12), Lavender.Dark)
		})
		board(c, 18, 18, cells...)
	})
}

func TestSnapshotLabels(t *testing.T) {
	snap(t, "check-radio-widgets", 220, 80, func(c *ui.Context) {
		on, off, choice := true, false, 1
		ui.Column(c).Padding(6).Gap(2).Children(func() {
			Checkbox(c, &on, "Show hidden files")
			Checkbox(c, &off, "Disabled option").Disabled(true)
			RadioGroup(c, func() {
				Radio(c, &choice, 1, "By name")
				Radio(c, &choice, 2, "By date")
			})
		})
	})
}

func TestSnapshotSlider(t *testing.T) {
	snap(t, "slider-states", 520, 40, func(c *ui.Context) {
		board(c, 120, 24,
			func(pn Pen) { DrawSlider(pn.Sub(0, 0, 120, 18), 0.3, 0, StateNormal, Lavender) },
			func(pn Pen) { DrawSlider(pn.Sub(0, 0, 120, 24), 0.5, 5, StateNormal, Lavender) },
			func(pn Pen) { DrawSlider(pn.Sub(0, 0, 120, 18), 0.7, 0, StateDisabled, Lavender) },
			func(pn Pen) { DrawSliderThumb(pn.Sub(10, 0, 11, 16), StatePressed, Lavender) },
		)
	})
}

func TestSnapshotProgress(t *testing.T) {
	snap(t, "progress-states", 520, 30, func(c *ui.Context) {
		board(c, 120, 14,
			func(pn Pen) { DrawProgress(pn, 0, Lavender) },
			func(pn Pen) { DrawProgress(pn, 0.4, Lavender) },
			func(pn Pen) { DrawProgress(pn, 1, Lavender) },
			func(pn Pen) { DrawProgressIndeterminate(pn, 3, Lavender) },
		)
	})
}

func TestSnapshotDisclosure(t *testing.T) {
	snap(t, "disclosure-states", 160, 28, func(c *ui.Context) {
		var cells []func(Pen)
		for _, s := range []State{StateNormal, StatePressed, StateDisabled} {
			for _, open := range []bool{false, true} {
				cells = append(cells, func(pn Pen) { DrawDisclosure(pn, open, s) })
			}
		}
		board(c, 12, 12, cells...)
	})
}

func TestSnapshotGroupAndTabs(t *testing.T) {
	snap(t, "group-tabs", 420, 130, func(c *ui.Context) {
		tab, on := 1, true
		ui.Row(c).Padding(8).Gap(10).AlignItems(ui.Start).Children(func() {
			GroupBox(c, "Options", func() {
				Checkbox(c, &on, "Sounds")
				Separator(c)
				Checkbox(c, &on, "Blink")
			}).Width(140)
			TabControl(c, &tab, []string{"General", "Views", "Labels"}, func(i int) {
				Checkbox(c, &on, "Tab content")
			}).Element().Width(240).Height(110)
		})
	})
}
