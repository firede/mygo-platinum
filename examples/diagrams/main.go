// Command diagrams shows every type of Mermaid diagram limu reads, drawn in
// the Platinum look by the diagram package: a list of the types, a pop-up
// menu of each type's examples, and the example's text beside its drawing,
// to edit. The pointer's item is named in the status bar. With -src, it
// shows a file of Mermaid instead of an example. With -png, it renders the
// window to a PNG without showing it, and with -frame too, only the drawing,
// in a window of the look with -frame's title, as large as the drawing:
//
//	go run ./examples/diagrams
//	go run ./examples/diagrams -png diagrams.png -type Sequence -example 2
//	go run ./examples/diagrams -src review.mmd -png review.png -frame Review -scale 2
package main

import (
	"flag"
	"fmt"
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
const width, height = 1040, 660

type app struct {
	types    []sampleType
	typ      int // the type chosen
	example  int // and its example
	src      string
	sc       *scene.Scene
	err      string
	split    float32
	scroll   ui.ScrollState
	over     string
	typeList []string
}

func newApp() *app {
	a := &app{types: samples(), split: 320}
	for _, t := range a.types {
		a.typeList = append(a.typeList, t.title)
	}
	a.choose(a.find("Flowchart"), 0)
	return a
}

// find returns the index of the type of the title, or 0.
func (a *app) find(title string) int {
	for i, t := range a.types {
		if t.title == title {
			return i
		}
	}
	return 0
}

// choose shows the example of the type, and lays it out.
func (a *app) choose(typ, example int) {
	a.typ, a.example = typ, example
	a.src = a.types[typ].examples[example].src
	a.sc = nil
	a.scroll = ui.ScrollState{}
	a.layout()
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
	t := a.types[a.typ]
	ui.Column(c).Fill().AlignItems(ui.Stretch).Background(platinum.Gray2).Children(func() {
		ui.Row(c).Grow(1).MinHeight(0).Padding(8).Gap(8).Children(func() {
			ui.Column(c).Width(170).AlignSelf(ui.Stretch).AlignItems(ui.Stretch).Children(func() {
				list := platinum.ListBox(c, &a.typ, a.typeList).Label("Types")
				list.Element().Grow(1)
				if list.Changed() {
					a.choose(a.typ, 0)
				}
			})
			ui.Column(c).Grow(1).AlignSelf(ui.Stretch).Gap(8).AlignItems(ui.Stretch).Children(func() {
				names := make([]string, len(t.examples))
				for i, e := range t.examples {
					names[i] = e.title
				}
				ui.Row(c).Children(func() {
					if platinum.PopupMenu(c, &a.example, "Example:", names).Changed() {
						a.choose(a.typ, a.example)
					}
				})
				platinum.Split(c, &a.split, func() {
					ui.Column(c).Fill().Children(func() {
						if platinum.TextArea(c, &a.src).Label(t.title).Element().Fill().Changed() {
							a.layout()
						}
					})
				}, func() {
					ui.Column(c).Fill().Children(func() {
						platinum.ScrollView(c, &a.scroll, func() {
							ui.Box(c).Background(platinum.White).Padding(8).Children(func() {
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
			})
		})
		status := fmt.Sprintf("%s, %d of %d", t.title, a.example+1, len(t.examples))
		switch {
		case a.err != "":
			status = a.err
		case a.over != "":
			status = a.over
		}
		platinum.StatusBar(c, status)
	})
}

// framed returns a view of the scene alone, on white paper in a window of
// the look with the title, on the desk's gray, and the size of the view.
func framed(sc *scene.Scene, title string) (func(*ui.Context), int, int) {
	const margin, paper = 12, 12
	st := platinum.WindowState{
		X: margin, Y: margin,
		W: int(sc.Width) + 2*paper + 2*platinum.FrameWidth,
		H: platinum.TitleBarHeight + int(sc.Height) + 2*paper + platinum.FrameWidth,
	}
	view := func(c *ui.Context) {
		platinum.Use(c, nil)
		ui.Box(c).Fill().Background(platinum.Gray5).Children(func() {
			platinum.Window(c, &st, title, platinum.WindowOptions{Active: true, NoGrow: true}, func() {
				ui.Box(c).Fill().Background(platinum.White).Padding(paper).Children(func() {
					diagram.View(c, sc)
				})
			})
		})
	}
	return view, st.W + 2*margin, st.H + 2*margin
}

func main() {
	out := flag.String("png", "", "render the window to this PNG file and exit")
	scale := flag.Float64("scale", 1, "the scale of the PNG")
	typ := flag.String("type", "Flowchart", "the type of diagram to show, by its title")
	example := flag.Int("example", 1, "the example of the type to show, from 1")
	src := flag.String("src", "", "a file of Mermaid to show instead of an example")
	frame := flag.String("frame", "", "with -png, render only the drawing, in a window of this title")
	flag.Parse()
	a := newApp()
	t := a.find(*typ)
	if a.types[t].title != *typ {
		log.Fatalf("no type %q", *typ)
	}
	if *example < 1 || *example > len(a.types[t].examples) {
		log.Fatalf("%s has examples 1 to %d", *typ, len(a.types[t].examples))
	}
	a.choose(t, *example-1)
	if *src != "" {
		data, err := os.ReadFile(*src)
		if err != nil {
			log.Fatal(err)
		}
		a.src = string(data)
		a.layout()
		if a.err != "" {
			log.Fatal(a.err)
		}
	}
	if *out != "" {
		view, w, h := a.view, width, height
		if *frame != "" {
			view, w, h = framed(a.sc, *frame)
		}
		img := ui.Render(view, w, h, float32(*scale))
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
		mygo.NewWindow(mygo.WindowOptions{Title: "Diagrams", Width: width, Height: height, Content: ui.View(a.view)})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
