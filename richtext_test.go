package platinum

import (
	"reflect"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestRichText(t *testing.T) {
	freeze(t)
	doc := NewRichText("Plain ")
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(8).AlignItems(ui.Stretch).Children(func() {
			FormatBar(c, doc)
			RichTextArea(c, doc).Label("Document").Element().Size(260, 90)
		})
	}
	tt := ui.NewTester(view, 280, 150)
	r, _ := tt.Find("Document")
	tt.ClickAt(r.X+r.W-30, r.Y+8) // focus, at the end
	tt.Key(ui.Cmd, ui.KeyDown)
	tt.Key(ui.Cmd, ui.KeyB)
	tt.Type("bold")
	tt.Key(ui.Cmd, ui.KeyB)
	tt.Type(" and ")
	tt.Key(ui.Cmd, ui.KeyU)
	tt.Type("underlined")
	tt.Frame()
	runs := doc.Runs()
	if len(runs) != 4 || runs[1] != (Run{"bold", Bold}) || runs[3] != (Run{"underlined", Underline}) {
		t.Fatalf("runs %+v", runs)
	}
	// Select "Plain" and make it italic from the format bar.
	tt.Key(ui.Cmd, ui.KeyUp)
	for i := 0; i < 5; i++ {
		tt.Key(ui.Shift, ui.KeyRight)
	}
	tt.Frame()
	if err := tt.Click("Italic"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if doc.Runs()[0] != (Run{"Plain", Italic}) {
		t.Fatalf("runs %+v", doc.Runs())
	}
	if err := tt.Click("Align Center"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if doc.Align() != AlignCenter {
		t.Fatal("not centered")
	}
	compareGolden(t, "richtext", tt.Image())
	// Undo takes the centering away, then the italic.
	x, y := findCenter(t, tt, "Document")
	tt.ClickAt(x, y)
	tt.Key(ui.Cmd, ui.KeyZ)
	tt.Frame()
	if doc.Align() != AlignLeft || doc.Runs()[0].Style != Italic {
		t.Fatalf("undo left %v and %+v", doc.Align(), doc.Runs()[0])
	}
	tt.Key(ui.Cmd, ui.KeyZ)
	tt.Frame()
	if doc.Runs()[0].Style != 0 {
		t.Fatalf("undo left %+v", doc.Runs()[0])
	}
}

func TestRichTextParagraphAlign(t *testing.T) {
	freeze(t)
	doc := NewRichText("Title\nThe first paragraph runs on, wrapping.\nThe end")
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(8).AlignItems(ui.Stretch).Children(func() {
			FormatBar(c, doc)
			RichTextArea(c, doc).Label("Document").Element().Size(200, 90)
		})
	}
	tt := ui.NewTester(view, 220, 150)
	r, _ := tt.Find("Document")
	tt.ClickAt(r.X+20, r.Y+8) // in the title
	tt.Frame()
	if err := tt.Click("Align Center"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	// The caret to the last paragraph, aligned right.
	tt.ClickAt(r.X+20, r.Y+8)
	tt.Key(ui.Cmd, ui.KeyDown)
	tt.Frame()
	if doc.Align() != AlignLeft {
		t.Fatalf("the last paragraph is aligned %v before it was aligned", doc.Align())
	}
	if err := tt.Click("Align Right"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if got := doc.Aligns(); len(got) != 3 || got[0] != AlignCenter || got[1] != AlignLeft || got[2] != AlignRight {
		t.Fatalf("aligns %v", got)
	}
	compareGolden(t, "richtext-paragraphs", tt.Image())
	// A return in the last paragraph makes another aligned as it is.
	tt.ClickAt(r.X+r.W-10, r.Y+r.H-10)
	tt.Key(ui.Cmd, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	tt.Type("!")
	tt.Frame()
	if got := doc.Aligns(); len(got) != 4 || got[3] != AlignRight {
		t.Fatalf("after a return, aligns %v", got)
	}
	// Joining the first two paragraphs keeps the first one's alignment.
	tt.Key(ui.Cmd, ui.KeyUp)
	tt.Key(ui.Cmd, ui.KeyRight)
	tt.Key(0, ui.KeyDelete)
	tt.Frame()
	if got := doc.Aligns(); len(got) != 3 || got[0] != AlignCenter || got[1] != AlignRight {
		t.Fatalf("after joining, aligns %v (%q)", got, doc.Text)
	}
	// Undo brings the paragraphs and their alignments back.
	tt.Key(ui.Cmd, ui.KeyZ)
	tt.Frame()
	if got := doc.Aligns(); len(got) != 4 || got[0] != AlignCenter || got[1] != AlignLeft {
		t.Fatalf("after undo, aligns %v (%q)", got, doc.Text)
	}
}

func TestRichTextSizes(t *testing.T) {
	freeze(t)
	doc := NewRichText("")
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Column(c).Padding(8).AlignItems(ui.Stretch).Children(func() {
			RichTextArea(c, doc).Label("Document").Element().Size(220, 110)
		})
	}
	tt := ui.NewTester(view, 240, 130)
	r, _ := tt.Find("Document")
	tt.ClickAt(r.X+20, r.Y+8)
	tt.Key(ui.Cmd|ui.Shift, ui.KeyPeriod)
	tt.Key(ui.Cmd, ui.KeyB)
	tt.Type("Menu")
	tt.Key(ui.Cmd|ui.Shift, ui.KeyComma)
	tt.Key(ui.Cmd, ui.KeyB)
	tt.Type(" for today")
	tt.Key(0, ui.KeyEnter)
	tt.Type("Soup, ")
	tt.Key(ui.Cmd, ui.KeyI)
	tt.Type("bread")
	tt.Frame()
	want := []Run{{"Menu", Bold.Scaled(2)}, {" for today\nSoup, ", 0}, {"bread", Italic}}
	if got := doc.Runs(); !reflect.DeepEqual(got, want) {
		t.Fatalf("runs %+v", got)
	}
	// A click on the second line, under the taller first, finds it.
	tt.ClickAt(r.X+4, r.Y+float32(4+2*TextLineHeight+AppFont.Height()+2))
	tt.Frame()
	if a, _ := doc.editor().selection(); a != len([]rune("Menu for today\n")) {
		t.Fatalf("the click put the caret at %d", a)
	}
	compareGolden(t, "richtext-sizes", tt.Image())
	// The Size menu checks the size of the selection, and sets it.
	doc.editor().anchor, doc.editor().caret = 0, 4
	m := SizeMenu(doc)
	if !m.Items[1].Checked || m.Items[0].Checked {
		t.Fatalf("the Size menu checks %+v", m.Items)
	}
	m.Items[2].Action()
	if doc.Size() != 3 || doc.Style() != Bold.Scaled(3) {
		t.Fatalf("size %d, style %v", doc.Size(), doc.Style())
	}
}
