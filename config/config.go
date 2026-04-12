// Package config handles API key discovery and vox configuration.
//
// Key discovery order:
//  1. OPENAI_API_KEY environment variable
//  2. ~/.vox/config file (written by vox login)
//  3. Shell profile files: ~/.bashrc, ~/.zshrc, ~/.bash_profile, ~/.profile
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	configDir  = ".vox"
	configFile = "config"
)

// FindAPIKey returns the OpenAI API key by searching in priority order:
//  1. OPENAI_API_KEY env var
//  2. ~/.vox/config
//  3. Shell profile files (~/.bashrc, ~/.zshrc, ~/.bash_profile, ~/.profile)
//
// Returns an empty string if no key is found anywhere.
func FindAPIKey() string {
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		return key
	}
	if key := keyFromVoxConfig(); key != "" {
		return key
	}
	if key := keyFromShellProfiles(); key != "" {
		return key
	}
	return ""
}

// keyFromVoxConfig reads OPENAI_API_KEY from ~/.vox/config.
func keyFromVoxConfig() string {
	path, err := voxConfigPath()
	if err != nil {
		return ""
	}
	return readKeyFromFile(path, "OPENAI_API_KEY")
}

// keyFromShellProfiles scans common shell profile files for an exported OPENAI_API_KEY.
func keyFromShellProfiles() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	profiles := []string{
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".bash_profile"),
		filepath.Join(home, ".profile"),
	}
	for _, p := range profiles {
		if key := parseExportFromFile(p, "OPENAI_API_KEY"); key != "" {
			return key
		}
	}
	return ""
}

// parseExportFromFile scans a shell script file for:
//
//	export KEY=value
//	export KEY="value"
//	export KEY='value'
//
// Returns the value if found, empty string otherwise.
func parseExportFromFile(path, key string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	prefix := "export " + key + "="
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		val := strings.TrimPrefix(line, prefix)
		val = stripInlineComment(val)
		val = stripQuotes(val)
		if val != "" {
			return val
		}
	}
	return ""
}

// readKeyFromFile reads a key=value config file and returns the value for the given key.
func readKeyFromFile(path, key string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	prefix := key + "="
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, prefix) {
			val := strings.TrimPrefix(line, prefix)
			val = stripQuotes(val)
			if val != "" {
				return val
			}
		}
	}
	return ""
}

// SaveAPIKey writes the API key to ~/.vox/config.
func SaveAPIKey(key string) error {
	path, err := voxConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("creating ~/.vox: %w", err)
	}

	// Read existing config, replace or append OPENAI_API_KEY line.
	lines := []string{}
	existing, err := os.ReadFile(path)
	if err == nil {
		for _, line := range strings.Split(string(existing), "\n") {
			if !strings.HasPrefix(strings.TrimSpace(line), "OPENAI_API_KEY=") {
				lines = append(lines, line)
			}
		}
	}
	lines = append(lines, "OPENAI_API_KEY="+key)

	content := strings.Join(lines, "\n")
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return os.WriteFile(path, []byte(content), 0600)
}

// voxConfigPath returns the path to ~/.vox/config.
func voxConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDir, configFile), nil
}

// stripQuotes removes surrounding single or double quotes from a string.
func stripQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') ||
			(s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// stripInlineComment removes trailing # comments from a shell value.
// e.g. `sk-abc123 # my key` → `sk-abc123`
func stripInlineComment(s string) string {
	// Only strip if not inside quotes.
	if len(s) > 0 && (s[0] == '"' || s[0] == '\'') {
		return s
	}
	if idx := strings.Index(s, " #"); idx != -1 {
		return strings.TrimSpace(s[:idx])
	}
	return s
}
