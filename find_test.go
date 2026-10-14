package platinum

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestFindReplace(t *testing.T) {
	text := "The cat sat. A Cat, a cattle."
	if a, b := Find(text, "cat", 0, FindOptions{}); a != 4 || b != 7 {
		t.Errorf("first cat at %d,%d", a, b)
	}
	if a, _ := Find(text, "cat", 5, FindOptions{MatchCase: true}); a != 22 {
		t.Errorf("next lowercase cat at %d", a)
	}
	if a, _ := Find(text, "cat", 16, FindOptions{WholeWord: true}); a != 4 {
		t.Errorf("whole word cat after 16 wraps to %d", a)
	}
	got, n := ReplaceAll(text, "cat", "dog", FindOptions{WholeWord: true})
	if got != "The dog sat. A dog, a cattle." || n != 2 {
		t.Errorf("replaced %d: %q", n, got)
	}
	freeze(t)
	open := true
	st := &FindState{Query: "cat", With: "dog", Note: "1 found"}
	var acts []FindAction
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Box(c).Fill().Children(func() {
			if a := FindReplace(c, &open, st); a != FindNone {
				acts = append(acts, a)
			}
		})
	}
	tt := ui.NewTester(view, 400, 260)
	tt.Frame()
	compareGolden(t, "find-replace", tt.Image())
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if err := tt.Click("Replace All"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(acts) != 2 || acts[0] != FindNext || acts[1] != ReplaceEvery {
		t.Fatalf("actions %v", acts)
	}
	// The note is read out when it shows, and again after each search,
	// even if it says the same.
	tt.Frame()
	if got := tt.Announcements(); len(got) != 3 || got[0] != "1 found" || got[2] != "1 found" {
		t.Fatalf("announced %q", got)
	}
}
