// Command desktop is a classic desktop built with the Platinum controls: a
// menu bar, icons, windows that move, resize and roll up, a browser of the
// desk's own little disk, and a few small applications — a calculator, a
// note pad, a text editor, a sliding puzzle, an alarm clock, an appearance
// control panel and an about box.
//
//	go run ./examples/desktop
//	go run ./examples/desktop -png desktop.png   # a picture, without a window
package main

import (
	"flag"
	"image"
	"image/png"
	"log"
	"os"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
)

// The size of the desk, in DIPs.
const width, height = 900, 620

func main() {
	out := flag.String("png", "", "render a scene of the desk to this PNG file and exit")
	capture := flag.String("capture", "", "open the scene in a window, save what the window shows to this PNG file and exit")
	scale := flag.Float64("scale", 1, "the scale of the PNG")
	flag.Parse()
	if *out != "" {
		img := renderScene(float32(*scale))
		f, err := os.Create(*out)
		if err != nil {
			log.Fatal(err)
		}
		if err := png.Encode(f, img); err != nil {
			log.Fatal(err)
		}
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
		return
	}
	d := newDesk(time.Now)
	if *capture != "" {
		d = scene()
	}
	// The desk draws its own menus. Native menu accelerators would close
	// the host window before its documents can ask to save their changes.
	mygo.App.SetMenu(nil)
	mygo.App.WhenReady(func() {
		w := mygo.NewWindow(mygo.WindowOptions{
			Title:   "Platinum Desk",
			Width:   width,
			Height:  height,
			Content: ui.View(d.view),
		})
		if *capture != "" {
			go func() {
				// Let the window show, then take what it drew.
				time.Sleep(1500 * time.Millisecond)
				data, err := w.CapturePage()
				if err == nil {
					err = os.WriteFile(*capture, data, 0o644)
				}
				if err != nil {
					log.Print(err)
				}
				mygo.App.Quit()
			}()
		}
	})
	d.quit = func() { mygo.App.Quit() }
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

// sceneTime is the time the scenes of the desk show, for pictures and
// tests that come out the same every time.
var sceneTime = time.Date(2026, 10, 14, 9, 41, 0, 0, time.UTC)

// scene returns a desk with a few windows open, at sceneTime.
func scene() *desk {
	pt.SetClock(func() time.Time { return sceneTime })
	d := newDesk(func() time.Time { return sceneTime })
	d.size = [2]int{width, height - pt.MenuBarHeight}
	f := d.openFolder(d.disk)
	f.frame = pt.WindowState{X: 24, Y: 20, W: 360, H: 220}
	j := d.launch(appJotter)
	j.frame = pt.WindowState{X: 420, Y: 250, W: 240, H: 200}
	k := d.launch(appCalculator)
	k.frame.X, k.frame.Y = 620, 40
	for _, key := range []string{"1", "2", "×", "3", "+", "6", "="} {
		k.calc.press(key)
	}
	// Two folders docked at the bottom as pop-up windows, one open.
	for _, c := range d.disk.children {
		switch c.name {
		case "Documents":
			w := d.openFolder(c)
			w.frame.W = 250
			d.dock(w)
			w.dock.X, w.dock.H = 610, 104
		case "Applications":
			w := d.openFolder(c)
			w.frame.W = 200
			d.dock(w)
			w.dock.X, w.dock.Open = 390, false
		}
	}
	p := d.launch(appPuzzle)
	p.frame.X, p.frame.Y = 120, 290
	return d
}

func renderScene(scale float32) *image.RGBA {
	d := scene()
	tt := ui.NewTester(d.view, width, height)
	tt.SetScale(scale)
	tt.Frame()
	return tt.Image()
}
