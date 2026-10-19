package main

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// The sample draws; text that does not read keeps the last drawing and says why.
func TestLayout(t *testing.T) {
	a := newApp()
	if a.sc == nil || a.err != "" {
		t.Fatalf("sample: scene %v, error %q", a.sc != nil, a.err)
	}
	ui.Render(a.view, width, height, 1)
	sc := a.sc
	a.src = "pie\n  \"a\": 1"
	a.layout()
	if a.sc != sc || a.err == "" {
		t.Errorf("unsupported: scene kept %v, error %q", a.sc == sc, a.err)
	}
	a.src = sample
	a.layout()
	if a.sc == sc || a.err != "" {
		t.Errorf("again: new scene %v, error %q", a.sc != sc, a.err)
	}
}
