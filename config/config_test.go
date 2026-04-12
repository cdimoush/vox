package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindAPIKey_EnvVar(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-env-key")
	got := FindAPIKey()
	if got != "sk-env-key" {
		t.Errorf("expected sk-env-key, got %q", got)
	}
}

func TestFindAPIKey_VoxConfig(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "") // clear env
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := filepath.Join(home, ".vox")
	os.MkdirAll(dir, 0700)
	os.WriteFile(filepath.Join(dir, "config"), []byte("OPENAI_API_KEY=sk-from-vox-config\n"), 0600)

	got := FindAPIKey()
	if got != "sk-from-vox-config" {
		t.Errorf("expected sk-from-vox-config, got %q", got)
	}
}

func TestFindAPIKey_Bashrc(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "") // clear env
	home := t.TempDir()
	t.Setenv("HOME", home)

	os.WriteFile(filepath.Join(home, ".bashrc"), []byte(`
# shell config
export PATH=/usr/bin:$PATH
export OPENAI_API_KEY=sk-from-bashrc
export OTHER=stuff
`), 0644)

	got := FindAPIKey()
	if got != "sk-from-bashrc" {
		t.Errorf("expected sk-from-bashrc, got %q", got)
	}
}

func TestFindAPIKey_NotFound(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	home := t.TempDir()
	t.Setenv("HOME", home)

	got := FindAPIKey()
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFindAPIKey_EnvTakesPrecedenceOverConfig(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-env-wins")
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := filepath.Join(home, ".vox")
	os.MkdirAll(dir, 0700)
	os.WriteFile(filepath.Join(dir, "config"), []byte("OPENAI_API_KEY=sk-config-loses\n"), 0600)

	got := FindAPIKey()
	if got != "sk-env-wins" {
		t.Errorf("expected sk-env-wins, got %q", got)
	}
}

func TestParseExportFromFile_Variants(t *testing.T) {
	cases := []struct {
		line string
		want string
	}{
		{`export OPENAI_API_KEY=sk-plain`, "sk-plain"},
		{`export OPENAI_API_KEY="sk-double-quoted"`, "sk-double-quoted"},
		{`export OPENAI_API_KEY='sk-single-quoted'`, "sk-single-quoted"},
		{`export OPENAI_API_KEY=sk-with-comment # my key`, "sk-with-comment"},
		{`export OTHER=something`, ""},
		{`# export OPENAI_API_KEY=sk-commented-out`, ""},
	}

	for _, tc := range cases {
		f, _ := os.CreateTemp("", "profile-*.sh")
		f.WriteString(tc.line + "\n")
		f.Close()

		got := parseExportFromFile(f.Name(), "OPENAI_API_KEY")
		os.Remove(f.Name())

		if got != tc.want {
			t.Errorf("line %q: expected %q, got %q", tc.line, tc.want, got)
		}
	}
}

func TestSaveAPIKey(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := SaveAPIKey("sk-saved"); err != nil {
		t.Fatalf("SaveAPIKey: %v", err)
	}

	// Read it back.
	got := keyFromVoxConfig()
	if got != "sk-saved" {
		t.Errorf("expected sk-saved, got %q", got)
	}
}

func TestSaveAPIKey_Overwrites(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	SaveAPIKey("sk-old")
	SaveAPIKey("sk-new")

	got := keyFromVoxConfig()
	if got != "sk-new" {
		t.Errorf("expected sk-new, got %q", got)
	}
}
