package tmux

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/vbrdnk/tmx/config"

	"github.com/fatih/color"
)

// ResolveSession creates a new session if it doesn't exist and then attaches to it
func ResolveSession(dir string) {
	cfg, err := config.ParseConfig()
	if err != nil {
		log.Printf("Error reading config file: %v", err)
	}

	// Determine session name
	sessionName := determineSessionName(dir, cfg)

	// Check if session exists, create if it doesn't
	if !sessionExists(sessionName) {
		newSession(sessionName, dir, cfg)
	}

	AttachToSession(sessionName)
}

func isAttached() bool {
	_, tmuxRunning := os.LookupEnv("TMUX")
	return tmuxRunning
}

// attachToSession attaches to an existing tmux session
func AttachToSession(sessionName string) {
	var tc *TmuxCommand

	tmuxRunning := isAttached()
	if !tmuxRunning {
		tc = NewTmuxCommand("attach-session", "-t", sessionName)
	} else {
		tc = NewTmuxCommand("switch-client", "-t", sessionName)
	}

	err := tc.ExecuteWithIO()
	if err != nil {
		color.Red(fmt.Sprintf("Error attaching to %s tmux session: %v\n", sessionName, err))
	}
}

func KillSession(sessionName string) {
	tc := NewTmuxCommand("kill-session", "-t", sessionName)

	err := tc.ExecuteWithIO()
	if err != nil {
		color.Red(fmt.Sprintf("Error killing %s tmux session: %v\n", sessionName, err))
	}
}

// SessionExists checks if a tmux session exists
func sessionExists(sessionName string) bool {
	tc := NewTmuxCommand("has-session", "-t", sessionName)
	return tc.Execute() == nil
}

// NewSession creates a new tmux session with the given name in the specified directory
func newSession(sessionName string, dir string, cfg *config.Config) {
	color.Green(fmt.Sprintf("Creating new session: %s in directory: %s\n", sessionName, dir))

	var commands []*TmuxCommand

	// Handle case with no config
	if cfg == nil {
		color.Green("Using default configuration (no config file found)\n")
		commands = append(commands, NewTmuxCommand("new-session", "-ds", sessionName, "-c", dir))
	} else {
		commands = createSessionCommands(sessionName, dir, cfg)
	}

	if len(commands) == 0 {
		return
	}

	for _, cmd := range commands {
		cmd.ExecuteVerbose()
	}

	color.Green(fmt.Sprintf("Successfully started tmux session: %s\n", sessionName))
}

// createSessionCommands generates commands for creating a session based on config
func createSessionCommands(sessionName string, dir string, cfg *config.Config) []*TmuxCommand {
	var commands []*TmuxCommand

	// Try to find a matching workspace
	for _, ws := range cfg.Workspace {
		if filepath.Base(dir) == filepath.Base(ws.Directory) {
			sessionName = createSessionName(ws.Name)

			// Create first window with new-session
			firstWindow := true
			for _, window := range ws.Windows {
				if firstWindow {
					commands = append(commands, NewTmuxCommand("new-session", "-ds", sessionName, "-c", dir, "-n", window))
					firstWindow = false
				} else {
					commands = append(commands, NewTmuxCommand("neww", "-t", sessionName, "-c", dir, "-n", window))
				}
			}
			return commands
		}
	}

	// No matching workspace found, create a default session
	color.Yellow("No matching workspace found. Creating default session...\n")
	return []*TmuxCommand{NewTmuxCommand("new-session", "-ds", sessionName, "-c", dir)}
}
