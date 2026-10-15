package platinum

import (
	"encoding/json"
	"io/fs"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	"github.com/firede/platinum-fonts/bitmap"
	"github.com/firede/platinum-fonts/cjk12/ja"
	"github.com/firede/platinum-fonts/cjk12/zhhans"
)

// TestCatalogs checks the controls' translations against the English
// catalog: the same keys, messages that parse, with the same arguments,
// and each said in its own language rather than falling back to English.
func TestCatalogs(t *testing.T) {
	read := func(name string) map[string]string {
		data, err := fs.ReadFile(l10nFiles, "l10n/"+name)
		if err != nil {
			t.Fatal(err)
		}
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		msgs := map[string]string{}
		for k, v := range raw {
			if s, ok := v.(string); ok && k[0] != '@' {
				msgs[k] = s
			}
		}
		return msgs
	}
	args := func(name, key, msg string) []string {
		m, err := i18n.ParseMessage(msg)
		if err != nil {
			t.Errorf("%s: %s: %v", name, key, err)
			return nil
		}
		var out []string
		for _, a := range m.Args() {
			out = append(out, a.Name)
		}
		slices.Sort(out)
		return out
	}
	en := read("en.json")
	langs := map[string]bool{}
	for _, l := range catalog().Languages() {
		langs[l] = true
	}
	for _, tag := range []string{"zh-Hans", "zh-Hant", "ja", "ko"} {
		if !langs[tag] {
			t.Errorf("the catalog has no %s", tag)
		}
		tr := read(tag + ".json")
		for key, msg := range en {
			got, ok := tr[key]
			if !ok {
				t.Errorf("%s lacks %s", tag, key)
				continue
			}
			if !slices.Equal(args(tag, key, got), args("en", key, msg)) {
				t.Errorf("%s: %s takes %v, the English %v", tag, key, args(tag, key, got), args("en", key, msg))
			}
			if _, lang, ok := catalog().Message(tag, key); !ok || lang != tag {
				t.Errorf("%s: %s is said in %q", tag, key, lang)
			}
		}
		for key := range tr {
			if _, ok := en[key]; !ok {
				t.Errorf("%s has %s, which English has not", tag, key)
			}
		}
	}
}

// TestSnapshotLanguages draws the controls whose words and dates follow
// the window's language in Simplified Chinese and Japanese, with the East
// Asian faces of each.
func TestSnapshotLanguages(t *testing.T) {
	freeze(t)
	t.Cleanup(func() { SetEastAsian(nil, nil) })
	day := time.Date(2026, 10, 15, 21, 41, 0, 0, time.UTC)
	page, query := 1, ""
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Row(c).Fill().Padding(10).Gap(12).AlignItems(ui.Start).Children(func() {
			Calendar(c, &day)
			ui.Column(c).Gap(10).AlignItems(ui.Start).Children(func() {
				DatePicker(c, &day)
				TimeField(c, &day, 1)
				Pagination(c, &page, 5)
				SearchField(c, &query)
			})
		})
	}
	for _, tc := range []struct {
		tag           string
		golden        string
		regular, bold func() *bitmap.Face
	}{
		{"zh-Hans", "languages-zh-hans", zhhans.Regular, zhhans.Bold},
		{"ja", "languages-ja", ja.Regular, ja.Bold},
	} {
		SetEastAsian(tc.regular, tc.bold)
		tt := ui.NewTester(view, 420, 200)
		tt.SetLocale(tc.tag)
		tt.Frame()
		compareGolden(t, tc.golden, tt.Image())
	}
}

// TestSnapshotSaveDialogChinese draws the save dialog, and its alert over
// a file of the same name, in Simplified Chinese.
func TestSnapshotSaveDialogChinese(t *testing.T) {
	freeze(t)
	SetEastAsian(zhhans.Regular, zhhans.Bold)
	t.Cleanup(func() { SetEastAsian(nil, nil) })
	roots := fileTree()
	open := true
	st := &FileDialogState{Path: []*TreeNode{roots[0], roots[0].Children[0]}, Name: "Letter"}
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Box(c).Fill().Children(func() {
			SaveDialog(c, &open, st, roots, "存储为：")
		})
	}
	tt := ui.NewTester(view, 400, 390)
	tt.SetLocale("zh-Hans")
	tt.Frame()
	compareGolden(t, "save-dialog-zh-hans", tt.Image())
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if !tt.HasText("要替换现有的“Letter”吗？") {
		t.Fatalf("no alert of the file there: %q", tt.Texts())
	}
	compareGolden(t, "save-dialog-replace-zh-hans", tt.Image())
}

// TestTimeLayout checks how each language writes a time in a cycle of 12
// hours — the half of the day after the minutes in English, before the
// hours in Chinese, Japanese and Korean — and of 24.
func TestTimeLayout(t *testing.T) {
	for _, tc := range []struct {
		locale string
		twelve bool
		order  []int
		seps   []string
		pad    bool
	}{
		{"en", true, []int{0, 1, 2}, []string{":", " "}, false},
		{"en", false, []int{0, 1}, []string{":"}, true},
		{"zh-Hans", true, []int{2, 0, 1}, []string{"", ":"}, false},
		{"zh-Hans", false, []int{0, 1}, []string{":"}, false},
		{"zh-Hant", true, []int{2, 0, 1}, []string{"", ":"}, false},
		{"ja", true, []int{2, 0, 1}, []string{"", ":"}, false},
		{"ja", false, []int{0, 1}, []string{":"}, false},
		{"ko", true, []int{2, 0, 1}, []string{" ", ":"}, false},
		{"ko", false, []int{0, 1}, []string{":"}, true},
	} {
		l := timeLayoutOf(tc.locale, tc.twelve)
		if !slices.Equal(l.order, tc.order) || !slices.Equal(l.seps, tc.seps) || l.pad != tc.pad {
			t.Errorf("%s (12 hours %v): order %v, separators %q, two digits %v; want %v, %q, %v",
				tc.locale, tc.twelve, l.order, l.seps, l.pad, tc.order, tc.seps, tc.pad)
		}
	}
}

// TestSnapshotRegions draws the controls of dates and times with regional
// formats the user set apart from the language: English with Monday first
// and 24 hours, Chinese with Sunday first and 12.
func TestSnapshotRegions(t *testing.T) {
	freeze(t)
	t.Cleanup(func() { SetEastAsian(nil, nil) })
	SetEastAsian(zhhans.Regular, zhhans.Bold)
	day := time.Date(2026, 10, 15, 21, 41, 0, 0, time.UTC)
	view := func(c *ui.Context) {
		Use(c, nil)
		ui.Row(c).Fill().Padding(10).Gap(12).AlignItems(ui.Start).Children(func() {
			Calendar(c, &day)
			ui.Column(c).Gap(10).AlignItems(ui.Start).Children(func() {
				DatePicker(c, &day)
				TimeField(c, &day, 1)
			})
		})
	}
	for _, tc := range []struct {
		tag, golden string
		region      i18n.Region
	}{
		{"en", "region-en-monday-24", i18n.Region{Locale: "en-GB", FirstWeekday: time.Monday, HourCycle: 24}},
		{"zh-Hans", "region-zh-hans-sunday-12", i18n.Region{FirstWeekday: time.Sunday, HourCycle: 12}},
	} {
		tt := ui.NewTester(view, 400, 180)
		tt.SetLocale(tc.tag)
		tt.SetRegion(tc.region)
		tt.Frame()
		compareGolden(t, tc.golden, tt.Image())
	}
}
