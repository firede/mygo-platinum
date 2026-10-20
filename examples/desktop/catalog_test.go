package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
)

// TestCatalogNames checks the pages of the Controls against the library:
// every function a page names is one of its controls, and every control —
// an exported function taking a *ui.Context first, in the library or in its
// diagram package — has a page.
func TestCatalogNames(t *testing.T) {
	notControls := []string{"Use", "ThemeOf", "EditCommand"}
	var controls []string
	fset := token.NewFileSet()
	// The library's controls by their names, and those of its diagram
	// package by the package's name and theirs.
	for dir, prefix := range map[string]string{"../..": "", "../../diagram": "diagram."} {
		files, err := filepath.Glob(dir + "/*.go")
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range files {
			if strings.HasSuffix(f, "_test.go") {
				continue
			}
			file, err := parser.ParseFile(fset, f, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Recv != nil || !fn.Name.IsExported() || slices.Contains(notControls, fn.Name.Name) || len(fn.Type.Params.List) == 0 {
					continue
				}
				if star, ok := fn.Type.Params.List[0].Type.(*ast.StarExpr); ok {
					if sel, ok := star.X.(*ast.SelectorExpr); ok && sel.Sel.Name == "Context" {
						controls = append(controls, prefix+fn.Name.Name)
					}
				}
			}
		}
	}
	shown := map[string]bool{}
	for _, p := range catalogPages() {
		for _, name := range strings.Split(p.name, ", ") {
			// SizeMenu builds a menu, not a control of a window.
			if !slices.Contains(controls, name) && name != "SizeMenu" {
				t.Errorf("the page %q names %s, which is no control of the library", p.name, name)
			}
			shown[name] = true
		}
	}
	for _, name := range controls {
		if !shown[name] {
			t.Errorf("%s has no page in the Controls", name)
		}
	}
}

// TestCatalogPages builds every page of the Controls in each language of
// the desk, and checks the first against its golden image.
func TestCatalogPages(t *testing.T) {
	for _, lang := range languages {
		d := newDeskIn(func() time.Time { return sceneTime }, lang)
		d.windows = nil
		w := d.launch(appCatalog)
		w.frame.X, w.frame.Y = 4, 4
		tt := ui.NewTester(d.view, 612, 410)
		tt.SetLocale(lang)
		for i, p := range catalogPages() {
			w.catalog.chosen = i
			tt.Frame()
			if !tt.HasText(p.name) {
				t.Errorf("%s: the page %q does not show its name", lang, p.name)
			}
		}
		if lang == "en" || lang == "zh-Hans" {
			w.catalog.chosen = 0
			tt.Frame()
			golden(t, "controls-"+strings.ToLower(lang), tt.Image())
		}
	}
}

// The compact IconView page clips its second row and lets the user scroll
// its icon and name fully into view in either language.
func TestCatalogIconViewScroll(t *testing.T) {
	for _, lang := range []string{"en", "zh-Hans"} {
		t.Run(lang, func(t *testing.T) {
			d := newDeskIn(func() time.Time { return sceneTime }, lang)
			d.windows = nil
			w := d.launch(appCatalog)
			w.frame.X, w.frame.Y = 4, 4
			for i, p := range catalogPages() {
				if p.name == "IconView" {
					w.catalog.chosen = i
				}
			}
			tt := ui.NewTester(d.view, 640, 410)
			tt.SetLocale(lang)
			before, ok := tt.Find(d.say().Puzzle())
			if !ok || before.H >= 64 {
				t.Fatalf("the compact view should show part of its second row: %+v", before)
			}
			golden(t, "controls-iconview-"+strings.ToLower(lang), tt.Image())
			tt.Scroll(before.X+20, before.Y+10, 0, 1000)
			settled(t, tt, "the catalog's icons scrolled")
			after, ok := tt.Find(d.say().Puzzle())
			if !ok || after.H != 64 || after.Y >= before.Y || w.catalog.icon != 1 {
				t.Fatalf("scroll did not reveal the second row: before=%+v after=%+v choice=%d", before, after, w.catalog.icon)
			}
			golden(t, "controls-iconview-scrolled-"+strings.ToLower(lang), tt.Image())
		})
	}
}

// The Window page's windows follow the pointer all the way while dragged,
// and keep their titles as they change places.
func TestCatalogWindowDrag(t *testing.T) {
	d := newDeskIn(func() time.Time { return sceneTime }, "en")
	d.windows = nil
	w := d.launch(appCatalog)
	w.frame.X, w.frame.Y = 4, 4
	for i, p := range catalogPages() {
		if p.name == "Window" {
			w.catalog.chosen = i
		}
	}
	tt := ui.NewTester(d.view, 640, 410)
	tt.Frame()
	back, front := w.catalog.windows[0], w.catalog.windows[1]
	r, _ := tt.Find(d.say().Untitled())
	// The window in front, by its title bar, from where the one behind is.
	x := r.X - float32(back.X) + float32(front.X) + 100
	y := r.Y - float32(back.Y) + float32(front.Y) + 6
	tt.Press(x, y)
	for i := 1; i <= 6; i++ {
		tt.Move(x+float32(5*i), y+float32(3*i))
		tt.Frame()
	}
	tt.Release(x+30, y+18)
	tt.Frame()
	if front.X != 60 || front.Y != 26 {
		t.Fatalf("the drag stopped short: the window is at %d,%d", front.X, front.Y)
	}
	tt.ClickAt(r.X+8, r.Y+r.H/2) // brings the untitled window forward
	tt.Frame()
	if w.catalog.windows[1] != back {
		t.Fatal("the click did not bring the untitled window forward")
	}
	if got, _ := tt.Find(d.say().Untitled()); got != r {
		t.Fatalf("the untitled window's title moved from %+v to %+v", r, got)
	}
}

// The ZoomRects page's button shows the rectangles, by the frames' clock,
// which moves here as the test says.
func TestCatalogZoomRects(t *testing.T) {
	defer pt.SetClock(pt.SetClock(nil))
	d := newDeskIn(func() time.Time { return sceneTime }, "en")
	d.windows = nil
	w := d.launch(appCatalog)
	w.frame.X, w.frame.Y = 4, 4
	for i, p := range catalogPages() {
		if p.name == "ZoomRects" {
			w.catalog.chosen = i
		}
	}
	tt := ui.NewTester(d.view, 640, 410)
	tt.Frame()
	before := bytes.Clone(tt.Image().Pix)
	if err := tt.Click(d.say().CatShow()); err != nil {
		t.Fatal(err)
	}
	tt.Advance(pt.ZoomDuration / 3)
	if bytes.Equal(before, tt.Image().Pix) {
		t.Fatal("no rectangles showed a third of the way")
	}
	tt.Advance(pt.ZoomDuration)
	if !bytes.Equal(before, tt.Image().Pix) {
		t.Fatal("the rectangles stayed after they were done")
	}
}

// TestCatalogKeepsState works the controls of a few pages, which keep what
// the user did — a folder of the tree opened, a window brought forward, the
// mixed check box chosen — and shows an alert, in front of every window.
func TestCatalogKeepsState(t *testing.T) {
	d := newDeskIn(func() time.Time { return sceneTime }, "en")
	d.windows = nil
	w := d.launch(appCatalog)
	w.frame.X, w.frame.Y = 4, 4
	k := w.catalog
	tt := ui.NewTester(d.view, 900, 620)
	page := func(name string) {
		for i, p := range catalogPages() {
			if p.name == name {
				k.chosen = i
			}
		}
		tt.Frame()
	}

	page("TreeView")
	if err := tt.Click("Documents"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyRight) // opens it
	tt.Frame()
	tt.Frame()
	if !tt.HasText("Letter") {
		t.Error("the Documents folder did not stay open")
	}

	page("Window")
	front := k.windows[len(k.windows)-1]
	r, _ := tt.Find("untitled") // the window behind, its left edge clear
	tt.ClickAt(r.X+8, r.Y+r.H/2)
	tt.Frame()
	if k.windows[0] != front {
		t.Error("a click on the window behind did not bring it forward")
	}

	page("Checkbox")
	if err := tt.Click("Three"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if k.mixedShown {
		t.Error("the mixed check box stayed mixed when chosen")
	}

	// An alert of the Controls is in front of the desk's windows too.
	page("Alert")
	if err := tt.Click("Show"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !d.modalUp() {
		t.Error("the Controls' alert left the windows active")
	}
}
