package main

import (
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

// appName returns the name of the application a window belongs to, and its
// picture.
func appName(w *win) (string, []string) {
	switch w.app {
	case appCalculator:
		return "Calculator", miniCalc
	case appJotter:
		return "Jotter", miniNote
	case appScribe:
		return "Scribe", miniPage
	case appPuzzle:
		return "Puzzle", miniTiles
	case appClock:
		return "Alarm Clock", miniClock
	case appLedger:
		return "Ledger", miniGrid
	case appAppearance, appDateTime, appSound, appMouse:
		return w.title, miniPanel
	}
	return "Desk", miniDesk // the desk's own windows: folders, About
}

// rightMenus returns the clock and the application menu, at the right of
// the menu bar.
func (d *desk) rightMenus() []pt.Menu {
	name, pic := "Desk", miniDesk
	if f := d.front(); f != nil {
		name, pic = appName(f)
	}
	var apps []pt.MenuItem
	seen := map[string]bool{}
	for i := len(d.windows) - 1; i >= 0; i-- {
		w := d.windows[i]
		n, _ := appName(w)
		if seen[n] {
			continue
		}
		seen[n] = true
		apps = append(apps, pt.MenuItem{Title: n, Checked: n == name, Action: func() {
			w.frame.Collapsed = false
			d.raise(w)
		}})
	}
	if !seen["Desk"] {
		apps = append(apps, pt.MenuItem{Title: "Desk", Checked: name == "Desk"})
	}
	shade := func(hide func(n string) bool) func() {
		return func() {
			for _, w := range d.windows {
				n, _ := appName(w)
				w.frame.Collapsed = hide(n)
			}
		}
	}
	items := append([]pt.MenuItem{
		{Title: "Hide " + name, Action: shade(func(n string) bool { return n == name })},
		{Title: "Hide Others", Action: shade(func(n string) bool { return n != name })},
		{Title: "Show All", Action: shade(func(string) bool { return false })},
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
		return d.now().Format("Mon, Jan 2, 2006")
	}
	return d.now().Format("3:04 PM")
}
