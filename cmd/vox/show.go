package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/cdimoush/vox/history"
)

func cmdShow() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("Usage: vox show <n>")
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
	fmt.Fprintf(os.Stderr, "[%s]\n\n", relativeTime(entry.Timestamp))
	fmt.Println(entry.Text)
	return nil
}
