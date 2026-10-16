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
// the days in a grid of six weeks from the first day of the user's week,
// the chosen one highlighted and today outlined. A click chooses a day; with the keyboard focus, the arrows move
// the choice by a day or a week, and Page Up and Down by a month.
func Calendar(c *ui.Context, day *time.Time) *CalendarWidget {
	th := ThemeOf(c)
	w := &CalendarWidget{}
	region, locale := formats(c)
	firstDay := region.FirstWeekday
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
		t := gridStart(month, firstDay).AddDate(0, 0, row*7+col)
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
	today := frameNow(c)
	dim := inactive(c)
	w.el.Draw(func(p *ui.Painter, r ui.Rect) {
		pn := NewPen(p, r).Dimmed(dim)
		DrawCalendar(pn, locale, firstDay, month, chosen, today, focused, th.Accent)
		if focused {
			DrawFocusRing(pn, 0, th.Accent.Dark)
		}
	})
	report(w.el, w.changed)
	return w
}

// gridStart returns the first day of the week, first, on or before the
// first of month.
func gridStart(month time.Time, first time.Weekday) time.Time {
	return month.AddDate(0, 0, -((int(month.Weekday()) - int(first) + 7) % 7))
}

func sameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}

// Changed reports whether the user chose another day since the last frame.
func (w *CalendarWidget) Changed() bool { return report(w.el, w.changed) }

// Element returns the calendar's element.
func (w *CalendarWidget) Element() ui.Element { return w.el }

// Label names the calendar.
func (w *CalendarWidget) Label(s string) *CalendarWidget { w.el.Label(s); return w }

var (
	monthBack = []string{"...#", "..##", ".###", "####", ".###", "..##", "...#"}
	monthNext = []string{"#...", "##..", "###.", "####", "###.", "##..", "#..."}
)

// DrawCalendar draws the calendar of month in pn's box, in the language of
// locale, its weeks starting on first, chosen highlighted and today
// outlined.
func DrawCalendar(pn Pen, locale string, first time.Weekday, month, chosen, today time.Time, focused bool, a Accent) {
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
	names := i18n.Weekdays(locale, i18n.Narrow, true)
	for i := range 7 {
		x := calendarPad + 2 + i*CalendarCellW
		pn.TextCentered(x, calendarHeader+2, CalendarCellW, calendarDays, names[(int(first)+i)%7], AppFont, Gray8)
	}
	pn.HLine(calendarPad+2, calendarHeader+calendarDays+1, 7*CalendarCellW, Gray4)
	start := gridStart(month, first)
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
//
// The counterpart among MyGo's own controls is ui.DateInput.
func DatePicker(c *ui.Context, day *time.Time) *DatePickerWidget {
	th := ThemeOf(c)
	w := &DatePickerWidget{}
	_, locale := formats(c)
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
	report(w.el, w.changed)
	return w
}

// Changed reports whether the user chose another day since the last frame.
func (w *DatePickerWidget) Changed() bool { return report(w.el, w.changed) }

// Element returns the picker's element.
func (w *DatePickerWidget) Element() ui.Element { return w.el }

type timeState struct {
	seg     int    // the part chosen: 0 hours, 1 minutes, 2 AM or PM
	typed   string // digits typed into the part
	changed bool
	picked  time.Time // the time the keys set between builds
	rect    ui.Rect
}

// timeLayout is how the language of locale writes a time in an hour
// cycle of 12 or 24: the order of its parts — 0 the hours, 1 the minutes,
// 2 the half of the day, which a cycle of 24 has not — what is written
// between them, and whether the hours take two digits.
type timeLayout struct {
	order []int
	seps  []string
	pad   bool
}

func timeLayoutOf(locale string, twelve bool) timeLayout {
	cycle := 24
	if twelve {
		cycle = 12
	}
	p := i18n.NewDateFormatter(locale, i18n.DateOptions{Skeleton: "jmm", HourCycle: cycle}).Pattern()
	h := strings.IndexAny(p, "hHkK")
	m := strings.IndexByte(p, 'm')
	l := timeLayout{order: []int{0, 1}, seps: []string{":"}}
	if h < 0 || m < h {
		return l
	}
	end := len(p) - len(strings.TrimLeft(p[h:], "hHkK")) // after the hours' letters
	l.pad = end-h >= 2
	if sep := strings.Trim(p[end:m], "'"); sep != "" {
		l.seps[0] = sep
	}
	if !twelve {
		return l
	}
	if a := strings.IndexByte(p, 'a'); a >= 0 && a < h {
		l.order = []int{2, 0, 1}
		l.seps = []string{strings.Trim(p[a+1:h], "'"), l.seps[0]}
	} else {
		l.order = []int{0, 1, 2}
		l.seps = append(l.seps, " ")
	}
	return l
}

// timeParts returns the text of the parts of t, as "9", "41" and "AM", or
// "09" and "41" in a cycle of 24 hours; the half of the day in the
// language of locale.
func timeParts(t time.Time, locale string, l timeLayout) [3]string {
	twelve := len(l.order) == 3
	h := t.Hour()
	if twelve {
		if h = h % 12; h == 0 {
			h = 12
		}
	}
	hours := fmt.Sprint(h)
	if l.pad {
		hours = fmt.Sprintf("%02d", h)
	}
	if !twelve {
		return [3]string{hours, fmt.Sprintf("%02d", t.Minute()), ""}
	}
	ampm, pm := i18n.DayPeriods(locale, i18n.Abbreviated)
	if t.Hour() >= 12 {
		ampm = pm
	}
	return [3]string{hours, fmt.Sprintf("%02d", t.Minute()), ampm}
}

// TimeField builds a field of the time of day of *t, as "9:41 AM" or
// "21:41", in the hour cycle of the user's regional formats, with little
// arrows at its right. Its parts — hours, minutes, and AM or PM — are
// chosen one at a time, by a click or with Left and Right; Up, Down and the
// arrows step the part chosen (minutes by step), digits type it, and A and
// P set the half of the day.
//
// The counterpart among MyGo's own controls is ui.TimeInput.
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
	region, locale := formats(c)
	twelve := region.HourCycle == 12
	// The parts as the language writes them: the half of the day after the
	// minutes, or before the hours, or none.
	l := timeLayoutOf(locale, twelve)
	n := len(l.order)
	if st.seg >= n {
		st.seg = 0 // the cycle changed under it
	}
	// atHour returns *t at hour h of its day: the time field sets the
	// time, never the date.
	atHour := func(h int) time.Time {
		return time.Date(t.Year(), t.Month(), t.Day(), (h%24+24)%24, t.Minute(), t.Second(), 0, t.Location())
	}
	add := func(d int) {
		switch st.seg {
		case 0:
			set(atHour(t.Hour() + d))
		case 1:
			m := (t.Minute() + d*step + 60*step) % 60
			m = m / step * step
			set(time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), m, t.Second(), 0, t.Location()))
		case 2:
			set(atHour(t.Hour() + 12))
		}
	}
	join := func(parts [3]string) string {
		s := ""
		for i, p := range l.order {
			s += parts[p]
			if i < len(l.seps) {
				s += l.seps[i]
			}
		}
		return s
	}
	parts := timeParts(*t, locale, l)
	shown := make([]string, n)
	for i, p := range l.order {
		shown[i] = parts[p]
	}
	text := join(parts)
	// As wide as the widest time.
	am, pm := i18n.DayPeriods(locale, i18n.Abbreviated)
	width := 0
	for _, period := range []string{am, pm} {
		width = max(width, AppFont.Measure(join([3]string{"23", "59", period})))
	}
	width += 2*textInset + 10
	at := func(seg int) int { return slices.Index(l.order, seg) }
	row.Role(ui.RoleGroup).Label(say(c, "time")).Value(text)
	row.Children(func() {
		field := ui.Box(c).Size(float32(width), TextFieldHeight).Focusable().FocusRing(false).
			Role(ui.RoleTextField).Label(say(c, "time")).Value(text)
		if field.Shortcut(0, ui.KeyLeft) {
			st.seg, st.typed = l.order[max(at(st.seg)-1, 0)], ""
		}
		if field.Shortcut(0, ui.KeyRight) {
			st.seg, st.typed = l.order[min(at(st.seg)+1, n-1)], ""
		}
		if field.Shortcut(0, ui.KeyUp) {
			add(1)
		}
		if field.Shortcut(0, ui.KeyDown) {
			add(-1)
		}
		// Where each part shown starts, for the pointer and the highlight.
		xs := make([]int, n+1)
		xs[0] = textInset + 1
		for i := range n {
			sep := ""
			if i < len(l.seps) {
				sep = l.seps[i]
			}
			xs[i+1] = xs[i] + AppFont.Measure(shown[i]+sep) + 1
		}
		field.HandleInput(func(ev ui.InputEvent) bool {
			switch ev.Kind {
			case ui.InputPointerDown:
				st.seg, st.typed = l.order[0], ""
				for i := 1; i < n; i++ {
					if int(ev.X) >= xs[i]-2 {
						st.seg = l.order[i]
					}
				}
				return false // and take the focus
			case ui.InputText:
				r := []rune(ev.Text)
				if len(r) != 1 {
					return true
				}
				switch {
				case twelve && st.seg == 2 && (r[0] == 'a' || r[0] == 'A') && t.Hour() >= 12,
					twelve && st.seg == 2 && (r[0] == 'p' || r[0] == 'P') && t.Hour() < 12:
					set(atHour(t.Hour() + 12))
				case r[0] >= '0' && r[0] <= '9' && st.seg < 2:
					st.typed += string(r)
					v := 0
					fmt.Sscan(st.typed, &v)
					if st.seg == 0 {
						h := v
						if twelve {
							h = v % 12
							if v > 12 || v == 0 {
								h, st.typed = int(r[0]-'0')%12, string(r)
							}
							if t.Hour() >= 12 {
								h += 12
							}
						} else if v > 23 {
							h, st.typed = int(r[0]-'0'), string(r)
						}
						set(time.Date(t.Year(), t.Month(), t.Day(), h, t.Minute(), t.Second(), 0, t.Location()))
					} else {
						m := v
						if m > 59 {
							m, st.typed = int(r[0]-'0'), string(r)
						}
						set(time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), m, t.Second(), 0, t.Location()))
					}
					if len(st.typed) >= 2 && st.seg == 0 {
						st.seg, st.typed = 1, ""
					} else if len(st.typed) >= 2 {
						st.seg, st.typed = l.order[min(at(st.seg)+1, n-1)], ""
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
			DrawTimeField(pn, shown, l.seps, xs, at(seg), focused, inact, th.Accent)
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
// half of the day if the time has one, in the order the language writes
// them, seps between them — from xs, part seg chosen while focused, in
// pn's box.
func DrawTimeField(pn Pen, parts, seps []string, xs []int, seg int, focused, inactive bool, a Accent) {
	DrawTextFrame(pn, false)
	if inactive {
		pn.Frame(1, 1, pn.W-2, pn.H-2, Gray8)
		pn = pn.Dimmed(true) // its text gray as well
	}
	y := (pn.H-AppFont.CapHeight())/2 - (AppFont.Ascent() - AppFont.CapHeight())
	if focused && seg >= 0 && seg+1 < len(xs) {
		pn.Rect(xs[seg]-1, 4, xs[seg+1]-xs[seg]-1, pn.H-8, a.Light)
	}
	for i, part := range parts {
		x := pn.Text(xs[i], y, part, AppFont, Black)
		if i < len(seps) {
			pn.Text(x, y, seps[i], AppFont, Black)
		}
	}
}
