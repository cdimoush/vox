package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cdimoush/vox/config"
)

func cmdLogin() error {
	fmt.Fprint(os.Stderr, "Enter your OpenAI API key: ")

	reader := bufio.NewReader(os.Stdin)
	key, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}
	key = strings.TrimSpace(key)

	if key == "" {
		return fmt.Errorf("no key entered")
	}
	if !strings.HasPrefix(key, "sk-") {
		return fmt.Errorf("invalid key: expected it to start with sk-")
	}

	// Save to ~/.vox/config.
	if err := config.SaveAPIKey(key); err != nil {
		return fmt.Errorf("saving key: %w", err)
	}

	// Also suggest writing to the shell profile so it's available system-wide.
	profile := detectShellProfile()
	fmt.Fprintf(os.Stderr, "✓ Key saved to ~/.vox/config\n")
	if profile != "" {
		short := "~/" + filepath.Base(filepath.Dir(profile)) + "/" + filepath.Base(profile)
		if filepath.Dir(profile) == os.Getenv("HOME") {
			short = "~/" + filepath.Base(profile)
		}
		fmt.Fprintf(os.Stderr, "\nTo make it permanent in new shells, add to %s:\n", short)
		fmt.Fprintf(os.Stderr, "  export OPENAI_API_KEY=%s\n", key)
		fmt.Fprintf(os.Stderr, "Then run: source %s\n", short)
	}
	return nil
}

// detectShellProfile returns the most appropriate shell profile file to suggest.
// Preference: ~/.zshrc > ~/.bashrc > ~/.bash_profile > ~/.profile
func detectShellProfile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	candidates := []string{
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".bash_profile"),
		filepath.Join(home, ".profile"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
