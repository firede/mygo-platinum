package platinum

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
)

// Calendar metrics, from docs/spec.md.
const (
	CalendarCellW  = 22 // the width of a day of a calendar
	CalendarCellH  = 16 // its height
	calendarHeader = 22 // the month and its arrows
	calendarDays   = 14 // the row of weekday initials
	calendarPad    = 4
)

// CalendarWidth and CalendarHeight are the size of a calendar.
const (
	CalendarWidth  = 7*CalendarCellW + 2*calendarPad + 4
	CalendarHeight = calendarHeader + calendarDays + 6*CalendarCellH + calendarPad + 4
)

type calendarState struct {
	month   time.Time // the first of the month shown
	changed bool
	picked  time.Time // the day clicked between builds, for the next one
}

// CalendarWidget is a calendar being built.
type CalendarWidget struct {
	el      ui.Element
	changed bool
}

// Calendar builds a month of days choosing *day: the month and year
// between arrows that turn the month, the initials of the weekdays, and
// the days in a grid of six weeks, the chosen one highlighted and today
// outlined. A click chooses a day; with the keyboard focus, the arrows move
// the choice by a day or a week, and Page Up and Down by a month.
func Calendar(c *ui.Context, day *time.Time) *CalendarWidget {
	th := ThemeOf(c)
	w := &CalendarWidget{}
	locale := c.Locale()
	w.el = ui.Box(c).Size(CalendarWidth, CalendarHeight).Focusable().FocusRing(false).
		Role(ui.RoleGroup).Label(say(c, "calendar")).Value(i18n.FormatDate(locale, *day, i18n.DateOptions{Date: i18n.LongStyle}))
	st := ui.Local(w.el, "calendar", func() calendarState { return calendarState{} })
	if st.changed {
		*day = st.picked
	}
	w.changed, st.changed = st.changed, false
	first := func(t time.Time) time.Time { return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()) }
	if st.month.IsZero() {
		st.month = first(*day)
	}
	set := func(t time.Time) {
		if !t.Equal(*day) {
			*day = t
			w.changed = true
		}
		st.month = first(t)
	}
	for _, k := range []struct {
		key ui.Key
		d   func(time.Time) time.Time
	}{
		{ui.KeyLeft, func(t time.Time) time.Time { return t.AddDate(0, 0, -1) }},
		{ui.KeyRight, func(t time.Time) time.Time { return t.AddDate(0, 0, 1) }},
		{ui.KeyUp, func(t time.Time) time.Time { return t.AddDate(0, 0, -7) }},
		{ui.KeyDown, func(t time.Time) time.Time { return t.AddDate(0, 0, 7) }},
		{ui.KeyPageUp, func(t time.Time) time.Time { return t.AddDate(0, -1, 0) }},
		{ui.KeyPageDown, func(t time.Time) time.Time { return t.AddDate(0, 1, 0) }},
	} {
		if w.el.Shortcut(0, k.key) {
			set(k.d(*day))
		}
	}
	month := st.month
	w.el.HandleInput(func(ev ui.InputEvent) bool {
		if ev.Kind != ui.InputPointerDown || ev.Button != 0 {
			return false
		}
		x, y := int(ev.X), int(ev.Y)
		if y < calendarHeader {
			if x < 22 {
				st.month = st.month.AddDate(0, -1, 0)
			} else if x >= CalendarWidth-22 {
				st.month = st.month.AddDate(0, 1, 0)
			}
			return false
		}
		row := (y - calendarHeader - calendarDays - 2) / CalendarCellH
		col := (x - calendarPad - 2) / CalendarCellW
		if row < 0 || row > 5 || col < 0 || col > 6 || y < calendarHeader+calendarDays+2 {
			return false
		}
		t := gridStart(month).AddDate(0, 0, row*7+col)
		t = time.Date(t.Year(), t.Month(), t.Day(), day.Hour(), day.Minute(), day.Second(), 0, day.Location())
		if !t.Equal(*day) {
			*day = t
			st.picked = t
			st.changed = true
		}
		st.month = first(t)
		return false // and take the focus
	})
	focused := w.el.Focused()
	chosen := *day
	today := wallNow()
	dim := inactive(c)
	w.el.Draw(func(p *ui.Painter, r ui.Rect) {
		pn := NewPen(p, r).Dimmed(dim)
		DrawCalendar(pn, locale, month, chosen, today, focused, th.Accent)
		if focused {
			DrawFocusRing(pn, 0, th.Accent.Dark)
		}
	})
	return w
}

// gridStart returns the Sunday on or before the first of month.
func gridStart(month time.Time) time.Time {
	return month.AddDate(0, 0, -int(month.Weekday()))
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}

// Changed reports whether the user chose another day since the last frame.
func (w *CalendarWidget) Changed() bool { return w.changed }

// Element returns the calendar's element.
func (w *CalendarWidget) Element() ui.Element { return w.el }

// Label names the calendar.
func (w *CalendarWidget) Label(s string) *CalendarWidget { w.el.Label(s); return w }

var (
	monthBack = []string{"...#", "..##", ".###", "####", ".###", "..##", "...#"}
	monthNext = []string{"#...", "##..", "###.", "####", "###.", "##..", "#..."}
)

// DrawCalendar draws the calendar of month in pn's box, in the language of
// locale, chosen highlighted and today outlined.
func DrawCalendar(pn Pen, locale string, month, chosen, today time.Time, focused bool, a Accent) {
	w, h := pn.W, pn.H
	DrawContentRing(pn)
	pn.Frame(1, 1, w-2, h-2, Black)
	pn.Rect(2, 2, w-4, h-4, White)
	// The header: the month between its arrows, on a raised band.
	hd := pn.Sub(2, 2, w-4, calendarHeader-2)
	hd.Rect(0, 0, hd.W, hd.H, Gray2)
	hd.Bevel(0, 0, hd.W, hd.H, White, Gray7)
	pn.HLine(2, calendarHeader, w-4, Black)
	hd.Bitmap(6, (hd.H-7)/2, monthBack, Black)
	hd.Bitmap(hd.W-10, (hd.H-7)/2, monthNext, Black)
	hd.TextCentered(0, 0, hd.W, hd.H, i18n.FormatDate(locale, month, i18n.DateOptions{Skeleton: "yMMMM"}), SystemFont, Black)
	// The weekdays.
	for i, d := range i18n.Weekdays(locale, i18n.Narrow, true) {
		x := calendarPad + 2 + i*CalendarCellW
		pn.TextCentered(x, calendarHeader+2, CalendarCellW, calendarDays, d, AppFont, Gray8)
	}
	pn.HLine(calendarPad+2, calendarHeader+calendarDays+1, 7*CalendarCellW, Gray4)
	start := gridStart(month)
	for i := 0; i < 42; i++ {
		d := start.AddDate(0, 0, i)
		x := calendarPad + 2 + i%7*CalendarCellW
		y := calendarHeader + calendarDays + 2 + i/7*CalendarCellH
		ink := Black
		if d.Month() != month.Month() {
			ink = Gray6
		}
		if sameDay(d, chosen) {
			bg := Gray4
			if focused {
				bg, ink = a.Dark, White
			}
			pn.Rect(x+1, y, CalendarCellW-2, CalendarCellH, bg)
		}
		if sameDay(d, today) {
			pn.Frame(x+1, y, CalendarCellW-2, CalendarCellH, Black)
		}
		pn.TextCentered(x, y, CalendarCellW, CalendarCellH, fmt.Sprint(d.Day()), AppFont, ink)
	}
}

type pickerState struct {
	open bool
}

// DatePickerWidget is a date picker being built.
type DatePickerWidget struct {
	el      ui.Element
	changed bool
}

// DatePicker builds a pop-up button showing *day, which opens a calendar
// under it to choose another; choosing a day, Escape or a click elsewhere
// closes it. Space, Return and Down open it from the keyboard.
func DatePicker(c *ui.Context, day *time.Time) *DatePickerWidget {
	th := ThemeOf(c)
	w := &DatePickerWidget{}
	locale := c.Locale()
	date := func(d time.Time) string { return i18n.FormatDate(locale, d, i18n.DateOptions{Skeleton: "yMMMd"}) }
	text := date(*day)
	widest := 0
	for _, d := range []time.Time{time.Date(2000, 9, 30, 0, 0, 0, 0, time.UTC), time.Date(2000, 12, 30, 0, 0, 0, 0, time.UTC)} {
		widest = max(widest, SystemFont.Measure(date(d)))
	}
	btnW := widest + popupTextLeft + 8 + PopupArrowBox
	w.el = ui.ButtonBase(c).FocusRing(false).Role(ui.RolePopUpButton).Label(text).Size(float32(btnW), PopupHeight)
	st := ui.Local(w.el, "datepicker", func() pickerState { return pickerState{} })
	if w.el.Clicked() || w.el.Shortcut(0, ui.KeyDown) {
		st.open = !st.open
	}
	w.el.Expanded(st.open)
	pressed := w.el.Pressed()
	focused := w.el.FocusVisible()
	w.el.Draw(func(p *ui.Painter, r ui.Rect) {
		pn := NewPen(p, r)
		s := StateNormal
		if pressed {
			s = StatePressed
		}
		DrawPopupButton(pn, text, s)
		if focused {
			DrawFocusRing(pn, 3, th.Accent.Dark)
		}
	})
	if st.open {
		anchor := w.el
		ui.Overlay(c, func() {
			panel := ui.Box(c).AttachTo(anchor, ui.AnchorBottomLeft, ui.AnchorTopLeft).Margin(2, 0, 0, 0)
			panel.Children(func() {
				cal := Calendar(c, day)
				cal.Element().AutoFocus()
				if cal.Changed() {
					w.changed = true
				}
				// A click on a day chooses it and closes the calendar; the
				// keys move the choice until Return.
				if cal.Element().Clicked() || cal.Element().Shortcut(0, ui.KeyEnter) {
					st.open = false
				}
			})
			if panel.PressedOutside() || panel.OverlayShortcut(0, ui.KeyEscape) {
				st.open = false
			}
		})
	}
	return w
}

// Changed reports whether the user chose another day since the last frame.
func (w *DatePickerWidget) Changed() bool { return w.changed }

// Element returns the picker's element.
func (w *DatePickerWidget) Element() ui.Element { return w.el }

type timeState struct {
	seg     int    // the part chosen: 0 hours, 1 minutes, 2 AM or PM
	typed   string // digits typed into the part
	changed bool
	picked  time.Time // the time the keys set between builds
	rect    ui.Rect
}

// timeParts returns the text of the parts of t, as "9", "41" and "AM", the
// half of the day in the language of locale.
func timeParts(t time.Time, locale string) [3]string {
	h := t.Hour() % 12
	if h == 0 {
		h = 12
	}
	ampm, pm := i18n.DayPeriods(locale, i18n.Abbreviated)
	if t.Hour() >= 12 {
		ampm = pm
	}
	return [3]string{fmt.Sprint(h), fmt.Sprintf("%02d", t.Minute()), ampm}
}

// TimeField builds a field of the time of day of *t, as "9:41 AM", with
// little arrows at its right. Its parts — hours, minutes, AM or PM — are
// chosen one at a time, by a click or with Left and Right; Up, Down and the
// arrows step the part chosen (minutes by step), digits type it, and A and
// P set the half of the day.
func TimeField(c *ui.Context, t *time.Time, step int) ui.Element {
	th := ThemeOf(c)
	step = max(step, 1)
	row := ui.Row(c).Gap(3).AlignItems(ui.Center)
	st := ui.Local(row, "time", func() timeState { return timeState{} })
	if st.changed {
		*t = st.picked
		st.changed = false
	}
	set := func(nt time.Time) {
		*t = nt
		st.picked, st.changed = nt, true
	}
	add := func(d int) {
		switch st.seg {
		case 0:
			set(t.Add(time.Duration(d) * time.Hour))
		case 1:
			m := (t.Minute() + d*step + 60*step) % 60
			m = m / step * step
			set(time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), m, t.Second(), 0, t.Location()))
		case 2:
			set(t.Add(12 * time.Hour))
		}
	}
	locale := c.Locale()
	parts := timeParts(*t, locale)
	// The parts as the language writes them: the half of the day after the
	// minutes, or before the hours.
	order, seps := timeOrder(locale)
	var shown [3]string
	for i, p := range order {
		shown[i] = parts[p]
	}
	text := shown[0] + seps[0] + shown[1] + seps[1] + shown[2]
	// As wide as the widest time.
	width := 0
	am, pm := i18n.DayPeriods(locale, i18n.Abbreviated)
	for _, period := range []string{am, pm} {
		widest := [3]string{"12", "59", period}
		width = max(width, AppFont.Measure(widest[order[0]]+seps[0]+widest[order[1]]+seps[1]+widest[order[2]]))
	}
	width += 2*textInset + 10
	at := func(seg int) int { return slices.Index(order[:], seg) }
	row.Role(ui.RoleGroup).Label(say(c, "time")).Value(text)
	row.Children(func() {
		field := ui.Box(c).Size(float32(width), TextFieldHeight).Focusable().FocusRing(false).
			Role(ui.RoleTextField).Label(say(c, "time")).Value(text)
		if field.Shortcut(0, ui.KeyLeft) {
			st.seg, st.typed = order[max(at(st.seg)-1, 0)], ""
		}
		if field.Shortcut(0, ui.KeyRight) {
			st.seg, st.typed = order[min(at(st.seg)+1, 2)], ""
		}
		if field.Shortcut(0, ui.KeyUp) {
			add(1)
		}
		if field.Shortcut(0, ui.KeyDown) {
			add(-1)
		}
		// Where each part shown starts, for the pointer and the highlight.
		xs := [4]int{textInset + 1}
		for i := range 3 {
			sep := ""
			if i < 2 {
				sep = seps[i]
			}
			xs[i+1] = xs[i] + AppFont.Measure(shown[i]+sep) + 1
		}
		field.HandleInput(func(ev ui.InputEvent) bool {
			switch ev.Kind {
			case ui.InputPointerDown:
				st.seg, st.typed = order[0], ""
				for i := 1; i < 3; i++ {
					if int(ev.X) >= xs[i]-2 {
						st.seg = order[i]
					}
				}
				return false // and take the focus
			case ui.InputText:
				r := []rune(ev.Text)
				if len(r) != 1 {
					return true
				}
				switch {
				case st.seg == 2 && (r[0] == 'a' || r[0] == 'A') && t.Hour() >= 12,
					st.seg == 2 && (r[0] == 'p' || r[0] == 'P') && t.Hour() < 12:
					set(t.Add(12 * time.Hour))
				case r[0] >= '0' && r[0] <= '9' && st.seg < 2:
					st.typed += string(r)
					n := 0
					fmt.Sscan(st.typed, &n)
					if st.seg == 0 {
						h := n % 12
						if n > 12 || n == 0 {
							h, st.typed = int(r[0]-'0')%12, string(r)
						}
						if t.Hour() >= 12 {
							h += 12
						}
						set(time.Date(t.Year(), t.Month(), t.Day(), h, t.Minute(), t.Second(), 0, t.Location()))
					} else {
						m := n
						if m > 59 {
							m, st.typed = int(r[0]-'0'), string(r)
						}
						set(time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), m, t.Second(), 0, t.Location()))
					}
					if len(st.typed) >= 2 && st.seg == 0 {
						st.seg, st.typed = 1, ""
					} else if len(st.typed) >= 2 {
						st.seg, st.typed = order[min(at(st.seg)+1, 2)], ""
					}
				}
				return true
			}
			return false
		})
		field.TextCaret(ui.Rect{})
		focused := field.Focused()
		seg := st.seg
		inact := inactive(c)
		field.Draw(func(p *ui.Painter, r ui.Rect) {
			pn := NewPen(p, r)
			DrawTimeField(pn, shown, seps, xs, at(seg), focused, inact, th.Accent)
			if focused {
				DrawFocusRing(pn, 0, th.Accent.Dark)
			}
		})
		var d float64
		if Stepper(c, &d, -1, 1, 1).Label(say(c, "changeTime")).Changed() {
			add(int(d))
		}
	})
	return row
}

// DrawTimeField draws a time field showing parts — the hours, minutes and
// half of the day, in the order the language writes them, seps after the
// first two — from xs, part seg chosen while focused, in pn's box.
func DrawTimeField(pn Pen, parts [3]string, seps [2]string, xs [4]int, seg int, focused, inactive bool, a Accent) {
	DrawTextFrame(pn, false)
	if inactive {
		pn.Frame(1, 1, pn.W-2, pn.H-2, Gray8)
		pn = pn.Dimmed(true) // its text gray as well
	}
	y := (pn.H-AppFont.CapHeight())/2 - (AppFont.Ascent() - AppFont.CapHeight())
	if focused {
		pn.Rect(xs[seg]-1, 4, xs[seg+1]-xs[seg]-1, pn.H-8, a.Light)
	}
	for i, part := range parts {
		x := pn.Text(xs[i], y, part, AppFont, Black)
		if i < 2 {
			pn.Text(x, y, seps[i], AppFont, Black)
		}
	}
}

// timeOrder returns the order the parts of a time — 0 the hours, 1 the
// minutes, 2 the half of the day — are written in the language of locale,
// and what is written after the first two.
func timeOrder(locale string) (order [3]int, seps [2]string) {
	p := i18n.NewDateFormatter(locale, i18n.DateOptions{Skeleton: "hm"}).Pattern()
	a, h := strings.IndexByte(p, 'a'), strings.IndexAny(p, "hK")
	if a >= 0 && h > a {
		return [3]int{2, 0, 1}, [2]string{strings.Trim(p[a+1:h], "'"), ":"}
	}
	return [3]int{0, 1, 2}, [2]string{":", " "}
}
