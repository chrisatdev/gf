package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

// goos is the operating system identifier. Overridden in tests to simulate
// different platforms without recompiling.
var goos = runtime.GOOS

// openFn is the function used to launch the browser command.
// Override in tests to capture calls without spawning a real process.
var openFn = func(program, rawURL string) error {
	return exec.Command(program, rawURL).Run()
}

// Open opens rawURL in the system's default browser.
//
// Supported platforms:
//   - linux  → xdg-open
//   - darwin → open
//
// Returns a clear error on unsupported operating systems.
func Open(rawURL string) error {
	switch goos {
	case "linux":
		return openFn("xdg-open", rawURL)
	case "darwin":
		return openFn("open", rawURL)
	default:
		return fmt.Errorf("gf browser: unsupported OS %q", goos)
	}
}
