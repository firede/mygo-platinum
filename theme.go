package platinum

import (
	"sync"

	"github.com/egoist/mygo/ui"
)

// The gray ramp of the Platinum look. Every control is drawn from these
// twelve grays, black and white, and the theme's accent; docs/spec.md
// says which part of which control takes which.
var (
	White  = ui.RGB(0xFF, 0xFF, 0xFF)
	Gray1  = ui.RGB(0xEE, 0xEE, 0xEE) // inner highlights, menus
	Gray2  = ui.RGB(0xDD, 0xDD, 0xDD) // the face of windows and controls
	Gray3  = ui.RGB(0xCC, 0xCC, 0xCC) // pressed faces of small controls
	Gray4  = ui.RGB(0xBB, 0xBB, 0xBB) // tracks of scroll bars and sliders
	Gray5  = ui.RGB(0xAA, 0xAA, 0xAA) // inner shadows
	Gray6  = ui.RGB(0x99, 0x99, 0x99) // disabled frames
	Gray7  = ui.RGB(0x88, 0x88, 0x88) // outer shadows
	Gray8  = ui.RGB(0x77, 0x77, 0x77) // disabled text
	Gray9  = ui.RGB(0x66, 0x66, 0x66) // pressed faces of buttons
	Gray10 = ui.RGB(0x55, 0x55, 0x55) // bevel button outlines, zoom rectangles
	Gray11 = ui.RGB(0x44, 0x44, 0x44) // pressed shadows
	Gray12 = ui.RGB(0x22, 0x22, 0x22) // a gray of the color picker
	Black  = ui.RGB(0x00, 0x00, 0x00)
)

// Accent is a color variation: the color of highlighted menu items, selected
// rows, the keyboard focus, the filled part of progress bars and the thumbs
// of sliders.
type Accent struct {
	// Name is what an appearance control panel calls it.
	Name string
	// Dark fills highlighted menu items and selected rows under white text,
	// and draws the focus ring.
	Dark ui.Color
	// Mid fills progress bars and slider thumbs.
	Mid ui.Color
	// Light highlights selected text under black text.
	Light ui.Color
}

// The accents that ship with the library.
var (
	Lavender = Accent{Name: "Lavender", Dark: ui.RGB(0x4C, 0x4C, 0xA8), Mid: ui.RGB(0x88, 0x88, 0xDD), Light: ui.RGB(0xCC, 0xCC, 0xFF)}
	Graphite = Accent{Name: "Graphite", Dark: ui.RGB(0x4A, 0x4A, 0x4A), Mid: ui.RGB(0x8C, 0x8C, 0x8C), Light: ui.RGB(0xCC, 0xCC, 0xCC)}
	Harbor   = Accent{Name: "Harbor", Dark: ui.RGB(0x1F, 0x5A, 0x8C), Mid: ui.RGB(0x5C, 0x99, 0xCC), Light: ui.RGB(0xBB, 0xDD, 0xFF)}
	Moss     = Accent{Name: "Moss", Dark: ui.RGB(0x3A, 0x6B, 0x2A), Mid: ui.RGB(0x77, 0xAA, 0x5E), Light: ui.RGB(0xCC, 0xEE, 0xBB)}
	Ember    = Accent{Name: "Ember", Dark: ui.RGB(0x99, 0x3D, 0x1F), Mid: ui.RGB(0xDD, 0x88, 0x55), Light: ui.RGB(0xFF, 0xDD, 0xBB)}
	Orchid   = Accent{Name: "Orchid", Dark: ui.RGB(0x7A, 0x2E, 0x7A), Mid: ui.RGB(0xBB, 0x77, 0xBB), Light: ui.RGB(0xEE, 0xCC, 0xEE)}
)

// Accents lists the accents that ship with the library, Lavender first.
func Accents() []Accent {
	return []Accent{Lavender, Graphite, Harbor, Moss, Ember, Orchid}
}

// Theme is what a window's Platinum controls take their look from.
type Theme struct {
	Accent Accent
	// Background is the face of windows and dialogs.
	Background ui.Color
}

// DefaultTheme returns the theme controls use where Use set none: the gray
// ramp with the Lavender accent.
func DefaultTheme() *Theme {
	return &Theme{Accent: Lavender, Background: Gray2}
}

var (
	themesMu sync.Mutex
	themes   = map[*ui.Context]*Theme{}
)

// Use makes the window of c draw its Platinum controls with th, and sets the
// MyGo theme of the window to match (its background, text, selection and
// focus colors), so that MyGo's own elements sit well among them. Call it at
// the top of the view, every frame; a nil th is DefaultTheme.
func Use(c *ui.Context, th *Theme) {
	if th == nil {
		th = DefaultTheme()
	}
	themesMu.Lock()
	themes[c] = th
	themesMu.Unlock()
	nextFrame(c)
	mt := *ui.LightTheme()
	mt.Background = th.Background
	mt.Surface = Gray2
	mt.SurfaceHover = Gray1
	mt.SurfacePressed = Gray3
	mt.Border = Black
	mt.Text = Black
	mt.TextMuted = Gray8
	mt.Accent = th.Accent.Dark
	mt.AccentHover = th.Accent.Dark
	mt.AccentPressed = th.Accent.Dark
	mt.AccentText = White
	mt.Selection = th.Accent.Light
	mt.Focus = th.Accent.Dark
	mt.Scrollbar = ui.Transparent
	mt.Radius = 0
	c.SetTheme(&mt)
}

// ThemeOf returns the theme the window of c draws with.
func ThemeOf(c *ui.Context) *Theme {
	themesMu.Lock()
	th := themes[c]
	themesMu.Unlock()
	if th == nil {
		th = defaultTheme
	}
	return th
}

var defaultTheme = DefaultTheme()
