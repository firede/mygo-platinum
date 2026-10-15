package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// TestCatalogNames checks the pages of the Controls against the library:
// every function a page names is one of its controls, and every control —
// an exported function taking a *ui.Context first — has a page.
func TestCatalogNames(t *testing.T) {
	notControls := []string{"Use", "ThemeOf", "EditCommand"}
	files, err := filepath.Glob("../../*.go")
	if err != nil {
		t.Fatal(err)
	}
	var controls []string
	fset := token.NewFileSet()
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
					controls = append(controls, fn.Name.Name)
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
