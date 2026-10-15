package platinum

import (
	"strings"
	"unicode"

	"github.com/egoist/mygo/ui"
)

// FindOptions says how Find matches.
type FindOptions struct {
	MatchCase, WholeWord bool
}

// Find returns where the first match of query in text starts and ends, in
// runes, at or after from — wrapping round to the start — or -1, -1.
func Find(text, query string, from int, o FindOptions) (int, int) {
	if query == "" {
		return -1, -1
	}
	t, q := []rune(text), []rune(query)
	if !o.MatchCase {
		t, q = []rune(strings.ToLower(text)), []rune(strings.ToLower(query))
	}
	word := func(i int) bool { return i >= 0 && i < len(t) && (unicode.IsLetter(t[i]) || unicode.IsDigit(t[i])) }
	match := func(i int) bool {
		if i+len(q) > len(t) {
			return false
		}
		for k := range q {
			if t[i+k] != q[k] {
				return false
			}
		}
		return !o.WholeWord || !word(i-1) && !word(i+len(q))
	}
	from = min(max(from, 0), len(t))
	for n := 0; n < len(t); n++ {
		i := (from + n) % len(t)
		if match(i) {
			return i, i + len(q)
		}
	}
	return -1, -1
}

// ReplaceAll replaces every match of query in text with with, and returns
// the text and how many it replaced.
func ReplaceAll(text, query, with string, o FindOptions) (string, int) {
	n := 0
	from := 0
	for {
		a, b := Find(text, query, from, o)
		if a < 0 || a < from {
			return text, n
		}
		r := []rune(text)
		text = string(r[:a]) + with + string(r[b:])
		from = a + len([]rune(with))
		n++
	}
}

// FindAction is what the user asked of a find and replace dialog.
type FindAction int

// The actions of a find and replace dialog.
const (
	FindNone FindAction = iota
	FindNext
	ReplaceOne
	ReplaceEvery
)

// FindState is what a find and replace dialog edits, kept by the app.
type FindState struct {
	Query, With string
	Options     FindOptions
	// Note is what the dialog says under its fields, as "3 replaced";
	// screen readers read it out when it shows.
	Note      string
	announced string // the note read out, or "\x00" to read the next
}

// FindReplace shows a find and replace dialog while *open: fields of what to
// find and what to replace it with, the options to match case and whole
// words, and the buttons Replace All, Replace and Find (the default). It
// returns what the user pressed, for the app to do it to its text, as with
// Find and ReplaceAll, and the select methods of its text area.
func FindReplace(c *ui.Context, open *bool, st *FindState) FindAction {
	act := FindNone
	Dialog(c, open, say(c, "findAndReplace"), 340, func() {
		for _, f := range []struct {
			label string
			v     *string
		}{{say(c, "findLabel"), &st.Query}, {say(c, "replaceLabel"), &st.With}} {
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				ui.Box(c).Width(62).Children(func() { Label(c, f.label) })
				tf := TextField(c, f.v).Width(240).Label(strings.TrimSuffix(f.label, ":"))
				if f.v == &st.Query {
					tf.Element().AutoFocus()
				}
			})
		}
		ui.Row(c).Gap(14).Padding(0, 0, 0, 68).Children(func() {
			Checkbox(c, &st.Options.MatchCase, say(c, "matchCase"))
			Checkbox(c, &st.Options.WholeWord, say(c, "wholeWord"))
		})
		if st.Note != "" {
			ui.Row(c).Padding(0, 0, 0, 68).Children(func() { TextIn(c, st.Note, AppFont, Gray9) })
		}
		if st.Note != st.announced {
			if st.Note != "" {
				c.Announce(st.Note)
			}
			st.announced = st.Note
		}
		ui.Row(c).Gap(10).Justify(ui.End).AlignItems(ui.Center).Children(func() {
			if Button(c, say(c, "replaceAll")).Disabled(st.Query == "").Clicked() {
				act = ReplaceEvery
			}
			if Button(c, say(c, "replace")).Disabled(st.Query == "").Clicked() {
				act = ReplaceOne
			}
			if Button(c, say(c, "find")).Default().Disabled(st.Query == "").Clicked() {
				act = FindNext
			}
		})
	})
	if act != FindNone {
		st.announced = "\x00" // the note the app gives it is news, if the same
	}
	return act
}
