// Package history manages append-only JSONL transcription storage.
package history

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Entry represents a single transcription record.
type Entry struct {
	Timestamp string  `json:"ts"`
	Text      string  `json:"text"`
	DurationS float64 `json:"duration_s"`
}

// Store manages reading and writing history entries to a JSONL file.
type Store struct {
	path string
}

// NewStore creates a Store that reads and writes to the given file path.
func NewStore(path string) *Store {
	return &Store{path: path}
}

// DefaultPath returns the default history file path: ~/.vox/history.jsonl.
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".vox", "history.jsonl")
	}
	return filepath.Join(home, ".vox", "history.jsonl")
}

// Append adds an entry to the history file. It creates the parent directory
// and file if they do not exist.
func (s *Store) Append(e Entry) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(e)
}

// List returns history entries in reverse chronological order (most recent first).
// If n > 0, at most n entries are returned. If n == 0, all entries are returned.
// If the history file does not exist, an empty slice and nil error are returned.
func (s *Store) List(n int) ([]Entry, error) {
	f, err := os.Open(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Entry{}, nil
		}
		return nil, err
	}
	defer f.Close()

	var entries []Entry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var e Entry
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Reverse to most-recent-first order.
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	if n > 0 && n < len(entries) {
		entries = entries[:n]
	}

	return entries, nil
}

// Clear removes the history file. If the file does not exist, no error is returned.
func (s *Store) Clear() error {
	err := os.Remove(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
