// Package paste handles clipboard write + best-effort keystroke injection
// to paste transcribed text at the cursor.
//
// The contract (system-design.md §7): if clipboard write succeeds, Paste
// MUST return nil — paste injection is a bonus. A failure to simulate the
// paste keystroke is logged and swallowed.
//
// Runtime behavior per platform is deliberately command-based (shells out
// to xdotool / wtype / ydotool / osascript) so this package has no CGO.
// The richer CGEventPost path on macOS lives behind the `ui` build tag
// and the human handoff — see HANDOFF.md.
package paste

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/cdimoush/vox/clipboard"
)

// Method enumerates paste strategies. Values match config.paste.method.
type Method string

const (
	MethodAuto          Method = "auto"
	MethodClipboardOnly Method = "clipboard-only"
	MethodXdotool       Method = "xdotool"
	MethodYdotool       Method = "ydotool"
	MethodWtype         Method = "wtype"
)

// Injector builds the command that simulates the paste keystroke. Runner
// is injectable so tests can capture invocations without spawning real
// processes.
type Injector struct {
	Method  Method
	Runner  func(name string, args ...string) error
	LookPath func(string) (string, error)
}

// NewInjector returns an Injector bound to the real exec.LookPath and
// exec.Command-based runner.
func NewInjector(m Method) *Injector {
	return &Injector{
		Method:   m,
		Runner:   func(name string, args ...string) error { return exec.Command(name, args...).Run() },
		LookPath: exec.LookPath,
	}
}

// Paste writes text to the clipboard, then attempts (best-effort) to
// simulate the platform paste keystroke. Returns an error ONLY when the
// clipboard write fails.
func (i *Injector) Paste(text string) error {
	if err := clipboard.Write(text); err != nil {
		return fmt.Errorf("clipboard: %w", err)
	}
	if i.Method == MethodClipboardOnly {
		return nil
	}
	if err := i.inject(); err != nil {
		// Best-effort: log to stderr and keep going.
		fmt.Fprintf(os.Stderr, "vox: paste injection skipped: %v\n", err)
	}
	return nil
}

// injectHook, when non-nil, is tried before command-based injection.
// Set by platform-specific init() functions (e.g. paste_darwin_ui.go
// registers CGEventPost when built with the ui tag).
var injectHook func() error

// inject chooses and runs a paste keystroke command. Returns an error
// only if every applicable strategy fails.
func (i *Injector) inject() error {
	// Try the platform-specific hook first (e.g. CGEventPost on macOS).
	if injectHook != nil {
		if err := injectHook(); err == nil {
			return nil
		}
		// Fall through to command-based injection.
	}

	cmds := i.candidates()
	if len(cmds) == 0 {
		return fmt.Errorf("no paste injector available for %s", runtime.GOOS)
	}
	var lastErr error
	for _, c := range cmds {
		if _, err := i.LookPath(c.name); err != nil {
			lastErr = fmt.Errorf("%s not in PATH", c.name)
			continue
		}
		if err := i.Runner(c.name, c.args...); err != nil {
			lastErr = fmt.Errorf("%s failed: %w", c.name, err)
			continue
		}
		return nil
	}
	return lastErr
}

type cmd struct {
	name string
	args []string
}

// candidates returns the ordered list of paste commands to try, given
// Method and the current OS. The first one found in PATH wins.
func (i *Injector) candidates() []cmd {
	switch runtime.GOOS {
	case "darwin":
		// osascript keystroke is the no-CGO path. CGEventPost (richer,
		// does not require System Events) is a human-handoff item.
		return []cmd{
			{"osascript", []string{"-e", `tell application "System Events" to keystroke "v" using command down`}},
		}
	case "linux":
		switch i.Method {
		case MethodXdotool:
			return []cmd{{"xdotool", []string{"key", "ctrl+v"}}}
		case MethodYdotool:
			return []cmd{{"ydotool", []string{"key", "29:1", "47:1", "47:0", "29:0"}}}
		case MethodWtype:
			return []cmd{{"wtype", []string{"-M", "ctrl", "v"}}}
		case MethodAuto, "":
			// On Wayland, prefer wtype then ydotool. On X11, xdotool.
			if os.Getenv("WAYLAND_DISPLAY") != "" {
				return []cmd{
					{"wtype", []string{"-M", "ctrl", "v"}},
					{"ydotool", []string{"key", "29:1", "47:1", "47:0", "29:0"}},
				}
			}
			return []cmd{{"xdotool", []string{"key", "ctrl+v"}}}
		}
	}
	return nil
}
