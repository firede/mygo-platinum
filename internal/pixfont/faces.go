package pixfont

import _ "embed"

//go:embed system.txt
var systemSrc string

// System is the bold face of menus, buttons, labels and window titles.
var System = MustParse("system", systemSrc)

//go:embed app.txt
var appSrc string

// App is the regular face of document text, lists and icon names.
var App = MustParse("app", appSrc)
