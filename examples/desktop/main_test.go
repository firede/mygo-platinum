package main

import (
	"os"
	"testing"
	"time"

	pt "github.com/firede/mygo-platinum"
)

// TestMain stops the library's clock for every test, at the moment
// PLATINUM_TEST_NOW (RFC 3339) gives or a fixed one: no test reads the wall
// clock, and tests that show a time set their own.
func TestMain(m *testing.M) {
	now := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
	if s := os.Getenv("PLATINUM_TEST_NOW"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			now = t
		}
	}
	pt.SetClock(func() time.Time { return now })
	pt.MacKeys = true // menus draw the same on every system
	os.Exit(m.Run())
}
