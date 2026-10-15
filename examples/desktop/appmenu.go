package main

import (
	"github.com/egoist/mygo/i18n"
	pt "github.com/firede/mygo-platinum"
)

// The 12×12 pictures of the applications, as the application menu shows
// them, drawn for the desk.
var (
	miniDesk = []string{
		"............",
		"####........",
		"#oo#######..",
		"#oooooooo#..",
		"#o++++++o#..",
		"#o++++++o#..",
		"#o++++++o#..",
		"#o++++++o#..",
		"#oooooooo#..",
		"##########..",
		"............",
		"............",
	}
	miniCalc = []string{
		".##########.",
		".#::::::::#.",
		".#:wwwwww:#.",
		".#::::::::#.",
		".#:w:w:w::#.",
		".#::::::::#.",
		".#:w:w:w::#.",
		".#::::::::#.",
		".#:w:w:w::#.",
		".#::::::::#.",
		".##########.",
		"............",
	}
	miniNote = []string{
		"..#.#.#.#...",
		".#########..",
		".#wwwwwww#..",
		".#w:::::w#..",
		".#wwwwwww#..",
		".#w:::::w#..",
		".#wwwwwww#..",
		".#w:::::w#..",
		".#wwwwwww#..",
		".#########..",
		"............",
		"............",
	}
	miniPage = []string{
		".#######....",
		".#wwwww##...",
		".#wwwww#w#..",
		".#w:::w####.",
		".#wwwwwwww#.",
		".#w::::::w#.",
		".#wwwwwwww#.",
		".#w::::::w#.",
		".#wwwwwwww#.",
		".##########.",
		"............",
		"............",
	}
	miniGrid = []string{
		"############",
		"#oo#ww#ww#w#",
		"############",
		"#ww#ww#ww#w#",
		"############",
		"#ww#ww#ww#w#",
		"############",
		"#ww#ww#ww#w#",
		"############",
		"............",
		"............",
		"............",
	}
	miniTiles = []string{
		"#####.#####.",
		"#+++#.#ooo#.",
		"#+++#.#ooo#.",
		"#####.#####.",
		"............",
		"#####.#####.",
		"#ooo#.#:::#.",
		"#ooo#.#:::#.",
		"#####.#####.",
		"............",
		"............",
		"............",
	}
	miniClock = []string{
		"...######...",
		"..#wwwwww#..",
		".#www#www#..",
		"#wwww#wwww#.",
		"#wwww#wwww#.",
		"#wwww###ww#.",
		"#wwwwwwwww#.",
		"#wwwwwwwww#.",
		".#wwwwwww#..",
		"..#wwwww#...",
		"...#####....",
		"............",
	}
	miniPanel = []string{
		"##########..",
		"#::::::::#..",
		"#:######:#..",
		"#:#oooo#:#..",
		"#:#oooo#:#..",
		"#:######:#..",
		"#::::::::#..",
		"##########..",
		"...####.....",
		"..######....",
		"............",
		"............",
	}
)

// titleOf returns the title of w's window in the desk's language: a
// document's name, a folder's, or the name of what the window shows.
func (d *desk) titleOf(w *win) string {
	m := d.say()
	switch {
	case w.doc != nil:
		return w.title
	case w.folder != nil && w.folder.trash:
		return m.Wastebasket()
	case w.folder != nil:
		return w.folder.dir.name
	case w.info != nil:
		return m.InfoTitle(w.info.name)
	}
	return d.appTitle(w.app)
}

// appTitle returns the name of an application or a window of the desk.
func (d *desk) appTitle(a appKind) string {
	m := d.say()
	switch a {
	case appCalculator:
		return m.Calculator()
	case appJotter:
		return m.Jotter()
	case appScribe:
		return m.Scribe()
	case appPuzzle:
		return m.Puzzle()
	case appClock:
		return m.AlarmClock()
	case appLedger:
		return m.Ledger()
	case appAppearance:
		return m.Appearance()
	case appDateTime:
		return m.DateTime()
	case appSound:
		return m.Sound()
	case appMouse:
		return m.Mouse()
	case appLanguage:
		return m.Language()
	case appAbout:
		return m.AboutThisDesk()
	case appFind:
		return m.FindFile()
	}
	return m.Desk()
}

// appPic returns the small picture of an application.
func appPic(a appKind) []string {
	switch a {
	case appCalculator:
		return miniCalc
	case appJotter:
		return miniNote
	case appScribe:
		return miniPage
	case appPuzzle:
		return miniTiles
	case appClock:
		return miniClock
	case appLedger:
		return miniGrid
	case appAppearance, appDateTime, appSound, appMouse, appLanguage:
		return miniPanel
	}
	return miniDesk
}

// appName returns the name of the application a window belongs to, and its
// picture: the desk's for its own windows, folders and About.
func (d *desk) appName(w *win) (string, []string) {
	switch w.app {
	case appCalculator, appJotter, appScribe, appPuzzle, appClock, appLedger,
		appAppearance, appDateTime, appSound, appMouse, appLanguage:
		return d.appTitle(w.app), appPic(w.app)
	}
	return d.say().Desk(), miniDesk
}

// rightMenus returns the clock and the application menu, at the right of
// the menu bar.
func (d *desk) rightMenus() []pt.Menu {
	m := d.say()
	name, pic := m.Desk(), miniDesk
	if f := d.front(); f != nil {
		name, pic = d.appName(f)
	}
	var apps []pt.MenuItem
	seen := map[string]bool{}
	for i := len(d.windows) - 1; i >= 0; i-- {
		w := d.windows[i]
		n, _ := d.appName(w)
		if seen[n] {
			continue
		}
		seen[n] = true
		apps = append(apps, pt.MenuItem{Title: n, Checked: n == name, Action: func() {
			w.frame.Collapsed = false
			d.raise(w)
		}})
	}
	if !seen[m.Desk()] {
		apps = append(apps, pt.MenuItem{Title: m.Desk(), Checked: name == m.Desk()})
	}
	shade := func(hide func(n string) bool) func() {
		return func() {
			for _, w := range d.windows {
				n, _ := d.appName(w)
				w.frame.Collapsed = hide(n)
			}
		}
	}
	items := append([]pt.MenuItem{
		{Title: m.HideApp(name), Action: shade(func(n string) bool { return n == name })},
		{Title: m.HideOthers(), Action: shade(func(n string) bool { return n != name })},
		{Title: m.ShowAll(), Action: shade(func(string) bool { return false })},
		pt.Divider(),
	}, apps...)
	return []pt.Menu{
		{Title: d.clockTitle(), Right: true, Action: func() { d.showDate = !d.showDate }},
		{Title: name, Icon: pic, IconTitle: true, Right: true, Items: items},
	}
}

// clockTitle is what the clock in the menu bar shows: the time, or the
// date, which a click on it shows instead, as the classic clock did.
func (d *desk) clockTitle() string {
	if d.showDate {
		return i18n.FormatDate(d.formats, d.now(), i18n.DateOptions{Skeleton: "yMMMEd"})
	}
	return i18n.FormatDate(d.formats, d.now(), i18n.DateOptions{Skeleton: "jm"})
}
