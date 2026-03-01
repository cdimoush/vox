package transcribe

import (
	"context"
	"errors"
	"testing"
)

func TestTranscribeMissingAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	_, _, err := Transcribe(context.Background(), "somefile.wav")
	if !errors.Is(err, ErrNoAPIKey) {
		t.Fatalf("expected ErrNoAPIKey, got: %v", err)
	}
}

func TestTranscribeMissingFile(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key-not-real")
	_, _, err := Transcribe(context.Background(), "/nonexistent/path/audio.wav")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
