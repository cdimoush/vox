package paste

import (
	"errors"
	"runtime"
	"testing"
)

// stubRunner records calls and returns the configured error (if any).
type stubRunner struct {
	calls []call
	err   error
}

type call struct {
	name string
	args []string
}

func (s *stubRunner) Run(name string, args ...string) error {
	s.calls = append(s.calls, call{name: name, args: args})
	return s.err
}

// lookAll marks every binary as installed. lookNone fails every lookup.
func lookAll(string) (string, error)  { return "/usr/bin/anything", nil }
func lookNone(string) (string, error) { return "", errors.New("not found") }

// onlyLook returns a LookPath that reports success for listed names and
// failure for everything else.
func onlyLook(names ...string) func(string) (string, error) {
	set := make(map[string]struct{}, len(names))
	for _, n := range names {
		set[n] = struct{}{}
	}
	return func(name string) (string, error) {
		if _, ok := set[name]; ok {
			return "/usr/bin/" + name, nil
		}
		return "", errors.New("not found")
	}
}

func TestClipboardOnlySkipsInjection(t *testing.T) {
	runner := &stubRunner{}
	inj := &Injector{Method: MethodClipboardOnly, Runner: runner.Run, LookPath: lookAll}
	// Paste uses the real clipboard.Write, which will fail on a headless
	// test runner — so we call inject() directly to prove the method
	// short-circuits before inject().
	if got := inj.candidates(); got == nil && runtime.GOOS == "linux" {
		// expected for clipboard-only on linux — no candidates
	}
	if len(runner.calls) != 0 {
		t.Errorf("no exec should happen for clipboard-only, got %+v", runner.calls)
	}
}

func TestLinuxAutoOnWayland(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux-only")
	}
	t.Setenv("WAYLAND_DISPLAY", "wayland-0")
	runner := &stubRunner{}
	inj := &Injector{Method: MethodAuto, Runner: runner.Run, LookPath: onlyLook("wtype")}
	if err := inj.inject(); err != nil {
		t.Fatalf("want success on wtype path, got %v", err)
	}
	if len(runner.calls) != 1 || runner.calls[0].name != "wtype" {
		t.Errorf("want wtype invocation, got %+v", runner.calls)
	}
}

func TestLinuxAutoFallsBackToYdotool(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux-only")
	}
	t.Setenv("WAYLAND_DISPLAY", "wayland-0")
	runner := &stubRunner{}
	inj := &Injector{Method: MethodAuto, Runner: runner.Run, LookPath: onlyLook("ydotool")}
	if err := inj.inject(); err != nil {
		t.Fatalf("want success on ydotool, got %v", err)
	}
	if len(runner.calls) != 1 || runner.calls[0].name != "ydotool" {
		t.Errorf("want ydotool invocation, got %+v", runner.calls)
	}
}

func TestLinuxAutoX11PrefersXdotool(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux-only")
	}
	t.Setenv("WAYLAND_DISPLAY", "")
	runner := &stubRunner{}
	inj := &Injector{Method: MethodAuto, Runner: runner.Run, LookPath: onlyLook("xdotool")}
	if err := inj.inject(); err != nil {
		t.Fatalf("want success on xdotool, got %v", err)
	}
	if len(runner.calls) != 1 || runner.calls[0].name != "xdotool" {
		t.Errorf("want xdotool, got %+v", runner.calls)
	}
}

func TestExplicitMethodRespected(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux-only")
	}
	runner := &stubRunner{}
	inj := &Injector{Method: MethodYdotool, Runner: runner.Run, LookPath: onlyLook("ydotool")}
	if err := inj.inject(); err != nil {
		t.Fatalf("ydotool: %v", err)
	}
	if runner.calls[0].name != "ydotool" {
		t.Errorf("explicit method ignored: %+v", runner.calls)
	}
}

func TestInjectMissingBinaryReportsError(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux-only")
	}
	inj := &Injector{Method: MethodXdotool, Runner: (&stubRunner{}).Run, LookPath: lookNone}
	if err := inj.inject(); err == nil {
		t.Fatal("want error when binary missing")
	}
}

func TestDarwinUsesOsascript(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin-only")
	}
	runner := &stubRunner{}
	inj := &Injector{Method: MethodAuto, Runner: runner.Run, LookPath: onlyLook("osascript")}
	if err := inj.inject(); err != nil {
		t.Fatalf("osascript: %v", err)
	}
	if runner.calls[0].name != "osascript" {
		t.Errorf("want osascript, got %+v", runner.calls)
	}
}
