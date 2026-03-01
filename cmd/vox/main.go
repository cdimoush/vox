package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/cdimoush/vox/transcribe"
)

const version = "v0.1.0-dev"

func main() {
	var err error
	if len(os.Args) < 2 {
		err = run()
	} else {
		switch os.Args[1] {
		case "file":
			err = cmdFile()
		case "ls":
			err = cmdLs()
		case "cp":
			err = cmdCp()
		case "show":
			err = cmdShow()
		case "clear":
			err = cmdClear()
		case "--version", "-v":
			fmt.Println("vox " + version)
			return
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\nUsage: vox [file|ls|cp|show|clear]\n", os.Args[1])
			os.Exit(1)
		}
	}
	if err != nil {
		// In JSON mode, cmdFile already wrote JSON to stdout.
		// Only print to stderr for non-JSON errors.
		var je *jsonError
		if !errors.As(err, &je) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(exitCode(err))
	}
}

// exitCode maps errors to exit codes:
//
//	0 = success
//	1 = general error (file not found, bad args, bad format)
//	2 = API error (transcription failed, rate limit, timeout)
//	3 = no API key
func exitCode(err error) int {
	if errors.Is(err, transcribe.ErrNoAPIKey) {
		return 3
	}
	if errors.Is(err, transcribe.ErrAPI) {
		return 2
	}
	return 1
}
