package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/cdimoush/vox/clipboard"
	"github.com/cdimoush/vox/history"
	"github.com/cdimoush/vox/transcribe"
)

// fileResult is the JSON output structure for --json mode.
type fileResult struct {
	Text     string  `json:"text"`
	Duration float64 `json:"duration_s"`
	Chunks   int     `json:"chunks"`
	Error    string  `json:"error,omitempty"`
}

// jsonError is returned from cmdFile when --json mode has already written
// the error as JSON to stdout. main.go should exit with the right code
// but NOT print "Error: ..." to stderr.
type jsonError struct {
	wrapped error
}

func (e *jsonError) Error() string { return e.wrapped.Error() }
func (e *jsonError) Unwrap() error { return e.wrapped }

// parseFileFlags extracts --json and --format from os.Args (after the path).
func parseFileFlags() (bool, string) {
	jsonMode := false
	format := "ogg"
	for _, arg := range os.Args[3:] {
		switch {
		case arg == "--json":
			jsonMode = true
		case strings.HasPrefix(arg, "--format="):
			format = strings.TrimPrefix(arg, "--format=")
		}
	}
	return jsonMode, format
}

func cmdFile() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("Usage: vox file <path> [--json] [--format=ogg]")
	}

	filePath := os.Args[2]
	jsonMode, format := parseFileFlags()

	// Stdin mode: read all of stdin into a temp file.
	if filePath == "-" {
		tmp, err := os.CreateTemp("", "vox-stdin-*."+format)
		if err != nil {
			return wrapErr(jsonMode, fmt.Errorf("creating temp file: %w", err))
		}
		defer os.Remove(tmp.Name())

		if _, err := io.Copy(tmp, os.Stdin); err != nil {
			tmp.Close()
			return wrapErr(jsonMode, fmt.Errorf("reading stdin: %w", err))
		}
		tmp.Close()
		filePath = tmp.Name()
	} else {
		if _, err := os.Stat(filePath); err != nil {
			return wrapErr(jsonMode, fmt.Errorf("file not found: %s", filePath))
		}
	}

	if os.Getenv("OPENAI_API_KEY") == "" {
		return wrapErr(jsonMode, transcribe.ErrNoAPIKey)
	}

	if !jsonMode {
		if _, err := clipboard.Detect(); err != nil {
			return err
		}
	}

	// Ctrl+C aborts transcription.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	go func() {
		_, ok := <-sigCh
		if !ok {
			return
		}
		cancel()
	}()

	// Spinner (only in non-JSON mode).
	var spinnerWg sync.WaitGroup
	spinnerDone := make(chan struct{})
	if !jsonMode {
		spinnerWg.Add(1)
		go func() {
			defer spinnerWg.Done()
			i := 0
			for {
				select {
				case <-spinnerDone:
					fmt.Fprintf(os.Stderr, "\r  \r")
					return
				default:
					fmt.Fprintf(os.Stderr, "\r%s Transcribing...", spinnerFrames[i%len(spinnerFrames)])
					i++
					time.Sleep(80 * time.Millisecond)
				}
			}
		}()
	}

	text, duration, err := transcribeWithContext(ctx, filePath)
	close(spinnerDone)
	spinnerWg.Wait()

	if err != nil {
		return wrapErr(jsonMode, err)
	}

	trimmed := strings.TrimSpace(text)

	// Determine chunk count from duration.
	chunks := 1
	if duration > transcribe.ChunkThreshold {
		chunks = int(duration/transcribe.ChunkDuration) + 1
	}

	if jsonMode {
		result := fileResult{
			Text:     trimmed,
			Duration: duration,
			Chunks:   chunks,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetEscapeHTML(false)
		return enc.Encode(result)
	}

	// Normal mode: raw text to stdout, UI feedback to stderr.
	fmt.Fprintln(os.Stdout, trimmed)
	fmt.Fprintf(os.Stderr, "\n\"%s\"\n\n", trimmed)

	if err := clipboard.Write(trimmed); err != nil {
		return fmt.Errorf("clipboard: %w", err)
	}
	fmt.Fprintln(os.Stderr, "✓ Copied to clipboard")

	store := history.NewStore(history.DefaultPath())
	entry := history.Entry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Text:      trimmed,
		DurationS: duration,
	}
	if err := store.Append(entry); err != nil {
		return fmt.Errorf("saving history: %w", err)
	}

	return nil
}

// wrapErr handles errors in JSON vs normal mode.
// In JSON mode: writes error JSON to stdout and returns a jsonError
// (so main.go knows to skip stderr output but still exit with correct code).
// In normal mode: returns the error as-is.
func wrapErr(jsonMode bool, err error) error {
	if !jsonMode {
		return err
	}
	result := fileResult{
		Error: err.Error(),
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(result)
	return &jsonError{wrapped: err}
}
