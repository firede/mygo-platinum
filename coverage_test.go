package platinum

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// snapshotCoverage lists, for every control of the package, the golden
// images that show its states. TestSnapshotCoverage finds the controls
// itself — every exported function taking a *ui.Context first — so that a
// new control without snapshots fails it.
var snapshotCoverage = map[string]struct {
	states  string
	goldens []string
}{
	"Button":           {"normal pressed disabled default focused", []string{"button-states", "focus-states"}},
	"Checkbox":         {"off on mixed × normal pressed disabled, focused", []string{"checkbox-states", "check-radio-widgets", "focus-states"}},
	"Radio":            {"off on × normal pressed disabled, focused", []string{"radio-states", "check-radio-widgets", "focus-states"}},
	"RadioGroup":       {"a group of radios", []string{"check-radio-widgets"}},
	"Slider":           {"normal ticks disabled pressed thumb, focused", []string{"slider-states", "focus-states"}},
	"ProgressBar":      {"empty partial full indeterminate", []string{"progress-states"}},
	"Disclosure":       {"closed open × normal pressed disabled, focused", []string{"disclosure-states", "focus-states"}},
	"GroupBox":         {"titled group", []string{"group-tabs"}},
	"Separator":        {"etched line", []string{"group-tabs"}},
	"TabControl":       {"normal selected pressed disabled focused, pane", []string{"tabs-states", "group-tabs", "focus-states"}},
	"MenuBar":          {"open menu, highlighted item, submenu, disabled item, check mark", []string{"menubar-open"}},
	"PopupMenu":        {"normal pressed disabled open focused", []string{"popup-states", "popup-open", "focus-states"}},
	"TextField":        {"text placeholder disabled focused selection; accented letters", []string{"textfield-states", "textfield-focused", "focus-states", "accents"}},
	"TextArea":         {"wrapped lines, selection, focused; first lines indented", []string{"textarea-focused", "first-line-indent"}},
	"SearchField":      {"magnifier, clear button", []string{"stepper-status"}},
	"ListBox":          {"focused and unfocused selection, disabled, scrolled", []string{"listbox-states", "listbox-focused", "focus-states"}},
	"ScrollBar":        {"normal, thumb, arrow and track pressed, inactive, horizontal", []string{"scrollbar-states"}},
	"Window":           {"active inactive, boxes normal and pressed, utility, controls in the background, path menu", []string{"window-frames", "titlebox-states", "iconview-utility", "inactive-window", "window-path-menu"}},
	"Alert":            {"caution alert with default and cancel buttons; note and stop icons", []string{"alert-caution", "alert-icons"}},
	"Dialog":           {"movable modal dialog", []string{"dialog-modal"}},
	"Label":            {"bold text", []string{"dialog-modal"}},
	"TextIn":           {"text in a face and color", []string{"split-scrollview"}},
	"Paragraph":        {"wrapped text", []string{"alert-caution"}},
	"BevelButton":      {"normal on off disabled pressed, icon, menu, focused", []string{"bevel-states", "focus-states"}},
	"IconButton":       {"icon alone, on", []string{"bevel-states"}},
	"Segmented":        {"chosen segment, focused", []string{"bevel-states", "focus-states"}},
	"Toolbar":          {"items, separator, toggle, overflow", []string{"bevel-states", "toolbar-overflow"}},
	"ToolbarSeparator": {"etched line", []string{"bevel-states"}},
	"Stepper":          {"normal, up pressed, disabled, focused", []string{"stepper-status", "focus-states"}},
	"NumberField":      {"number and arrows", []string{"stepper-status"}},
	"ChasingArrows":    {"four steps", []string{"stepper-status"}},
	"Placard":          {"raised band", []string{"stepper-status"}},
	"Badge":            {"inverted label", []string{"stepper-status"}},
	"Tag":              {"outlined label", []string{"stepper-status"}},
	"ImageWell":        {"empty well", []string{"well-balloon"}},
	"Balloon":          {"balloon with tail", []string{"well-balloon"}},
	"ContextMenu":      {"open menu, disabled item", []string{"context-menu"}},
	"TreeView":         {"open and closed items, icons, focused selection", []string{"treeview"}},
	"Table":            {"sorted header, multiple selection, stripes, resized column", []string{"table"}},
	"IconView":         {"selected focused icon", []string{"iconview-utility"}},
	"Split":            {"divider", []string{"split-scrollview"}},
	"ScrollView":       {"scrolled content", []string{"split-scrollview"}},
	"Calendar":         {"month, chosen day focused, other months' days", []string{"calendar", "datepicker-open"}},
	"DatePicker":       {"button, open calendar", []string{"datepicker-open", "combobox-open"}},
	"TimeField":        {"hours and minutes, focused", []string{"timefield"}},
	"ComboBox":         {"text, arrow pressed, filtered menu", []string{"combobox-open"}},
	"Sheet":            {"headers, numbers right-aligned, range and cursor focused", []string{"sheet"}},
	"FormulaBar":       {"cell name and its text", []string{"sheet"}},
	"FindReplace":      {"fields, options, note, default Find", []string{"find-replace"}},
	"Ruler":            {"scale, margins dragged", []string{"ruler"}},
	"ProgressDialog":   {"message, bar, detail, cancel", []string{"progress-dialog-banner"}},
	"Banner":           {"note icon, text, close box", []string{"progress-dialog-banner"}},
	"EmptyState":       {"empty folder, title, detail", []string{"paths-pages"}},
	"PathPopup":        {"folder button with menu arrow", []string{"paths-pages"}},
	"Breadcrumb":       {"path with arrows, last one black", []string{"paths-pages"}},
	"Pagination":       {"arrows and page", []string{"paths-pages"}},
	"Card":             {"raised panel", []string{"paths-pages"}},
	"StatusBar":        {"placards", []string{"paths-pages"}},
	"ColumnView":       {"columns, items holding others marked, focused column", []string{"columnview"}},
	"DocumentTabs":     {"tabs, selected tab and pane", []string{"document-tabs"}},
	"FontPicker":       {"families, sizes, styles, styled sample", []string{"fontpicker", "pagesetup"}},
	"PageSetup":        {"paper, orientation toggles, scale, margins, preview", []string{"pagesetup"}},
	"ZoomRects":        {"outlines a third of the way and two thirds", []string{"zoomrects"}},
	"RichTextArea":     {"italic, bold and underlined runs, centered, selection; paragraphs aligned apart; text at two sizes", []string{"richtext", "richtext-paragraphs", "richtext-sizes"}},
	"FormatBar":        {"style toggles on, alignment", []string{"richtext"}},
	"ColorWell":        {"well pressed, picker with palette and sliders", []string{"colorpicker"}},
	"OpenDialog":       {"folder pop-up, list with a file chosen, Desktop, Cancel and Open", []string{"open-dialog"}},
	"SaveDialog":       {"name selected under its prompt, list unfocused; asking to replace", []string{"save-dialog", "save-dialog-replace"}},
	"WindowHeader":     {"count at the left, room at the right", []string{"window-header"}},
	"PopupWindow":      {"open in front, closed behind, tab focused, tab pressed", []string{"popup-window", "popup-window-focus", "popup-window-pressed"}},
}

// notControls are exported functions taking a context that build nothing.
var notControls = map[string]bool{"Use": true, "ThemeOf": true, "EditCommand": true, "SpringOpen": true}

func TestSnapshotCoverage(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi os.FileInfo) bool { return !strings.HasSuffix(fi.Name(), "_test.go") }, 0)
	if err != nil {
		t.Fatal(err)
	}
	var controls []string
	for _, f := range pkgs["platinum"].Files {
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Recv != nil || !fd.Name.IsExported() || notControls[fd.Name.Name] || len(fd.Type.Params.List) == 0 {
				continue
			}
			star, ok := fd.Type.Params.List[0].Type.(*ast.StarExpr)
			if !ok {
				continue
			}
			if sel, ok := star.X.(*ast.SelectorExpr); ok && sel.Sel.Name == "Context" {
				controls = append(controls, fd.Name.Name)
			}
		}
	}
	sort.Strings(controls)
	used := map[string]bool{}
	for _, name := range controls {
		cov, ok := snapshotCoverage[name]
		if !ok {
			t.Errorf("control %s has no snapshot listed in snapshotCoverage", name)
			continue
		}
		for _, g := range cov.goldens {
			used[g] = true
			if _, err := os.Stat(filepath.Join("testdata", "golden", g+".png")); err != nil {
				t.Errorf("control %s: golden %s is missing", name, g)
			}
		}
	}
	for name := range snapshotCoverage {
		if !sort.StringsAreSorted(controls) || !contains(controls, name) {
			t.Errorf("snapshotCoverage lists %s, which is not a control", name)
		}
	}
	// Every golden image belongs to a control; one that does not is stale.
	files, _ := filepath.Glob(filepath.Join("testdata", "golden", "*.png"))
	for _, f := range files {
		g := strings.TrimSuffix(filepath.Base(f), ".png")
		if !used[g] {
			t.Errorf("golden %s shows no control listed in snapshotCoverage", g)
		}
	}
	if len(controls) < 40 {
		t.Errorf("found only %d controls; the search for them is broken", len(controls))
	}
}

func contains(list []string, s string) bool {
	i := sort.SearchStrings(list, s)
	return i < len(list) && list[i] == s
}

// TestExampleCoverage requires an example of its own, Example<Control>,
// for every control of the snapshot coverage, so that each has godoc
// that compiles and runs.
func TestExampleCoverage(t *testing.T) {
	fset := token.NewFileSet()
	have := map[string]bool{}
	files, _ := filepath.Glob("example*_test.go")
	for _, f := range files {
		file, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range file.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok && strings.HasPrefix(fn.Name.Name, "Example") {
				name := strings.TrimPrefix(fn.Name.Name, "Example")
				if i := strings.Index(name, "_"); i >= 0 {
					name = name[:i]
				}
				have[name] = true
			}
		}
	}
	for control := range snapshotCoverage {
		if !have[control] {
			t.Errorf("control %s has no example of its own", control)
		}
	}
}
