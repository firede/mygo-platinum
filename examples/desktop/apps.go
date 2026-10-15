package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
	"unicode"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
	"github.com/firede/mygo-platinum/examples/desktop/locales"
)

// jotter is a pad of notes, a page at a time.
type jotter struct {
	pages []string
	page  int
}

const jotterPages = 8

// newJotterWindow opens the pad, its first note written in the desk's
// language: the user's words from then on.
func newJotterWindow(d *desk) *win {
	j := &jotter{pages: make([]string, jotterPages)}
	j.pages[0] = d.say().JotterNotes()
	w := &win{app: appJotter, jot: j}
	w.frame = pt.WindowState{W: 240, H: 220}
	w.content = jotterView
	return w
}

func (j *jotter) turn(d int) {
	j.page = (j.page + d + len(j.pages)) % len(j.pages)
}

var dogEar = []string{
	"#...............",
	"##..............",
	"#:#.............",
	"#::#............",
	"#:::#...........",
	"#::::#..........",
	"#:::::#.........",
	"#::::::#........",
	"#:::::::#.......",
	"#::::::::#......",
	"#:::::::::#.....",
	"#::::::::::#....",
	"#:::::::::::#...",
	"#::::::::::::#..",
	"#:::::::::::::#.",
	"################",
}

func jotterView(c *ui.Context, w *win) {
	m := w.desk.say()
	j := w.jot
	th := pt.ThemeOf(c)
	ui.Column(c).Fill().AlignItems(ui.Stretch).Background(ui.RGB(0xFF, 0xF6, 0xC8)).Children(func() {
		pt.TextArea(c.Key(fmt.Sprint("page", j.page)), &j.pages[j.page]).Label(m.Page()).Element().Grow(1).FillWidth()
		ui.Row(c).Height(20).Padding(0, 8, 0, 4).Gap(8).AlignItems(ui.Center).Children(func() {
			corner := ui.Box(c).Size(16, 16).Role(ui.RoleButton).Label(m.TurnPage()).Focusable().FocusRing(false)
			if corner.Clicked() {
				if corner.ClickModifiers()&ui.Shift != 0 {
					j.turn(-1)
				} else {
					j.turn(1)
				}
			}
			pressed := corner.Pressed()
			corner.Draw(func(p *ui.Painter, r ui.Rect) {
				pn := pt.NewPen(p, r)
				for y, row := range dogEar {
					for x := 0; x < len(row); x++ {
						switch row[x] {
						case '#':
							pn.Dot(x, y, pt.Black)
						case ':':
							if pressed {
								pn.Dot(x, y, th.Accent.Mid)
							} else {
								pn.Dot(x, y, ui.RGB(0xEE, 0xDD, 0x99))
							}
						}
					}
				}
			})
			pt.TextIn(c, m.PageOf(j.page+1, len(j.pages)), pt.AppFont, pt.Black)
		})
	})
}

// scribe is a plain text editor of one document.
type scribe struct {
	file    *node
	doc     *pt.RichText
	n       int // the number of an untitled document
	ruler   pt.RulerState
	noRuler bool
	find    pt.FindState
	finding bool
	// saved is the document as last saved or opened, to tell whether it
	// has changed since; closing is set when a save is to close it.
	saved   string
	closing bool
}

// fingerprint is the document's text, styles and alignments, to compare.
func (s *scribe) fingerprint() string {
	return fmt.Sprint(s.doc.Runs(), s.doc.Aligns())
}

// changed reports whether the document changed since it was last saved or
// opened.
func (s *scribe) changed() bool { return s.fingerprint() != s.saved }

func newScribeWindow(d *desk, f *node) *win {
	s := &scribe{file: f, doc: pt.NewRichText("")}
	title := d.say().Untitled()
	if f != nil {
		s.doc = pt.NewRichText(f.text)
		if f.runs != nil {
			s.doc.SetRuns(f.runs)
			s.doc.SetAligns(f.aligns)
		}
		title = f.name
	} else {
		d.untitled++
		s.n = d.untitled
		if s.n > 1 {
			title = d.say().UntitledN(s.n)
		}
	}
	s.saved = s.fingerprint()
	w := &win{app: appScribe, title: title, doc: s}
	w.frame = pt.WindowState{W: 420, H: 300}
	w.content = scribeView
	return w
}

// save writes the document back to its file; an untitled document asks
// where first, in the save dialog.
func (s *scribe) save(d *desk) {
	if s.file == nil {
		d.saving = s
		d.saveDlg.Name = d.say().Untitled()
		if s.n > 1 {
			d.saveDlg.Name = d.say().UntitledN(s.n)
		}
		if len(d.saveDlg.Path) == 0 {
			d.saveDlg.Path = d.treePath(d.documents())
		}
		return
	}
	s.write(d)
}

// saveAsDialog asks where to save the document under a new name, in the
// folder its file is in, its name offered.
func (s *scribe) saveAsDialog(d *desk) {
	if s.file == nil {
		s.save(d)
		return
	}
	d.saving = s
	d.saveDlg.Name = s.file.name
	if s.file.parent != nil {
		d.saveDlg.Path = d.treePath(s.file.parent)
	}
}

// saveAs makes the document a file named name in dir, replacing one of the
// name there, and writes it.
func (s *scribe) saveAs(d *desk, dir *node, name string) {
	for _, c := range dir.children {
		if c.name == name && c.kind == kindDocument {
			c.remove()
		}
	}
	s.file = dir.add(&node{name: name, kind: kindDocument})
	for _, w := range d.windows {
		if w.doc == s {
			w.title = name
		}
	}
	s.write(d)
}

func (s *scribe) write(d *desk) {
	s.file.text = s.doc.Text
	s.file.runs, s.file.aligns = s.doc.Runs(), s.doc.Aligns()
	s.file.modified = d.now()
	s.saved = s.fingerprint()
	if s.closing {
		// Saved on its way to closing: close it.
		for _, w := range d.windows {
			if w.doc == s {
				d.close(w)
				break
			}
		}
	}
}

func scribeView(c *ui.Context, w *win) {
	m := w.desk.say()
	s := w.doc
	showRuler := !s.noRuler
	ui.Column(c).Fill().AlignItems(ui.Stretch).Children(func() {
		items := []pt.ToolItem{
			{Label: m.Save(), Icon: iconSave, IconOnly: true, Action: func() { s.save(w.desk) }},
			{Separator: true},
			{Label: m.Find(), Icon: iconFind, IconOnly: true, Action: func() { s.finding = true }},
			{Label: m.Ruler(), Icon: iconRuler, IconOnly: true, On: &showRuler},
			{Separator: true},
		}
		pt.Toolbar(c, append(items, pt.FormatItems(s.doc)...))
		s.noRuler = !showRuler
		if s.ruler.Right == 0 {
			s.ruler = pt.RulerState{Left: 8, Right: max(w.frame.W-60, 120)}
		}
		if !s.noRuler {
			pt.Ruler(c, &s.ruler)
		}
		// The document fills the window, its scroll bar against the
		// window's right side, down to the row of the resize corner.
		area := pt.RichTextArea(c, s.doc).Label(m.Document())
		left, wrap := 8, 0
		if !s.noRuler {
			left, wrap = s.ruler.Left, s.ruler.Right-s.ruler.Left
		}
		area.Document(left, wrap).Element().Grow(1).FillWidth()
		if !s.noRuler {
			area.Indent(s.ruler.First)
		}
		ui.Row(c).Height(pt.GrowBoxSize).AlignItems(ui.Stretch).Children(func() {
			pt.Placard(c, wordCount(m, s.doc.Text)).Grow(1).Height(pt.GrowBoxSize)
			ui.Box(c).Width(pt.GrowBoxSize)
		})
		switch pt.FindReplace(c, &s.finding, &s.find) {
		case pt.FindNext:
			_, from := area.Selection()
			if a, b := pt.Find(s.doc.Text, s.find.Query, from, s.find.Options); a >= 0 {
				area.Select(a, b)
				s.find.Note = ""
			} else {
				s.find.Note = m.NotFound()
			}
		case pt.ReplaceOne:
			a, b := area.Selection()
			if a != b && strings.EqualFold(string([]rune(s.doc.Text)[a:b]), s.find.Query) {
				r := []rune(s.doc.Text)
				s.doc.Text = string(r[:a]) + s.find.With + string(r[b:])
			}
			if a, b := pt.Find(s.doc.Text, s.find.Query, a, s.find.Options); a >= 0 {
				area.Select(a, b)
			}
		case pt.ReplaceEvery:
			var n int
			s.doc.Text, n = pt.ReplaceAll(s.doc.Text, s.find.Query, s.find.With, s.find.Options)
			s.find.Note = m.Replaced(n)
		}
	})
}

// wordCount says how many words and lines text has, in the language of m:
// words between spaces, and each Han character and kana a word of its own,
// as Chinese and Japanese write no spaces.
func wordCount(m locales.Messages, text string) string {
	words, lines := 0, 1
	inWord := false
	for _, r := range text {
		if r == '\n' {
			lines++
		}
		switch {
		case unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana):
			inWord = false
			words++
		case r == ' ' || r == '\n' || r == '\t' || unicode.In(r, unicode.Punct) && r > 0x2FFF:
			inWord = false
		case !inWord:
			inWord = true
			words++
		}
	}
	return m.WordsLines(words, lines)
}

// The 16×16 icons of Scribe's toolbar.
var (
	iconSave = []string{
		"................",
		".#############..",
		".#:#wwwwwww#:#..",
		".#:#wwwwwww#:#..",
		".#:#wwwwwww#:##.",
		".#:#########::#.",
		".#::::::::::::#.",
		".#::#######:::#.",
		".#::#ww##ww#::#.",
		".#::#ww##ww#::#.",
		".#::#ww##ww#::#.",
		".##############.",
		"................",
		"................",
		"................",
		"................",
	}
	iconFind = []string{
		"................",
		".....####.......",
		"...##wwww##.....",
		"..#ww....ww#....",
		"..#w......w#....",
		".#w........w#...",
		".#w........w#...",
		".#w........w#...",
		"..#w......w#....",
		"..#ww....ww#....",
		"...##wwww###....",
		".....####.###...",
		"...........###..",
		"............###.",
		".............##.",
		"................",
	}
	iconRuler = []string{
		"................",
		"................",
		"................",
		"................",
		"................",
		"################",
		"#ww#ww#ww#ww#ww#",
		"#ww#ww#ww#ww#ww#",
		"#wwwww#wwwww#ww#",
		"#wwwwwwwwwwwwww#",
		"################",
		"................",
		"................",
		"................",
		"................",
		"................",
	}
)

// newLedgerWindow opens a sheet of figures, its labels written in the
// desk's language: the user's from then on.
func newLedgerWindow(d *desk) *win {
	m := d.say()
	st := pt.NewSheetState()
	for r, row := range [][]string{
		{m.Item(), m.Cost()}, {m.Rent(), "800"}, {m.Groceries(), "245.5"}, {m.LampOil(), "12"}, {m.Books(), "38"},
		{"", ""}, {m.Total(), "=SUM(B2:B5)"}, {m.Average(), "=AVERAGE(B2:B5)"},
	} {
		for col, v := range row {
			if v != "" {
				st.Cells[pt.Cell{Col: col, Row: r}] = v
			}
		}
	}
	st.Cursor, st.Anchor = pt.Cell{Col: 1, Row: 6}, pt.Cell{Col: 1, Row: 6}
	w := &win{app: appLedger, sheet: st}
	w.frame = pt.WindowState{W: 360, H: 240}
	w.content = func(c *ui.Context, w *win) {
		ui.Column(c).Fill().AlignItems(ui.Stretch).Children(func() {
			pt.FormulaBar(c, w.sheet)
			pt.Sheet(c, w.sheet, 99, 26).Element().Grow(1).FillWidth().Margin(0, 0, pt.GrowBoxSize-2, 0)
		})
	}
	return w
}

// puzzle is the fifteen puzzle: tiles 1 to 15 in a 4×4 frame with one
// hole, slid into order.
type puzzle struct {
	tiles [16]int // 0 is the hole
	moves int
}

func newPuzzleWindow() *win {
	pz := &puzzle{}
	pz.reset()
	pz.shuffle(1)
	w := &win{app: appPuzzle, puzzle: pz}
	w.frame = pt.WindowState{W: 184, H: 214}
	w.opts = pt.WindowOptions{NoGrow: true, NoZoom: true}
	w.content = puzzleView
	return w
}

func (pz *puzzle) reset() {
	for i := range pz.tiles {
		pz.tiles[i] = (i + 1) % 16
	}
	pz.moves = 0
}

func (pz *puzzle) hole() int {
	for i, t := range pz.tiles {
		if t == 0 {
			return i
		}
	}
	return -1
}

// slide moves tile i into the hole, if it is next to it.
func (pz *puzzle) slide(i int) bool {
	h := pz.hole()
	dx, dy := i%4-h%4, i/4-h/4
	if dx*dx+dy*dy != 1 {
		return false
	}
	pz.tiles[h], pz.tiles[i] = pz.tiles[i], 0
	pz.moves++
	return true
}

// shuffle makes random moves from the solved order, so the puzzle stays
// solvable.
func (pz *puzzle) shuffle(seed int64) {
	pz.reset()
	r := rand.New(rand.NewSource(seed))
	for n := 0; n < 200; n++ {
		h := pz.hole()
		var next []int
		for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			x, y := h%4+d[0], h/4+d[1]
			if x >= 0 && x < 4 && y >= 0 && y < 4 {
				next = append(next, y*4+x)
			}
		}
		pz.slide(next[r.Intn(len(next))])
	}
	pz.moves = 0
}

func (pz *puzzle) solved() bool {
	for i, t := range pz.tiles {
		if t != (i+1)%16 {
			return false
		}
	}
	return true
}

const tileSize = 36

func puzzleView(c *ui.Context, w *win) {
	m := w.desk.say()
	pz := w.puzzle
	th := pt.ThemeOf(c)
	ui.Column(c).Fill().Padding(10).Gap(6).AlignItems(ui.Center).Children(func() {
		board := ui.Box(c).Size(4*tileSize+4, 4*tileSize+4).Focusable().FocusRing(false).Role(ui.RoleGroup).Label(m.Tiles())
		// The arrows slide the tile on that side of the hole into it.
		h := pz.hole()
		for _, k := range []struct {
			key ui.Key
			dx  int
			dy  int
		}{{ui.KeyLeft, 1, 0}, {ui.KeyRight, -1, 0}, {ui.KeyUp, 0, 1}, {ui.KeyDown, 0, -1}} {
			if board.Shortcut(0, k.key) || w.active && c.Shortcut(0, k.key) {
				x, y := h%4+k.dx, h/4+k.dy
				if x >= 0 && x < 4 && y >= 0 && y < 4 {
					pz.slide(y*4 + x)
				}
			}
		}
		board.HandleInput(func(ev ui.InputEvent) bool {
			if ev.Kind == ui.InputPointerDown && ev.Button == 0 {
				x, y := int(ev.X-2)/tileSize, int(ev.Y-2)/tileSize
				if x >= 0 && x < 4 && y >= 0 && y < 4 {
					pz.slide(y*4 + x)
				}
				return false
			}
			return false
		})
		tiles := pz.tiles
		board.Draw(func(p *ui.Painter, r ui.Rect) {
			pn := pt.NewPen(p, r)
			pn.Bevel(0, 0, pn.W, pn.H, pt.Gray7, pt.White)
			pn.Frame(1, 1, pn.W-2, pn.H-2, pt.Black)
			pn.Rect(2, 2, pn.W-4, pn.H-4, pt.Gray5)
			for i, t := range tiles {
				if t == 0 {
					continue
				}
				x, y := 2+(i%4)*tileSize, 2+(i/4)*tileSize
				tp := pn.Sub(x, y, tileSize, tileSize)
				face := pt.Gray2
				if t == i+1 {
					face = th.Accent.Light // in its place
				}
				tp.Rect(0, 0, tileSize, tileSize, face)
				tp.Frame(0, 0, tileSize, tileSize, pt.Black)
				tp.Bevel(1, 1, tileSize-2, tileSize-2, pt.White, pt.Gray7)
				tp.TextCentered(0, 0, tileSize, tileSize, fmt.Sprint(t), pt.SystemFont, pt.Black)
			}
		})
		status := m.Moves(pz.moves)
		if pz.solved() {
			status = m.Solved()
		}
		pt.TextIn(c, status, pt.SystemFont, pt.Black)
	})
}

// alarmClock shows the time and rings an alarm.
type alarmClock struct {
	hour, minute int
	armed        bool
}

func newClockWindow(d *desk) *win {
	now := d.now()
	w := &win{app: appClock, clock: &alarmClock{hour: now.Hour(), minute: (now.Minute()/5*5 + 5) % 60}}
	w.frame = pt.WindowState{W: 220}
	w.opts = pt.WindowOptions{NoGrow: true, NoZoom: true}
	w.content = func(c *ui.Context, w *win) { clockView(c, w, d) }
	return w
}

func clockView(c *ui.Context, w *win, d *desk) {
	m := d.say()
	k := w.clock
	now := d.now()
	c.After(time.Second - time.Duration(now.Nanosecond()))
	ui.Column(c).FillWidth().Padding(10).Gap(8).AlignItems(ui.Stretch).Children(func() {
		t := now.Format("15:04:05")
		date := i18n.FormatDate(d.lang, now, i18n.DateOptions{Date: i18n.FullStyle})
		ui.Box(c).Height(40).Role(ui.RoleStatus).Label(m.Time()).Value(t).Draw(func(p *ui.Painter, r ui.Rect) {
			pn := pt.NewPen(p, r)
			// A sunken gray face with black figures.
			pt.DrawContentRing(pn)
			pn.Frame(1, 1, pn.W-2, pn.H-2, pt.Black)
			pn.Rect(2, 2, pn.W-4, pn.H-4, pt.Gray1)
			tw := pt.SystemFont.Measure(t) * 2
			pn.TextScaled((pn.W-tw)/2, (pn.H-18)/2, t, pt.SystemFont, pt.Black, 2)
		})
		ui.Row(c).Justify(ui.Center).Children(func() { pt.TextIn(c, date, pt.AppFont, pt.Black) })
		pt.Separator(c)
		ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
			pt.Checkbox(c, &k.armed, m.AlarmAt())
			hours := make([]string, 24)
			for i := range hours {
				hours[i] = fmt.Sprintf("%02d", i)
			}
			mins := make([]string, 12)
			for i := range mins {
				mins[i] = fmt.Sprintf("%02d", i*5)
			}
			m := k.minute / 5
			pt.PopupMenu(c, &k.hour, "", hours)
			pt.TextIn(c, ":", pt.SystemFont, pt.Black)
			pt.PopupMenu(c, &m, "", mins)
			k.minute = m * 5
		})
	})
}

// appearance is the control panel of colors and patterns.
type appearance struct {
	tab     int
	accent  string
	pattern int
}

func newAppearanceWindow(d *desk) *win {
	a := &appearance{accent: d.theme.Accent.Name, pattern: d.pattern}
	w := &win{app: appAppearance, appear: a}
	w.frame = pt.WindowState{W: 300, H: 250}
	w.opts = pt.WindowOptions{NoGrow: true, NoZoom: true}
	w.content = func(c *ui.Context, w *win) { appearanceView(c, w, d) }
	return w
}

func appearanceView(c *ui.Context, w *win, d *desk) {
	m := d.say()
	a := w.appear
	ui.Column(c).Fill().Padding(10).AlignItems(ui.Stretch).Children(func() {
		pt.TabControl(c, &a.tab, []string{m.Colors(), m.DesktopTab()}, func(i int) {
			switch i {
			case 0:
				pt.Label(c, m.HighlightColor())
				pt.RadioGroup(c, func() {
					for _, acc := range pt.Accents() {
						ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
							swatch := acc
							ui.Box(c).Size(24, 12).Draw(func(p *ui.Painter, r ui.Rect) {
								pn := pt.NewPen(p, r)
								pn.Frame(0, 0, pn.W, pn.H, pt.Black)
								pn.Rect(1, 1, 11, pn.H-2, swatch.Dark)
								pn.Rect(12, 1, 11, pn.H-2, swatch.Light)
							})
							if pt.Radio(c, &a.accent, acc.Name, accentName(m, acc.Name)).Changed() {
								d.theme.Accent = acc
							}
						})
					}
				})
			case 1:
				names := make([]string, len(patterns))
				for i, p := range patterns {
					names[i] = locales.Catalog.Format(d.lang, p.key, nil)
				}
				ui.Row(c).Gap(10).Children(func() {
					if pt.ListBox(c, &a.pattern, names).Changed() {
						d.pattern = a.pattern
					}
					pat := patterns[a.pattern]
					ui.Box(c).Size(64, 64).Role(ui.RoleImage).Label(m.Preview()).Draw(func(p *ui.Painter, r ui.Rect) {
						pn := pt.NewPen(p, r)
						pn.Frame(0, 0, pn.W, pn.H, pt.Black)
						fg := ui.RGB(pat.fg.R, pat.fg.G, pat.fg.B)
						bg := ui.RGB(pat.bg.R, pat.bg.G, pat.bg.B)
						pn.Sub(1, 1, pn.W-2, pn.H-2).Pattern(0, 0, pn.W-2, pn.H-2, pat.bits, fg, bg)
					})
				})
			}
		}).Element().Grow(1)
	})
}

func newAboutWindow(d *desk) *win {
	w := &win{app: appAbout}
	w.frame = pt.WindowState{W: 330, H: 200}
	w.opts = pt.WindowOptions{NoGrow: true, NoZoom: true}
	w.content = func(c *ui.Context, w *win) { aboutView(c, d) }
	return w
}

func aboutView(c *ui.Context, d *desk) {
	m := d.say()
	ui.Column(c).Fill().Padding(12).Gap(10).AlignItems(ui.Stretch).Children(func() {
		ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
			ui.Box(c).Size(32, 32).Draw(func(p *ui.Painter, r ui.Rect) { iconAbout.draw(p, r.X, r.Y, false) })
			ui.Column(c).Gap(2).Children(func() {
				pt.Label(c, m.PlatinumDesk())
				pt.TextIn(c, m.Version(), pt.AppFont, pt.Black)
			})
		})
		pt.Separator(c)
		const total = 64 << 20
		used := 12 << 20 // the desk itself
		ui.Column(c).Gap(4).AlignItems(ui.Stretch).Children(func() {
			row := func(name string, size int) {
				ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
					ui.Box(c).Width(100).Children(func() { pt.TextIn(c, name, pt.AppFont, pt.Black) })
					ui.Box(c).Width(60).Children(func() { pt.TextIn(c, humanSize(m, size), pt.AppFont, pt.Black) })
					pt.ProgressBar(c, float64(size)/float64(16<<20)).Width(120).Height(10).Label(name)
				})
			}
			row(m.Desk(), used)
			for _, w := range d.windows {
				if w.app == appAbout || w.app == appCabinet {
					continue
				}
				title := d.titleOf(w)
				size := 1<<20 + len(title)*96<<10
				used += size
				row(title, size)
			}
		})
		pt.TextIn(c, m.Memory(humanSize(m, total), humanSize(m, total-used)), pt.AppFont, pt.Black)
	})
}

// accentName returns the name of an accent in the language of m.
func accentName(m locales.Messages, name string) string {
	switch name {
	case pt.Lavender.Name:
		return m.AccentLavender()
	case pt.Graphite.Name:
		return m.AccentGraphite()
	case pt.Harbor.Name:
		return m.AccentHarbor()
	case pt.Moss.Name:
		return m.AccentMoss()
	case pt.Ember.Name:
		return m.AccentEmber()
	case pt.Orchid.Name:
		return m.AccentOrchid()
	}
	return name
}
