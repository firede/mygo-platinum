package platinum

import (
	"github.com/egoist/mygo/ui"
)

// Banner builds a notification band where it is laid out while *open —
// across the top of a window's content, as a rule, the content under it
// moved down: a raised panel as wide as its parent, with a small icon of
// kind, text and a close box that closes it, its text read out once by
// screen readers as it shows. It returns the band, or a nil Element while
// closed.
func Banner(c *ui.Context, open *bool, kind AlertKind, text string) ui.Element {
	if !*open {
		return ui.Element{}
	}
	panel := ui.Row(c).Height(BannerHeight).AlignSelf(ui.Stretch).Shrink(0).
		Role(ui.RoleStatus).Label(text).Padding(0, 8, 0, 8).Gap(8).AlignItems(ui.Center)
	said := ui.Local(panel, "banner", func() string { return "" })
	if *said != text {
		c.Announce(text) // screen readers read it out once
		*said = text
	}
	panel.Draw(func(p *ui.Painter, r ui.Rect) { DrawAlertFrame(NewPen(p, r)) })
	panel.Children(func() {
		ui.Box(c).Size(16, 16).Role(ui.RoleImage).Draw(func(p *ui.Painter, r ui.Rect) {
			DrawSmallAlertIcon(NewPen(p, r), kind)
		})
		ui.Box(c).Grow(1).AlignSelf(ui.Stretch).Draw(func(p *ui.Painter, r ui.Rect) {
			pn := NewPen(p, r)
			pn.TextMiddle(0, 0, pn.H, AppFont.Truncate(text, pn.W), AppFont, Black)
		})
		box := ui.ButtonBase(c).Size(TitleBoxSize, TitleBoxSize).FocusRing(false).Label(say(c, "close"))
		pressed := box.Pressed()
		box.Draw(func(p *ui.Painter, r ui.Rect) { DrawTitleBox(NewPen(p, r), PartClose, pressed) })
		if box.Clicked() {
			*open = false
		}
	})
	return panel
}

// BannerHeight is the height of a banner.
const BannerHeight = 30

// DrawSmallAlertIcon draws the 16×16 icon of an alert of kind, as banners
// show it, in pn's corner.
func DrawSmallAlertIcon(pn Pen, kind AlertKind) {
	switch kind {
	case AlertCaution:
		for y := 1; y < 15; y++ {
			half := (y - 1) / 2
			pn.HLine(8-half-1, y, 2*half+2, ui.RGB(0xFF, 0xCC, 0x00))
			pn.Dot(8-half-1, y, Black)
			pn.Dot(8+half, y, Black)
		}
		pn.HLine(1, 15, 14, Black)
		pn.Rect(7, 5, 2, 6, Black)
		pn.Rect(7, 12, 2, 2, Black)
	case AlertStop:
		pn.FillRound(0, 0, 16, 16, 4, Black)
		pn.FillRound(1, 1, 14, 14, 3, ui.RGB(0xCC, 0x22, 0x22))
		pn.Rect(7, 3, 2, 7, White)
		pn.Rect(7, 11, 2, 2, White)
	default:
		pn.FillRound(0, 0, 16, 13, 3, Black)
		pn.FillRound(1, 1, 14, 11, 2, White)
		pn.Rect(3, 13, 2, 2, Black)
		pn.Rect(7, 3, 2, 2, Black)
		pn.Rect(7, 6, 2, 4, Black)
	}
}

// EmptyState builds what a list or a window shows while it has nothing to
// show: one line of gray text, centered in the room it is given.
func EmptyState(c *ui.Context, text string) ui.Element {
	col := ui.Column(c).Grow(1).AlignItems(ui.Center).Justify(ui.Center).Role(ui.RoleGroup).Label(text)
	col.Children(func() { TextIn(c, text, AppFont, Gray8) })
	return col
}
