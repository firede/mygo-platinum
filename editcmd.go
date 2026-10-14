package platinum

import (
	"sync"

	"github.com/egoist/mygo/ui"
)

// The text field or area with the keyboard focus in each window, for the
// edit commands of menus drawn by the library.
var (
	editMu     sync.Mutex
	editTarget = map[*ui.Context]editReg{}
	frameGen   = map[*ui.Context]int{}
)

type editReg struct {
	gen int
	run func(cmd string)
}

// nextFrame counts the frames of a window; Use calls it.
func nextFrame(c *ui.Context) {
	editMu.Lock()
	frameGen[c]++
	editMu.Unlock()
}

func registerEditTarget(c *ui.Context, run func(cmd string)) {
	editMu.Lock()
	editTarget[c] = editReg{gen: frameGen[c], run: run}
	editMu.Unlock()
}

// EditCommand performs an edit command — "undo", "redo", "cut", "copy",
// "paste", "delete" or "selectAll" — on the text field or text area that
// has the keyboard focus in the window of c, as the items of an Edit menu
// do, and reports whether one had it.
func EditCommand(c *ui.Context, cmd string) bool {
	editMu.Lock()
	reg, ok := editTarget[c]
	gen := frameGen[c]
	editMu.Unlock()
	// The field must have had the focus in this frame or the last.
	if !ok || reg.gen < gen-1 {
		return false
	}
	reg.run(cmd)
	return true
}
