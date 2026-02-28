package main

import (
	"testing"
	"time"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello world this is a long string", 15, "hello world ..."},
		{"", 10, ""},
		{"exactly10!", 10, "exactly10!"},
		{"line1\nline2\nline3", 60, "line1 line2 line3"},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}

func TestRelativeTime(t *testing.T) {
	tests := []struct {
		name string
		ts   string
		want string
	}{
		{
			"5 minutes ago",
			time.Now().Add(-5 * time.Minute).UTC().Format(time.RFC3339),
			"5m ago",
		},
		{
			"3 hours ago",
			time.Now().Add(-3 * time.Hour).UTC().Format(time.RFC3339),
			"3h ago",
		},
		{
			"2 days ago",
			time.Now().Add(-2 * 24 * time.Hour).UTC().Format(time.RFC3339),
			"2d ago",
		},
		{
			"invalid timestamp",
			"invalid-timestamp",
			"invalid-timestamp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := relativeTime(tt.ts)
			if got != tt.want {
				t.Errorf("relativeTime(%q) = %q, want %q", tt.ts, got, tt.want)
			}
		})
	}
}
