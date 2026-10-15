package platinum

import (
	"strings"
	"syscall/js"
)

// macKeyboard returns whether the browser runs on a Mac, where MyGo takes
// ⌘ for Control in shortcuts, so that menus show ⌘ as the key to press.
func macKeyboard() bool {
	p := js.Global().Get("navigator").Get("platform")
	return p.Type() == js.TypeString && strings.HasPrefix(p.String(), "Mac")
}
