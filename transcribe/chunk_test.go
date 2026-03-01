package transcribe

import (
	"os/exec"
	"testing"
)

func hasSoX(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("sox"); err != nil {
		t.Skip("sox not installed, skipping")
	}
}

func TestGetDurationMissingFile(t *testing.T) {
	hasSoX(t)
	_, err := GetDuration("/nonexistent/audio.wav")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestChunkFileShort(t *testing.T) {
	// A file shorter than threshold should still produce one chunk.
	hasSoX(t)

	// Generate a 2-second sine wave for testing.
	tmp := t.TempDir() + "/test.wav"
	cmd := exec.Command("sox", "-n", "-r", "16000", "-c", "1", tmp, "synth", "2", "sine", "440")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to create test audio: %s: %v", string(out), err)
	}

	dur, err := GetDuration(tmp)
	if err != nil {
		t.Fatalf("GetDuration: %v", err)
	}
	if dur < 1.5 || dur > 2.5 {
		t.Fatalf("expected ~2s duration, got %.2f", dur)
	}

	// ChunkFile with a short duration should still produce one chunk.
	chunks, err := ChunkFile(tmp, dur)
	if err != nil {
		t.Fatalf("ChunkFile: %v", err)
	}
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
}

func TestChunkFileMultiple(t *testing.T) {
	hasSoX(t)

	// Generate a 620-second (>10min) audio file to test multiple chunks.
	tmp := t.TempDir() + "/long.wav"
	cmd := exec.Command("sox", "-n", "-r", "8000", "-c", "1", tmp, "synth", "620", "sine", "440")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to create test audio: %s: %v", string(out), err)
	}

	dur, err := GetDuration(tmp)
	if err != nil {
		t.Fatalf("GetDuration: %v", err)
	}

	chunks, err := ChunkFile(tmp, dur)
	if err != nil {
		t.Fatalf("ChunkFile: %v", err)
	}
	defer func() {
		for _, c := range chunks {
			exec.Command("rm", c).Run()
		}
	}()

	// 620s / 300s = 2 full chunks + 1 partial = 3 chunks.
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks for 620s audio, got %d", len(chunks))
	}
}
