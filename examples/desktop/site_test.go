package main

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// TestSitePictures draws the pictures of the desk's web page, each against
// its golden image; the pictures of the desk in its languages are its
// scenes at twice the scale, which TestDesktopScenes checks.
func TestSitePictures(t *testing.T) {
	pics := sitePictures()
	for _, lang := range languages {
		if pics["scene-"+lang+".png"] == nil {
			t.Errorf("no picture of the desk in %s", lang)
		}
	}
	for name, img := range pics {
		if strings.HasPrefix(name, "scene-") {
			continue
		}
		golden(t, "site-"+strings.TrimSuffix(name, ".png"), img)
	}
	for _, pic := range sharePictures {
		if b := pics[pic.name+".png"].Bounds(); b.Dx() != int(float32(pic.w)*pic.scale) || b.Dy() != int(float32(pic.h)*pic.scale) {
			t.Errorf("%s is %v, want %d×%d at %v", pic.name, b.Size(), pic.w, pic.h, pic.scale)
		}
	}
}

// TestGreet checks the desk as a web page opens it: its disk's window
// and, in front, the Read Me.
func TestGreet(t *testing.T) {
	d := newDeskIn(func() time.Time { return sceneTime }, "en")
	d.greet()
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	var titles []string
	for _, w := range d.windows {
		titles = append(titles, d.titleOf(w))
	}
	if !slices.Equal(titles, []string{"Platinum Disk", "Read Me"}) {
		t.Errorf("windows %q, want the disk's and the Read Me", titles)
	}
}

// TestAboutSourceLink clicks the address in the About box, which opens
// the desk's source.
func TestAboutSourceLink(t *testing.T) {
	d := newDeskIn(func() time.Time { return sceneTime }, "en")
	d.launch(appAbout)
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	if err := tt.Click("github.com/firede/mygo-platinum"); err != nil {
		t.Fatal(err)
	}
	if got := tt.OpenedURLs(); !slices.Equal(got, []string{sourceURL}) {
		t.Errorf("opened %q, want %q", got, sourceURL)
	}
}
