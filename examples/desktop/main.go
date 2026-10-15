// Command desktop is a classic desktop built with the Platinum controls: a
// menu bar, icons, windows that move, resize and roll up, a browser of the
// desk's own little disk, and a few small applications — a calculator, a
// note pad, a text editor, a sliding puzzle, an alarm clock, a catalog of
// the library's controls, control panels and an about box. It speaks English, Simplified and
// Traditional Chinese, Japanese and Korean, as its Language control panel
// chooses.
//
//	go run ./examples/desktop
//	go run ./examples/desktop -png desktop.png               # a picture, without a window
//	go run ./examples/desktop -png desktop.png -lang zh-Hans # the same, in Chinese
//	go run ./examples/desktop -site dir                      # the pictures of its web page
package main

import (
	"flag"
	"image"
	"image/png"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
	"github.com/firede/mygo-platinum/examples/desktop/locales"
)

// The size of the desk, in DIPs.
const width, height = 900, 620

func main() {
	out := flag.String("png", "", "render a scene of the desk to this PNG file and exit")
	capture := flag.String("capture", "", "open the scene in a window, save what the window shows to this PNG file and exit")
	scale := flag.Float64("scale", 1, "the scale of the PNG")
	lang := flag.String("lang", "", "the language of the PNG, as zh-Hans")
	site := flag.String("site", "", "render the pictures of the desk's web page to this directory and exit")
	flag.Parse()
	if *site != "" {
		writeSite(*site)
		return
	}
	if *out != "" {
		savePNG(*out, renderScene(float32(*scale), *lang))
		return
	}
	d := newDesk(time.Now)
	if *capture != "" {
		d = scene("")
	} else if runtime.GOOS == "js" {
		// In a browser the desk is a web page's, seen once: it opens with
		// its disk and the Controls.
		d.greet()
	}
	d.setLocale = func(tag string) { mygo.App.SetLocale(tag) }
	// The desk draws its own menus. Native menu accelerators would close
	// the host window before its documents can ask to save their changes.
	mygo.App.SetMenu(nil)
	mygo.App.WhenReady(func() {
		d.chosenLang = chosenLanguage(mygo.App.Locale(), mygo.App.PreferredLanguages(), mygo.App.Languages())
		w := mygo.NewWindow(mygo.WindowOptions{
			Title:   locales.PlatinumDesk(),
			Width:   width,
			Height:  height,
			Content: ui.View(d.view),
		})
		mygo.App.OnLocaleChanged(func(string) { w.SetTitle(locales.PlatinumDesk()) })
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

// greet opens the disk's window and, in front, the Controls.
func (d *desk) greet() {
	f := d.openFolder(d.disk)
	f.frame.X, f.frame.Y = 24, 20
	k := d.launch(appCatalog)
	k.frame.X, k.frame.Y = 120, 70
}

// sceneTime is the time the scenes of the desk show, for pictures and
// tests that come out the same every time.
var sceneTime = time.Date(2026, 10, 14, 9, 41, 0, 0, time.UTC)

// scene returns a desk with a few windows open, at sceneTime, its
// documents written in the language of lang.
func scene(lang string) *desk {
	pt.SetClock(func() time.Time { return sceneTime })
	d := newDeskIn(func() time.Time { return sceneTime }, lang)
	d.lang = lang
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
		switch c.key {
		case "documents":
			w := d.openFolder(c)
			w.frame.W = 250
			d.dock(w)
			w.dock.X, w.dock.H = 610, 104
		case "applications":
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

// savePNG writes img to the file at path, or ends the program.
func savePNG(path string, img image.Image) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func renderScene(scale float32, lang string) *image.RGBA {
	d := scene(lang)
	tt := ui.NewTester(d.view, width, height)
	tt.SetLocale(lang)
	tt.SetScale(scale)
	tt.Frame()
	return tt.Image()
}
