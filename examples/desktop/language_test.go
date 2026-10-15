package main

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/firede/platinum-fonts/cjk12/ja"
	"github.com/firede/platinum-fonts/cjk12/ko"
	"github.com/firede/platinum-fonts/cjk12/zhhans"
	"github.com/firede/platinum-fonts/cjk12/zhhant"

	pt "github.com/firede/mygo-platinum"
	"github.com/firede/mygo-platinum/examples/desktop/locales"
)

// TestDesktopScenes draws the scene of the desk in each East Asian
// language it speaks, its documents written in it, and compares each with
// its golden image.
func TestDesktopScenes(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	t.Cleanup(func() { pt.SetEastAsian(nil, nil) })
	for _, lang := range []string{"zh-Hans", "zh-Hant", "ja", "ko"} {
		golden(t, "scene-"+strings.ToLower(lang), renderScene(1, lang))
	}
}

// TestDesktopLanguagePanel chooses languages in the Language control
// panel: the menus, the icons, the windows' titles and the desk's own
// folders follow, a copy the user made keeps its name, and the system's
// language brings English back.
func TestDesktopLanguagePanel(t *testing.T) {
	defer pt.SetClock(pt.SetClock(func() time.Time { return sceneTime }))
	t.Cleanup(func() { pt.SetEastAsian(nil, nil) })
	d := newDesk(func() time.Time { return sceneTime })
	var apps *node
	for _, c := range d.disk.children {
		if c.key == "applications" {
			apps = c
		}
	}
	d.duplicate(apps)
	tt := ui.NewTester(d.view, width, height)
	d.setLocale = tt.SetLocale
	d.launch(appCalculator)
	d.launch(appLanguage)
	tt.Frame()
	for _, s := range []string{"File", "Calculator", "Language", "Platinum Disk", "Wastebasket"} {
		if !tt.HasText(s) {
			t.Fatalf("no %q in English: %q", s, tt.Texts())
		}
	}
	if err := tt.Click("简体中文"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if d.chosenLang != "zh-Hans" {
		t.Fatalf("the panel chose %q", d.chosenLang)
	}
	for _, s := range []string{"文件", "计算器", "语言", "Platinum 磁盘", "废纸篓", "系统语言"} {
		if !tt.HasText(s) {
			t.Errorf("no %q in Chinese: %q", s, tt.Texts())
		}
	}
	for _, c := range d.disk.children {
		switch {
		case c == apps && c.name != "应用程序":
			t.Errorf("Applications is %q in Chinese", c.name)
		case c.key == "" && c.kind == kindFolder && c.name != "Applications copy":
			t.Errorf("the user's copy is %q in Chinese", c.name)
		}
	}
	if err := tt.Click("日本語"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText("ファイル") || !tt.HasText("計算機") {
		t.Errorf("no menus in Japanese: %q", tt.Texts())
	}
	if err := tt.Click("システムの言語"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if d.chosenLang != "" || !tt.HasText("File") || !tt.HasText("Calculator") {
		t.Errorf("the system's language did not bring English back: chose %q, %q", d.chosenLang, tt.Texts())
	}
}

// TestEastAsianFaces checks the forms the desk draws East Asian characters
// in, for each language.
func TestEastAsianFaces(t *testing.T) {
	ptr := func(f any) uintptr { return reflect.ValueOf(f).Pointer() }
	for _, tc := range []struct {
		locale string
		want   any
	}{
		{"", zhhans.Regular}, {"en", zhhans.Regular}, {"zh-Hans", zhhans.Regular}, {"zh-CN", zhhans.Regular},
		{"zh-Hant", zhhant.Regular}, {"zh-TW", zhhant.Regular}, {"zh-Hant-HK", zhhant.Regular},
		{"ja", ja.Regular}, {"ja-JP", ja.Regular}, {"ko", ko.Regular}, {"ko-KR", ko.Regular},
	} {
		regular, _ := eastAsianFaces(tc.locale)
		if ptr(regular) != ptr(tc.want) {
			t.Errorf("%q draws in the wrong forms", tc.locale)
		}
	}
}

// TestChosenLanguage checks what the Language control panel shows chosen:
// the system's language unless the app shows another than it would pick.
func TestChosenLanguage(t *testing.T) {
	own := []string{"en", "ja", "ko", "zh-Hans", "zh-Hant"}
	for _, tc := range []struct {
		shown     string
		preferred []string
		want      string
	}{
		{"zh-Hans", []string{"zh-Hans-CN", "en"}, ""},
		{"ja", []string{"zh-Hans-CN", "en"}, "ja"},
		{"en", []string{"de-DE"}, ""},
		{"ko", []string{"de-DE"}, "ko"},
		{"", nil, ""},
	} {
		if got := chosenLanguage(tc.shown, tc.preferred, own); got != tc.want {
			t.Errorf("%s shown, %v preferred: chose %q, want %q", tc.shown, tc.preferred, got, tc.want)
		}
	}
}

// TestWordCount counts words between spaces, and each Han character and
// kana as a word.
func TestWordCount(t *testing.T) {
	for _, tc := range []struct{ lang, text, want string }{
		{"en", "Two words\nand a line", "5 words, 2 lines"},
		{"en", "One", "1 word, 1 line"},
		{"zh-Hans", "新桌面已经运转起来了。", "10 个字，1 行"},
		{"ja", "かなと漢字", "5 語、1 行"},
		{"ko", "새 데스크가 잘", "단어 3개, 1줄"},
	} {
		if got := wordCount(locales.In(tc.lang), tc.text); got != tc.want {
			t.Errorf("%s %q: %q, want %q", tc.lang, tc.text, got, tc.want)
		}
	}
}
