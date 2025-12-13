package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/vbrdnk/tmx/pkg/config"
)

// SessionManager handles tmux session lifecycle operations
type SessionManager struct {
	config *config.Config
}

// NewSessionManager creates a new SessionManager instance
func NewSessionManager(cfg *config.Config) *SessionManager {
	return &SessionManager{
		config: cfg,
	}
}

// ResolveSession creates a new session if it doesn't exist and then attaches to it
func (sm *SessionManager) ResolveSession(dir string) error {
	// Determine session name
	sessionName := sm.determineSessionName(dir)

	// Check if session exists, create if it doesn't
	if !sm.sessionExists(sessionName) {
		if err := sm.createSession(sessionName, dir); err != nil {
			return err
		}
	}

	return sm.AttachToSession(sessionName)
}

// AttachToSession attaches to an existing tmux session
func (sm *SessionManager) AttachToSession(sessionName string) error {
	var tc *TmuxCommand
	tmuxRunning := TmuxRunning()

	if !tmuxRunning {
		tc = NewTmuxCommand("attach-session", "-t", sessionName)
	} else {
		tc = NewTmuxCommand("switch-client", "-t", sessionName)
	}

	return tc.ExecuteWithIO()
}

// KillSession terminates a tmux session
func (sm *SessionManager) KillSession(sessionName string) error {
	return NewTmuxCommand("kill-session", "-t", sessionName).ExecuteWithIO()
}

// ListSessions lists all active tmux sessions
func (sm *SessionManager) ListSessions() error {
	return NewTmuxCommand("list-sessions").ExecuteWithIO()
}

// sessionExists checks if a tmux session exists
func (sm *SessionManager) sessionExists(sessionName string) bool {
	tc := NewTmuxCommand("has-session", "-t", sessionName)
	return tc.Execute() == nil
}

// createSession creates a new tmux session with the given name in the specified directory
func (sm *SessionManager) createSession(sessionName string, dir string) error {
	color.Green(fmt.Sprintf("Creating new session: %s in directory: %s\n", sessionName, dir))

	var commands []*TmuxCommand

	// Handle case with no config
	if sm.config == nil {
		color.Green("Using default configuration (no config file found)\n")
		commands = append(commands, NewTmuxCommand("new-session", "-ds", sessionName, "-c", dir))
	} else {
		commands = sm.buildSessionCommands(sessionName, dir)
	}

	if len(commands) == 0 {
		return fmt.Errorf("no commands generated for session creation")
	}

	for _, cmd := range commands {
		if err := cmd.ExecuteVerbose(); err != nil {
			return fmt.Errorf("failed to execute session command: %w", err)
		}
	}

	color.Green(fmt.Sprintf("Successfully started tmux session: %s\n", sessionName))
	return nil
}

// buildSessionCommands generates commands for creating a session based on config
func (sm *SessionManager) buildSessionCommands(sessionName string, dir string) []*TmuxCommand {
	var commands []*TmuxCommand

	// Try to find a matching workspace
	for _, ws := range sm.config.Workspace {
		if filepath.Base(dir) == filepath.Base(ws.Directory) {
			sessionName = sm.createSessionName(ws.Name)

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

// determineSessionName tries to find a matching workspace in config or falls back to dir basename
func (sm *SessionManager) determineSessionName(dir string) string {
	// If no config, use directory name
	if sm.config == nil {
		return sm.createSessionName(filepath.Base(dir))
	}

	// Try to find a matching workspace
	for _, ws := range sm.config.Workspace {
		if filepath.Base(dir) == filepath.Base(ws.Directory) {
			return sm.createSessionName(ws.Name)
		}
	}

	// Default to directory name if no match found
	return sm.createSessionName(filepath.Base(dir))
}

// createSessionName creates a valid tmux session name from a directory name
func (sm *SessionManager) createSessionName(dirName string) string {
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

// TmuxRunning checks if currently running inside a tmux session
func TmuxRunning() bool {
	_, tmuxRunning := os.LookupEnv("TMUX")
	return tmuxRunning
}
