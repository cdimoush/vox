package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/cdimoush/vox/history"
)

func cmdClear() error {
	store := history.NewStore(history.DefaultPath())
	entries, err := store.List(0)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "No history to clear.")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Delete all %d transcriptions? [y/N] ", len(entries))

	response, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return err
	}

	if strings.TrimSpace(strings.ToLower(response)) == "y" {
		if err := store.Clear(); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "✓ History cleared")
	} else {
		fmt.Fprintln(os.Stderr, "Cancelled.")
	}

	return nil
}
