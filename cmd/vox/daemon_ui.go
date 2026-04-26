//go:build ui

package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/cdimoush/vox/daemon"
	"github.com/cdimoush/vox/daemon/config"
	"github.com/cdimoush/vox/daemon/hotkey"
	"github.com/cdimoush/vox/daemon/level"
	"github.com/cdimoush/vox/daemon/overlay"
	"github.com/cdimoush/vox/daemon/tray"
)

// startUISubsystems launches display-dependent goroutines when built
// with -tags=ui. Each subsystem is best-effort — a failure to start
// one does not bring down the daemon.
func startUISubsystems(d *daemon.Daemon, cfg config.Config, quit chan struct{}) {
	// --- Hotkey ---
	if isWayland() {
		printWaylandInstructions(cfg.Hotkey)
	} else {
		go func() {
			if err := hotkey.Run(d, cfg.Hotkey); err != nil {
				log.Printf("hotkey: %v (use `vox ui toggle` instead)", err)
			}
		}()
	}

	// --- Mic-level monitor ---
	// Runs continuously; the overlay ignores levels when hidden.
	// Uses context so it stops on daemon shutdown.
	levelCtx, levelCancel := context.WithCancel(context.Background())
	levels := level.Monitor(levelCtx)
	go func() { <-quit; levelCancel() }()

	// --- Overlay ---
	if cfg.Overlay.Enabled {
		go overlay.Run(d, levels, cfg.Overlay.Width, cfg.Overlay.Height)
	}

	// --- System tray ---
	go tray.Run(d, quit)
}

// isWayland returns true when running under a Wayland compositor.
func isWayland() bool {
	return os.Getenv("WAYLAND_DISPLAY") != ""
}

// printWaylandInstructions logs compositor keybinding setup for the
// configured hotkey. Wayland forbids global key grabs, so the user
// must point their compositor at `vox ui toggle`.
func printWaylandInstructions(cfg config.HotkeyConfig) {
	combo := formatHotkey(cfg)
	log.Printf("wayland detected — global hotkey disabled")
	log.Printf("configure your compositor to bind %s → `vox ui toggle`:", combo)
	log.Println("")
	log.Printf("  GNOME:    Settings → Keyboard → Custom Shortcuts")
	log.Printf("            Name: Vox Toggle  |  Command: vox ui toggle")
	log.Printf("  Sway:     bindsym %s exec vox ui toggle", combo)
	log.Printf("  Hyprland: bind = %s, exec, vox ui toggle", combo)
	log.Println("")
}

func formatHotkey(cfg config.HotkeyConfig) string {
	parts := make([]string, 0, len(cfg.Modifiers)+1)
	for _, m := range cfg.Modifiers {
		switch strings.ToLower(m) {
		case "ctrl":
			parts = append(parts, "Ctrl")
		case "shift":
			parts = append(parts, "Shift")
		case "alt":
			parts = append(parts, "Alt")
		case "super", "win", "mod4":
			parts = append(parts, "$mod")
		default:
			parts = append(parts, m)
		}
	}
	key := cfg.Key
	if len(key) > 0 {
		key = strings.ToUpper(key[:1]) + key[1:]
	}
	parts = append(parts, key)
	return strings.Join(parts, "+")
}
