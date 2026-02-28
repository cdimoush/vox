package history

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAppendCreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "deep", "history.jsonl")
	store := NewStore(path)

	err := store.Append(Entry{
		Timestamp: "2026-02-28T14:30:00Z",
		Text:      "hello world",
		DurationS: 1.5,
	})
	if err != nil {
		t.Fatalf("Append: %v", err)
	}

	// Verify directory was created.
	info, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("stat dir: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected directory to be created")
	}

	// Verify file exists and contains valid JSONL.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty file")
	}
}

func TestAppendAndList(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	store := NewStore(path)

	entries := []Entry{
		{Timestamp: "2026-02-28T10:00:00Z", Text: "first", DurationS: 1.0},
		{Timestamp: "2026-02-28T11:00:00Z", Text: "second", DurationS: 2.0},
		{Timestamp: "2026-02-28T12:00:00Z", Text: "third", DurationS: 3.0},
	}
	for _, e := range entries {
		if err := store.Append(e); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	// List all — should be in reverse order.
	all, err := store.List(0)
	if err != nil {
		t.Fatalf("List(0): %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(all))
	}
	if all[0].Text != "third" {
		t.Errorf("expected first element to be 'third', got %q", all[0].Text)
	}
	if all[2].Text != "first" {
		t.Errorf("expected last element to be 'first', got %q", all[2].Text)
	}

	// List 2 — should be the 2 most recent.
	top2, err := store.List(2)
	if err != nil {
		t.Fatalf("List(2): %v", err)
	}
	if len(top2) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(top2))
	}
	if top2[0].Text != "third" || top2[1].Text != "second" {
		t.Errorf("unexpected entries: %v", top2)
	}
}

func TestListEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent", "history.jsonl")
	store := NewStore(path)

	entries, err := store.List(0)
	if err != nil {
		t.Fatalf("List on nonexistent file: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty slice, got %d entries", len(entries))
	}
}

func TestClear(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	store := NewStore(path)

	if err := store.Append(Entry{Timestamp: "2026-02-28T10:00:00Z", Text: "data", DurationS: 1.0}); err != nil {
		t.Fatalf("Append: %v", err)
	}

	if err := store.Clear(); err != nil {
		t.Fatalf("Clear: %v", err)
	}

	entries, err := store.List(0)
	if err != nil {
		t.Fatalf("List after Clear: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty list after clear, got %d entries", len(entries))
	}

	// Clear again on nonexistent file should not error.
	if err := store.Clear(); err != nil {
		t.Fatalf("Clear on nonexistent: %v", err)
	}
}

func TestRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	store := NewStore(path)

	want := Entry{
		Timestamp: "2026-02-28T14:30:00Z",
		Text:      "Move the contact sensor config into YAML",
		DurationS: 12.4,
	}

	if err := store.Append(want); err != nil {
		t.Fatalf("Append: %v", err)
	}

	entries, err := store.List(0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	got := entries[0]
	if got.Timestamp != want.Timestamp {
		t.Errorf("Timestamp: got %q, want %q", got.Timestamp, want.Timestamp)
	}
	if got.Text != want.Text {
		t.Errorf("Text: got %q, want %q", got.Text, want.Text)
	}
	if got.DurationS != want.DurationS {
		t.Errorf("DurationS: got %v, want %v", got.DurationS, want.DurationS)
	}
}
