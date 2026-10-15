package main

import (
	"time"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
)

// settings are what the control panels set.
type settings struct {
	offset    time.Duration // the desk's clock from the computer's
	volume    float64
	muted     bool
	alert     int
	tracking  float64
	doubleHit string
}

func newPanelWindow(d *desk, app appKind) *win {
	w := &win{app: app}
	w.opts = pt.WindowOptions{NoGrow: true, NoZoom: true}
	switch app {
	case appDateTime:
		w.frame = pt.WindowState{W: 390}
		w.content = func(c *ui.Context, w *win) { dateTimeView(c, d) }
	case appSound:
		w.frame = pt.WindowState{W: 300}
		w.content = func(c *ui.Context, w *win) { soundView(c, d) }
	case appMouse:
		w.frame = pt.WindowState{W: 300}
		w.content = func(c *ui.Context, w *win) { mouseView(c, d) }
	}
	return w
}

func dateTimeView(c *ui.Context, d *desk) {
	m := d.say()
	now := d.now()
	day := now
	ui.Column(c).FillWidth().Padding(10).Gap(10).AlignItems(ui.Stretch).Children(func() {
		pt.Card(c, func() {
			ui.Row(c).Gap(12).AlignItems(ui.Start).Children(func() {
				pt.Calendar(c, &day)
				ui.Column(c).Gap(8).Children(func() {
					pt.Label(c, m.TimeLabel())
					pt.TimeField(c, &day, 1)
					pt.TextIn(c, i18n.FormatDate(d.formats, now, i18n.DateOptions{Skeleton: "MMMEd"}), pt.AppFont, pt.Black)
				})
			})
		})
		ui.Row(c).Justify(ui.End).Children(func() {
			if pt.Button(c, m.UseComputersClock()).Clicked() {
				d.settings.offset = 0
				day = d.now()
			}
		})
	})
	// The day and time chosen move the desk's clock.
	if !day.Equal(now) {
		d.settings.offset += day.Sub(now)
	}
}

func soundView(c *ui.Context, d *desk) {
	m := d.say()
	s := &d.settings
	ui.Column(c).FillWidth().Padding(10).Gap(10).AlignItems(ui.Stretch).Children(func() {
		pt.Card(c, func() {
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				pt.Label(c, m.VolumeLabel())
				pt.Slider(c, &s.volume, 0, 7).Width(150).Ticks(8).Step(1).Disabled(s.muted).Label(m.Volume())
			})
			pt.Checkbox(c, &s.muted, m.Mute())
		})
		pt.GroupBox(c, m.AlertSound(), func() {
			sounds := []string{m.SoundChime(), m.SoundClick(), m.SoundDroplet(), m.SoundGlass(), m.SoundWhistle()}
			pt.ListBox(c, &s.alert, sounds).Element().Size(250, 66).Label(m.AlertSound())
		})
	})
}

func mouseView(c *ui.Context, d *desk) {
	m := d.say()
	s := &d.settings
	ui.Column(c).FillWidth().Padding(10).Gap(10).AlignItems(ui.Stretch).Children(func() {
		pt.GroupBox(c, m.TrackingSpeed(), func() {
			pt.Slider(c, &s.tracking, 0, 6).Width(240).Ticks(7).Step(1).Label(m.TrackingSpeed())
			ui.Row(c).Gap(150).Children(func() {
				pt.TextIn(c, m.Slow(), pt.AppFont, pt.Black)
				pt.TextIn(c, m.Fast(), pt.AppFont, pt.Black)
			})
		})
		pt.GroupBox(c, m.DoubleClickSpeed(), func() {
			ui.Row(c).Gap(16).Children(func() {
				pt.RadioGroup(c, func() {
					ui.Row(c).Gap(14).Children(func() {
						pt.Radio(c, &s.doubleHit, "slow", m.Slow())
						pt.Radio(c, &s.doubleHit, "medium", m.Medium())
						pt.Radio(c, &s.doubleHit, "fast", m.Fast())
					})
				})
			})
		})
	})
}

var (
	iconDateTime = newIcon(func(k canvas) {
		k.box(4, 6, 24, 22, white)
		k.rect(5, 7, 22, 5, red)
		k.rect(9, 3, 2, 6, black)
		k.rect(21, 3, 2, 6, black)
		for y := 14; y < 26; y += 4 {
			for x := 7; x < 25; x += 4 {
				k.rect(x, y, 2, 2, gray7)
			}
		}
	})
	iconSound = newIcon(func(k canvas) {
		k.box(6, 11, 6, 10, gray4)
		for i := 0; i < 8; i++ {
			k.rect(11+i, 11-i, 1, 10+2*i, gray4)
			k.rect(11+i, 11-i, 1, 1, black)
			k.rect(11+i, 20+i, 1, 1, black)
		}
		k.rect(19, 3, 1, 26, black)
		k.line(23, 10, 25, 8, black)
		k.line(23, 16, 27, 16, black)
		k.line(23, 22, 25, 24, black)
	})
	iconMouse = newIcon(func(k canvas) {
		k.box(9, 8, 14, 21, gray2)
		k.rect(10, 9, 12, 6, gray4)
		k.rect(10, 15, 12, 1, black)
		k.line(16, 8, 16, 3, black)
		k.line(16, 3, 22, 1, black)
	})
)
