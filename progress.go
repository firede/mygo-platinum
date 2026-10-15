package platinum

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Progress bar metrics, from docs/spec.md.
const (
	ProgressHeight = 14 // the height of a progress bar
	stripeWidth    = 8  // the width of each stripe of an indeterminate bar
	stripeSpeed    = 24 // how many pixels a second the stripes move
)

// ProgressBar builds a progress bar showing value, from 0 to 1, of work
// done: a sunken channel filled from the left with the accent. A negative
// value shows work of unknown length, as diagonal stripes moving along the
// bar.
func ProgressBar(c *ui.Context, value float64) ui.Element {
	th := ThemeOf(c)
	el := ui.Box(c).Height(ProgressHeight).Width(160).Role(ui.RoleProgress)
	if value >= 0 {
		el.Range(0, 1, min(value, 1))
	}
	el.Draw(func(p *ui.Painter, r ui.Rect) {
		pn := NewPen(p, r)
		if value < 0 {
			t := paintNow(p)
			if !frozen() {
				p.After(time.Second / stripeSpeed)
			}
			DrawProgressIndeterminate(pn, int(t.UnixMilli()*stripeSpeed/1000), th.Accent)
			return
		}
		DrawProgress(pn, value, th.Accent)
	})
	return el
}

// Indeterminate is the value ProgressBar takes for work of unknown length.
const Indeterminate = -1

func progressChannel(pn Pen) {
	pn.Frame(0, 0, pn.W, pn.H, Black)
	pn.Rect(1, 1, pn.W-2, pn.H-2, Gray4)
	pn.HLine(1, 1, pn.W-2, Gray7)
	pn.VLine(1, 2, pn.H-3, Gray7)
	pn.HLine(2, pn.H-2, pn.W-3, Gray1)
	pn.VLine(pn.W-2, 2, pn.H-4, Gray1)
}

// DrawProgress draws a progress bar at f (0 to 1) of its length.
func DrawProgress(pn Pen, f float64, a Accent) {
	f = min(max(f, 0), 1)
	progressChannel(pn)
	w := int(f*float64(pn.W-2) + 0.5)
	if w <= 0 {
		return
	}
	h := pn.H - 2
	pn.Rect(1, 1, w, h, a.Mid)
	pn.HLine(1, 1, w, a.Light)
	pn.HLine(1, 2, w, a.Light.Mix(a.Mid, 0.5))
	pn.HLine(1, h-1, w, a.Mid.Mix(a.Dark, 0.5))
	pn.HLine(1, h, w, a.Dark)
	if w < pn.W-2 {
		pn.VLine(1+w, 1, h, a.Dark)
	}
}

// DrawProgressIndeterminate draws a progress bar of unknown length, its
// stripes moved by phase pixels.
func DrawProgressIndeterminate(pn Pen, phase int, a Accent) {
	progressChannel(pn)
	w, h := pn.W-2, pn.H-2
	pn.Rect(1, 1, w, h, a.Light)
	period := 2 * stripeWidth
	pn.Clip(1, 1, w, h, func() {
		for j := 0; j < h; j++ {
			// Each row is the one above moved a pixel left: stripes lean
			// to the right, as they move right.
			off := ((phase-j)%period + period) % period
			for x := off - period; x < w; x += period {
				pn.HLine(1+x, 1+j, stripeWidth, a.Mid)
			}
		}
	})
	pn.HLine(1, 1, w, White.Alpha(0.5))
	pn.HLine(1, h, w, a.Dark.Alpha(0.6))
}

// ProgressDialog shows a movable modal dialog of work in progress while
// *open: message over a progress bar of value (Indeterminate for work of
// unknown length) and detail under it, and a Cancel button that Escape and
// Command-period press too. It reports whether the user cancelled, which
// closes it.
func ProgressDialog(c *ui.Context, open *bool, title, message string, value float64, detail string) bool {
	cancelled := false
	was := *open
	Dialog(c, open, title, 320, func() {
		Label(c, message)
		ProgressBar(c, value).Width(290).Label(message)
		if detail != "" {
			TextIn(c, detail, AppFont, Gray9)
		}
		ui.Row(c).Justify(ui.End).Children(func() {
			if Button(c, say(c, "cancel")).Cancel().Default().Clicked() {
				cancelled = true
			}
		})
	})
	// Escape closes the dialog itself; any way it closes is a cancel.
	if cancelled || was && !*open {
		*open = false
		cancelled = true
	}
	return cancelled
}
