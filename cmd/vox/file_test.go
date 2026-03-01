package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/cdimoush/vox/transcribe"
)

func TestCmdFileNoArgs(t *testing.T) {
	orig := os.Args
	defer func() { os.Args = orig }()

	os.Args = []string{"vox", "file"}
	err := cmdFile()
	if err == nil {
		t.Fatal("expected error for missing path argument")
	}
	if !strings.Contains(err.Error(), "Usage") {
		t.Errorf("expected usage message, got: %v", err)
	}
}

func TestCmdFileNonexistentFile(t *testing.T) {
	orig := os.Args
	defer func() { os.Args = orig }()

	os.Args = []string{"vox", "file", "/nonexistent/path/audio.wav"}
	err := cmdFile()
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestCmdFileMissingAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	tmpFile, err := os.CreateTemp("", "vox-test-*.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	orig := os.Args
	defer func() { os.Args = orig }()

	os.Args = []string{"vox", "file", tmpFile.Name()}
	err = cmdFile()
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
	if !errors.Is(err, transcribe.ErrNoAPIKey) {
		t.Errorf("expected ErrNoAPIKey, got: %v", err)
	}
}

func TestCmdFileMissingAPIKeyJSON(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")

	tmpFile, err := os.CreateTemp("", "vox-test-*.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	orig := os.Args
	defer func() { os.Args = orig }()

	os.Args = []string{"vox", "file", tmpFile.Name(), "--json"}
	err = cmdFile()
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
	// Should be wrapped in jsonError but still unwrap to ErrNoAPIKey.
	if !errors.Is(err, transcribe.ErrNoAPIKey) {
		t.Errorf("expected ErrNoAPIKey through jsonError, got: %v", err)
	}
	// Should be a jsonError.
	var je *jsonError
	if !errors.As(err, &je) {
		t.Errorf("expected jsonError wrapper, got: %T", err)
	}
}

func TestExitCodeMapping(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"no api key", transcribe.ErrNoAPIKey, 3},
		{"api error", transcribe.ErrAPI, 2},
		{"wrapped api error", fmt.Errorf("transcription: %w", transcribe.ErrAPI), 2},
		{"json wrapped no api key", &jsonError{wrapped: transcribe.ErrNoAPIKey}, 3},
		{"json wrapped api error", &jsonError{wrapped: fmt.Errorf("fail: %w", transcribe.ErrAPI)}, 2},
		{"general error", errors.New("file not found"), 1},
		{"json wrapped general", &jsonError{wrapped: errors.New("bad format")}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exitCode(tt.err)
			if got != tt.want {
				t.Errorf("exitCode(%v) = %d, want %d", tt.err, got, tt.want)
			}
		})
	}
}

func TestParseFileFlags(t *testing.T) {
	orig := os.Args
	defer func() { os.Args = orig }()

	os.Args = []string{"vox", "file", "test.ogg", "--json", "--format=mp3"}
	jsonMode, format := parseFileFlags()
	if !jsonMode {
		t.Error("expected jsonMode=true")
	}
	if format != "mp3" {
		t.Errorf("expected format=mp3, got %s", format)
	}
}

func TestParseFileFlagsDefaults(t *testing.T) {
	orig := os.Args
	defer func() { os.Args = orig }()

	os.Args = []string{"vox", "file", "test.ogg"}
	jsonMode, format := parseFileFlags()
	if jsonMode {
		t.Error("expected jsonMode=false")
	}
	if format != "ogg" {
		t.Errorf("expected format=ogg, got %s", format)
	}
}

func TestJsonErrorUnwrap(t *testing.T) {
	inner := transcribe.ErrNoAPIKey
	je := &jsonError{wrapped: inner}

	if !errors.Is(je, transcribe.ErrNoAPIKey) {
		t.Error("jsonError should unwrap to ErrNoAPIKey")
	}

	// Verify main.go would suppress stderr.
	var check *jsonError
	if !errors.As(je, &check) {
		t.Error("errors.As should find jsonError")
	}
}
