package tmux

import (
	"path/filepath"
	"strings"
	"tmx/config"
)

// createSessionName creates a valid tmux session name from a directory name
func createSessionName(dirName string) string {
	// Replace characters that tmux doesn't like in session names
	name := strings.ReplaceAll(dirName, ".", "_")
	name = strings.ReplaceAll(name, ":", "_")

	// Additional cleaning if needed
	invalidChars := []string{" ", "/", "\\", "$", "#", "&", "*", "(", ")", "{", "}", "[", "]", "@", "!"}
	for _, char := range invalidChars {
		name = strings.ReplaceAll(name, char, "_")
	}
	return name
}

// determineSessionName tries to find a matching workspace in config or falls back to dir basename
func determineSessionName(dir string, cfg *config.Config) string {
	// If no config, use directory name
	if cfg == nil {
		return createSessionName(filepath.Base(dir))
	}

	// Try to find a matching workspace
	for _, ws := range cfg.Workspace {
		if filepath.Base(dir) == filepath.Base(ws.Directory) {
			return createSessionName(ws.Name)
		}
	}

	// Default to directory name if no match found
	return createSessionName(filepath.Base(dir))
}
