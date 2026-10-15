package main

import (
	"bytes"
	"flag"
	"image"
	"image/draw"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
)

var update = flag.Bool("update", false, "write the golden images anew")

func golden(t *testing.T, name string, got *image.RGBA) {
	t.Helper()
	path := filepath.Join("testdata", "golden", name+".png")
	data, err := os.ReadFile(path)
	if err != nil {
		if *update {
			writePNG(t, path, got)
			return
		}
		t.Fatalf("golden %s: %v", path, err)
	}
	m, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	want := image.NewRGBA(m.Bounds())
	draw.Draw(want, want.Bounds(), m, m.Bounds().Min, draw.Src)
	if want.Bounds() != got.Bounds() || !bytes.Equal(want.Pix, got.Pix) {
		out := filepath.Join(os.TempDir(), "desk-"+name+".png")
		writePNG(t, out, got)
		if *update {
			writePNG(t, path, got)
			return
		}
		t.Fatalf("%s differs from its golden; got %s", name, out)
	}
}

func writePNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	os.MkdirAll(filepath.Dir(path), 0o755)
	var b bytes.Buffer
	png.Encode(&b, img)
	if err := os.WriteFile(path, b.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

func center(t *testing.T, tt *ui.Tester, label string) (float32, float32) {
	t.Helper()
	r, ok := tt.Find(label)
	if !ok {
		t.Fatalf("nothing shows %q; texts %q", label, tt.Texts())
	}
	return r.X + r.W/2, r.Y + r.H/2
}

func doubleClick(t *testing.T, tt *ui.Tester, label string) {
	t.Helper()
	x, y := center(t, tt, label)
	tt.ClickAt(x, y)
	tt.ClickAt(x, y)
	tt.Frame()
}

func chooseMenu(t *testing.T, tt *ui.Tester, menu, item string) {
	t.Helper()
	x, y := center(t, tt, menu)
	tt.ClickAt(x, y)
	tt.Frame()
	ix, iy := center(t, tt, item)
	tt.ClickAt(ix, iy)
	tt.Frame()
	tt.Frame()
}

// TestDesktopScene renders the scene of the desk and compares it with its
// golden image.
func TestDesktopScene(t *testing.T) {
	defer pt.SetClock(pt.SetClock(nil))
	golden(t, "scene", renderScene(1, ""))
}

func TestDesktopCloseButtonRedraws(t *testing.T) {
	for _, collapsed := range []bool{false, true} {
		d := newDesk(func() time.Time { return sceneTime })
		w := d.launch(appCalculator)
		w.frame.Collapsed = collapsed
		tt := ui.NewTester(d.view, width, height)
		x := float32(w.frame.X + 13)
		y := float32(w.frame.Y + pt.MenuBarHeight + 9)
		tt.ClickAt(x, y)
		// Do not send another input or force a Frame: the close event must
		// remove both the window and its rendered elements on its own.
		if d.find(appCalculator) != nil {
			t.Fatalf("collapsed %v: one click did not close the window", collapsed)
		}
		if _, visible := tt.Find("Calculator"); visible {
			t.Fatalf("collapsed %v: the closed window still shows", collapsed)
		}
	}
}

func TestCalculatorPhysicalOperators(t *testing.T) {
	d := newDesk(func() time.Time { return sceneTime })
	w := d.launch(appCalculator)
	tt := ui.NewTester(d.view, width, height)
	for _, test := range []struct {
		key  ui.Key
		want string
	}{
		{ui.KeyEqual, "5"},
		{ui.Key8, "6"},
	} {
		tt.Key(0, ui.KeyEscape)
		tt.Key(0, ui.Key2)
		// macOS reports the shifted symbol as KeyUnknown, retaining its
		// ANSI key position. Ordinary shortcut tests do not model this.
		tt.KeyAt(ui.Shift, test.key)
		tt.Key(0, ui.Key3)
		tt.Key(0, ui.KeyEnter)
		if w.calc.shown != test.want {
			t.Fatalf("2, Shift+%v, 3, Enter: got %q, want %q", test.key, w.calc.shown, test.want)
		}
		// Clicking a keypad button moves focus to it; shifted operators
		// must keep working without applying an operator twice.
		x, y := center(t, tt, "C")
		tt.ClickAt(x, y)
		tt.Key(0, ui.Key2)
		tt.KeyAt(ui.Shift, test.key)
		tt.Key(0, ui.Key3)
		x, y = center(t, tt, "=")
		tt.ClickAt(x, y)
		if w.calc.shown != test.want {
			t.Fatalf("with keypad focus: got %q, want %q", w.calc.shown, test.want)
		}
	}
	// A command chord or an unrelated shifted key is not an operator.
	tt.KeyAt(ui.Cmd|ui.Shift, ui.KeyEqual)
	tt.KeyAt(ui.Shift, ui.Key1)
	if w.calc.shown != "6" || w.calc.op != "" {
		t.Fatal("unrelated keys changed the result")
	}
	d.launch(appScribe)
	tt.Frame()
	tt.KeyAt(ui.Shift, ui.KeyEqual)
	if w.calc.shown != "6" || w.calc.op != "" {
		t.Fatal("a background calculator handled input")
	}
}

// TestDesktopSmoke drives the desk as a user would: opens the disk, a
// folder and the calculator from icons, calculates with the pointer and
// the keyboard, drags a window, uses the menus, the puzzle, the appearance
// control panel and the text editor, and compares the end with its golden
// image.
func TestDesktopSmoke(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()

	// The disk, then Applications, then the calculator.
	doubleClick(t, tt, "Platinum Disk")
	if d.front() == nil || d.front().folder == nil {
		t.Fatal("double-clicking the disk opened no window")
	}
	doubleClick(t, tt, "Applications")
	doubleClick(t, tt, "Calculator")
	calc := d.find(appCalculator)
	if calc == nil {
		t.Fatal("the calculator did not open")
	}
	for _, k := range []string{"1", "2", "+", "3", "0", "="} {
		x, y := center(t, tt, k)
		tt.ClickAt(x, y)
		tt.Frame()
	}
	if calc.calc.shown != "42" {
		t.Fatalf("12+30 shows %q", calc.calc.shown)
	}
	// The keyboard: 6 × 7 =.
	tt.Key(0, ui.KeyEscape)
	tt.Key(0, ui.Key6)
	tt.Key(ui.Shift, ui.Key8)
	tt.Key(0, ui.Key7)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if calc.calc.shown != "42" {
		t.Fatalf("6×7 shows %q", calc.calc.shown)
	}
	// Drag the calculator by its title bar.
	x0, y0 := calc.frame.X, calc.frame.Y
	tx, ty := float32(x0+60), float32(y0+pt.MenuBarHeight+8)
	tt.Press(tx, ty)
	tt.Move(tx+100, ty+40)
	tt.Release(tx+100, ty+40)
	tt.Frame()
	if calc.frame.X != x0+100 || calc.frame.Y != y0+40 {
		t.Fatalf("the calculator moved to %d,%d from %d,%d", calc.frame.X, calc.frame.Y, x0, y0)
	}

	// Command-` goes round the windows.
	front := d.front()
	tt.Key(ui.Cmd, ui.KeyBackquote)
	tt.Frame()
	if d.front() == front || d.front() != d.windows[len(d.windows)-1] {
		t.Fatal("Command-` did not bring another window to the front")
	}
	for d.front() != front {
		tt.Key(ui.Cmd, ui.KeyBackquote)
		tt.Frame()
	}

	// Command-W closes the windows, front first.
	for i := 0; i < 3; i++ {
		tt.Key(ui.Cmd, ui.KeyW)
		tt.Frame()
		tt.Frame()
	}
	if len(d.windows) != 0 {
		t.Fatalf("%d windows still open", len(d.windows))
	}

	// The puzzle from the Desk menu; the arrows slide its tiles.
	chooseMenu(t, tt, "Desk", "Puzzle")
	pz := d.find(appPuzzle)
	if pz == nil || d.front() != pz {
		t.Fatal("the Desk menu did not open the puzzle")
	}
	// Whichever side of the hole has a tile, an arrow slides it.
	for _, k := range []ui.Key{ui.KeyLeft, ui.KeyRight, ui.KeyUp, ui.KeyDown} {
		tt.Key(0, k)
		tt.Frame()
	}
	if pz.puzzle.moves < 2 {
		t.Fatalf("the arrows made %d moves", pz.puzzle.moves)
	}
	chooseMenu(t, tt, "Puzzle", "Solve")
	if !pz.puzzle.solved() {
		t.Fatal("Solve did not solve the puzzle")
	}

	// The appearance control panel: another accent and pattern.
	chooseMenu(t, tt, "Desk", "Control Panels")
	x, y := center(t, tt, "Appearance")
	tt.ClickAt(x, y)
	tt.Frame()
	tt.Frame()
	if d.find(appAppearance) == nil {
		t.Fatal("the control panel did not open")
	}
	x, y = center(t, tt, "Moss")
	tt.ClickAt(x, y)
	tt.Frame()
	if d.theme.Accent.Name != "Moss" {
		t.Fatalf("the accent is %s", d.theme.Accent.Name)
	}
	if err := tt.Click("Desktop"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	x, y = center(t, tt, "Teal Bricks")
	tt.ClickAt(x, y)
	tt.Frame()
	if d.pattern != 2 {
		t.Fatalf("the pattern is %d", d.pattern)
	}
	tt.Key(ui.Cmd, ui.KeyW)
	tt.Frame()
	tt.Frame()
	if d.find(appAppearance) != nil {
		t.Fatal("Command-W did not close the control panel")
	}

	// A document in the text editor: type, save.
	doubleClick(t, tt, "Platinum Disk")
	doubleClick(t, tt, "Read Me")
	ed := d.front()
	if ed == nil || ed.doc == nil {
		t.Fatal("the document did not open")
	}
	x, y = center(t, tt, "Document")
	tt.ClickAt(x, y)
	tt.Key(ui.Cmd, ui.KeyUp)
	tt.Type("Hello! ")
	tt.Key(ui.Cmd, ui.KeyS)
	tt.Frame()
	tt.Frame()
	if !strings.HasPrefix(ed.doc.file.text, "Hello! Welcome") {
		t.Fatalf("the saved document begins %q", ed.doc.file.text[:20])
	}
	golden(t, "smoke-end", tt.Image())
}

// TestDesktopWastebasket moves a document to the wastebasket and empties
// it, through the alert asking first.
func TestDesktopWastebasket(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	doubleClick(t, tt, "Platinum Disk")
	x, y := center(t, tt, "Read Me")
	tt.ClickAt(x, y)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyBackspace)
	tt.Frame()
	tt.Frame()
	if len(d.trash) != 1 || d.disk.has("Read Me") {
		t.Fatalf("trash %d, disk still has it: %v", len(d.trash), d.disk.has("Read Me"))
	}
	chooseMenu(t, tt, "Special", "Empty Wastebasket…")
	if !d.emptying {
		t.Fatal("no alert asked to empty the wastebasket")
	}
	golden(t, "empty-alert", tt.Image())
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if len(d.trash) != 0 {
		t.Fatal("OK did not empty the wastebasket")
	}
}

// TestDesktopAlarm rings the alarm clock when its time comes.
func TestDesktopAlarm(t *testing.T) {
	now := sceneTime
	defer pt.SetClock(pt.SetClock(func() time.Time { return now }))
	d := newDesk(func() time.Time { return now })
	tt := ui.NewTester(d.view, width, height)
	chooseMenu(t, tt, "Desk", "Alarm Clock")
	clock := d.find(appClock)
	if clock == nil {
		t.Fatal("the alarm clock did not open")
	}
	x, y := center(t, tt, "Alarm at")
	tt.ClickAt(x, y)
	tt.Frame()
	if !clock.clock.armed {
		t.Fatal("the alarm is not set")
	}
	golden(t, "alarm-clock", tt.Image())
	now = now.Add(4 * time.Minute)
	tt.Frame()
	tt.Frame()
	if !d.ringing || !tt.HasText("It is time!") {
		t.Fatal("the alarm did not ring")
	}
}

// TestDesktopOffice uses the ledger, the editor's find and replace, and a
// context menu.
func TestDesktopOffice(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	// The ledger: B7 sums B2:B5; change B2 and the total follows.
	chooseMenu(t, tt, "Desk", "Ledger")
	l := d.find(appLedger)
	if l == nil {
		t.Fatal("the ledger did not open")
	}
	if got := l.sheet.Value(pt.Cell{Col: 1, Row: 6}); got != "1095.5" {
		t.Fatalf("the total is %s", got)
	}
	x, y := center(t, tt, "Sheet")
	tt.ClickAt(x, y) // focus the sheet
	l.sheet.Cursor, l.sheet.Anchor = pt.Cell{Col: 1, Row: 1}, pt.Cell{Col: 1, Row: 1}
	tt.Frame()
	tt.Type("900")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if got := l.sheet.Value(pt.Cell{Col: 1, Row: 6}); got != "1195.5" {
		t.Fatalf("after editing, the total is %s", got)
	}
	golden(t, "ledger", tt.Image())
	tt.Key(ui.Cmd, ui.KeyW)
	tt.Frame()
	tt.Frame()

	// The editor: find and replace every "the".
	chooseMenu(t, tt, "Desk", "Scribe")
	ed := d.find(appScribe)
	ed.doc.doc.Text = "the cat and the hat"
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyF)
	tt.Frame()
	tt.Frame()
	if !ed.doc.finding {
		t.Fatal("Command-F did not open find and replace")
	}
	tt.Type("the")
	tt.Key(0, ui.KeyTab)
	tt.Type("a")
	tt.Frame()
	if err := tt.Click("Replace All"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if ed.doc.doc.Text != "a cat and a hat" {
		t.Fatalf("replaced into %q", ed.doc.doc.Text)
	}
	golden(t, "scribe-find", tt.Image())
	tt.Key(0, ui.KeyEscape)
	tt.Frame()

	// A context menu on the disk's icon opens it.
	if err := tt.RightClick("Platinum Disk"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	x, y = center(t, tt, "Open")
	tt.ClickAt(x, y)
	tt.Frame()
	tt.Frame()
	if f := d.front(); f == nil || f.folder == nil || f.folder.dir != d.disk {
		t.Fatal("the context menu did not open the disk")
	}
	// Its list view is a table that sorts by size.
	chooseMenu(t, tt, "View", "as List")
	x, y = center(t, tt, "Size")
	tt.ClickAt(x, y)
	tt.Frame()
	golden(t, "cabinet-list", tt.Image())
}

// TestDesktopControlPanels sets the desk's clock in Date & Time and mutes
// the sound.
func TestDesktopControlPanels(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	chooseMenu(t, tt, "Desk", "Control Panels")
	x, y := center(t, tt, "Date & Time")
	tt.ClickAt(x, y)
	tt.Frame()
	tt.Frame()
	// The time's up arrow, with the hours chosen: up an hour.
	r, ok := tt.Find("Change time")
	if !ok {
		t.Fatal("no little arrows of the time")
	}
	tt.ClickAt(r.X+6, r.Y+4)
	tt.Frame()
	tt.Frame()
	if got := d.now().Hour(); got != 10 {
		t.Fatalf("the desk's clock says %d o'clock", got)
	}
	d.find(appDateTime).frame.X = 20
	chooseMenu(t, tt, "Desk", "Control Panels")
	x, y = center(t, tt, "Sound")
	tt.ClickAt(x, y)
	tt.Frame()
	tt.Frame()
	x, y = center(t, tt, "Mute")
	tt.ClickAt(x, y)
	tt.Frame()
	if !d.settings.muted {
		t.Fatal("Mute is not checked")
	}
	d.find(appSound).frame.X, d.find(appSound).frame.Y = 420, 20
	tt.Frame()
	golden(t, "control-panels", tt.Image())
}

// TestDesktopScalesByWholePixels checks the scene at scale 2 is the scene
// at scale 1 with each pixel a 2×2 block.
func TestDesktopScalesByWholePixels(t *testing.T) {
	defer pt.SetClock(pt.SetClock(nil))
	one, two := renderScene(1, ""), renderScene(2, "")
	bad := 0
	for y := 0; y < 2*height; y++ {
		for x := 0; x < 2*width; x++ {
			if two.RGBAAt(x, y) != one.RGBAAt(x/2, y/2) {
				bad++
			}
		}
	}
	if bad > 0 {
		t.Errorf("%d pixels at scale 2 are not blocks of scale 1's", bad)
	}
}

// TestDesktopDragAndDrop drags a document into a folder, and another to
// the Wastebasket.
func TestDesktopDragAndDrop(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	doubleClick(t, tt, "Platinum Disk")
	drag := func(from, to string) {
		t.Helper()
		x0, y0 := center(t, tt, from)
		x1, y1 := center(t, tt, to)
		tt.Press(x0, y0)
		for i := 1; i <= 4; i++ {
			f := float32(i) / 4
			tt.Move(x0+(x1-x0)*f, y0+(y1-y0)*f)
			tt.Frame()
		}
		tt.Release(x1, y1)
		tt.Frame()
	}
	drag("Read Me", "Documents")
	if d.disk.has("Read Me") {
		t.Fatal("Read Me is still on the disk")
	}
	var docs *node
	for _, n := range d.disk.children {
		if n.name == "Documents" {
			docs = n
		}
	}
	if !docs.has("Read Me") {
		t.Fatal("Read Me did not move into Documents")
	}
	drag("System Folder", "Wastebasket")
	if len(d.trash) != 1 || d.trash[0].name != "System Folder" {
		t.Fatalf("the Wastebasket holds %d", len(d.trash))
	}
}

// TestDesktopApplicationMenu switches applications from the application
// menu, and hides the others.
func TestDesktopApplicationMenu(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	chooseMenu(t, tt, "Desk", "Puzzle")
	chooseMenu(t, tt, "Desk", "Calculator")
	if !tt.HasText("9:41 AM") {
		t.Fatal("no clock in the menu bar")
	}
	r, ok := tt.Find("Calculator")
	if !ok {
		t.Fatal("no application menu")
	}
	// The application menu is the rightmost title.
	tt.ClickAt(r.X+r.W/2, 10)
	tt.Frame()
	if _, ok := tt.Find("Hide Others"); !ok {
		t.Fatalf("the application menu did not open: %q", tt.Texts())
	}
	x, y := center(t, tt, "Hide Others")
	tt.ClickAt(x, y)
	tt.Frame()
	tt.Frame()
	if !d.find(appPuzzle).frame.Collapsed || d.find(appCalculator).frame.Collapsed {
		t.Fatal("Hide Others did not roll up the puzzle alone")
	}
}

// TestDesktopColumns browses the disk in columns and opens a document from
// them.
func TestDesktopColumns(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	doubleClick(t, tt, "Platinum Disk")
	d.front().frame.W = 460
	chooseMenu(t, tt, "View", "as Columns")
	x, y := center(t, tt, "Documents")
	tt.ClickAt(x, y)
	tt.Frame()
	tt.Frame()
	x, y = center(t, tt, "Letter")
	tt.ClickAt(x, y)
	tt.Frame()
	golden(t, "cabinet-columns", tt.Image())
	tt.ClickAt(x, y)
	tt.ClickAt(x, y)
	tt.Frame()
	tt.Frame()
	if f := d.front(); f == nil || f.doc == nil || f.title != "Letter" {
		t.Fatal("a double click in the columns did not open the letter")
	}
}

// TestDesktopFindFile finds items by name and kind, and opens one.
func TestDesktopFindFile(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	chooseMenu(t, tt, "File", "Find File…")
	if d.find(appFind) == nil {
		t.Fatal("Find File did not open")
	}
	// Typing finds nothing yet; Find does.
	tt.Type("re")
	tt.Frame()
	if tt.HasText("Recipes") {
		t.Fatal("Find File searched before Find was clicked")
	}
	if err := tt.Click("Find"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText("Recipes") || !tt.HasText("Read Me") {
		t.Fatalf("found %q", tt.Texts())
	}
	golden(t, "find-file", tt.Image())
	if said := tt.Announcements(); len(said) == 0 || said[len(said)-1] != "2 items found" {
		t.Fatalf("announced %q", said)
	}
	x, y := center(t, tt, "Recipes")
	tt.ClickAt(x, y)
	tt.ClickAt(x, y)
	tt.Frame()
	tt.Frame()
	if f := d.front(); f == nil || f.title != "Recipes" {
		t.Fatal("a double click did not open Recipes")
	}
	// Return in the field finds too, the Find button being the default.
	ff := d.find(appFind)
	d.raise(ff)
	tt.Frame()
	x, y = center(t, tt, "Search for")
	tt.ClickAt(x, y)
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("to")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if !tt.HasText("To Do") {
		t.Fatal("Return in the field did not find")
	}
}

// TestScribeKeepsStyles saves a styled, aligned document and opens it
// again as it was.
func TestScribeKeepsStyles(t *testing.T) {
	d := newDesk(func() time.Time { return sceneTime })
	w := newScribeWindow(d, nil)
	runs := []pt.Run{{Text: "Menu", Style: pt.Bold}, {Text: "\nSoup of the day"}}
	w.doc.doc.SetRuns(runs)
	w.doc.doc.SetAligns([]pt.Align{pt.AlignCenter})
	w.doc.saveAs(d, d.documents(), "Menu")
	again := newScribeWindow(d, w.doc.file)
	if got := again.doc.doc.Runs(); !reflect.DeepEqual(got, runs) {
		t.Fatalf("reopened with runs %+v", got)
	}
	if got := again.doc.doc.Aligns(); !reflect.DeepEqual(got, []pt.Align{pt.AlignCenter, pt.AlignLeft}) {
		t.Fatalf("reopened with aligns %v", got)
	}
}

// TestScribeSizeMenu makes a heading larger from Scribe's Size menu.
func TestScribeSizeMenu(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	chooseMenu(t, tt, "Desk", "Scribe")
	ed := d.find(appScribe)
	x, y := center(t, tt, "Document")
	tt.ClickAt(x, y)
	tt.Type("Recipes\nSoup of the day")
	tt.Key(ui.Cmd, ui.KeyUp)
	for range len("Recipes") {
		tt.Key(ui.Shift, ui.KeyRight)
	}
	tt.Frame()
	mx, my := center(t, tt, "Size")
	tt.ClickAt(mx, my)
	tt.Frame()
	ix, iy := center(t, tt, "22 Point")
	tt.ClickAt(ix, iy)
	tt.Frame()
	tt.Frame()
	if got := ed.doc.doc.Size(); got != 2 {
		t.Fatalf("the heading is at scale %d", got)
	}
	// The menu again, checking the size chosen.
	tt.ClickAt(mx, my)
	tt.Frame()
	golden(t, "scribe-size", tt.Image())
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
}

// TestDesktopPopupWindow docks the disk's window at the bottom of the
// desk, closes it to its tab, opens it again and tears it off.
func TestDesktopPopupWindow(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	doubleClick(t, tt, "Platinum Disk")
	disk := d.front()
	chooseMenu(t, tt, "View", "as Pop-up Window")
	if !disk.popup || !disk.dock.Open {
		t.Fatal("the window did not dock")
	}
	golden(t, "desk-popup", tt.Image())
	// The middle of its tab, which stands on the bottom of the desk.
	tab := func() (float32, float32) {
		h := pt.PopupTabHeight
		if disk.dock.Open {
			h += disk.dock.H - 1
		}
		return float32(disk.dock.X + 40), float32(height - h + pt.PopupTabHeight/2)
	}
	// A click on its tab closes it to the tab, and another opens it.
	for _, open := range []bool{false, true} {
		tt.ClickAt(tab())
		tt.Frame()
		if disk.dock.Open != open {
			t.Fatalf("a click on the tab left it open %v", disk.dock.Open)
		}
	}
	// Dragging the tab up tears it off, a window again.
	x, y := tab()
	tt.Press(x, y)
	tt.Move(x, y-20)
	tt.Move(x, y-60)
	tt.Release(x, y-60)
	tt.Frame()
	tt.Frame()
	if disk.popup {
		t.Fatal("dragging the tab up did not tear the window off")
	}
}

// TestDesktopFocusNamed tabs round the scene of the desk with each of its
// apps in front, both ways: each stop the keyboard focus makes must have a
// name for assistive technology, and Tab must come back to the first.
func TestDesktopFocusNamed(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	for _, app := range []appKind{appCalculator, appJotter, appScribe, appPuzzle, appClock, appAppearance, appAbout, appLedger, appDateTime, appSound, appMouse, appLanguage, appFind, appInfo} {
		d := scene("")
		if app == appInfo {
			d.getInfo(d.disk) // Get Info is not launched, but asked of an item
		} else {
			d.launch(app)
		}
		tt := ui.NewTester(d.view, width, height)
		tt.Frame()
		for _, mods := range []ui.Modifiers{0, ui.Shift} {
			// Round the desk, back to the first stop: nothing traps Tab.
			first, back := "", false
			for i := 0; i < 80 && !back; i++ {
				tt.Key(mods, ui.KeyTab)
				tt.Frame()
				named := ""
				for _, s := range tt.Texts() {
					if tt.Focused(s) {
						named = s
						break
					}
				}
				if named == "" {
					t.Errorf("%s in front: focus stop %d (shift %v) has no name", d.front().title, i+1, mods != 0)
				}
				if i == 0 {
					first = named
				} else if named == first {
					back = true
				}
			}
			if !back {
				t.Errorf("%s in front: 80 tabs (shift %v) did not come back to %q", d.front().title, mods != 0, first)
			}
		}
	}
}

// TestDesktopRandomInput clicks, double-clicks, drags and presses keys at
// random all over the scene of the desk, with apps opening and closing as
// it goes: nothing may panic.
func TestDesktopRandomInput(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	rng := rand.New(rand.NewSource(1998))
	d := scene("")
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	keys := []ui.Key{ui.KeyTab, ui.KeyEnter, ui.KeyEscape, ui.KeySpace, ui.KeyUp, ui.KeyDown, ui.KeyLeft, ui.KeyRight, ui.KeyBackspace, ui.KeyW, ui.KeyN, ui.KeyO, ui.KeyZ, ui.KeyA, ui.KeyD, ui.KeyY, ui.KeyI, ui.KeyS}
	apps := []appKind{appCalculator, appJotter, appScribe, appPuzzle, appClock, appAppearance, appAbout, appLedger, appDateTime, appSound, appMouse, appLanguage, appFind}
	for i := 0; i < 250; i++ {
		x, y := float32(rng.Intn(width)), float32(rng.Intn(height))
		switch rng.Intn(7) {
		case 0, 1:
			tt.ClickAtWith([]ui.Modifiers{0, ui.Shift, ui.Cmd}[rng.Intn(3)], x, y)
		case 2:
			tt.ClickAt(x, y)
			tt.ClickAt(x, y)
		case 3:
			tt.Press(x, y)
			tt.Move(x+float32(rng.Intn(80)-40), y+float32(rng.Intn(80)-40))
			tt.Release(x+float32(rng.Intn(80)-40), y+float32(rng.Intn(80)-40))
		case 4:
			tt.Key([]ui.Modifiers{0, ui.Shift, ui.Cmd}[rng.Intn(3)], keys[rng.Intn(len(keys))])
		case 5:
			tt.Type("x1")
		case 6:
			d.launch(apps[rng.Intn(len(apps))])
		}
		tt.Frame()
	}
}

// TestDesktopClock clicks the clock in the menu bar, which shows the date,
// and again, the time.
func TestDesktopClock(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	clock := sceneTime.Format("3:04 PM")
	date := sceneTime.Format("Mon, Jan 2, 2006")
	if err := tt.Click(clock); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText(date) {
		t.Fatalf("a click on the clock did not show the date: %q", tt.Texts())
	}
	golden(t, "clock-date", tt.Image())
	if err := tt.Click(date); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText(clock) {
		t.Fatal("a second click did not show the time")
	}
}

// TestDesktopSpringOpen drags a document onto the tab of a closed pop-up
// window, which springs open after a moment, and drops it in the window.
func TestDesktopSpringOpen(t *testing.T) {
	now := sceneTime
	defer pt.SetClock(pt.SetClock(func() time.Time { return now }))
	d := newDesk(func() time.Time { return now })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	var docs *node
	for _, n := range d.disk.children {
		if n.name == "Documents" {
			docs = n
		}
	}
	pop := d.openFolder(docs)
	d.dock(pop)
	pop.dock.X, pop.dock.Open = 500, false
	disk := d.openFolder(d.disk)
	disk.frame = pt.WindowState{X: 20, Y: 20, W: 360, H: 220}
	tt.Frame()
	x0, y0 := center(t, tt, "Read Me")
	x1, y1 := float32(pop.dock.X+40), float32(height-pt.PopupTabHeight/2)
	tt.Press(x0, y0)
	for i := 1; i <= 4; i++ {
		f := float32(i) / 4
		tt.Move(x0+(x1-x0)*f, y0+(y1-y0)*f)
		tt.Frame()
	}
	if pop.dock.Open {
		t.Fatal("the window opened before the drag rested on its tab")
	}
	for i := 0; i < 4; i++ {
		now = now.Add(200 * time.Millisecond)
		tt.Move(x1+float32(i%2), y1)
		tt.Frame()
	}
	if !pop.dock.Open {
		t.Fatal("the window did not spring open under the drag")
	}
	// Into the window, and dropped there.
	y2 := float32(height - pop.dock.H/2)
	tt.Move(x1, y2)
	tt.Frame()
	tt.Release(x1, y2)
	tt.Frame()
	tt.Frame()
	if !docs.has("Read Me") || d.disk.has("Read Me") {
		t.Fatal("Read Me did not move into the pop-up window's folder")
	}
	// Dropped on the tab of a closed one, an item goes in at once.
	pop.dock.Open = false
	tt.Frame()
	x0, y0 = center(t, tt, "System Folder")
	tt.Press(x0, y0)
	for i := 1; i <= 4; i++ {
		f := float32(i) / 4
		tt.Move(x0+(x1-x0)*f, y0+(y1-y0)*f)
		tt.Frame()
	}
	tt.Release(x1, y1)
	tt.Frame()
	tt.Frame()
	if !docs.has("System Folder") || pop.dock.Open {
		t.Fatal("System Folder dropped on the closed tab did not go in, or the window opened")
	}
}

// TestScribeSaveAndOpen saves an untitled document through the save
// dialog, into Documents, and opens it again through the open dialog.
func TestScribeSaveAndOpen(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	chooseMenu(t, tt, "Desk", "Scribe")
	x, y := center(t, tt, "Document")
	tt.ClickAt(x, y)
	tt.Type("Shopping list")
	tt.Key(ui.Cmd, ui.KeyS)
	tt.Frame()
	tt.Frame()
	if d.saving == nil {
		t.Fatal("Command-S on an untitled document did not ask where to save it")
	}
	tt.Type("Groceries")
	tt.Frame()
	golden(t, "scribe-save", tt.Image())
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	tt.Frame()
	docs := d.documents()
	if !docs.has("Groceries") || d.saving != nil {
		t.Fatalf("not saved into Documents, or the dialog stayed")
	}
	if f := d.front(); f == nil || f.title != "Groceries" {
		t.Fatal("the window did not take the document's name")
	}
	// Closed, and opened again from the open dialog.
	tt.Key(ui.Cmd, ui.KeyW)
	tt.Frame()
	chooseMenu(t, tt, "Desk", "Scribe")
	chooseMenu(t, tt, "File", "Open…")
	if !d.opening {
		t.Fatal("File > Open… did not show the open dialog")
	}
	// From the keyboard: into the disk, down to Documents and into it, and
	// Groceries, its first document, opened.
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	tt.Frame()
	if f := d.front(); f == nil || f.doc == nil || f.doc.doc.Text != "Shopping list" {
		t.Fatal("the open dialog did not open the document")
	}
}

// TestDesktopPathMenu Command-clicks the title of a folder's window and
// chooses the disk from its path menu, which opens the disk's window.
func TestDesktopPathMenu(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	w := d.openFolder(d.documents())
	w.frame = pt.WindowState{X: 100, Y: 60, W: 300, H: 200}
	tt.Frame()
	tt.ClickAtWith(ui.Cmd, float32(100+150), float32(pt.MenuBarHeight+60+9))
	tt.Frame()
	if !tt.HasText("Platinum Disk") {
		t.Fatal("no path menu")
	}
	// The disk's item, the second, from the keyboard.
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	for i := 0; i < 3; i++ {
		tt.Frame()
	}
	if f := d.front(); f == nil || f.folder == nil || f.folder.dir != d.disk {
		t.Fatal("choosing the disk from the path menu did not open its window")
	}
}

// TestDesktopSpringLoadedFolder rests a drag on a folder's icon: after a
// moment the folder's window springs open under the drag.
func TestDesktopSpringLoadedFolder(t *testing.T) {
	now := sceneTime
	defer pt.SetClock(pt.SetClock(func() time.Time { return now }))
	d := newDesk(func() time.Time { return now })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	disk := d.openFolder(d.disk)
	disk.frame = pt.WindowState{X: 20, Y: 20, W: 360, H: 220}
	tt.Frame()
	x0, y0 := center(t, tt, "Read Me")
	x1, y1 := center(t, tt, "Documents")
	tt.Press(x0, y0)
	for i := 1; i <= 4; i++ {
		f := float32(i) / 4
		tt.Move(x0+(x1-x0)*f, y0+(y1-y0)*f)
		tt.Frame()
	}
	// The folder under the drag, opened: a second window.
	opened := func() bool { return len(d.windows) > 1 }
	if opened() {
		t.Fatal("the folder sprang open at once")
	}
	for i := 0; i < 8; i++ {
		now = now.Add(200 * time.Millisecond)
		tt.Move(x1+float32(i%2), y1)
		tt.Frame()
	}
	tt.Frame()
	if !opened() {
		t.Fatal("the folder did not spring open under the drag")
	}
	tt.Release(x1, y1)
	tt.Frame()
}

// TestDesktopDropOnDisk drags a document from a folder onto the disk's
// icon on the desktop, which puts it at the top of the disk.
func TestDesktopDropOnDisk(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	docs := d.openFolder(d.documents())
	docs.frame = pt.WindowState{X: 20, Y: 20, W: 360, H: 220}
	tt.Frame()
	x0, y0 := center(t, tt, "Letter")
	x1, y1 := center(t, tt, "Platinum Disk") // the desktop's icon, the first
	tt.Press(x0, y0)
	for i := 1; i <= 4; i++ {
		f := float32(i) / 4
		tt.Move(x0+(x1-x0)*f, y0+(y1-y0)*f)
		tt.Frame()
	}
	tt.Release(x1, y1)
	tt.Frame()
	tt.Frame()
	if !d.disk.has("Letter") || d.documents().has("Letter") {
		t.Fatal("Letter dropped on the disk's icon did not go to the top of the disk")
	}
}

// TestDesktopPutAway moves a document from Documents to the Wastebasket,
// opens the Wastebasket and puts the document away, back in Documents.
func TestDesktopPutAway(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	docs := d.documents()
	w := d.openFolder(docs)
	w.frame = pt.WindowState{X: 20, Y: 20, W: 360, H: 220}
	tt.Frame()
	x, y := center(t, tt, "Letter")
	tt.ClickAt(x, y)
	tt.Key(ui.Cmd, ui.KeyBackspace)
	tt.Frame()
	tt.Frame()
	if docs.has("Letter") || len(d.trash) != 1 {
		t.Fatal("Letter did not go to the Wastebasket")
	}
	d.openTrash()
	tt.Frame()
	x, y = center(t, tt, "Letter")
	tt.ClickAt(x, y)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyY)
	tt.Frame()
	tt.Frame()
	if !docs.has("Letter") || len(d.trash) != 0 {
		t.Fatal("Put Away did not put Letter back in Documents")
	}
	if tt.HasText("Letter") && d.front().folder.dir.has("Letter") {
		t.Fatal("the Wastebasket's window still shows Letter")
	}
}

// TestDesktopGetInfo chooses a document and Get Info, which opens a
// window telling about it; again brings the same window forward.
func TestDesktopGetInfo(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	w := d.openFolder(d.disk)
	w.frame = pt.WindowState{X: 20, Y: 20, W: 360, H: 220}
	tt.Frame()
	x, y := center(t, tt, "Read Me")
	tt.ClickAt(x, y)
	tt.Key(ui.Cmd, ui.KeyI)
	tt.Frame()
	tt.Frame()
	f := d.front()
	if f == nil || f.info == nil || f.info.name != "Read Me" || !tt.HasText("text document") {
		t.Fatal("Get Info did not show the document's info")
	}
	f.frame.X, f.frame.Y = 420, 60
	tt.Frame()
	golden(t, "get-info", tt.Image())
	n := len(d.windows)
	d.raise(w)
	tt.Frame()
	x, y = center(t, tt, "Read Me")
	tt.ClickAt(x, y)
	tt.Key(ui.Cmd, ui.KeyI)
	tt.Frame()
	if len(d.windows) != n || d.front().info == nil {
		t.Fatal("Get Info again did not bring the same window forward")
	}
	// The disk chosen on the desktop tells about itself too.
	x, y = center(t, tt, "Platinum Disk") // the desktop's icon, the first
	tt.ClickAt(x, y)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyI)
	tt.Frame()
	tt.Frame()
	if f := d.front(); f == nil || f.info != d.disk {
		t.Fatal("Get Info on the disk's icon did not tell about the disk")
	}
}

// TestDesktopDuplicate duplicates a folder twice: copies of it and all it
// holds, named "copy" and "copy 2".
func TestDesktopDuplicate(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	w := d.openFolder(d.disk)
	w.frame = pt.WindowState{X: 20, Y: 20, W: 360, H: 300}
	tt.Frame()
	tt.Frame()
	// Documents chosen once (a second click there would be a double
	// click, opening it), then duplicated twice.
	x, y := center(t, tt, "Documents")
	tt.ClickAt(x, y)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyD)
	tt.Frame()
	tt.Frame()
	chooseMenu(t, tt, "File", "Duplicate")
	tt.Frame()
	tt.Frame()
	if !d.disk.has("Documents copy") || !d.disk.has("Documents copy 2") {
		t.Fatal("Duplicate did not make the copies")
	}
	for _, c := range d.disk.children {
		if c.name == "Documents copy" && (!c.has("Letter") || c.children[0].parent != c) {
			t.Fatal("the copy does not hold copies of what the folder holds")
		}
	}
}

// TestDesktopClickAfterResize clicks an icon in the frame right after its
// window is resized: the icons are laid out at the new width already, so
// the click chooses the icon and moves nothing.
func TestDesktopClickAfterResize(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	w := d.openFolder(d.disk)
	w.frame = pt.WindowState{X: 20, Y: 20, W: 360, H: 300}
	tt.Frame()
	x, y := center(t, tt, "Documents")
	tt.ClickAt(x, y)
	tt.Frame()
	if !d.disk.has("Documents") || w.folder.selected() == nil || w.folder.selected().name != "Documents" {
		t.Fatal("a click after a resize did not just choose the icon")
	}
}

// TestScribeSaveAs saves a document that has a file under a new name, with
// Save As…, beside the first.
func TestScribeSaveAs(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	var letter *node
	for _, n := range d.documents().children {
		if n.name == "Letter" {
			letter = n
		}
	}
	d.openNode(letter)
	tt.Frame()
	tt.Key(ui.Cmd|ui.Shift, ui.KeyS)
	tt.Frame()
	tt.Frame()
	if d.saving == nil || d.saveDlg.Name != "Letter" {
		t.Fatalf("Save As… did not ask, offering the name (%q)", d.saveDlg.Name)
	}
	tt.Type("Letter 2")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	tt.Frame()
	if !d.documents().has("Letter 2") || !d.documents().has("Letter") {
		t.Fatal("Save As… did not save a second document beside the first")
	}
	if f := d.front(); f == nil || f.title != "Letter 2" {
		t.Fatal("the window did not take the new name")
	}
}

// TestScribeAskToSave closes a document with changes: an alert asks
// whether to save them; Cancel keeps the window, Don't Save closes it, and
// Save writes the document first.
func TestScribeAskToSave(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	var letter *node
	for _, n := range d.documents().children {
		if n.name == "Letter" {
			letter = n
		}
	}
	d.openNode(letter)
	tt.Frame()
	w := d.front()
	x, y := center(t, tt, "Document")
	tt.ClickAt(x, y)
	tt.Key(ui.Cmd, ui.KeyUp)
	tt.Type("P.S. ")
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyW)
	tt.Frame()
	tt.Frame()
	if d.asking != w || !tt.HasText("Don’t Save") {
		t.Fatal("closing a changed document did not ask to save it")
	}
	golden(t, "scribe-ask-save", tt.Image())
	tt.Key(0, ui.KeyEscape) // Cancel
	tt.Frame()
	tt.Frame()
	if d.asking != nil || d.front() != w {
		t.Fatal("Cancel did not keep the window")
	}
	// Save writes it, and closes the window.
	tt.Key(ui.Cmd, ui.KeyW)
	tt.Frame()
	tt.Key(0, ui.KeyEnter) // Save, the default
	tt.Frame()
	tt.Frame()
	tt.Frame()
	if !strings.HasPrefix(letter.text, "P.S. ") || d.find(appScribe) != nil {
		t.Fatalf("Save did not write the letter (%q) and close", letter.text[:10])
	}
	// Unchanged, a document closes without asking.
	d.openNode(letter)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyW)
	tt.Frame()
	tt.Frame()
	if d.asking != nil || d.find(appScribe) != nil {
		t.Fatal("an unchanged document asked, or did not close")
	}
}

// TestQuitAsksToSave quits with a changed document open: the desk asks
// about it rather than quitting.
func TestQuitAsksToSave(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	quit := false
	d.quit = func() { quit = true }
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	chooseMenu(t, tt, "Desk", "Scribe")
	x, y := center(t, tt, "Document")
	tt.ClickAt(x, y)
	tt.Type("Unsaved")
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyQ)
	tt.Frame()
	tt.Frame()
	if quit || d.asking == nil {
		t.Fatalf("quit %v with a changed document, asking %v", quit, d.asking != nil)
	}
}

// TestDesktopOptionDragCopies drags a document onto a folder holding
// Option: a copy goes into the folder, and the document stays.
func TestDesktopOptionDragCopies(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	w := d.openFolder(d.disk)
	w.frame = pt.WindowState{X: 20, Y: 20, W: 360, H: 220}
	tt.Frame()
	x0, y0 := center(t, tt, "Read Me")
	x1, y1 := center(t, tt, "Documents")
	tt.Press(x0, y0)
	for i := 1; i <= 4; i++ {
		f := float32(i) / 4
		tt.Move(x0+(x1-x0)*f, y0+(y1-y0)*f)
		tt.Frame()
	}
	tt.HoldModifiers(ui.Alt)
	tt.Frame()
	tt.Release(x1, y1)
	tt.Frame()
	tt.Frame()
	tt.HoldModifiers(0)
	if !d.disk.has("Read Me") || !d.documents().has("Read Me") {
		t.Fatal("an Option-drag did not copy the document into the folder")
	}
}

// TestCalculatorCopyPaste copies the Calculator's display, and pastes a
// number into it.
func TestCalculatorCopyPaste(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	k := d.launch(appCalculator)
	tt.Frame()
	for _, key := range []string{"6", "×", "7", "="} {
		k.calc.press(key)
	}
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyC)
	tt.Frame()
	if got := tt.Clipboard(); got != "42" {
		t.Fatalf("copied %q", got)
	}
	k.calc.press("C")
	tt.SetClipboard("1,234.5")
	tt.Key(ui.Cmd, ui.KeyV)
	tt.Frame()
	if k.calc.shown != "1234.5" {
		t.Fatalf("pasting made the display %q", k.calc.shown)
	}
}

// TestDesktopSortList sorts a folder's list view from View > Sort List.
func TestDesktopSortList(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	w := d.openFolder(d.disk)
	w.frame = pt.WindowState{X: 20, Y: 20, W: 360, H: 220}
	tt.Frame()
	chooseMenu(t, tt, "View", "as List")
	x, y := center(t, tt, "View")
	tt.ClickAt(x, y)
	tt.Frame()
	sx, sy := center(t, tt, "Sort List")
	tt.Move(sx, sy)
	tt.Frame()
	tt.Frame()
	kx, ky := center(t, tt, "by Kind")
	tt.ClickAt(kx, ky)
	tt.Frame()
	tt.Frame()
	if w.folder.table.SortColumn != 2 {
		t.Fatalf("sorted by column %d", w.folder.table.SortColumn)
	}
}

// TestDesktopAsButtons shows a folder as buttons, and opens one with a
// single click.
func TestDesktopAsButtons(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	d := newDesk(func() time.Time { return sceneTime })
	tt := ui.NewTester(d.view, width, height)
	tt.Frame()
	w := d.openFolder(d.disk)
	w.frame = pt.WindowState{X: 20, Y: 20, W: 380, H: 240}
	tt.Frame()
	chooseMenu(t, tt, "View", "as Buttons")
	if !w.folder.buttons {
		t.Fatal("View > as Buttons did not show buttons")
	}
	golden(t, "cabinet-buttons", tt.Image())
	x, y := center(t, tt, "Documents")
	tt.ClickAt(x, y)
	tt.Frame()
	tt.Frame()
	if f := d.front(); f == nil || f.folder == nil || f.folder.dir != d.documents() {
		t.Fatal("a click on a button did not open its folder")
	}
}
