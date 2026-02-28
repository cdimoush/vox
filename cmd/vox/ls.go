package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/cdimoush/vox/history"
)

func cmdLs() error {
	n := 20
	args := os.Args[2:]

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--all":
			n = 0
		case "-n":
			if i+1 >= len(args) {
				return fmt.Errorf("-n requires a number\n\nUsage: vox ls [-n N] [--all]")
			}
			i++
			val, err := strconv.Atoi(args[i])
			if err != nil || val < 1 {
				return fmt.Errorf("invalid value for -n: %s\n\nUsage: vox ls [-n N] [--all]", args[i])
			}
			n = val
		default:
			return fmt.Errorf("unknown flag: %s\n\nUsage: vox ls [-n N] [--all]", args[i])
		}
	}

	store := history.NewStore(history.DefaultPath())
	entries, err := store.List(n)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "No history yet.")
		return nil
	}

	fmt.Fprintf(os.Stdout, "%-4s%-12s%s\n", "#", "When", "Text")
	for i, e := range entries {
		when := relativeTime(e.Timestamp)
		text := truncate(e.Text, 60)
		fmt.Fprintf(os.Stdout, "%-4d%-12s%s\n", i+1, when, text)
	}

	return nil
}
