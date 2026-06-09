package session

import (
	"os"
	"os/exec"
	"path/filepath"
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
					Windows:   []config.WindowConfig{{Name: "editor"}, {Name: "server"}},
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
					Windows:   []config.WindowConfig{{Name: "editor"}},
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
					Windows:   []config.WindowConfig{{Name: "editor"}, {Name: "server"}, {Name: "logs"}},
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

	t.Run("WithWindowCommand", func(t *testing.T) {
		cfg := &config.Config{
			Workspace: []config.WorkspaceConfig{
				{
					Directory: "/path/to/project",
					Name:      "My Project",
					Windows: []config.WindowConfig{
						{Name: "editor"},
						{Name: "git", Command: "git pull"},
					},
				},
			},
		}
		sm := NewSessionManager(cfg)

		commands := sm.buildSessionCommands("testsession", "/path/to/project")

		// 2 windows + 1 run-shell (sleep) + 1 send-keys for the git window = 4 commands
		if len(commands) != 4 {
			t.Errorf("Expected 4 commands, got %d", len(commands))
		}

		// Last command should be send-keys for the git window command
		last := commands[len(commands)-1]
		if last.args[0] != "send-keys" {
			t.Errorf("Expected last command to be send-keys, got %s", last.args[0])
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

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "commit", "--allow-empty", "-m", "init"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git command %v failed: %s\n%s", args, err, out)
		}
	}
}

func createWorktree(t *testing.T, mainDir, wtDir, branch string) {
	t.Helper()
	cmd := exec.Command("git", "worktree", "add", wtDir, "-b", branch)
	cmd.Dir = mainDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git worktree add failed: %s\n%s", err, out)
	}
}

func TestFindMatchingWorkspace(t *testing.T) {
	t.Run("ExactMatch", func(t *testing.T) {
		cfg := &config.Config{
			Workspace: []config.WorkspaceConfig{
				{
					Directory: "/git/client-web",
					Name:      "client-web",
					Windows:   []config.WindowConfig{{Name: "editor"}, {Name: "server"}},
				},
			},
		}
		sm := NewSessionManager(cfg)
		ws := sm.findMatchingWorkspace("/git/client-web")
		if ws == nil {
			t.Fatal("expected workspace match")
		}
		if ws.Name != "client-web" {
			t.Errorf("expected name 'client-web', got %q", ws.Name)
		}
	})

	t.Run("NoMatch", func(t *testing.T) {
		cfg := &config.Config{
			Workspace: []config.WorkspaceConfig{
				{
					Directory: "/git/client-web",
					Name:      "client-web",
					Windows:   []config.WindowConfig{{Name: "editor"}},
				},
			},
		}
		sm := NewSessionManager(cfg)
		ws := sm.findMatchingWorkspace("/git/random-project")
		if ws != nil {
			t.Error("expected no match")
		}
	})

	t.Run("WorktreeFallback", func(t *testing.T) {
		mainDir := filepath.Join(t.TempDir(), "client-web")
		if err := os.Mkdir(mainDir, 0o755); err != nil {
			t.Fatal(err)
		}
		initGitRepo(t, mainDir)

		wtDir := filepath.Join(t.TempDir(), "client-web-my-feature")
		createWorktree(t, mainDir, wtDir, "my-feature")

		cfg := &config.Config{
			Workspace: []config.WorkspaceConfig{
				{
					Directory: mainDir,
					Name:      "client-web",
					Windows:   []config.WindowConfig{{Name: "editor"}, {Name: "server"}},
				},
			},
		}
		sm := NewSessionManager(cfg)
		ws := sm.findMatchingWorkspace(wtDir)
		if ws == nil {
			t.Fatal("expected workspace match via worktree fallback")
		}
		if ws.Name != "client-web" {
			t.Errorf("expected name 'client-web', got %q", ws.Name)
		}
	})

	t.Run("WorktreeNoConfigMatch", func(t *testing.T) {
		mainDir := filepath.Join(t.TempDir(), "unconfigured-repo")
		if err := os.Mkdir(mainDir, 0o755); err != nil {
			t.Fatal(err)
		}
		initGitRepo(t, mainDir)

		wtDir := filepath.Join(t.TempDir(), "unconfigured-repo-feature")
		createWorktree(t, mainDir, wtDir, "feature")

		cfg := &config.Config{
			Workspace: []config.WorkspaceConfig{
				{
					Directory: "/git/client-web",
					Name:      "client-web",
					Windows:   []config.WindowConfig{{Name: "editor"}},
				},
			},
		}
		sm := NewSessionManager(cfg)
		ws := sm.findMatchingWorkspace(wtDir)
		if ws != nil {
			t.Error("expected no match even with worktree fallback")
		}
	})

	t.Run("NotAWorktree", func(t *testing.T) {
		cfg := &config.Config{
			Workspace: []config.WorkspaceConfig{
				{
					Directory: "/git/client-web",
					Name:      "client-web",
					Windows:   []config.WindowConfig{{Name: "editor"}},
				},
			},
		}
		sm := NewSessionManager(cfg)
		ws := sm.findMatchingWorkspace("/git/client-web-my-feature")
		if ws != nil {
			t.Error("expected no match when not a worktree")
		}
	})

	t.Run("NilConfig", func(t *testing.T) {
		sm := NewSessionManager(nil)
		ws := sm.findMatchingWorkspace("/git/anything")
		if ws != nil {
			t.Error("expected nil with nil config")
		}
	})
}

func TestDetermineSessionNameWithWorktree(t *testing.T) {
	mainDir := filepath.Join(t.TempDir(), "client-web")
	if err := os.Mkdir(mainDir, 0o755); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, mainDir)

	wtDir := filepath.Join(t.TempDir(), "client-web-my-feature")
	createWorktree(t, mainDir, wtDir, "my-feature")

	cfg := &config.Config{
		Workspace: []config.WorkspaceConfig{
			{
				Directory: mainDir,
				Name:      "client-web",
				Windows:   []config.WindowConfig{{Name: "editor"}},
			},
		},
	}

	t.Run("WorktreeUsesDirectoryName", func(t *testing.T) {
		sm := NewSessionManager(cfg)
		result := sm.determineSessionName(wtDir)
		expected := "client-web-my-feature"
		if result != expected {
			t.Errorf("determineSessionName() = %q, want %q", result, expected)
		}
	})

	t.Run("DirectMatchUsesWorkspaceName", func(t *testing.T) {
		sm := NewSessionManager(cfg)
		result := sm.determineSessionName(mainDir)
		expected := "client-web"
		if result != expected {
			t.Errorf("determineSessionName() = %q, want %q", result, expected)
		}
	})
}

func TestBuildSessionCommandsWithWorktree(t *testing.T) {
	mainDir := filepath.Join(t.TempDir(), "client-web")
	if err := os.Mkdir(mainDir, 0o755); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, mainDir)

	wtDir := filepath.Join(t.TempDir(), "client-web-my-feature")
	createWorktree(t, mainDir, wtDir, "my-feature")

	cfg := &config.Config{
		Workspace: []config.WorkspaceConfig{
			{
				Directory: mainDir,
				Name:      "client-web",
				Windows: []config.WindowConfig{
					{Name: "editor", Command: "nvim"},
					{Name: "server"},
				},
			},
		},
	}

	t.Run("WorktreeInheritsFullConfig", func(t *testing.T) {
		sm := NewSessionManager(cfg)
		commands := sm.buildSessionCommands("client-web-my-feature", wtDir)

		// 2 windows + 1 run-shell (sleep) + 1 send-keys for nvim = 4 commands
		if len(commands) != 4 {
			t.Errorf("expected 4 commands, got %d", len(commands))
		}

		if len(commands) > 0 && commands[0].args[0] != "new-session" {
			t.Error("expected first command to be new-session")
		}

		// Verify session is created in worktree directory, not main repo
		found := false
		for _, arg := range commands[0].args {
			if arg == wtDir {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected session directory to be the worktree path, got args: %v", commands[0].args)
		}
	})
}
