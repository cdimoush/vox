package transcribe

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLiveAudioRoundTrip exercises the real OpenAI Whisper API with the
// committed audio fixture. Asserts the transcript is similar (not exact)
// to the canonical expected text, since Whisper output may vary slightly
// between calls (capitalization, punctuation).
//
// Gated on VOX_LIVE_TESTS=1 so default `go test ./...` stays network-free.
// Requires a real OpenAI API key in the environment. Run as:
//
//	VOX_LIVE_TESTS=1 go test ./transcribe -run LiveAudioRoundTrip -v
func TestLiveAudioRoundTrip(t *testing.T) {
	if os.Getenv("VOX_LIVE_TESTS") != "1" {
		t.Skip("set VOX_LIVE_TESTS=1 to run live tests against OpenAI")
	}

	audioPath := filepath.Join("..", "testdata", "audio", "testaudio.m4a")
	expectedPath := filepath.Join("..", "testdata", "audio", "testaudio.expected.txt")

	if _, err := os.Stat(audioPath); err != nil {
		t.Fatalf("audio fixture missing: %v", err)
	}

	got, _, err := Transcribe(context.Background(), audioPath)
	if err != nil {
		t.Fatalf("transcribe: %v", err)
	}

	expectedBytes, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("read expected: %v", err)
	}

	gotN := normalize(got)
	wantN := normalize(string(expectedBytes))
	if gotN != wantN {
		t.Errorf("transcript not similar to expected:\n  got:      %q\n  expected: %q\n  got (norm):      %q\n  expected (norm): %q",
			got, string(expectedBytes), gotN, wantN)
	}
}

// normalize strips case, punctuation, and collapses whitespace so that
// "The brown fox jumped over the lazy dog." and "the brown fox jumped
// over the lazy dog" compare equal.
func normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var sb strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z',
			r >= '0' && r <= '9',
			r == ' ', r == '\t', r == '\n':
			sb.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(sb.String()), " ")
}
