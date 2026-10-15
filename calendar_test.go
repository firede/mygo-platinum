package platinum

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

var calDay = time.Date(2026, 10, 14, 9, 41, 0, 0, time.UTC)

func TestCalendarKeysAndClicks(t *testing.T) {
	freeze(t)
	day := calDay
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() { Calendar(c, &day) })
	}
	tt := ui.NewTester(view, 200, 200)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyRight)
	tt.Key(0, ui.KeyDown)
	tt.Frame()
	if day.Day() != 22 {
		t.Fatalf("Right and Down chose %v", day)
	}
	tt.Key(0, ui.KeyPageDown)
	tt.Frame()
	if day.Month() != time.November {
		t.Fatalf("Page Down chose %v", day)
	}
	compareGolden(t, "calendar", tt.Image())
	// The back arrow, then a click on the 1st of October: a Thursday in
	// the first week.
	r, _ := tt.Find("Calendar")
	tt.ClickAt(r.X+8, r.Y+8)
	tt.Frame()
	tt.ClickAt(r.X+calendarPad+2+4*CalendarCellW+8, r.Y+calendarHeader+calendarDays+2+6)
	tt.Frame()
	if day.Month() != time.October || day.Day() != 1 {
		t.Fatalf("clicked %v", day)
	}
}

func TestDatePickerAndCombo(t *testing.T) {
	freeze(t)
	day := calDay
	font := ""
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Gap(10).Children(func() {
			DatePicker(c, &day)
			ComboBox(c, &font, []string{"Sans", "Serif", "Slab", "Mono"}).Label("Font")
		})
	}
	tt := ui.NewTester(view, 260, 260)
	if err := tt.Click("Oct 14, 2026"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if _, ok := tt.Find("Calendar"); !ok {
		t.Fatal("the calendar did not open")
	}
	compareGolden(t, "datepicker-open", tt.Image())
	tt.Key(0, ui.KeyRight)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if day.Day() != 15 {
		t.Fatalf("chose %v", day)
	}
	if _, ok := tt.Find("Calendar"); ok {
		t.Fatal("Return did not close the calendar")
	}
	// The combo box: type, open the matching items, choose.
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyTab)
	tt.Type("s")
	tt.Key(0, ui.KeyDown)
	tt.Frame()
	if _, ok := tt.Find("Mono"); ok {
		t.Fatal("Mono does not start with s")
	}
	compareGolden(t, "combobox-open", tt.Image())
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	tt.Frame()
	if font != "Serif" {
		t.Fatalf("chose %q", font)
	}
}

func TestTimeField(t *testing.T) {
	at := calDay // 9:41 AM
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(10).Children(func() { TimeField(c, &at, 5) })
	}
	tt := ui.NewTester(view, 200, 60)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyUp) // the hours
	tt.Frame()
	if at.Hour() != 10 {
		t.Fatalf("Up made %v", at)
	}
	tt.Key(0, ui.KeyRight) // the minutes
	tt.TypeKey(0, ui.Key0, "0")
	tt.TypeKey(0, ui.Key5, "5")
	tt.Frame()
	if at.Minute() != 5 {
		t.Fatalf("typing 05 made %v", at)
	}
	// Typing two digits went on to AM or PM.
	tt.TypeKey(0, ui.KeyP, "p")
	tt.Frame()
	if at.Hour() != 22 {
		t.Fatalf("p made %v", at)
	}
	compareGolden(t, "timefield", tt.Image())
	r, _ := tt.Find("Change time")
	tt.ClickAt(r.X+6, r.Y+r.H-4) // the down arrow: an hour less... of PM
	tt.Frame()
	if at.Hour() != 10 {
		t.Fatalf("the down arrow made %v", at)
	}
	// The field sets the time of its day: the hours go round it, past
	// midnight either way, and the half of the day turns within it.
	at = calDay
	tt = ui.NewTester(view, 200, 60)
	tt.Key(0, ui.KeyTab) // the hours
	for range 15 {
		tt.Key(0, ui.KeyUp)
	}
	tt.Frame()
	if at.Hour() != 0 || at.Day() != calDay.Day() {
		t.Fatalf("15 hours up from 9 made %v, want 0:41 of %v", at, calDay.Format(time.DateOnly))
	}
	tt.Key(0, ui.KeyDown)
	tt.Frame()
	if at.Hour() != 23 || at.Day() != calDay.Day() {
		t.Fatalf("an hour down from 0 made %v, want 23:41 of the same day", at)
	}
}
