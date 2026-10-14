package main

import (
	"github.com/egoist/mygo/ui"
	pt "github.com/firede/mygo-platinum"
	"strconv"
	"strings"
)

// calculator is a four-function calculator.
type calculator struct {
	entry    string  // the number being typed, or "" after an operator
	acc      float64 // the value to its left
	op       string  // the operator pending
	shown    string  // what the display shows
	fresh    bool    // the display shows a result; typing starts anew
	flashKey string  // the key the keyboard pressed, shown pressed briefly
}

func newCalculatorWindow() *win {
	w := &win{app: appCalculator, title: "Calculator", calc: &calculator{shown: "0"}}
	w.frame = pt.WindowState{W: 156, H: 196}
	w.opts = pt.WindowOptions{NoGrow: true, NoZoom: true}
	w.content = calculatorView
	return w
}

func format(v float64) string {
	s := strconv.FormatFloat(v, 'g', 10, 64)
	if strings.Contains(s, "e") || len(s) > 14 {
		s = strconv.FormatFloat(v, 'e', 4, 64)
	}
	return s
}

// press acts on a key of the calculator: a digit, ".", an operator, "=",
// "C" or "←".
func (k *calculator) press(key string) {
	switch {
	case key >= "0" && key <= "9" || key == ".":
		if k.fresh {
			k.entry, k.fresh = "", false
			k.op = ""
		}
		if key == "." && strings.Contains(k.entry, ".") || len(k.entry) >= 12 {
			return
		}
		if k.entry == "0" && key != "." {
			k.entry = ""
		}
		if k.entry == "" && key == "." {
			k.entry = "0"
		}
		k.entry += key
		k.shown = k.entry
	case key == "C":
		*k = calculator{shown: "0"}
	case key == "←":
		if k.entry != "" && !k.fresh {
			k.entry = k.entry[:len(k.entry)-1]
			k.shown = k.entry
			if k.shown == "" {
				k.shown = "0"
			}
		}
	case key == "=":
		k.apply()
		k.op, k.fresh = "", true
	default: // + − × ÷
		if k.entry != "" || k.op == "" {
			k.apply()
		}
		k.op, k.fresh = key, false
	}
}

// apply folds the entry into the accumulator with the pending operator.
func (k *calculator) apply() {
	v := k.acc
	if k.entry != "" {
		v, _ = strconv.ParseFloat(k.entry, 64)
	}
	switch k.op {
	case "+":
		k.acc += v
	case "−":
		k.acc -= v
	case "×":
		k.acc *= v
	case "÷":
		if v == 0 {
			k.shown, k.acc, k.entry = "Error", 0, ""
			return
		}
		k.acc /= v
	default:
		k.acc = v
	}
	k.entry = ""
	k.shown = format(k.acc)
}

// keyboard maps the keys of the keyboard to those of the calculator.
var calcKeyboard = []struct {
	mods ui.Modifiers
	key  ui.Key
	calc string
}{
	{0, ui.Key0, "0"}, {0, ui.Key1, "1"}, {0, ui.Key2, "2"}, {0, ui.Key3, "3"}, {0, ui.Key4, "4"},
	{0, ui.Key5, "5"}, {0, ui.Key6, "6"}, {0, ui.Key7, "7"}, {0, ui.Key8, "8"}, {0, ui.Key9, "9"},
	{0, ui.KeyPeriod, "."}, {0, ui.KeyMinus, "−"}, {0, ui.KeySlash, "÷"}, {ui.Shift, ui.Key8, "×"},
	{ui.Shift, ui.KeyEqual, "+"}, {0, ui.KeyEqual, "="}, {0, ui.KeyEnter, "="}, {0, ui.KeyEscape, "C"},
	{0, ui.KeyBackspace, "←"},
}

// shiftedOperator handles platforms that report '+' and '*' as unknown
// layout keys, leaving their ANSI key positions in the input event.
func (k *calculator) shiftedOperator(ev ui.InputEvent) bool {
	if ev.Kind != ui.InputKeyDown || ev.Mods != ui.Shift || ev.Key != ui.KeyUnknown {
		return false
	}
	var key string
	switch ev.PhysicalKey {
	case ui.KeyEqual:
		key = "+"
	case ui.Key8:
		key = "×"
	default:
		return false
	}
	k.press(key)
	k.flashKey = key
	return true
}

func calculatorView(c *ui.Context, w *win) {
	k := w.calc
	pressed := map[string]bool{}
	if k.flashKey != "" {
		pressed[k.flashKey] = true
		k.flashKey = ""
	}
	if w.active {
		for _, kb := range calcKeyboard {
			if c.Shortcut(kb.mods, kb.key) {
				k.press(kb.calc)
				pressed[kb.calc] = true
			}
		}
	}
	input := func(ev ui.InputEvent) bool {
		return w.active && k.shiftedOperator(ev)
	}
	key := func(label string, width, height int) {
		b := pt.Button(c, label).Width(width)
		b.Element().HandleInput(input)
		if height > 0 {
			b.Element().Height(float32(height))
		}
		if pressed[label] {
			b.Flash()
		}
		if b.Clicked() {
			k.press(label)
		}
	}
	row := func(keys ...string) {
		ui.Row(c).Gap(4).Children(func() {
			for _, l := range keys {
				key(l, 28, 0)
			}
		})
	}
	body := ui.Column(c).Fill().Padding(10).Gap(8).AlignItems(ui.Stretch).
		Focusable().FocusRing(false).HandleInput(input)
	if w.active && !body.FocusWithin() {
		body.Focus()
	}
	body.Children(func() {
		// The display: a sunken panel with the number at its right.
		shown := k.shown
		ui.Box(c).Height(24).Role(ui.RoleStatus).Label("Display").Value(shown).Draw(func(p *ui.Painter, r ui.Rect) {
			pn := pt.NewPen(p, r)
			pn.Bevel(0, 0, pn.W, pn.H, pt.Gray7, pt.White)
			pn.Frame(1, 1, pn.W-2, pn.H-2, pt.Black)
			pn.Rect(2, 2, pn.W-4, pn.H-4, ui.RGB(0xDD, 0xE8, 0xCC))
			tw := pt.SystemFont.Measure(shown)
			pn.TextMiddle(pn.W-tw-6, 0, pn.H, shown, pt.SystemFont, pt.Black)
		})
		ui.Column(c).Gap(4).Children(func() {
			row("C", "←", "÷", "×")
			row("7", "8", "9", "−")
			row("4", "5", "6", "+")
			ui.Row(c).Gap(4).Children(func() {
				ui.Column(c).Gap(4).Children(func() {
					row("1", "2", "3")
					ui.Row(c).Gap(4).Children(func() {
						key("0", 60, 0)
						key(".", 28, 0)
					})
				})
				key("=", 28, 44)
			})
		})
	})
}
