// Command flowchart edits a flowchart's text beside the drawing of it, in
// the Platinum look; with -png, it renders the window to a PNG without
// showing it:
//
//	go run ./examples/flowchart
//	go run ./examples/flowchart -png flowchart.png -scale 2
package main

import (
	"flag"
	"image/png"
	"log"
	"os"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/firede/limu"
	"github.com/firede/limu/scene"
	platinum "github.com/firede/mygo-platinum"
	"github.com/firede/mygo-platinum/diagram"
)

// The size of the window, in DIPs.
const width, height = 960, 600

const sample = `flowchart TD
  A([Customer]) -->|places| B[Order received]
  B --> C{In stock?}
  C -->|yes| D[[Pack items]]
  C -->|no| E[(Backorder)]
  E -.-> F>Notify customer]
  D ==> G((Ship))
  subgraph Warehouse
    D
    G
    H{{Quality check}}
  end
  G --> H
  B --> I[/Invoice/]
`

type app struct {
	src    string
	sc     *scene.Scene
	err    string
	split  float32
	scroll ui.ScrollState
	over   string
}

func newApp() *app {
	a := &app{src: sample, split: 300}
	a.layout()
	return a
}

// layout lays the text out again; text that does not read keeps the last drawing, and
// says why.
func (a *app) layout() {
	sc, err := limu.Layout(a.src, diagram.Layout())
	if err != nil {
		a.err = err.Error()
		return
	}
	a.sc, a.err = sc, ""
}

func (a *app) view(c *ui.Context) {
	platinum.Use(c, nil)
	ui.Column(c).Size(width, height).Background(platinum.Gray2).Children(func() {
		platinum.Split(c, &a.split, func() {
			ui.Column(c).Fill().Padding(8).Children(func() {
				if platinum.TextArea(c, &a.src).Label("Flowchart").Element().Fill().Changed() {
					a.layout()
				}
			})
		}, func() {
			ui.Column(c).Fill().Padding(8).Children(func() {
				platinum.ScrollView(c, &a.scroll, func() {
					ui.Box(c).Padding(8).Children(func() {
						if a.sc == nil {
							return
						}
						w := diagram.View(c, a.sc)
						a.over = ""
						if it := w.Hovered(); it != nil {
							a.over = string(it.Role) + " " + it.ID
						}
					})
				}).Fill()
			})
		}).Grow(1).MinHeight(0)
		status := "Ready"
		switch {
		case a.err != "":
			status = a.err
		case a.over != "":
			status = a.over
		}
		platinum.StatusBar(c, status)
	})
}

func main() {
	out := flag.String("png", "", "render the window to this PNG file and exit")
	scale := flag.Float64("scale", 1, "the scale of the PNG")
	flag.Parse()
	a := newApp()
	if *out != "" {
		img := ui.Render(a.view, width, height, float32(*scale))
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
	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{Title: "Flowchart", Width: width, Height: height, Content: ui.View(a.view)})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
