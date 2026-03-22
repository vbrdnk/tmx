package history

import (
	"os"
	"path/filepath"
	"strings"
)

const defaultHistoryFile = ".local/share/tmx/history"

// filePath returns the path to the history file
func filePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, defaultHistoryFile), nil
}

// Record adds sessionName to the history file, deduplicating and capping at max entries.
// Errors are silently ignored so history failures never interrupt normal workflow.
func Record(sessionName string, max int) {
	path, err := filePath()
	if err != nil {
		return
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}

	entries := load(path)

	// Remove any existing entry for this session (dedup)
	filtered := entries[:0]
	for _, e := range entries {
		if e != sessionName {
			filtered = append(filtered, e)
		}
	}

	// Append as most recent
	filtered = append(filtered, sessionName)

	// Cap to max entries (keep the tail — most recent)
	if len(filtered) > max {
		filtered = filtered[len(filtered)-max:]
	}

	content := strings.Join(filtered, "\n") + "\n"
	os.WriteFile(path, []byte(content), 0o644) //nolint:errcheck
}

// Load returns up to max recent session names, newest first.
func Load(max int) []string {
	path, err := filePath()
	if err != nil {
		return nil
	}
	entries := load(path)

	// Cap to max
	if len(entries) > max {
		entries = entries[len(entries)-max:]
	}

	// Reverse so newest is first
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	return entries
}

// load reads the history file and returns non-empty lines
func load(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var entries []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			entries = append(entries, line)
		}
	}
	return entries
}
