// Package recorder wraps SoX rec for audio capture.
package recorder

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// Result holds the output of a recording session.
type Result struct {
	FilePath string
	Duration time.Duration
}

// Record captures audio using SoX rec until the context is cancelled.
// The caller is responsible for deleting the temporary WAV file when done.
func Record(ctx context.Context) (Result, error) {
	if _, err := exec.LookPath("rec"); err != nil {
		return Result{}, fmt.Errorf("rec (SoX) not found\n\nInstall with:\n  macOS:  brew install sox\n  Linux:  sudo apt-get install sox")
	}

	tmpFile, err := os.CreateTemp("", "vox-*.wav")
	if err != nil {
		return Result{}, fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()

	cmd := exec.CommandContext(ctx, "rec", "-S", "-r", "16000", "-c", "1", "-b", "16", tmpPath)

	// Send SIGINT instead of SIGKILL so rec can finalize the WAV header.
	cmd.Cancel = func() error {
		return cmd.Process.Signal(os.Interrupt)
	}
	cmd.WaitDelay = 3 * time.Second

	stderr, err := cmd.StderrPipe()
	if err != nil {
		os.Remove(tmpPath)
		return Result{}, fmt.Errorf("getting stderr pipe: %w", err)
	}

	start := time.Now()

	if err := cmd.Start(); err != nil {
		os.Remove(tmpPath)
		return Result{}, fmt.Errorf("starting rec: %w", err)
	}

	// Read stderr in a goroutine to display volume meter.
	// SoX progress uses \r (not \n) between updates, so we split on both.
	go func() {
		scanner := bufio.NewScanner(stderr)
		scanner.Split(scanCRLF)
		for scanner.Scan() {
			line := scanner.Text()
			if level, ok := parseVolume(line); ok {
				bar := renderBar(level, 30)
				fmt.Fprintf(os.Stderr, "\r  %s", bar)
			}
		}
	}()

	_ = cmd.Wait()
	elapsed := time.Since(start)

	// Clear the volume bar line.
	fmt.Fprintln(os.Stderr)

	return Result{FilePath: tmpPath, Duration: elapsed}, nil
}

// scanCRLF is a bufio.SplitFunc that splits on \n or \r.
// SoX progress output uses \r between updates and \n for headers.
func scanCRLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	// Find earliest \r or \n.
	cr := bytes.IndexByte(data, '\r')
	lf := bytes.IndexByte(data, '\n')

	switch {
	case cr >= 0 && (lf < 0 || cr < lf):
		return cr + 1, data[:cr], nil
	case lf >= 0:
		return lf + 1, data[:lf], nil
	}

	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil // request more data
}
