package platinum

import (
	"image"
	"reflect"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestRichTextComposition(t *testing.T) {
	freeze(t)
	for _, tc := range []struct {
		name   string
		width  int
		style  TextStyle
		align  Align
		prefix string
	}{
		{"plain", 240, 0, AlignLeft, "A"},
		{"styled", 240, Bold.Scaled(2), AlignCenter, "A"},
		{"wrapped", 70, 0, AlignLeft, "AAAAAAAA"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			doc := NewRichText(tc.prefix + "Z")
			tester := func(d *RichText) *ui.Tester {
				return ui.NewTester(func(c *ui.Context) {
					Use(c, nil)
					RichTextArea(c, d).Label("Document").Element().Size(float32(tc.width), 90)
				}, tc.width+20, 110)
			}
			tt := tester(doc)
			tt.Key(0, ui.KeyTab)
			tt.Key(ui.Cmd, ui.KeyDown)
			tt.Key(0, ui.KeyLeft)
			doc.ed.typing = tc.style
			doc.SetAlign(tc.align)
			originalRuns := doc.Runs()
			undoCount := len(doc.ed.undo)
			for _, caret := range []int{5, 2} {
				tt.Compose("nihao", caret)
				tt.Frame()
				if doc.Text != tc.prefix+"Z" || !reflect.DeepEqual(doc.Runs(), originalRuns) || len(doc.ed.undo) != undoCount {
					t.Fatal("preedit changed the document or its undo history")
				}
				// The preview must look like real underlined text, with the
				// insertion caret at the input method's requested offset.
				want := NewRichText("")
				want.SetRuns([]Run{{tc.prefix, 0}, {"nihao", tc.style | Underline}, {"Z", 0}})
				want.SetAligns([]Align{tc.align})
				want.ed.caret = len([]rune(tc.prefix)) + caret
				want.ed.anchor = want.ed.caret
				ref := tester(want)
				ref.Key(0, ui.KeyTab)
				ref.Frame()
				if diff := diffImages(ref.Image(), tt.Image()); diff != "" {
					t.Fatalf("preedit rendering at offset %d: %s", caret, diff)
				}
				gotCaret, active := tt.TextCaret()
				wantCaret, _ := ref.TextCaret()
				if !active || gotCaret != wantCaret {
					t.Fatalf("candidate anchor=%v active=%v, want %v", gotCaret, active, wantCaret)
				}
				assertPaintedCaret(t, tt.Image(), gotCaret)
			}
			tt.Compose("", 0)
			tt.Frame()
			if doc.Text != tc.prefix+"Z" || len(doc.ed.undo) != undoCount {
				t.Fatal("canceling preedit changed the document")
			}
			tt.Compose("nihao", 5)
			tt.Type("你好")
			tt.Frame()
			if doc.Text != tc.prefix+"你好Z" {
				t.Fatalf("commit produced %q", doc.Text)
			}
			tt.Key(ui.Cmd, ui.KeyZ)
			tt.Frame()
			if doc.Text != tc.prefix+"Z" || !reflect.DeepEqual(doc.Runs(), originalRuns) {
				t.Fatalf("undo did not restore the original document: %q", doc.Text)
			}
		})
	}
}

func TestCompositionReplacesSelection(t *testing.T) {
	freeze(t)
	doc := NewRichText("AoldZ")
	tt := ui.NewTester(func(c *ui.Context) {
		Use(c, nil)
		RichTextArea(c, doc).Label("Document").Element().Size(200, 80)
	}, 220, 100)
	tt.Key(0, ui.KeyTab)
	doc.ed.anchor, doc.ed.caret = 1, 4
	tt.Compose("nihao", 2)
	tt.Frame()
	if doc.Text != "AoldZ" || doc.ed.selectedText() != "old" {
		t.Fatal("preedit modified the selected document text")
	}
	caret, active := tt.TextCaret()
	wantX := float32(4 + AppFont.Measure("Ani") + 1)
	if !active || caret.X != wantX {
		t.Fatalf("candidate anchor=%v, want X=%g", caret, wantX)
	}
	assertPaintedCaret(t, tt.Image(), caret)
	tt.Compose("", 0)
	tt.Frame()
	if doc.ed.selectedText() != "old" {
		t.Fatal("canceling preedit lost the selection")
	}
	tt.Compose("nihao", 5)
	tt.Type("你好")
	tt.Frame()
	if doc.Text != "A你好Z" {
		t.Fatalf("commit produced %q", doc.Text)
	}
	tt.Key(ui.Cmd, ui.KeyZ)
	tt.Frame()
	if doc.Text != "AoldZ" {
		t.Fatalf("undo produced %q", doc.Text)
	}
}

func assertPaintedCaret(t *testing.T, img *image.RGBA, caret ui.Rect) {
	t.Helper()
	for y := int(caret.Y); y < int(caret.Y+caret.H); y++ {
		if got := img.RGBAAt(int(caret.X), y); got != rgba(Black) {
			t.Fatalf("candidate anchor differs from the painted caret at %d,%d: %v", int(caret.X), y, got)
		}
	}
}

func TestTextCompositionCaret(t *testing.T) {
	freeze(t)
	value := "A"
	tt := ui.NewTester(func(c *ui.Context) {
		Use(c, nil)
		TextField(c, &value).Label("Name").Width(60)
	}, 80, 40)
	tt.Key(0, ui.KeyTab)
	tt.Key(0, ui.KeyEnd)
	tt.Frame()
	for _, caretOffset := range []int{12, 2, 0} {
		tt.Compose(strings.Repeat("a", 12), caretOffset)
		tt.Frame()
		caret, active := tt.TextCaret()
		if !active || caret.X < 3 || caret.X >= 57 {
			t.Fatalf("candidate anchor outside the visible input: %v", caret)
		}
		assertPaintedCaret(t, tt.Image(), caret)
		if value != "A" {
			t.Fatalf("preedit changed the value to %q", value)
		}
	}
	tt.Compose("", 0)
	tt.Frame()
	caret, _ := tt.TextCaret()
	assertPaintedCaret(t, tt.Image(), caret)
}
