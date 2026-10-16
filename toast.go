package platinum

import (
	"github.com/egoist/mygo/ui"
)

// Toasts builds the toasts of the window — what c.Toast, c.ToastAction and
// c.AddToast of MyGo show — in the Platinum look: small alert boxes at the
// bottom of the window, centered, the newest at the bottom, each with the
// small icon of its Type ("error" a stop, "caution" or "warning" a
// caution, else a note), its title and description, its action as a push
// button and a close box. They show and go without fading, as nothing of
// the look fades. Call it once in the view, wherever.
func Toasts(c *ui.Context) ui.Element {
	return ui.ToastViewportBase(c, func(viewport ui.Element, toasts []ui.Toast) {
		viewport.Padding(12).Gap(6).AlignItems(ui.Center)
		for _, t := range toasts {
			toast := ui.ToastBase(c, t)
			text := t.Title
			if t.Description != "" {
				text += " — " + t.Description
			}
			cw, _ := c.Size()
			w := min(AppFont.Measure(text), round(cw)-120)
			toast.Root.Row().Height(BannerHeight).Padding(0, 8, 0, 8).Gap(8).AlignItems(ui.Center)
			toast.Root.Draw(func(p *ui.Painter, r ui.Rect) { DrawAlertFrame(NewPen(p, r)) })
			toast.Root.Children(func() {
				ui.Box(c).Size(16, 16).Draw(func(p *ui.Painter, r ui.Rect) {
					DrawSmallAlertIcon(NewPen(p, r), toastKind(t.Type))
				})
				ui.Box(c).Width(float32(w)).AlignSelf(ui.Stretch).Draw(func(p *ui.Painter, r ui.Rect) {
					pn := NewPen(p, r)
					pn.TextMiddle(0, 0, pn.H, AppFont.Truncate(text, pn.W), AppFont, Black)
				})
				if t.Action != "" {
					b := toast.ActionButton().Label(t.Action)
					pressed := b.Pressed()
					b.Size(float32(max(SystemFont.Measure(t.Action)+24, 58)), ButtonHeight)
					b.Draw(func(p *ui.Painter, r ui.Rect) {
						s := StateNormal
						if pressed {
							s = StatePressed
						}
						DrawButton(NewPen(p, r), t.Action, s)
					})
				}
				box := toast.CloseButton().Size(TitleBoxSize, TitleBoxSize).Label(say(c, "close"))
				pressed := box.Pressed()
				box.Draw(func(p *ui.Painter, r ui.Rect) { DrawTitleBox(NewPen(p, r), PartClose, pressed) })
			})
		}
	})
}

// toastKind returns the kind of alert whose icon a toast of Type typ shows.
func toastKind(typ string) AlertKind {
	switch typ {
	case "error", "stop":
		return AlertStop
	case "caution", "warning":
		return AlertCaution
	}
	return AlertNote
}
