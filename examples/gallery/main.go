// Command gallery shows every control of the Platinum library in a window,
// or, with -png, renders them to a PNG without one:
//
//	go run ./examples/gallery
//	go run ./examples/gallery -png examples/gallery/testdata/golden/page-1.png -scale 1
//	go run ./examples/gallery -png docs/gallery@2x.png -scale 2
package main

import (
	"flag"
	"image/png"
	"log"
	"os"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
)

// The size of the gallery, in DIPs.
const width, height = 756, 476

func main() {
	out := flag.String("png", "", "render the gallery to this PNG file and exit")
	scale := flag.Float64("scale", 1, "the scale of the PNG")
	page := flag.Int("page", 1, "the page of the gallery the PNG shows, 1 to 5")
	capture := flag.String("capture", "", "show the page in a window, save what the window drew to this PNG file and exit")
	flag.Parse()
	g := newGallery()
	g.page = *page - 1
	if *out != "" {
		pt.SetClock(func() time.Time { return time.Unix(0, 0) })
		img := ui.Render(g.view, width, height, float32(*scale))
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
	if *capture != "" {
		pt.SetClock(func() time.Time { return time.Unix(0, 0) })
	}
	mygo.App.WhenReady(func() {
		w := mygo.NewWindow(mygo.WindowOptions{
			Title:   "Platinum Gallery",
			Width:   width,
			Height:  height,
			Content: ui.View(g.view),
		})
		if *capture != "" {
			go func() {
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
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
