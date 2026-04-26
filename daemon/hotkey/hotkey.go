//go:build ui && linux

// Package hotkey registers a global hotkey that calls daemon.Toggle().
//
// Uses a custom X11 implementation that handles NumLock/CapsLock/ScrollLock
// modifier states — the golang.design/x/hotkey library does not.
package hotkey

/*
#cgo LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/keysym.h>

// Lock modifier masks that must be combined with the target modifiers
// to ensure the grab works regardless of NumLock/CapsLock/ScrollLock state.
static unsigned int lockMasks[] = {
	0,
	LockMask,                        // CapsLock
	Mod2Mask,                        // NumLock
	LockMask | Mod2Mask,             // CapsLock + NumLock
};
static int nLockMasks = 4;

static Display *dpy;
static Window   root;

int hk_init() {
	dpy = XOpenDisplay(NULL);
	if (!dpy) return -1;
	root = DefaultRootWindow(dpy);
	return 0;
}

int hk_grab(unsigned int mod, unsigned int keysym) {
	int keycode = XKeysymToKeycode(dpy, keysym);
	if (!keycode) return -1;
	for (int i = 0; i < nLockMasks; i++) {
		XGrabKey(dpy, keycode, mod | lockMasks[i], root,
			False, GrabModeAsync, GrabModeAsync);
	}
	XSelectInput(dpy, root, KeyPressMask | KeyReleaseMask);
	XFlush(dpy);
	return keycode;
}

void hk_ungrab(unsigned int mod, int keycode) {
	for (int i = 0; i < nLockMasks; i++) {
		XUngrabKey(dpy, keycode, mod | lockMasks[i], root);
	}
	XFlush(dpy);
}

// hk_wait blocks until a KeyPress event arrives. Returns 0 on success.
int hk_wait() {
	XEvent ev;
	while (1) {
		XNextEvent(dpy, &ev);
		if (ev.type == KeyPress) return 0;
		// Ignore KeyRelease and other events.
	}
}

void hk_close() {
	if (dpy) XCloseDisplay(dpy);
}
*/
import "C"

import (
	"fmt"
	"log"
	"strings"

	"github.com/cdimoush/vox/daemon"
	"github.com/cdimoush/vox/daemon/config"
)

// Run registers the global hotkey described by cfg and blocks forever,
// calling d.Toggle() on every keydown. Returns on registration failure.
func Run(d *daemon.Daemon, cfg config.HotkeyConfig) error {
	mod, err := parseMods(cfg.Modifiers)
	if err != nil {
		return fmt.Errorf("hotkey config: %w", err)
	}
	keysym, err := parseKey(cfg.Key)
	if err != nil {
		return fmt.Errorf("hotkey config: %w", err)
	}

	if C.hk_init() != 0 {
		return fmt.Errorf("cannot open X11 display")
	}
	defer C.hk_close()

	keycode := C.hk_grab(C.uint(mod), C.uint(keysym))
	if keycode < 0 {
		return fmt.Errorf("cannot grab hotkey (keysym 0x%x)", keysym)
	}
	defer C.hk_ungrab(C.uint(mod), keycode)

	log.Printf("hotkey registered: %s (mod=0x%x keysym=0x%x, NumLock-aware)",
		formatCombo(cfg), mod, keysym)

	for {
		if C.hk_wait() != 0 {
			return fmt.Errorf("X11 event error")
		}
		d.Toggle()
	}
}

func parseMods(names []string) (uint32, error) {
	var mod uint32
	for _, name := range names {
		switch strings.ToLower(name) {
		case "ctrl", "control":
			mod |= C.ControlMask
		case "shift":
			mod |= C.ShiftMask
		case "alt":
			mod |= C.Mod1Mask
		case "super", "win", "mod4":
			mod |= C.Mod4Mask
		default:
			return 0, fmt.Errorf("unknown modifier %q (valid: ctrl, shift, alt, super)", name)
		}
	}
	return mod, nil
}

func parseKey(name string) (uint32, error) {
	switch strings.ToLower(name) {
	case "space":
		return C.XK_space, nil
	case "return", "enter":
		return C.XK_Return, nil
	case "tab":
		return C.XK_Tab, nil
	case "escape", "esc":
		return C.XK_Escape, nil
	case "delete":
		return C.XK_Delete, nil
	default:
		if len(name) == 1 {
			c := name[0]
			if c >= 'a' && c <= 'z' {
				return uint32(c), nil
			}
			if c >= 'A' && c <= 'Z' {
				return uint32(c + 32), nil
			}
			if c >= '0' && c <= '9' {
				return uint32(c), nil
			}
		}
		return 0, fmt.Errorf("unknown key %q", name)
	}
}

func formatCombo(cfg config.HotkeyConfig) string {
	parts := make([]string, 0, len(cfg.Modifiers)+1)
	for _, m := range cfg.Modifiers {
		parts = append(parts, strings.Title(strings.ToLower(m)))
	}
	parts = append(parts, strings.Title(cfg.Key))
	return strings.Join(parts, "+")
}
