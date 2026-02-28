package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/cdimoush/vox/clipboard"
	"github.com/cdimoush/vox/history"
)

func cmdFile() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("Usage: vox file <path>")
	}

	filePath := os.Args[2]

	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("file not found: %s", filePath)
	}

	if os.Getenv("OPENAI_API_KEY") == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable not set\n\nSet it with:\n  export OPENAI_API_KEY=your-key")
	}

	if _, err := clipboard.Detect(); err != nil {
		return err
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

	// Spinner.
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

	text, err := transcribeWithContext(ctx, filePath)
	close(spinnerDone)
	spinnerWg.Wait()

	if err != nil {
		return fmt.Errorf("transcription failed: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\n\"%s\"\n\n", strings.TrimSpace(text))

	if err := clipboard.Write(strings.TrimSpace(text)); err != nil {
		return fmt.Errorf("clipboard: %w", err)
	}
	fmt.Fprintln(os.Stderr, "✓ Copied to clipboard")

	store := history.NewStore(history.DefaultPath())
	entry := history.Entry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Text:      strings.TrimSpace(text),
		DurationS: 0,
	}
	if err := store.Append(entry); err != nil {
		return fmt.Errorf("saving history: %w", err)
	}

	return nil
}
