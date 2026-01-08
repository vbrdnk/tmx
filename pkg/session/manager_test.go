package session

import (
	"testing"

	"github.com/vbrdnk/tmx/pkg/config"
)

func TestNewSessionManager(t *testing.T) {
	t.Run("WithNilConfig", func(t *testing.T) {
		sm := NewSessionManager(nil)
		if sm == nil {
			t.Fatal("NewSessionManager() returned nil")
		}
		if sm.config != nil {
			t.Error("Expected config to be nil")
		}
	})

	t.Run("WithConfig", func(t *testing.T) {
		cfg := &config.Config{}
		sm := NewSessionManager(cfg)
		if sm == nil {
			t.Fatal("NewSessionManager() returned nil")
		}
		if sm.config != cfg {
			t.Error("Expected config to match provided config")
		}
	})
}

func TestCreateSessionName(t *testing.T) {
	sm := NewSessionManager(nil)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple name",
			input:    "myproject",
			expected: "myproject",
		},
		{
			name:     "Name with dots",
			input:    "my.project",
			expected: "my_project",
		},
		{
			name:     "Name with colons",
			input:    "my:project",
			expected: "my_project",
		},
		{
			name:     "Name with spaces",
			input:    "my project",
			expected: "my_project",
		},
		{
			name:     "Name with special chars",
			input:    "my@project#123",
			expected: "my_project_123",
		},
		{
			name:     "Name with slashes",
			input:    "my/project/path",
			expected: "my_project_path",
		},
		{
			name:     "Complex name",
			input:    "my.project:v2.0 (beta)",
			expected: "my_project_v2_0__beta_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sm.createSessionName(tt.input)
			if result != tt.expected {
				t.Errorf("createSessionName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDetermineSessionName(t *testing.T) {
	t.Run("WithNilConfig", func(t *testing.T) {
		sm := NewSessionManager(nil)
		result := sm.determineSessionName("/path/to/myproject")

		// Should use directory basename
		expected := "myproject"
		if result != expected {
			t.Errorf("determineSessionName() = %q, want %q", result, expected)
		}
	})

	t.Run("WithMatchingWorkspace", func(t *testing.T) {
		cfg := &config.Config{
			Workspace: []config.WorkspaceConfig{
				{
					Directory: "/path/to/myproject",
					Name:      "Custom Project Name",
					Windows:   []string{"editor", "server"},
				},
			},
		}
		sm := NewSessionManager(cfg)
		result := sm.determineSessionName("/path/to/myproject")

		// Should use workspace name (sanitized)
		expected := "Custom_Project_Name"
		if result != expected {
			t.Errorf("determineSessionName() = %q, want %q", result, expected)
		}
	})

	t.Run("WithNonMatchingWorkspace", func(t *testing.T) {
		cfg := &config.Config{
			Workspace: []config.WorkspaceConfig{
				{
					Directory: "/path/to/otherproject",
					Name:      "Other Project",
					Windows:   []string{"editor"},
				},
			},
		}
		sm := NewSessionManager(cfg)
		result := sm.determineSessionName("/path/to/myproject")

		// Should fall back to directory basename
		expected := "myproject"
		if result != expected {
			t.Errorf("determineSessionName() = %q, want %q", result, expected)
		}
	})

	t.Run("WithDotInDirectoryName", func(t *testing.T) {
		sm := NewSessionManager(nil)
		result := sm.determineSessionName("/path/to/my.project")

		// Should sanitize dots
		expected := "my_project"
		if result != expected {
			t.Errorf("determineSessionName() = %q, want %q", result, expected)
		}
	})
}

func TestBuildSessionCommands(t *testing.T) {
	t.Run("WithoutMatchingWorkspace", func(t *testing.T) {
		cfg := &config.Config{
			Workspace: []config.WorkspaceConfig{},
		}
		sm := NewSessionManager(cfg)

		commands := sm.buildSessionCommands("testsession", "/path/to/project")

		// Should create a single default session command
		if len(commands) != 1 {
			t.Errorf("Expected 1 command, got %d", len(commands))
		}

		if len(commands) > 0 {
			// Check that it's a new-session command
			if len(commands[0].args) < 1 || commands[0].args[0] != "new-session" {
				t.Error("Expected first command to be new-session")
			}
		}
	})

	t.Run("WithMatchingWorkspaceAndWindows", func(t *testing.T) {
		cfg := &config.Config{
			Workspace: []config.WorkspaceConfig{
				{
					Directory: "/path/to/project",
					Name:      "My Project",
					Windows:   []string{"editor", "server", "logs"},
				},
			},
		}
		sm := NewSessionManager(cfg)

		commands := sm.buildSessionCommands("testsession", "/path/to/project")

		// Should create commands for each window (3 windows = 1 new-session + 2 neww)
		if len(commands) != 3 {
			t.Errorf("Expected 3 commands, got %d", len(commands))
		}

		if len(commands) > 0 {
			// First command should be new-session
			if commands[0].args[0] != "new-session" {
				t.Error("Expected first command to be new-session")
			}
		}

		if len(commands) > 1 {
			// Subsequent commands should be neww (new window)
			for i := 1; i < len(commands); i++ {
				if commands[i].args[0] != "neww" {
					t.Errorf("Expected command %d to be neww, got %s", i, commands[i].args[0])
				}
			}
		}
	})
}

func TestTmuxRunning(t *testing.T) {
	// This test just ensures the function works
	// The actual result depends on whether we're running in tmux
	result := TmuxRunning()

	// Result should be a boolean
	_ = result
}
