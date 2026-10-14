package main

import (
	"sort"
	"strings"
	"time"

	pt "github.com/firede/mygo-platinum"
)

// node is a file or folder of the desk's own little file system, which
// lives in memory: the desktop browses it, and the text editor reads and
// writes its documents.
type node struct {
	name     string
	kind     kind
	text     string   // a document's contents
	runs     []pt.Run // and its styles and alignments, as Scribe saved them
	aligns   []pt.Align
	app      appKind
	children []*node
	parent   *node
	from     *node  // the folder an item in the Wastebasket came from
	comment  string // what Get Info's Comments say
	modified time.Time
}

type kind int

const (
	kindFolder kind = iota
	kindDocument
	kindApp
	kindDisk
	kindPanel // a control panel
)

func (n *node) kindName() string {
	switch n.kind {
	case kindFolder:
		return "folder"
	case kindDocument:
		return "text document"
	case kindApp:
		return "application"
	case kindDisk:
		return "disk"
	case kindPanel:
		return "control panel"
	}
	return ""
}

// size returns the bytes a node takes, as the browser shows them.
func (n *node) size() int {
	switch n.kind {
	case kindDocument:
		return len(n.text)
	case kindApp, kindPanel:
		return 24*1024 + len(n.name)*512
	}
	s := 0
	for _, c := range n.children {
		s += c.size()
	}
	return s
}

func (n *node) add(c *node) *node {
	c.parent = n
	n.children = append(n.children, c)
	n.sort()
	return c
}

func (n *node) sort() {
	sort.SliceStable(n.children, func(i, j int) bool {
		return strings.ToLower(n.children[i].name) < strings.ToLower(n.children[j].name)
	})
}

func (n *node) path() string {
	if n.parent == nil {
		return n.name
	}
	return n.parent.path() + ":" + n.name
}

func (n *node) remove() {
	p := n.parent
	if p == nil {
		return
	}
	for i, c := range p.children {
		if c == n {
			p.children = append(p.children[:i], p.children[i+1:]...)
			break
		}
	}
	n.parent = nil
}

// newDisk returns the desk's disk, filled with its folders, applications
// and a few documents.
func newDisk(now time.Time) *node {
	d := &node{name: "Platinum Disk", kind: kindDisk, modified: now}
	apps := d.add(&node{name: "Applications", kind: kindFolder, modified: now})
	for _, a := range []struct {
		name string
		app  appKind
	}{{"Calculator", appCalculator}, {"Jotter", appJotter}, {"Ledger", appLedger}, {"Scribe", appScribe}, {"Puzzle", appPuzzle}, {"Alarm Clock", appClock}} {
		apps.add(&node{name: a.name, kind: kindApp, app: a.app, modified: now})
	}
	docs := d.add(&node{name: "Documents", kind: kindFolder, modified: now})
	docs.add(&node{name: "Letter", kind: kindDocument, modified: now, text: "Dear Ada,\n\nThe new desk is up and running. Every window has a striped title bar again, and the buttons press in when you click them.\n\nWrite soon,\nGrace"})
	docs.add(&node{name: "Recipes", kind: kindDocument, modified: now, text: "Pancakes\n- 2 cups flour\n- 2 eggs\n- 1 1/2 cups milk\n- a pinch of salt\n\nWhisk, rest ten minutes, fry in butter."})
	docs.add(&node{name: "To Do", kind: kindDocument, modified: now, text: "- Water the plants\n- Back up the disk\n- Finish the puzzle"})
	sys := d.add(&node{name: "System Folder", kind: kindFolder, modified: now})
	cp := sys.add(&node{name: "Control Panels", kind: kindFolder, modified: now})
	cp.add(&node{name: "Appearance", kind: kindPanel, app: appAppearance, modified: now})
	cp.add(&node{name: "Date & Time", kind: kindPanel, app: appDateTime, modified: now})
	cp.add(&node{name: "Mouse", kind: kindPanel, app: appMouse, modified: now})
	cp.add(&node{name: "Sound", kind: kindPanel, app: appSound, modified: now})
	d.add(&node{name: "Read Me", kind: kindDocument, modified: now, text: "Welcome to the Platinum desk.\n\nDouble-click an icon to open it. Drag windows by their title bars, resize them from the bottom right corner, and roll them up with the box at the right of the title bar.\n\nThe Appearance control panel changes the accent color and the desktop pattern."})
	return d
}

// root returns the outermost folder n is in, or n.
func (n *node) root() *node {
	for n.parent != nil {
		n = n.parent
	}
	return n
}

// clone returns a copy of n and of all it holds, outside any folder.
func (n *node) clone() *node {
	c := *n
	c.parent, c.from, c.children = nil, nil, nil
	c.runs = append([]pt.Run(nil), n.runs...)
	c.aligns = append([]pt.Align(nil), n.aligns...)
	for _, ch := range n.children {
		cc := ch.clone()
		cc.parent = &c
		c.children = append(c.children, cc)
	}
	return &c
}
