package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/conner/vox/clipboard"
	"github.com/conner/vox/history"
	"github.com/conner/vox/recorder"
	"github.com/conner/vox/transcribe"
)

// spinner frames for the transcription progress indicator.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func run() error {
	// Check dependencies up front.
	if os.Getenv("OPENAI_API_KEY") == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable not set\n\nSet it with:\n  export OPENAI_API_KEY=your-key")
	}
	if _, err := exec.LookPath("rec"); err != nil {
		return fmt.Errorf("rec (SoX) not found\n\nInstall with:\n  macOS:  brew install sox\n  Linux:  sudo apt-get install sox")
	}
	if _, err := clipboard.Detect(); err != nil {
		return err
	}

	// Set up two-phase signal handling:
	// Phase 1: Ctrl+C during recording stops recording → proceed to transcription
	// Phase 2: Ctrl+C during transcription aborts
	recCtx, recCancel := context.WithCancel(context.Background())
	defer recCancel()

	// Intercept SIGINT so first Ctrl+C stops recording.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	// Also stop recording when user presses Enter.
	go func() {
		buf := make([]byte, 1)
		os.Stdin.Read(buf)
		recCancel()
	}()

	// Forward first SIGINT to cancel recording context.
	go func() {
		sig, ok := <-sigCh
		if !ok {
			return
		}
		_ = sig
		recCancel()
	}()

	fmt.Fprintln(os.Stderr, "● Recording... (Enter to stop)")

	result, err := recorder.Record(recCtx)
	if err != nil {
		return fmt.Errorf("recording failed: %w", err)
	}
	defer os.Remove(result.FilePath)

	// Phase 2: second Ctrl+C aborts transcription.
	txCtx, txCancel := context.WithCancel(context.Background())
	defer txCancel()
	go func() {
		sig, ok := <-sigCh
		if !ok {
			return
		}
		_ = sig
		txCancel()
	}()

	// Start spinner.
	var spinnerWg sync.WaitGroup
	spinnerDone := make(chan struct{})
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

	// Transcribe.
	text, err := transcribeWithContext(txCtx, result.FilePath)
	close(spinnerDone)
	spinnerWg.Wait()

	if err != nil {
		return fmt.Errorf("transcription failed: %w", err)
	}

	// Print transcribed text in quotes.
	fmt.Fprintf(os.Stderr, "\n\"%s\"\n\n", strings.TrimSpace(text))

	// Copy to clipboard.
	if err := clipboard.Write(strings.TrimSpace(text)); err != nil {
		return fmt.Errorf("clipboard: %w", err)
	}
	fmt.Fprintln(os.Stderr, "✓ Copied to clipboard")

	// Save to history.
	store := history.NewStore(history.DefaultPath())
	entry := history.Entry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Text:      strings.TrimSpace(text),
		DurationS: result.Duration.Seconds(),
	}
	if err := store.Append(entry); err != nil {
		return fmt.Errorf("saving history: %w", err)
	}

	return nil
}

// transcribeWithContext runs transcription but respects context cancellation.
func transcribeWithContext(ctx context.Context, filePath string) (string, error) {
	type result struct {
		text string
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		text, err := transcribe.Transcribe(filePath)
		ch <- result{text, err}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case r := <-ch:
		return r.text, r.err
	}
}
