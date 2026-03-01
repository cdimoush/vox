package transcribe

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	// ChunkDuration is the length of each chunk in seconds (5 minutes).
	ChunkDuration = 300.0
	// ChunkThreshold is the minimum duration to trigger chunking (8 minutes).
	ChunkThreshold = 480.0
)

// GetDuration returns the audio duration in seconds using soxi -D.
func GetDuration(filePath string) (float64, error) {
	out, err := exec.Command("soxi", "-D", filePath).Output()
	if err != nil {
		return 0, fmt.Errorf("soxi -D %s: %w", filePath, err)
	}
	s := strings.TrimSpace(string(out))
	dur, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing duration %q: %w", s, err)
	}
	return dur, nil
}

// ChunkFile splits an audio file into segments of ChunkDuration seconds
// using sox trim. Returns a list of temporary file paths; caller must clean up.
// The totalDuration parameter avoids re-reading duration.
func ChunkFile(filePath string, totalDuration float64) ([]string, error) {
	ext := filepath.Ext(filePath)
	if ext == "" {
		ext = ".wav"
	}

	dir := filepath.Dir(filePath)
	base := strings.TrimSuffix(filepath.Base(filePath), ext)

	var chunks []string
	for start := 0.0; start < totalDuration; start += ChunkDuration {
		idx := len(chunks)
		outPath := filepath.Join(dir, fmt.Sprintf("%s_chunk%03d%s", base, idx, ext))

		// sox input output trim <start> <duration>
		args := []string{filePath, outPath, "trim",
			strconv.FormatFloat(start, 'f', 2, 64),
			strconv.FormatFloat(ChunkDuration, 'f', 2, 64),
		}
		cmd := exec.Command("sox", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			// Clean up any chunks we already created.
			for _, c := range chunks {
				os.Remove(c)
			}
			return nil, fmt.Errorf("sox trim at %.0fs: %s: %w", start, string(out), err)
		}
		chunks = append(chunks, outPath)
	}

	return chunks, nil
}
