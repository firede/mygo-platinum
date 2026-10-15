package platinum

import (
	"github.com/egoist/mygo/ui"
)

// Alert metrics, from docs/spec.md.
const (
	AlertIconSize = 32  // the icon of an alert
	AlertWidth    = 340 // the width of an alert box
	alertPadding  = 13
	alertGap      = 12
)

// AlertKind is the kind of news an alert brings, which its icon shows.
type AlertKind int

const (
	// AlertNote tells the user something.
	AlertNote AlertKind = iota
	// AlertCaution warns of what an action will do.
	AlertCaution
	// AlertStop says something went wrong.
	AlertStop
)

// Alert shows an alert box over the window while *open: an icon of kind,
// message in the bold face, detail under it in the regular face, and
// buttons along its bottom, the last one the default. It keeps the
// keyboard while it shows: Return and Enter press the default button, and
// Escape and Command-period a button named "Cancel". It returns the index
// of the button the user pressed, and closes, or -1.
//
// The counterpart among MyGo's own controls is ui.AlertDialog.
func Alert(c *ui.Context, open *bool, kind AlertKind, message, detail string, buttons ...string) int {
	chosen := -1
	if !*open {
		return chosen
	}
	ui.Overlay(c, func() {
		activeWithin(c, func() {
			back := ui.Column(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).Center().Modal()
			shown(c, back)
			back.Children(func() {
				box := ui.Column(c).Width(AlertWidth).Padding(alertPadding + 2).Gap(alertGap).
					Role(ui.RoleAlertDialog).Label(message)
				box.Draw(func(p *ui.Painter, r ui.Rect) { DrawAlertFrame(NewPen(p, r)) })
				box.Children(func() {
					ui.Row(c).Gap(alertGap + 4).AlignItems(ui.Start).Children(func() {
						ui.Box(c).Size(AlertIconSize, AlertIconSize).Role(ui.RoleImage).Draw(func(p *ui.Painter, r ui.Rect) {
							DrawAlertIcon(NewPen(p, r), kind)
						})
						ui.Column(c).Grow(1).Gap(6).Children(func() {
							Paragraph(c, message, SystemFont, AlertWidth-2*alertPadding-AlertIconSize-alertGap-8)
							if detail != "" {
								Paragraph(c, detail, AppFont, AlertWidth-2*alertPadding-AlertIconSize-alertGap-8)
							}
						})
					})
					ui.Row(c).Gap(10).Justify(ui.End).AlignItems(ui.Center).Children(func() {
						for i, label := range buttons {
							b := Button(c, label)
							if i == len(buttons)-1 {
								b.Default()
								b.Element().AutoFocus()
							}
							if label == "Cancel" || label == say(c, "cancel") {
								b.Cancel()
							}
							if b.Clicked() {
								chosen = i
							}
						}
					})
				})
			})
		})
	})
	if chosen >= 0 {
		*open = false
	}
	return chosen
}

// DrawAlertFrame draws the frame of an alert box filling pn's box: an
// outline and a thick, beveled border around a gray face.
func DrawAlertFrame(pn Pen) {
	w, h := pn.W, pn.H
	pn.Rect(2, h, w-1, 1, Gray9.Alpha(0.6))
	pn.Rect(w, 2, 1, h-1, Gray9.Alpha(0.6))
	pn.Rect(0, 0, w, h, Gray2)
	pn.Frame(0, 0, w, h, Black)
	pn.Bevel(1, 1, w-2, h-2, White, Gray7)
	pn.Bevel(3, 3, w-6, h-6, Gray7, White)
	pn.Bevel(4, 4, w-8, h-8, White, Gray5)
}

// DrawAlertIcon draws the 32×32 icon of an alert of kind in pn's corner.
func DrawAlertIcon(pn Pen, kind AlertKind) {
	switch kind {
	case AlertCaution:
		yellow := ui.RGB(0xFF, 0xCC, 0x00)
		// A triangle, its point at the top.
		for y := 2; y < 30; y++ {
			half := (y - 2) / 2
			x0 := 16 - half - 1
			pn.HLine(x0, y, 2*half+2, yellow)
			pn.Dot(x0, y, Black)
			pn.Dot(x0+2*half+1, y, Black)
			if y%2 == 1 && y > 2 {
				pn.Dot(x0+1, y, Black)
				pn.Dot(x0+2*half, y, Black)
			}
		}
		pn.HLine(2, 30, 28, Black)
		pn.HLine(3, 31, 28, Gray7)
		pn.Rect(15, 10, 2, 12, Black)
		pn.Rect(14, 11, 4, 8, Black)
		pn.Rect(15, 25, 2, 2, Black)
	case AlertStop:
		red := ui.RGB(0xCC, 0x22, 0x22)
		// An octagon, ringed in white.
		for y := 0; y < 32; y++ {
			cut := 0
			if y < 9 {
				cut = 9 - y
			} else if y > 22 {
				cut = y - 22
			}
			pn.HLine(cut, y, 32-2*cut, Black)
			if y >= 1 && y <= 30 {
				c2 := max(cut, 1)
				pn.HLine(c2+1, y, 30-2*c2, White)
			}
			if y >= 3 && y <= 28 {
				c3 := cut
				if y < 9 {
					c3 = max(9-y+1, 3)
				} else if y > 22 {
					c3 = max(y-22+1, 3)
				} else {
					c3 = 3
				}
				pn.HLine(c3, y, 32-2*c3, red)
			}
		}
		// A hand-drawn exclamation, in white.
		pn.Rect(14, 7, 4, 12, White)
		pn.Rect(14, 22, 4, 3, White)
	default:
		// A speech balloon with a dot of thought.
		pn.FillRound(1, 1, 30, 22, 4, Black)
		pn.FillRound(2, 2, 28, 20, 3, White)
		for i := 0; i < 6; i++ {
			pn.HLine(7+i, 22+i, 6-i, Black)
			pn.HLine(8+i, 22+i, 4-i, White)
		}
		pn.Rect(14, 5, 4, 3, Black)
		pn.Rect(13, 10, 5, 2, Black)
		pn.Rect(14, 12, 4, 6, Black)
		pn.Rect(12, 17, 8, 2, Black)
	}
}

type dialogState struct {
	frame  WindowState
	placed bool
}

// Dialog shows a movable modal dialog over the window while *open: a window
// frame with a title bar, which the user drags, but no boxes, around the
// controls fn builds, in a column. It keeps the keyboard while it shows;
// Escape closes it, unless fn's buttons take it. It first shows centered,
// as wide as width.
//
// The counterpart among MyGo's own controls is ui.Modal.
func Dialog(c *ui.Context, open *bool, title string, width int, fn func()) {
	if !*open {
		return
	}
	ui.Overlay(c, func() {
		activeWithin(c, func() {
			back := ui.Box(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).Modal()
			shown(c, back)
			st := ui.Local(back, "dialog", func() dialogState { return dialogState{} })
			if back.OverlayShortcut(0, ui.KeyEscape) {
				*open = false
			}
			back.Children(func() {
				Window(c, &st.frame, title, WindowOptions{Modal: true, Active: true}, func() {
					col := ui.Column(c).Padding(12).Gap(8)
					col.Children(fn)
				})
				if !st.placed {
					// Center it across the window, a fifth of the way down; it is
					// as tall as its content.
					cw, ch := c.Size()
					st.frame.W = width
					st.frame.X = (round(cw) - st.frame.W) / 2
					st.frame.Y = round(ch) / 5
					st.placed = true
				}
			})
		})
	})
}
