package recorder

import (
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
	}{
		{
			name:      "silent",
			line:      "In:0.00% 00:00:00.00 [00:00:00.00] Out:0     [      |      ]        Clip:0",
			wantLevel: 0.0,
			wantOK:    true,
		},
		{
			name:      "full level",
			line:      "In:0.00% 00:00:01.23 [00:00:00.00] Out:16.1k [======|======]        Clip:0",
			wantLevel: 1.0,
			wantOK:    true,
		},
		{
			name:      "half level",
			line:      "In:0.00% 00:00:01.23 [00:00:00.00] Out:16.1k [===   |===   ]        Clip:0",
			wantLevel: 0.5,
			wantOK:    true,
		},
		{
			name:      "one third",
			line:      "In:0.00% 00:00:01.23 [00:00:00.00] Out:16.1k [==    |==    ]        Clip:0",
			wantLevel: 1.0 / 3.0,
			wantOK:    true,
		},
		{
			name:   "no VU meter",
			line:   "Input File     : 'default' (coreaudio)",
			wantOK: false,
		},
		{
			name:   "no pipe in brackets",
			line:   "[00:00:00.00]",
			wantOK: false,
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
			diff := level - tt.wantLevel
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.01 {
				t.Errorf("parseVolume(%q): got level=%f, want ~%f", tt.line, level, tt.wantLevel)
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
