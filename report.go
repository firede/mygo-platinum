package platinum

import "github.com/egoist/mygo/ui"

// report returns happened, an outcome of the input of el that its caller
// asks about — a click, a change, a window raised or closed — and when it
// happened, has MyGo build the view again in this frame, as the Clicked
// and Changed of its own widgets do: what the caller does about it may
// change what it built before el, as the order of its windows, and the
// frame would show the view as it was until the next input.
func report(el ui.Element, happened bool) bool {
	if happened {
		el.SetChanged().Changed()
	}
	return happened
}

// resized asks for another frame when an element paints at a size other
// than the one it last painted at, old: what its build reads of its last
// paint — the rows a list shows, the length of a scroll bar's track —
// was of the old size, and the frame would show it until the next input.
func resized(c *ui.Context, old, r ui.Rect) {
	if old.W != r.W || old.H != r.H {
		c.Invalidate()
	}
}

// shown asks for another frame after the first a modal backdrop shows in:
// the focus moves into it once the frame is done, and what had the focus
// would show it still.
func shown(c *ui.Context, back ui.Element) {
	seen := ui.Local(back, "shown", func() bool { return false })
	if !*seen {
		*seen = true
		c.Invalidate()
	}
}
