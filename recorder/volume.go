package recorder

import (
	"strings"
)

// parseVolume extracts a linear volume level (0.0–1.0) from a SoX -S progress line.
// SoX outputs a VU meter like: [  ====|====  ] where = chars represent level.
// In mono recording, both sides of the | mirror each other.
func parseVolume(line string) (float64, bool) {
	// Find the last VU meter bracket pair: [  ====|====  ]
	// SoX progress lines contain multiple [...] groups; the VU meter is the one with |
	lastOpen := -1
	for i := len(line) - 1; i >= 0; i-- {
		if line[i] == ']' {
			// Find matching [
			for j := i - 1; j >= 0; j-- {
				if line[j] == '[' {
					inner := line[j+1 : i]
					if strings.Contains(inner, "|") {
						lastOpen = j
						break
					}
					break
				}
			}
			if lastOpen >= 0 {
				break
			}
		}
	}
	if lastOpen < 0 {
		return 0, false
	}

	// Extract the bracket content after lastOpen
	closeIdx := strings.Index(line[lastOpen:], "]")
	if closeIdx < 0 {
		return 0, false
	}
	inner := line[lastOpen+1 : lastOpen+closeIdx]

	// Split on | to get left channel
	pipeIdx := strings.Index(inner, "|")
	if pipeIdx < 0 {
		return 0, false
	}
	left := inner[:pipeIdx]

	// Count = chars (signal) vs total width of the half
	width := len(left)
	if width == 0 {
		return 0, false
	}
	filled := strings.Count(left, "=")
	return float64(filled) / float64(width), true
}

// renderBar renders a volume bar of the given width using block characters.
// level should be in the range 0.0–1.0.
func renderBar(level float64, width int) string {
	if level < 0 {
		level = 0
	}
	if level > 1 {
		level = 1
	}

	filled := int(level*float64(width) + 0.5)
	empty := width - filled

	var b strings.Builder
	b.Grow(width * 3) // UTF-8 block chars are multi-byte
	for range filled {
		b.WriteRune('█')
	}
	for range empty {
		b.WriteRune('░')
	}
	return b.String()
}
