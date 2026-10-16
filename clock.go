package platinum

import (
	"sync"
	"time"

	"github.com/egoist/mygo/ui"
)

var (
	clockMu sync.Mutex
	clock   func() time.Time
)

// SetClock makes every control take its time from now instead of the
// frame's — indeterminate progress bars and chasing arrows, caret blinks,
// menu and button flashes, calendars' today — for tests and screenshots
// that must come out the same whenever and wherever they run. Animations
// stand still at that moment. A nil now restores the frame's time. It
// returns the clock it replaces, to put back:
//
//	defer platinum.SetClock(platinum.SetClock(fixed))
func SetClock(now func() time.Time) (previous func() time.Time) {
	clockMu.Lock()
	previous, clock = clock, now
	clockMu.Unlock()
	return previous
}

// frameNow is the time of the frame being built, or the clock's.
func frameNow(c *ui.Context) time.Time {
	clockMu.Lock()
	f := clock
	clockMu.Unlock()
	if f != nil {
		return f()
	}
	return c.Now()
}

func paintNow(p *ui.Painter) time.Time {
	clockMu.Lock()
	f := clock
	clockMu.Unlock()
	if f != nil {
		return f()
	}
	return p.Now()
}

func frozen() bool {
	clockMu.Lock()
	defer clockMu.Unlock()
	return clock != nil
}

// eventNow is when an input came, by the frames' clock (a Tester's in
// tests), or the clock's.
func eventNow(ev ui.InputEvent) time.Time {
	if frozen() || ev.Time.IsZero() {
		return wallNow()
	}
	return ev.Time
}

// wallNow is the time for what runs between frames without an input, as
// the commands of the Edit menu.
func wallNow() time.Time {
	clockMu.Lock()
	f := clock
	clockMu.Unlock()
	if f != nil {
		return f()
	}
	return time.Now()
}
