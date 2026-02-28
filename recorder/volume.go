package recorder

import (
	"math"
	"strconv"
	"strings"
)

// parseVolume extracts a linear volume level (0.0–1.0) from a SoX stderr line.
// SoX outputs patterns like "[  -3.5dB]" or "[ -inf dB]".
func parseVolume(line string) (float64, bool) {
	open := strings.LastIndex(line, "[")
	close := strings.LastIndex(line, "]")
	if open < 0 || close < 0 || close <= open {
		return 0, false
	}

	inner := strings.TrimSpace(line[open+1 : close])

	// Handle "-inf dB" or "-inf"
	if strings.HasPrefix(inner, "-inf") {
		return 0.0, true
	}

	// Strip trailing "dB" if present.
	inner = strings.TrimSuffix(inner, "dB")
	inner = strings.TrimSuffix(inner, "dB ")
	inner = strings.TrimSpace(inner)

	db, err := strconv.ParseFloat(inner, 64)
	if err != nil {
		return 0, false
	}

	level := math.Pow(10, db/20.0)
	if level < 0 {
		level = 0
	}
	if level > 1 {
		level = 1
	}
	return level, true
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

	filled := int(math.Round(level * float64(width)))
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
