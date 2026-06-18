package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultsAreValid(t *testing.T) {
	if err := Defaults().Validate(); err != nil {
		t.Fatalf("defaults invalid: %v", err)
	}
}

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "nope.toml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Hotkey.Key != "space" {
		t.Errorf("want default key=space, got %q", cfg.Hotkey.Key)
	}
	if !cfg.Paste.AutoPaste {
		t.Errorf("want default AutoPaste=true")
	}
}

func TestLoadOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	body := `
[hotkey]
modifiers = ["alt"]
key = "m"

[paste]
auto_paste = false
method = "wtype"
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Hotkey.Key != "m" || len(cfg.Hotkey.Modifiers) != 1 || cfg.Hotkey.Modifiers[0] != "alt" {
		t.Errorf("hotkey override not applied: %+v", cfg.Hotkey)
	}
	if cfg.Paste.Method != "wtype" || cfg.Paste.AutoPaste {
		t.Errorf("paste override not applied: %+v", cfg.Paste)
	}
	// Overlay untouched → should keep defaults.
	if cfg.Overlay.Position != "top-center" {
		t.Errorf("overlay defaults lost: %+v", cfg.Overlay)
	}
}

func TestLoadInvalidMethodRejected(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	os.WriteFile(path, []byte(`[paste]`+"\n"+`method = "telepathy"`), 0o644)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid paste.method, got nil")
	}
}

func TestLoadInvalidVolumeRejected(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	os.WriteFile(path, []byte(`[audio]`+"\n"+`cue_volume = 2.0`), 0o644)
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for out-of-range volume")
	}
}

func TestLoadMalformedToml(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	os.WriteFile(path, []byte("this is = not [ valid toml"), 0o644)
	if _, err := Load(path); err == nil {
		t.Fatal("expected parse error")
	}
}
