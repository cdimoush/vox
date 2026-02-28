package main

import (
	"os"
	"strings"
	"testing"
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
	if !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Errorf("expected API key error, got: %v", err)
	}
}
