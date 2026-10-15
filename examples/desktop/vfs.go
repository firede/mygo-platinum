package main

import (
	"sort"
	"strings"
	"time"

	pt "github.com/firede/mygo-platinum"
	"github.com/firede/mygo-platinum/examples/desktop/locales"
)

// node is a file or folder of the desk's own little file system, which
// lives in memory: the desktop browses it, and the text editor reads and
// writes its documents.
type node struct {
	name     string
	key      string // the message of its name, for the desk's own items, which follow the language
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

func (n *node) kindName(m locales.Messages) string {
	switch n.kind {
	case kindFolder:
		return m.KindFolder()
	case kindDocument:
		return m.KindDocument()
	case kindApp:
		return m.KindApp()
	case kindDisk:
		return m.KindDisk()
	case kindPanel:
		return m.KindPanel()
	}
	return ""
}

// relabel names the desk's own items of n, and n, in the language of
// locale, and sorts the folders by the new names.
func (n *node) relabel(locale string) {
	if n.key != "" {
		n.name = locales.Catalog.Format(locale, n.key, nil)
	}
	for _, c := range n.children {
		c.relabel(locale)
	}
	n.sort()
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
// and a few documents, in the language of locale. The documents are the
// user's: they keep their names and words when the language changes.
func newDisk(now time.Time, locale string) *node {
	m := locales.In(locale)
	own := func(key string, k kind, app appKind) *node {
		return &node{key: key, name: locales.Catalog.Format(locale, key, nil), kind: k, app: app, modified: now}
	}
	d := own("platinumDisk", kindDisk, 0)
	apps := d.add(own("applications", kindFolder, 0))
	for _, a := range []struct {
		key string
		app appKind
	}{{"calculator", appCalculator}, {"jotter", appJotter}, {"ledger", appLedger}, {"scribe", appScribe}, {"puzzle", appPuzzle}, {"alarmClock", appClock}} {
		apps.add(own(a.key, kindApp, a.app))
	}
	docs := d.add(own("documents", kindFolder, 0))
	docs.add(&node{name: m.Letter(), kind: kindDocument, modified: now, text: m.LetterText()})
	docs.add(&node{name: m.Recipes(), kind: kindDocument, modified: now, text: m.RecipesText()})
	docs.add(&node{name: m.ToDo(), kind: kindDocument, modified: now, text: m.ToDoText()})
	sys := d.add(own("systemFolder", kindFolder, 0))
	cp := sys.add(own("controlPanels", kindFolder, 0))
	cp.add(own("appearance", kindPanel, appAppearance))
	cp.add(own("dateTime", kindPanel, appDateTime))
	cp.add(own("language", kindPanel, appLanguage))
	cp.add(own("mouse", kindPanel, appMouse))
	cp.add(own("sound", kindPanel, appSound))
	d.add(&node{name: m.ReadMe(), kind: kindDocument, modified: now, text: m.ReadMeText()})
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
	c.key = "" // a copy is the user's: its name stays as it is
	c.runs = append([]pt.Run(nil), n.runs...)
	c.aligns = append([]pt.Align(nil), n.aligns...)
	for _, ch := range n.children {
		cc := ch.clone()
		cc.parent = &c
		c.children = append(c.children, cc)
	}
	return &c
}
