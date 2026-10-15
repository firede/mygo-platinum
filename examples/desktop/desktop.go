package main

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
)

// appKind is what a window of the desk shows.
type appKind int

const (
	appCabinet appKind = iota
	appCalculator
	appJotter
	appScribe
	appPuzzle
	appClock
	appAppearance
	appAbout
	appLedger
	appDateTime
	appSound
	appMouse
	appLanguage
	appFind
	appInfo
)

func appIcon(a appKind) *icon {
	switch a {
	case appCalculator:
		return iconCalculator
	case appJotter:
		return iconJotter
	case appScribe:
		return iconScribe
	case appPuzzle:
		return iconPuzzle
	case appClock:
		return iconClock
	case appAppearance:
		return iconAppearance
	case appAbout:
		return iconAbout
	case appLedger:
		return iconLedger
	case appDateTime:
		return iconDateTime
	case appSound:
		return iconSound
	case appMouse:
		return iconMouse
	case appLanguage:
		return iconLanguage
	}
	return iconFolder
}

// win is a window of the desk.
type win struct {
	id    int
	app   appKind
	title string
	frame pt.WindowState
	opts  pt.WindowOptions
	// popup docks the window at the bottom of the desk, where dock places
	// it, as a pop-up window.
	popup bool
	dock  pt.PopupWindowState
	// active is whether the window is in front, as last built.
	active bool

	// What the window shows, by its app.
	folder  *cabinet
	info    *node // what a Get Info window tells about
	calc    *calculator
	jot     *jotter
	doc     *scribe
	puzzle  *puzzle
	clock   *alarmClock
	appear  *appearance
	sheet   *pt.SheetState
	desk    *desk
	content func(c *ui.Context, w *win)
}

// deskIcon is an icon on the desktop.
type deskIcon struct {
	name     string
	icon     *icon
	x, y     int
	open     func()
	help     string // what its help balloon says
	grab     [2]float32
	dragging bool
}

// desk is the whole desktop: the menu bar, the icons, the windows and what
// they show.
type desk struct {
	theme    *pt.Theme
	untitled int  // the untitled documents opened, to number them
	showDate bool // the menu bar's clock shows the date
	// The save dialog, for saving it, and the open dialog.
	saving   *scribe
	asking   *win // a document window being closed with changes unsaved
	saveDlg  pt.FileDialogState
	opening  bool
	openDlg  pt.FileDialogState
	dragging bool         // an item is being dragged, as of the last frame
	dragMods ui.Modifiers // the modifier keys held while it was
	dropMods ui.Modifiers // and in the frame before this one
	dragSeen bool         // and in this one
	// The folder a drag rests on, and since when, to spring it open.
	springNode  *node
	springSince time.Time
	springSeen  bool
	pattern     int
	disk        *node
	trash       []*node
	windows     []*win // back to front
	nextID      int
	icons       []*deskIcon
	selIcon     int
	trashNode   *node // the Wastebasket, as Get Info tells about it
	now         func() time.Time
	size        [2]int // the desktop area, as last laid out
	bg          *ui.Bitmap
	bgFor       [3]int // the pattern and size bg was made for
	emptying    bool   // the alert asking to empty the wastebasket
	ringing     bool   // the alarm going off
	balloons    bool   // help balloons show
	// later holds what a build asked to do to the disk, done before the
	// next build: changing the folders while building them would build
	// some of their items twice.
	later    []func()
	settings settings
	// The zoom rectangles of a window opening from an icon.
	zoomFrom  ui.Rect
	zoomTo    ui.Rect
	zoomStart time.Time
	quit      func()
	// The language of the desk's window, as of this frame, and the one its
	// items were last named in.
	lang       string
	labelled   bool
	labelledIn string
	// The language chosen in the Language control panel, "" to follow the
	// system's, and what makes the app show it.
	chosenLang string
	setLocale  func(tag string)
}

// newDesk returns a desk, its documents written in the language the app
// shows as it starts.
func newDesk(now func() time.Time) *desk { return newDeskIn(now, i18n.Locale()) }

// newDeskIn returns a desk, its documents written in the language of lang.
func newDeskIn(now func() time.Time, lang string) *desk {
	d := &desk{theme: pt.DefaultTheme(), selIcon: -1}
	d.settings = settings{volume: 4, tracking: 3, doubleHit: "medium", alert: 2}
	// The desk's clock is the computer's, moved as Date & Time sets it.
	d.now = func() time.Time { return now().Add(d.settings.offset) }
	d.disk = newDisk(now(), lang)
	d.icons = []*deskIcon{
		{icon: iconDisk, open: func() { d.openFolder(d.disk) }},
		{icon: iconTrash, open: func() { d.openTrash() }},
	}
	return d
}

// front returns the window in front, or nil.
func (d *desk) front() *win {
	if len(d.windows) == 0 {
		return nil
	}
	return d.windows[len(d.windows)-1]
}

func (d *desk) raise(w *win) {
	for i, o := range d.windows {
		if o == w {
			d.windows = append(append(d.windows[:i:i], d.windows[i+1:]...), w)
			return
		}
	}
}

func (d *desk) close(w *win) {
	if w.doc != nil && w.doc.changed() && !w.doc.closing {
		d.asking = w // ask first whether to save the changes
		return
	}
	for i, o := range d.windows {
		if o == w {
			d.windows = append(d.windows[:i:i], d.windows[i+1:]...)
			return
		}
	}
}

// dock makes w a pop-up window at the bottom of the desk, open, where it
// was across.
func (d *desk) dock(w *win) {
	w.popup = true
	w.dock = pt.PopupWindowState{X: w.frame.X, W: w.frame.W, H: min(w.frame.H, 200), Open: true}
}

// undock makes w a window again, standing just above where it was docked.
func (d *desk) undock(w *win) {
	w.popup = false
	w.frame.X, w.frame.W = w.dock.X, w.dock.W
	w.frame.H = max(w.dock.H+pt.TitleBarHeight, w.frame.MinH)
	w.frame.Y = max(d.size[1]-w.frame.H-pt.PopupTabHeight-8, 4)
	w.frame.Collapsed = false
}

// open shows a new window, cascading it from the last.
func (d *desk) open(w *win) *win {
	d.nextID++
	w.id = d.nextID
	w.desk = d
	n := len(d.windows) % 6
	if w.frame.X == 0 && w.frame.Y == 0 {
		w.frame.X, w.frame.Y = 40+n*22, 16+n*22
	}
	d.windows = append(d.windows, w)
	if d.zoomFrom.W > 0 {
		// The window opens from the icon it was opened from.
		d.zoomTo = ui.Rect{X: float32(w.frame.X), Y: float32(w.frame.Y + pt.MenuBarHeight), W: float32(w.frame.W), H: float32(w.frame.H)}
		d.zoomStart = time.Now()
	}
	return w
}

// find returns the window of app, if one is open.
func (d *desk) find(app appKind) *win {
	for _, w := range d.windows {
		if w.app == app {
			return w
		}
	}
	return nil
}

// launch brings an app's window to the front, opening it if need be.
func (d *desk) launch(app appKind) *win {
	if w := d.find(app); w != nil && app != appScribe {
		d.raise(w)
		return w
	}
	switch app {
	case appCalculator:
		return d.open(newCalculatorWindow())
	case appJotter:
		return d.open(newJotterWindow(d))
	case appScribe:
		return d.open(newScribeWindow(d, nil))
	case appPuzzle:
		return d.open(newPuzzleWindow())
	case appClock:
		return d.open(newClockWindow(d))
	case appAppearance:
		return d.open(newAppearanceWindow(d))
	case appAbout:
		return d.open(newAboutWindow(d))
	case appLedger:
		return d.open(newLedgerWindow(d))
	case appDateTime, appSound, appMouse:
		return d.open(newPanelWindow(d, app))
	case appLanguage:
		return d.open(newLanguageWindow(d))
	case appFind:
		return d.open(newFindWindow(d))
	}
	return nil
}

// openNode opens what a node of the disk is.
func (d *desk) openNode(n *node) {
	switch n.kind {
	case kindFolder, kindDisk:
		d.openFolder(n)
	case kindDocument:
		for _, w := range d.windows {
			if w.doc != nil && w.doc.file == n {
				d.raise(w)
				return
			}
		}
		d.open(newScribeWindow(d, n))
	default:
		d.launch(n.app)
	}
}

func (d *desk) view(c *ui.Context) {
	for _, f := range d.later {
		f()
	}
	d.later = d.later[:0]
	d.lang = c.Locale()
	d.followLanguage()
	m := d.say()
	// A drag seen in the last frame lets the windows in the background
	// take its drop.
	d.dragging, d.dragSeen = d.dragSeen, false
	// The modifier keys as the drag went on, the frame before a drop.
	d.dropMods = d.dragMods
	if d.dragging {
		d.dragMods = c.Modifiers()
	}
	if !d.springSeen {
		d.springNode = nil // the drag left the folder it rested on
	}
	d.springSeen = false
	pt.Use(c, d.theme)
	// Command-` brings the window at the back to the front, going round
	// the windows.
	if c.Shortcut(ui.Cmd, ui.KeyBackquote) && len(d.windows) > 1 {
		d.raise(d.windows[0])
	}
	c.Root().Background(pt.Gray5)
	ui.Column(c).Fill().AlignItems(ui.Stretch).Children(func() {
		pt.MenuBar(c, d.menus(c))
		area := ui.Box(c).Grow(1).Clip()
		area.Draw(func(p *ui.Painter, r ui.Rect) {
			d.size = [2]int{int(r.W), int(r.H)}
			d.drawPattern(p, r)
		})
		if area.Pressed() && !area.Hovered() {
			d.selIcon = -1
		}
		area.Children(func() {
			d.iconsView(c)
			for _, w := range d.windows {
				d.windowView(c, w)
			}
		})
	})
	if d.zoomFrom.W > 0 && !pt.ZoomRects(c, d.zoomFrom, d.zoomTo, d.zoomStart) && !d.zoomStart.IsZero() {
		d.zoomFrom, d.zoomStart = ui.Rect{}, time.Time{}
	}
	d.alarmCheck(c)
	d.fileDialogs(c)
	if w := d.asking; w != nil {
		open := true
		switch pt.Alert(c, &open, pt.AlertCaution, m.SaveChanges(w.title), "", m.DontSave(), m.Cancel(), m.Save()) {
		case 0: // close, the changes lost
			w.doc.closing = true
			d.later = append(d.later, func() { d.close(w) })
			c.After(0)
		case 2: // save, then close
			w.doc.closing = true
			d.later = append(d.later, func() { w.doc.save(d) })
			c.After(0)
		}
		if !open {
			d.asking = nil
		}
	}
	if pt.Alert(c, &d.emptying, pt.AlertCaution, m.EmptyWastebasket(len(d.trash)), "", m.Cancel(), m.Ok()) == 1 {
		d.trash = nil
	}
	if pt.Alert(c, &d.ringing, pt.AlertNote, m.ItIsTime(), m.AlarmGoneOff(), m.Ok()) >= 0 {
		if w := d.find(appClock); w != nil {
			w.clock.armed = false
		}
	}
}

func (d *desk) windowView(c *ui.Context, w *win) {
	opts := w.opts
	// A modal dialog or alert in front leaves no window active under it.
	opts.Active = w == d.front() && !d.modalUp()
	opts.Dragging = d.dragging
	opts.ZoomTo = [4]int{4, 4, d.size[0] - 8, d.size[1] - 8}
	w.active = opts.Active
	if w.popup {
		pw := pt.PopupWindow(c.Key(w.id), &w.dock, d.titleOf(w), opts, func() {
			w.content(c, w)
		})
		pt.SpringOpen[*node](c, pw, &w.dock) // an icon dragged on the tab opens it
		if n, ok := ui.Drop[*node](pw.Tab()); ok && w.folder != nil && !w.folder.trash {
			dir := w.folder.dir
			copying := d.copying(c)
			d.later = append(d.later, func() { d.drop(n, dir, copying) }) // and dropped on it, goes in
			c.After(0)
		}
		if pw.Activated() && w != d.front() {
			d.raise(w)
		}
		if pw.TornOff() {
			d.undock(w)
		}
		if pw.Closed() {
			d.close(w)
		}
		return
	}
	// A folder's window has the path menu of the folders it is in.
	var outer []*node
	if w.folder != nil && !w.folder.trash {
		for n := w.folder.dir.parent; n != nil; n = n.parent {
			outer = append([]*node{n}, outer...)
		}
		for _, n := range outer {
			opts.Path = append(opts.Path, n.name)
		}
	}
	ww := pt.Window(c.Key(w.id), &w.frame, d.titleOf(w), opts, func() {
		w.content(c, w)
	})
	if i := ww.PathChosen(); i >= 0 && i < len(outer) {
		dir := outer[i]
		d.later = append(d.later, func() { d.openFolder(dir) })
		c.After(0)
	}
	if ww.Activated() && w != d.front() {
		d.raise(w)
	}
	if ww.Closed() {
		d.close(w)
	}
}

// The desktop patterns, 8×8, one bit a pixel, and the messages of their
// names.
var patterns = []struct {
	key    string
	bits   [8]byte
	fg, bg color.RGBA
}{
	{"patternLavenderWeave", [8]byte{0x88, 0x54, 0x22, 0x45, 0x88, 0x15, 0x22, 0x51}, color.RGBA{0x77, 0x77, 0xAA, 255}, color.RGBA{0x99, 0x99, 0xCC, 255}},
	{"patternGrayTweed", [8]byte{0x80, 0x10, 0x02, 0x20, 0x01, 0x08, 0x40, 0x04}, color.RGBA{0x88, 0x88, 0x88, 255}, color.RGBA{0xAA, 0xAA, 0xAA, 255}},
	{"patternTealBricks", [8]byte{0xFF, 0x80, 0x80, 0x80, 0xFF, 0x08, 0x08, 0x08}, color.RGBA{0x33, 0x77, 0x77, 255}, color.RGBA{0x55, 0x99, 0x99, 255}},
	{"patternRoseWaves", [8]byte{0x00, 0x18, 0x24, 0xC3, 0x00, 0x18, 0x24, 0xC3}, color.RGBA{0xAA, 0x66, 0x77, 255}, color.RGBA{0xCC, 0x99, 0xAA, 255}},
	{"patternPlainSlate", [8]byte{}, color.RGBA{}, color.RGBA{0x66, 0x77, 0x88, 255}},
}

// drawPattern tiles the desktop's pattern under everything, from a bitmap
// of the whole area made once for each pattern and size.
func (d *desk) drawPattern(p *ui.Painter, r ui.Rect) {
	w, h := int(r.W), int(r.H)
	if w <= 0 || h <= 0 {
		return
	}
	key := [3]int{d.pattern, w, h}
	if d.bg == nil || d.bgFor != key {
		pat := patterns[d.pattern]
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		for y := 0; y < h; y++ {
			row := pat.bits[y&7]
			for x := 0; x < w; x++ {
				c := pat.bg
				if row&(0x80>>(x&7)) != 0 {
					c = pat.fg
				}
				i := img.PixOffset(x, y)
				img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = c.R, c.G, c.B, 255
			}
		}
		if d.bg != nil && d.bgFor[1] == w && d.bgFor[2] == h {
			d.bg.Update(img)
		} else {
			d.bg = ui.NewBitmap(img).Pixelated()
		}
		d.bgFor = key
	}
	p.Image(d.bg, r, ui.FillBox)
}

const (
	deskIconW = 80
	deskIconH = 50
)

// iconsView builds the icons along the desktop's right edge.
func (d *desk) iconsView(c *ui.Context) {
	cw, _ := c.Size()
	for i, di := range d.icons {
		if di.x == 0 && di.y == 0 {
			di.x = max(int(cw)-deskIconW-12, 0)
			di.y = 12 + i*64
		}
		e := ui.Box(c.Key(fmt.Sprint("icon", i))).Absolute().Left(float32(di.x)).Top(float32(di.y)).Size(deskIconW, deskIconH).
			Focusable().FocusRing(false).Role(ui.RoleButton).Label(di.name)
		sel := d.selIcon == i
		from := ui.Rect{X: float32(di.x + (deskIconW-32)/2), Y: float32(di.y + pt.MenuBarHeight), W: 32, H: 32}
		if e.Shortcut(0, ui.KeyEnter) || e.Shortcut(ui.Cmd, ui.KeyO) || e.DoubleClicked() {
			d.zoomFrom = from
			di.open()
		}
		if di.icon == iconTrash {
			// Things dropped on the Wastebasket go into it.
			if n, ok := ui.Drop[*node](e); ok {
				d.later = append(d.later, func() { d.toTrash(n) })
				c.After(0)
			}
			if _, over := ui.DragOver[*node](e); over {
				sel = true
			}
		}
		if di.icon == iconDisk {
			// Things dropped on the disk go to its top; a drag resting on
			// it springs its window open.
			if n, ok := ui.Drop[*node](e); ok {
				copying := d.copying(c)
				d.later = append(d.later, func() { d.drop(n, d.disk, copying) })
				c.After(0)
			}
			if _, over := ui.DragOver[*node](e); over {
				sel = true
				d.springAt(c, d.disk)
			}
		}
		pt.ContextMenu(c, e, []pt.MenuItem{{Title: d.say().Open(), Action: di.open}, pt.Divider(),
			{Title: d.say().CleanUpDesktop(), Action: func() {
				for _, o := range d.icons {
					o.x, o.y = 0, 0
				}
			}}})
		if d.balloons {
			pt.Balloon(c, e, di.help)
		}
		if e.Focused() {
			d.selIcon = i
		}
		e.HandleInput(func(ev ui.InputEvent) bool {
			switch ev.Kind {
			case ui.InputPointerDown:
				d.selIcon = i
				di.grab = [2]float32{ev.X, ev.Y}
				di.dragging = false
				if ev.Clicks == 2 {
					d.zoomFrom = from
					di.open()
					return true
				}
				return false
			case ui.InputPointerMove:
				if ev.Button == 0 && (di.dragging || abs(int(ev.X-di.grab[0]))+abs(int(ev.Y-di.grab[1])) > 3) {
					di.dragging = true
					di.x += int(ev.X - di.grab[0])
					di.y += int(ev.Y - di.grab[1])
					return true
				}
			case ui.InputPointerUp:
				di.dragging = false
			}
			return false
		})
		name := di.name
		ic := di.icon
		if ic == iconTrash && len(d.trash) > 0 {
			ic = iconTrashFull
		}
		focused := e.FocusVisible()
		th := d.theme
		e.Draw(func(p *ui.Painter, r ui.Rect) {
			pn := pt.NewPen(p, r)
			ic.draw(p, float32(pn.X+(pn.W-32)/2), float32(pn.Y), sel)
			drawIconName(pn, name, 34, sel, th.Accent)
			if focused {
				pt.DrawFocusRing(pn.Sub((pn.W-34)/2, -1, 34, 34), 0, th.Accent.Dark)
			}
		})
	}
}

// drawIconName draws an icon's name centered under it, on a white plate,
// inverted when selected.
func drawIconName(pn pt.Pen, name string, y int, selected bool, a pt.Accent) {
	name = pt.AppFont.Truncate(name, pn.W-4)
	tw := pt.AppFont.Measure(name)
	x := (pn.W - tw) / 2
	bg, ink := pt.White, pt.Black
	if selected {
		bg, ink = pt.Black, pt.White
	}
	pn.Rect(x-2, y, tw+4, pt.AppFont.Height()+1, bg)
	pn.Text(x, y+1, name, pt.AppFont, ink)
}

// modalUp reports whether a modal dialog or alert is in front of the
// windows: the desk's alerts, the file dialogs, or Scribe's find and
// replace.
func (d *desk) modalUp() bool {
	if d.emptying || d.ringing || d.saving != nil || d.opening || d.asking != nil {
		return true
	}
	for _, w := range d.windows {
		if w.doc != nil && w.doc.finding {
			return true
		}
	}
	return false
}

// toTrash moves n into the Wastebasket, remembering where it was.
func (d *desk) toTrash(n *node) {
	n.from = n.parent
	n.remove()
	d.trash = append(d.trash, n)
}

// putAway takes n out of the Wastebasket, shown in window w, back to the
// folder it was in (or the disk's top, if that is gone too).
func (d *desk) putAway(w *win, n *node) {
	for i, t := range d.trash {
		if t == n {
			d.trash = append(d.trash[:i:i], d.trash[i+1:]...)
			break
		}
	}
	if w != nil && w.folder != nil {
		dir := w.folder.dir
		for i, c := range dir.children {
			if c == n {
				dir.children = append(dir.children[:i:i], dir.children[i+1:]...)
				break
			}
		}
		w.folder.sel = -1
	}
	to := n.from
	if to == nil || to.root() != d.disk {
		to = d.disk
	}
	n.from = nil
	to.add(n)
}

func (d *desk) openTrash() {
	t := &node{key: "wastebasket", name: d.say().Wastebasket(), kind: kindFolder}
	for _, n := range d.trash {
		t.children = append(t.children, n)
	}
	w := d.openFolder(t)
	w.folder.trash = true
}

// alarmCheck rings the alarm clock's alarm once its time comes.
func (d *desk) alarmCheck(c *ui.Context) {
	w := d.find(appClock)
	if w == nil || !w.clock.armed || d.ringing {
		return
	}
	now := d.now()
	at := time.Date(now.Year(), now.Month(), now.Day(), w.clock.hour, w.clock.minute, 0, 0, now.Location())
	if !now.Before(at) && now.Sub(at) < time.Minute {
		d.ringing = true
		w.clock.armed = false
	}
	c.After(time.Second)
}

func (d *desk) menus(c *ui.Context) []pt.Menu {
	m := d.say()
	front := d.front()
	frontApp := appKind(-1)
	if front != nil {
		frontApp = front.app
	}
	inCabinet := front != nil && front.folder != nil
	selected := (*node)(nil)
	if inCabinet {
		selected = front.folder.selected()
	}
	menus := []pt.Menu{
		{Title: m.Desk(), Icon: deskLogo, Items: []pt.MenuItem{
			{Title: m.AboutThisDeskMenu(), Action: func() { d.launch(appAbout) }},
			pt.Divider(),
			{Title: m.AlarmClock(), Action: func() { d.launch(appClock) }},
			{Title: m.Calculator(), Action: func() { d.launch(appCalculator) }},
			{Title: m.Jotter(), Action: func() { d.launch(appJotter) }},
			{Title: m.Ledger(), Action: func() { d.launch(appLedger) }},
			{Title: m.Puzzle(), Action: func() { d.launch(appPuzzle) }},
			{Title: m.Scribe(), Action: func() { d.launch(appScribe) }},
			pt.Divider(),
			{Title: m.ControlPanels(), Submenu: []pt.MenuItem{
				{Title: m.Appearance(), Action: func() { d.launch(appAppearance) }},
				{Title: m.DateTime(), Action: func() { d.launch(appDateTime) }},
				{Title: m.Language(), Action: func() { d.launch(appLanguage) }},
				{Title: m.Mouse(), Action: func() { d.launch(appMouse) }},
				{Title: m.Sound(), Action: func() { d.launch(appSound) }},
			}},
		}},
		{Title: m.MenuFile(), Items: []pt.MenuItem{
			{Title: m.NewFolder(), Key: ui.KeyN, Disabled: !inCabinet || front.folder.trash, Action: func() { front.folder.newFolder() }},
			{Title: m.Open(), Key: ui.KeyO, Disabled: selected == nil, Action: func() { d.openNode(selected) }},
			{Title: m.CloseWindow(), Key: ui.KeyW, Disabled: front == nil, Action: func() { d.close(front) }},
			pt.Divider(),
			{Title: m.FindFileMenu(), Action: func() { d.launch(appFind) }},
			pt.Divider(),
			{Title: m.MoveToWastebasket(), Key: ui.KeyBackspace, Disabled: selected == nil || front.folder.trash || selected.kind != kindDocument && selected.kind != kindFolder, Action: func() {
				d.toTrash(selected)
			}},
			{Title: m.PutAway(), Key: ui.KeyY, Disabled: selected == nil || !front.folder.trash, Action: func() {
				d.putAway(front, selected)
			}},
			{Title: m.GetInfo(), Key: ui.KeyI, Disabled: d.infoTarget(selected) == nil, Action: func() { d.getInfo(d.infoTarget(selected)) }},
			{Title: m.Duplicate(), Key: ui.KeyD, Disabled: selected == nil || front.folder.trash || selected.kind != kindDocument && selected.kind != kindFolder, Action: func() {
				n := selected
				d.later = append(d.later, func() { d.duplicate(n) }) // not while the folder is built
			}},
			pt.Divider(),
			{Title: m.Quit(), Key: ui.KeyQ, Disabled: d.quit == nil, Action: func() {
				// A document with changes asks first, as closing it does.
				for _, w := range d.windows {
					if w.doc != nil && w.doc.changed() {
						d.raise(w)
						d.asking = w
						return
					}
				}
				d.quit()
			}},
		}},
		{Title: m.MenuEdit(), Items: []pt.MenuItem{
			{Title: m.Undo(), Key: ui.KeyZ, Action: func() { pt.EditCommand(c, "undo") }},
			pt.Divider(),
			{Title: m.Cut(), Key: ui.KeyX, Action: func() { pt.EditCommand(c, "cut") }},
			{Title: m.Copy(), Key: ui.KeyC, Action: func() {
				// With no text to copy, the Calculator copies its display.
				if !pt.EditCommand(c, "copy") && front != nil && front.calc != nil {
					c.Services().WriteClipboard(front.calc.shown)
				}
			}},
			{Title: m.Paste(), Key: ui.KeyV, Action: func() {
				// And types the clipboard's number.
				if !pt.EditCommand(c, "paste") && front != nil && front.calc != nil {
					for _, r := range c.Services().ReadClipboard() {
						if r >= '0' && r <= '9' || r == '.' {
							front.calc.press(string(r))
						}
					}
				}
			}},
			{Title: m.SelectAll(), Key: ui.KeyA, Action: func() { pt.EditCommand(c, "selectAll") }},
		}},
		{Title: m.MenuView(), Items: []pt.MenuItem{
			{Title: m.AsIcons(), Checked: inCabinet && !front.folder.list && !front.folder.columns && !front.folder.buttons, Disabled: !inCabinet, Action: func() { front.folder.list, front.folder.columns, front.folder.buttons = false, false, false }},
			{Title: m.AsButtons(), Checked: inCabinet && front.folder.buttons, Disabled: !inCabinet || front.folder.trash, Action: func() { front.folder.list, front.folder.columns, front.folder.buttons = false, false, true }},
			{Title: m.AsList(), Checked: inCabinet && front.folder.list, Disabled: !inCabinet, Action: func() { front.folder.list, front.folder.columns, front.folder.buttons = true, false, false }},
			{Title: m.AsColumns(), Checked: inCabinet && front.folder.columns, Disabled: !inCabinet || front.folder.trash, Action: func() { front.folder.list, front.folder.columns, front.folder.buttons = false, true, false }},
			{Title: m.SortList(), Disabled: !inCabinet || !front.folder.list || front.folder.table == nil, Submenu: d.sortItems(front)},
			pt.Divider(),
			{Title: m.AsPopupWindow(), Checked: inCabinet && front.popup, Disabled: !inCabinet, Action: func() {
				if front.popup {
					d.undock(front)
				} else {
					d.dock(front)
				}
			}},
		}},
		{Title: m.MenuSpecial(), Items: []pt.MenuItem{
			{Title: m.CleanUpDesktop(), Action: func() {
				for _, di := range d.icons {
					di.x, di.y = 0, 0
				}
			}},
			{Title: m.EmptyWastebasketMenu(), Disabled: len(d.trash) == 0, Action: func() { d.emptying = true }},
			pt.Divider(),
			{Title: m.Restart(), Disabled: true},
			{Title: m.ShutDown(), Disabled: true},
		}},
	}
	switch frontApp {
	case appPuzzle:
		menus = append(menus, pt.Menu{Title: m.Puzzle(), Items: []pt.MenuItem{
			{Title: m.Shuffle(), Key: ui.KeyR, Action: func() { front.puzzle.shuffle(d.now().UnixNano()) }},
			{Title: m.Solve(), Action: func() { front.puzzle.reset() }},
		}})
	case appJotter:
		menus = append(menus, pt.Menu{Title: m.MenuPage(), Items: []pt.MenuItem{
			{Title: m.NextPage(), Key: ui.KeyBracketRight, Action: func() { front.jot.turn(1) }},
			{Title: m.PreviousPage(), Key: ui.KeyBracketLeft, Action: func() { front.jot.turn(-1) }},
		}})
	case appScribe:
		menus[2].Items = append(menus[2].Items, pt.Divider(),
			pt.MenuItem{Title: m.FindMenu(), Key: ui.KeyF, Action: func() { front.doc.finding = true }})
		for i, it := range menus[1].Items {
			if it.Title == m.Open() { // documents, from the open dialog
				menus[1].Items[i] = pt.MenuItem{Title: m.OpenMenu(), Key: ui.KeyO, Action: func() { d.opening = true }}
			}
		}
		menus[1].Items = append([]pt.MenuItem{
			{Title: m.Save(), Key: ui.KeyS, Action: func() { front.doc.save(d) }},
			{Title: m.SaveAsMenu(), Key: ui.KeyS, Mods: ui.Cmd | ui.Shift, Action: func() { front.doc.saveAsDialog(d) }},
			pt.Divider(),
		}, menus[1].Items...)
		menus = append(menus, pt.SizeMenu(front.doc.doc))
	}
	help := m.ShowBalloons()
	if d.balloons {
		help = m.HideBalloons()
	}
	menus = append(menus, pt.Menu{Title: m.MenuHelp(), Items: []pt.MenuItem{
		{Title: help, Action: func() { d.balloons = !d.balloons }},
	}})
	return append(menus, d.rightMenus()...)
}

var deskLogo = []string{
	"....##....",
	"...#oo#...",
	"..#oooo#..",
	".#oo++oo#.",
	"#oo+ww+oo#",
	"#oo+ww+oo#",
	".#oo++oo#.",
	"..#oooo#..",
	"...#oo#...",
	"....##....",
}

// fileDialogs shows the save dialog of a document being saved, and the
// open dialog, over the disk's files.
func (d *desk) fileDialogs(c *ui.Context) {
	if d.saving != nil {
		// New Folder makes one in the folder shown, the disk at the desktop.
		d.saveDlg.NewFolder = func(in *pt.TreeNode) string {
			dir := d.disk
			if in != nil {
				dir = in.Data.(*node)
			}
			name := d.say().UntitledFolder()
			for i := 2; dir.has(name); i++ {
				name = d.say().UntitledFolderN(i)
			}
			dir.add(&node{name: name, kind: kindFolder})
			return name
		}
		open := true
		if pt.SaveDialog(c, &open, &d.saveDlg, d.fileTree(), d.say().SaveAsLabel()) {
			dir := d.disk
			if p := d.saveDlg.Path; len(p) > 0 {
				dir = p[len(p)-1].Data.(*node)
			}
			s, name := d.saving, d.saveDlg.Name
			d.later = append(d.later, func() { s.saveAs(d, dir, name) })
			c.After(0)
		} else if !open {
			d.saving.closing = false // cancelled: the document stays open
		}
		if !open {
			d.saving = nil
		}
	}
	if d.opening {
		if n := pt.OpenDialog(c, &d.opening, &d.openDlg, d.fileTree()); n != nil {
			f := n.Data.(*node)
			d.later = append(d.later, func() { d.openNode(f) })
			c.After(0)
		}
	}
}

// fileTree is the disk as the file dialogs show it: its folders and
// documents, each node's Data the file.
func (d *desk) fileTree() []*pt.TreeNode {
	var tree func(n *node) *pt.TreeNode
	tree = func(n *node) *pt.TreeNode {
		t := &pt.TreeNode{Title: n.name, Data: n}
		if n.kind == kindFolder || n.kind == kindDisk {
			t.Children = []*pt.TreeNode{}
			for _, c := range n.children {
				if c.kind == kindFolder || c.kind == kindDocument {
					t.Children = append(t.Children, tree(c))
				}
			}
		}
		return t
	}
	return []*pt.TreeNode{tree(d.disk)}
}

// treePath is the path of the file dialogs down to dir: the disk, then
// the folders in it.
func (d *desk) treePath(dir *node) []*pt.TreeNode {
	var path []*pt.TreeNode
	for n := dir; n != nil; n = n.parent {
		path = append([]*pt.TreeNode{{Title: n.name, Data: n, Children: []*pt.TreeNode{}}}, path...)
	}
	return path
}

// documents is the Documents folder, or the disk where it has none.
func (d *desk) documents() *node {
	for _, c := range d.disk.children {
		if c.key == "documents" && c.kind == kindFolder {
			return c
		}
	}
	return d.disk
}

// copying reports whether a drop now copies rather than moves: Option is
// held, or was as the drag went on.
func (d *desk) copying(c *ui.Context) bool {
	return (c.Modifiers()|d.dropMods)&ui.Alt != 0
}

// sortItems are the Sort List submenu of a folder window shown as a list:
// by each of its columns, the one it is sorted by checked.
func (d *desk) sortItems(w *win) []pt.MenuItem {
	var items []pt.MenuItem
	m := d.say()
	for i, by := range []string{m.ByName(), m.BySize(), m.ByKind()} {
		checked := w != nil && w.folder != nil && w.folder.table != nil && w.folder.table.SortColumn == i
		items = append(items, pt.MenuItem{Title: by, Checked: checked, Action: func() {
			if w != nil && w.folder != nil && w.folder.table != nil {
				w.folder.table.SortColumn, w.folder.table.Descending = i, false
			}
		}})
	}
	return items
}
