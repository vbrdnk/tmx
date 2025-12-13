package session

import (
	"github.com/vbrdnk/tmx/pkg/config"
)

// This file maintains backward compatibility with the old API.
// New code should use SessionManager directly.

// ResolveSession creates a new session if it doesn't exist and then attaches to it
// Deprecated: Use SessionManager.ResolveSession instead
func ResolveSession(dir string, cfg *config.Config) {
	sm := NewSessionManager(cfg)
	sm.ResolveSession(dir)
}

// AttachToSession attaches to an existing tmux session
// Deprecated: Use SessionManager.AttachToSession instead
func AttachToSession(sessionName string) error {
	sm := NewSessionManager(nil)
	return sm.AttachToSession(sessionName)
}

// KillSession terminates a tmux session
// Deprecated: Use SessionManager.KillSession instead
func KillSession(sessionName string) error {
	sm := NewSessionManager(nil)
	return sm.KillSession(sessionName)
}

// ListSessions lists all active tmux sessions
// Deprecated: Use SessionManager.ListSessions instead
func ListSessions() error {
	sm := NewSessionManager(nil)
	return sm.ListSessions()
}
