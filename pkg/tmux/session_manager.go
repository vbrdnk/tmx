package tmux

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vbrdnk/tmx/pkg/config"

	"github.com/fatih/color"
	"github.com/vbrdnk/tmx/internal/utils"
)

// ResolveSession creates a new session if it doesn't exist and then attaches to it
func ResolveSession(dir string, cfg *config.Config) {
	// Determine session name
	sessionName := utils.DetermineSessionName(dir, cfg)

	// Check if session exists, create if it doesn't
	if !sessionExists(sessionName) {
		newSession(sessionName, dir, cfg)
	}

	AttachToSession(sessionName)
}

func TmuxRunning() bool {
	_, tmuxRunning := os.LookupEnv("TMUX")
	return tmuxRunning
}

// attachToSession attaches to an existing tmux session
func AttachToSession(sessionName string) error {
	var tc *TmuxCommand
	tmuxRunning := TmuxRunning()

	if !tmuxRunning {
		tc = NewTmuxCommand("attach-session", "-t", sessionName)
	} else {
		tc = NewTmuxCommand("switch-client", "-t", sessionName)
	}

	return tc.ExecuteWithIO()
}

func KillSession(sessionName string) error {
	return NewTmuxCommand("kill-session", "-t", sessionName).ExecuteWithIO()
}

func ListSessions() error {
	return NewTmuxCommand("list-sessions").ExecuteWithIO()
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
			sessionName = utils.CreateSessionName(ws.Name)

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
