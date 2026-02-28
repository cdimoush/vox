package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/conner/vox/clipboard"
	"github.com/conner/vox/history"
)

func cmdCp() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("Usage: vox cp <n>")
	}

	n, err := strconv.Atoi(os.Args[2])
	if err != nil {
		return fmt.Errorf("invalid entry number: %s", os.Args[2])
	}

	store := history.NewStore(history.DefaultPath())
	entries, err := store.List(0)
	if err != nil {
		return err
	}

	if n < 1 || n > len(entries) {
		return fmt.Errorf("entry #%d not found (have %d entries)", n, len(entries))
	}

	entry := entries[n-1]
	if err := clipboard.Write(entry.Text); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "✓ Copied #%d to clipboard\n", n)
	return nil
}
