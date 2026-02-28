package recorder

import (
	"math"
	"os/exec"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestRecDetection(t *testing.T) {
	_, err := exec.LookPath("rec")
	if err != nil {
		t.Skip("rec not installed")
	}
	// rec is available — no error expected.
}

func TestParseVolume(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantLevel float64
		wantOK    bool
		tolerance float64
	}{
		{
			name:      "normal dB value",
			line:      "In:0.00% 00:00:01.23 [  -3.5dB]  Out:...",
			wantLevel: math.Pow(10, -3.5/20.0),
			wantOK:    true,
			tolerance: 0.01,
		},
		{
			name:      "negative infinity",
			line:      "In:0.00% 00:00:00.00 [ -inf dB]",
			wantLevel: 0.0,
			wantOK:    true,
			tolerance: 0.001,
		},
		{
			name:      "zero dB",
			line:      "[  0.0dB]",
			wantLevel: 1.0,
			wantOK:    true,
			tolerance: 0.001,
		},
		{
			name:   "no brackets",
			line:   "no brackets here",
			wantOK: false,
		},
		{
			name:      "minus 20 dB",
			line:      "[ -20.0dB]",
			wantLevel: 0.1,
			wantOK:    true,
			tolerance: 0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, ok := parseVolume(tt.line)
			if ok != tt.wantOK {
				t.Fatalf("parseVolume(%q): got ok=%v, want ok=%v", tt.line, ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if math.Abs(level-tt.wantLevel) > tt.tolerance {
				t.Errorf("parseVolume(%q): got level=%f, want ~%f (±%f)", tt.line, level, tt.wantLevel, tt.tolerance)
			}
		})
	}
}

func TestRenderBar(t *testing.T) {
	tests := []struct {
		name       string
		level      float64
		width      int
		wantFilled int
		wantEmpty  int
	}{
		{"full", 1.0, 10, 10, 0},
		{"empty", 0.0, 10, 0, 10},
		{"half", 0.5, 10, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bar := renderBar(tt.level, tt.width)

			// Count runes, not bytes (block chars are multi-byte UTF-8).
			runeCount := utf8.RuneCountInString(bar)
			if runeCount != tt.width {
				t.Errorf("renderBar(%f, %d): got %d runes, want %d", tt.level, tt.width, runeCount, tt.width)
			}

			filled := strings.Count(bar, "█")
			empty := strings.Count(bar, "░")
			if filled != tt.wantFilled {
				t.Errorf("renderBar(%f, %d): got %d filled, want %d", tt.level, tt.width, filled, tt.wantFilled)
			}
			if empty != tt.wantEmpty {
				t.Errorf("renderBar(%f, %d): got %d empty, want %d", tt.level, tt.width, empty, tt.wantEmpty)
			}
		})
	}
}
