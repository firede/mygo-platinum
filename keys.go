//go:build !js

package platinum

import "runtime"

// macKeyboard returns whether the system marks its shortcuts as a Mac
// does.
func macKeyboard() bool { return runtime.GOOS == "darwin" }
