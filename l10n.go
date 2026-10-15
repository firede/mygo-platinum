package platinum

import (
	"embed"
	"io/fs"
	"sync"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
)

// The controls' own words, in English, Simplified and Traditional Chinese,
// Japanese and Korean: the catalog l10n/en.json and its translations. An
// app changes them, or gives them its other languages, with keys
// "platinum.<key>" in its catalogs, as "platinum.cancel".
//
//go:embed l10n/*.json
var l10nFiles embed.FS

var catalog = sync.OnceValue(func() *i18n.Catalog {
	files, err := fs.Sub(l10nFiles, "l10n")
	if err != nil {
		panic(err)
	}
	c, err := i18n.NewLibraryCatalog(files, "en", "platinum")
	if err != nil {
		panic(err)
	}
	return c
})

// say returns the message of key in the language of c's window.
func say(c *ui.Context, key string) string { return catalog().Format(c.Locale(), key, nil) }

// sayWith returns the message of key with args in the language of c's
// window.
func sayWith(c *ui.Context, key string, args map[string]any) string {
	return catalog().Format(c.Locale(), key, args)
}

// sayApp returns the message of key in the language the app shows, for
// what is made without a window, as menus.
func sayApp(key string) string { return catalog().Format(i18n.Locale(), key, nil) }

// sayAppWith returns the message of key with args in the language the app
// shows.
func sayAppWith(key string, args map[string]any) string {
	return catalog().Format(i18n.Locale(), key, args)
}

// formats returns the regional formats the controls of c write dates and
// times in — the first day of the week and the hour cycle — and the
// locale that puts them together with the window's language, as MyGo's
// own controls do: the user's, or where the system says nothing of them,
// as for a Tester, those CLDR gives the language.
func formats(c *ui.Context) (i18n.Region, string) {
	lang := c.Locale()
	if lang == "" {
		lang = "en"
	}
	r := c.Region()
	if r.HourCycle != 12 && r.HourCycle != 24 {
		r = i18n.RegionFor(lang)
		r.Locale = ""
	}
	return r, r.FormatLocale(lang)
}
