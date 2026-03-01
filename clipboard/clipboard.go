// Package clipboard detects and writes to the system clipboard.
package clipboard

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Detect finds the clipboard binary available on the system.
// It checks for pbcopy (macOS), xsel (Linux preferred), and xclip (Linux fallback)
// in that order, returning the first one found.
func Detect() (string, error) {
	for _, tool := range []string{"pbcopy", "xsel", "xclip"} {
		if _, err := exec.LookPath(tool); err == nil {
			return tool, nil
		}
	}
	return "", fmt.Errorf("no clipboard tool found\n\nInstall with:\n  macOS:  pbcopy is built-in\n  Linux:  sudo apt-get install xsel")
}

// Available reports whether clipboard copy is likely to work.
// Returns false on headless Linux (no DISPLAY or WAYLAND_DISPLAY)
// or when no clipboard tool is installed.
func Available() bool {
	if runtime.GOOS == "linux" {
		if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
			return false
		}
	}
	_, err := Detect()
	return err == nil
}

// Write copies text to the system clipboard using the detected clipboard tool.
func Write(text string) error {
	tool, err := Detect()
	if err != nil {
		return err
	}

	var args []string
	switch tool {
	case "xsel":
		args = []string{"--clipboard", "--input"}
	case "xclip":
		args = []string{"-selection", "clipboard"}
	}

	cmd := exec.Command(tool, args...)
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}
