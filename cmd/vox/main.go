package main

import (
	"fmt"
	"os"
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
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
