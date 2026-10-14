package platinum

import (
	"os"
	"testing"
	"time"
)

// testNow is the moment every test of the package sees, unless it sets one
// of its own: no test reads the wall clock. PLATINUM_TEST_NOW (RFC 3339)
// sets another, as scripts/verify.sh does to check that no result depends
// on it.
func testNow() time.Time {
	if s := os.Getenv("PLATINUM_TEST_NOW"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return t
		}
	}
	return time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
}

func TestMain(m *testing.M) {
	now := testNow()
	SetClock(func() time.Time { return now })
	os.Exit(m.Run())
}

// freeze sets the clock to a moment of the test's own, and puts the
// package's back after it.
func freeze(t *testing.T) {
	prev := SetClock(func() time.Time { return time.Unix(1000, 0) })
	t.Cleanup(func() { SetClock(prev) })
}
