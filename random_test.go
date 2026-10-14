package platinum

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

type memClipboard struct{ s string }

func (m *memClipboard) ReadClipboard() string   { return m.s }
func (m *memClipboard) WriteClipboard(s string) { m.s = s }

// TestRichTextRandomEdits drives a rich text document through thousands of
// random edits, keys, styles, sizes, alignments, undos and redos, laying it
// out and drawing it as it goes: nothing may panic, and the styles and
// alignments must stay one to a rune and at most one to a paragraph.
func TestRichTextRandomEdits(t *testing.T) {
	rng := rand.New(rand.NewSource(1999))
	keys := []ui.Key{ui.KeyLeft, ui.KeyRight, ui.KeyUp, ui.KeyDown, ui.KeyBackspace, ui.KeyDelete, ui.KeyEnter, ui.KeyHome, ui.KeyEnd}
	texts := []string{"a", "word ", "Ünïcödé", "\n", "two\nlines", "", "   ", "longer text that wraps past the width"}
	for round := 0; round < 8; round++ {
		d := NewRichText("Start\nof it")
		ed := d.editor()
		ed.wrap = 60 + rng.Intn(100)
		clip := &memClipboard{}
		img := ui.Render(func(c *ui.Context) {
			ui.Box(c).Fill().Draw(func(p *ui.Painter, r ui.Rect) {
				pn := NewPen(p, r)
				for i := 0; i < 300; i++ {
					switch rng.Intn(10) {
					case 0, 1:
						ed.replace(texts[rng.Intn(len(texts))], editTyping)
					case 2:
						mods := ui.Modifiers(0)
						if rng.Intn(3) == 0 {
							mods = ui.Shift
						}
						if rng.Intn(4) == 0 {
							mods |= ui.Cmd
						}
						ed.key(ui.InputEvent{Kind: ui.InputKeyDown, Key: keys[rng.Intn(len(keys))], Mods: mods}, clip)
					case 3:
						d.Toggle([]TextStyle{Bold, Italic, Underline}[rng.Intn(3)])
					case 4:
						d.SetSize(1 + rng.Intn(MaxScale))
					case 5:
						d.SetAlign(Align(rng.Intn(3)))
					case 6:
						if rng.Intn(2) == 0 {
							ed.undoEdit()
						} else {
							ed.redoEdit()
						}
					case 7:
						a, b := rng.Intn(len(ed.text)+1), rng.Intn(len(ed.text)+1)
						ed.anchor, ed.caret = a, b
					case 8:
						ed.command([]string{"selectAll", "copy", "cut", "paste", "undo", "redo"}[rng.Intn(6)], clip)
					case 9:
						ed.moveTo(ed.offsetAt(rng.Intn(len(ed.lines)+2)-1, rng.Intn(200)-10), rng.Intn(2) == 0)
						if rng.Intn(4) == 0 {
							ed.indent = rng.Intn(48)
						}
					}
					ed.lines = nil
					ed.layout()
					if len(ed.styles) != len(ed.text) {
						t.Fatalf("round %d step %d: %d styles for %d runes", round, i, len(ed.styles), len(ed.text))
					}
					if n := strings.Count(string(ed.text), "\n") + 1; len(ed.paras) > n {
						t.Fatalf("round %d step %d: %d alignments for %d paragraphs", round, i, len(ed.paras), n)
					}
					if ed.caret < 0 || ed.caret > len(ed.text) || ed.anchor < 0 || ed.anchor > len(ed.text) {
						t.Fatalf("round %d step %d: caret %d, anchor %d in %d runes", round, i, ed.caret, ed.anchor, len(ed.text))
					}
					ed.xOf(ed.caret)
					ed.lineAtY(0, rng.Intn(400)-20, TextLineHeight)
					ed.drawLines(pn, 4, 4, TextLineHeight, rng.Intn(len(ed.lines)), 0, Black, Gray4, true)
					_ = d.Runs()
					_ = d.Aligns()
					_ = d.Style()
				}
			})
		}, 240, 200, 1)
		if img == nil {
			t.Fatal("no image")
		}
	}
}

// TestListsRandomInput clicks, Shift- and Command-clicks and presses keys
// at random in a table, a list box and a tree while their rows come and go:
// nothing may panic.
func TestListsRandomInput(t *testing.T) {
	freeze(t)
	rng := rand.New(rand.NewSource(1984))
	rows := 12
	st := NewTableState()
	sel := 0
	roots := []*TreeNode{{Title: "One", Open: true, Children: []*TreeNode{{Title: "Two"}, {Title: "Three"}}}, {Title: "Four"}}
	var chosen *TreeNode
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Row(c).Padding(4).Gap(4).AlignItems(ui.Start).Children(func() {
			cols := []Column{{Title: "Name", Width: 60}, {Title: "Size", Width: 40, Right: true}}
			Table(c, st, cols, rows, func(r, col int) string { return itoa(r * (col + 1)) }).Label("Table").Element().Size(130, 120)
			items := make([]string, rows)
			for i := range items {
				items[i] = itoa(i)
			}
			ListBox(c, &sel, items).Label("List").Element().Size(80, 120)
			TreeView(c, roots, &chosen).Label("Tree").Element().Size(90, 120)
		})
	}
	tt := ui.NewTester(view, 330, 130)
	tt.Frame()
	keys := []ui.Key{ui.KeyUp, ui.KeyDown, ui.KeyLeft, ui.KeyRight, ui.KeyHome, ui.KeyEnd, ui.KeyPageUp, ui.KeyPageDown, ui.KeyEnter, ui.KeySpace, ui.KeyTab}
	for i := 0; i < 400; i++ {
		switch rng.Intn(4) {
		case 0:
			mods := []ui.Modifiers{0, ui.Shift, ui.Cmd}[rng.Intn(3)]
			tt.ClickAtWith(mods, float32(rng.Intn(330)), float32(rng.Intn(130)))
		case 1:
			mods := []ui.Modifiers{0, ui.Shift, ui.Cmd}[rng.Intn(3)]
			tt.Key(mods, keys[rng.Intn(len(keys))])
		case 2:
			rows = rng.Intn(20)
		case 3:
			if rng.Intn(2) == 0 {
				roots[0].Children = roots[0].Children[:rng.Intn(len(roots[0].Children)+1)]
			} else {
				roots[0].Children = append(roots[0].Children, &TreeNode{Title: "New"})
			}
		}
		tt.Frame()
	}
}
