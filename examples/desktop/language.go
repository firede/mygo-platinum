package main

import (
	"slices"
	"strings"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	"github.com/firede/platinum-fonts/bitmap"
	"github.com/firede/platinum-fonts/cjk12/ja"
	"github.com/firede/platinum-fonts/cjk12/ko"
	"github.com/firede/platinum-fonts/cjk12/zhhans"
	"github.com/firede/platinum-fonts/cjk12/zhhant"

	pt "github.com/firede/mygo-platinum"
	"github.com/firede/mygo-platinum/examples/desktop/locales"
)

// say returns the desk's messages in the language of its window.
func (d *desk) say() locales.Messages { return locales.In(d.lang) }

// eastAsianFaces returns the faces of the forms of East Asian characters
// the language of locale writes: Traditional Chinese for zh-Hant and the
// Chinese of Taiwan, Hong Kong and Macao, Japanese, Korean, else
// Simplified Chinese.
func eastAsianFaces(locale string) (regular, bold func() *bitmap.Face) {
	tag := strings.ToLower(locale)
	switch {
	case strings.HasPrefix(tag, "zh") && (strings.Contains(tag, "hant") || strings.Contains(tag, "-tw") || strings.Contains(tag, "-hk") || strings.Contains(tag, "-mo")):
		return zhhant.Regular, zhhant.Bold
	case strings.HasPrefix(tag, "ja"):
		return ja.Regular, ja.Bold
	case strings.HasPrefix(tag, "ko"):
		return ko.Regular, ko.Bold
	}
	return zhhans.Regular, zhhans.Bold
}

// followLanguage makes the desk speak the language of its window, d.lang,
// once it changes: the forms of East Asian characters, and the names of
// the desk's own items and icons.
func (d *desk) followLanguage() {
	if d.labelled && d.lang == d.labelledIn {
		return
	}
	d.labelled, d.labelledIn = true, d.lang
	pt.SetEastAsian(eastAsianFaces(d.lang))
	d.disk.relabel(d.lang)
	for _, n := range d.trash {
		n.relabel(d.lang)
	}
	if d.trashNode != nil {
		d.trashNode.relabel(d.lang)
	}
	m := d.say()
	for _, di := range d.icons {
		switch di.icon {
		case iconDisk:
			di.name, di.help = d.disk.name, m.DiskHelp()
		case iconTrash:
			di.name, di.help = m.Wastebasket(), m.WastebasketHelp()
		}
	}
}

// languages are the languages the desk speaks, in the order the Language
// control panel lists them.
var languages = []string{"en", "zh-Hans", "zh-Hant", "ja", "ko"}

// chosenLanguage returns the language the app keeps, chosen in the
// Language control panel, or "" while it follows the user's: of the
// languages the app shows, the one the system's would not have picked.
func chosenLanguage(shown string, preferred, own []string) string {
	picked := i18n.Match(preferred, own)
	if picked == "" && len(own) > 0 {
		picked = own[0] // the app's own language, which it falls back to
	}
	if shown == "" || shown == picked {
		return ""
	}
	return shown
}

func newLanguageWindow(d *desk) *win {
	w := &win{app: appLanguage}
	w.opts = pt.WindowOptions{NoGrow: true, NoZoom: true}
	w.frame = pt.WindowState{W: 300}
	w.content = func(c *ui.Context, w *win) { languageView(c, d) }
	return w
}

func languageView(c *ui.Context, d *desk) {
	m := d.say()
	choice := d.chosenLang
	ui.Column(c).FillWidth().Padding(10).Gap(10).AlignItems(ui.Stretch).Children(func() {
		pt.GroupBox(c, m.LanguageLabel(), func() {
			pt.RadioGroup(c, func() {
				ui.Column(c).Gap(6).Children(func() {
					pt.Radio(c, &choice, "", m.LanguageSystem())
					for _, tag := range languages {
						pt.Radio(c, &choice, tag, i18n.Autonym(tag))
					}
				})
			})
		})
		pt.Paragraph(c, m.LanguageNote(), pt.AppFont, 270)
	})
	if choice != d.chosenLang && slices.Contains(append([]string{""}, languages...), choice) {
		d.chosenLang = choice
		if d.setLocale != nil {
			d.setLocale(choice)
		}
	}
}

// iconLanguage is the Language control panel's: a globe.
var iconLanguage = newIcon(func(k canvas) {
	k.disc(32, 32, 24, blue)
	k.line(16, 5, 16, 27, black)
	k.line(5, 16, 27, 16, black)
	k.line(7, 10, 25, 10, black)
	k.line(7, 22, 25, 22, black)
	k.line(12, 6, 10, 16, black)
	k.line(10, 16, 12, 26, black)
	k.line(20, 6, 22, 16, black)
	k.line(22, 16, 20, 26, black)
})
