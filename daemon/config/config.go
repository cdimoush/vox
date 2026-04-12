// Package config loads the vox daemon configuration from ~/.vox/config.toml.
//
// The schema matches system-design.md §8. vox core (the CLI) does NOT read
// this file — only the daemon does. Unknown keys are ignored so older
// daemons keep working after schema additions.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

// Config is the fully-resolved daemon configuration.
type Config struct {
	Hotkey  HotkeyConfig  `toml:"hotkey"`
	Overlay OverlayConfig `toml:"overlay"`
	Audio   AudioConfig   `toml:"audio"`
	Paste   PasteConfig   `toml:"paste"`
}

type HotkeyConfig struct {
	Modifiers []string `toml:"modifiers"`
	Key       string   `toml:"key"`
}

type OverlayConfig struct {
	Enabled  bool   `toml:"enabled"`
	Position string `toml:"position"`
	Width    int    `toml:"width"`
	Height   int    `toml:"height"`
}

type AudioConfig struct {
	CuesEnabled bool    `toml:"cues_enabled"`
	CueVolume   float64 `toml:"cue_volume"`
}

type PasteConfig struct {
	AutoPaste bool   `toml:"auto_paste"`
	Method    string `toml:"method"` // auto | clipboard-only | xdotool | ydotool | wtype
}

// validMethods mirrors the enum in system-design.md §8.
var validMethods = map[string]struct{}{
	"auto":           {},
	"clipboard-only": {},
	"xdotool":        {},
	"ydotool":        {},
	"wtype":          {},
}

// Defaults returns the baseline config. Wayland overrides (overlay off) are
// applied by the daemon at runtime, not here — this stays platform-agnostic.
func Defaults() Config {
	return Config{
		Hotkey: HotkeyConfig{
			Modifiers: []string{"ctrl", "shift"},
			Key:       "space",
		},
		Overlay: OverlayConfig{
			Enabled:  true,
			Position: "top-center",
			Width:    200,
			Height:   40,
		},
		Audio: AudioConfig{
			CuesEnabled: true,
			CueVolume:   0.5,
		},
		Paste: PasteConfig{
			AutoPaste: true,
			Method:    "auto",
		},
	}
}

// DefaultPath returns ~/.vox/config.toml.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".vox", "config.toml"), nil
}

// Load reads and validates the config at path. A missing file is not an
// error — defaults are returned. An unreadable or malformed file IS.
func Load(path string) (Config, error) {
	cfg := Defaults()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return Config{}, fmt.Errorf("reading %s: %w", path, err)
	}
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing %s: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Validate checks enums and ranges. Called by Load; exported for tests.
func (c Config) Validate() error {
	if _, ok := validMethods[c.Paste.Method]; !ok {
		return fmt.Errorf("paste.method %q: must be one of auto, clipboard-only, xdotool, ydotool, wtype", c.Paste.Method)
	}
	if c.Audio.CueVolume < 0 || c.Audio.CueVolume > 1 {
		return fmt.Errorf("audio.cue_volume %.2f: must be between 0.0 and 1.0", c.Audio.CueVolume)
	}
	if c.Hotkey.Key == "" {
		return errors.New("hotkey.key: must not be empty")
	}
	return nil
}
