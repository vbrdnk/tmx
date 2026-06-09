package git

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// initGitRepo creates a git repo at dir with one commit so worktrees can be added.
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

func TestIsWorktree(t *testing.T) {
	t.Run("NonGitDirectory", func(t *testing.T) {
		dir := t.TempDir()
		if IsWorktree(dir) {
			t.Error("expected false for non-git directory")
		}
	})

	t.Run("MainRepo", func(t *testing.T) {
		dir := t.TempDir()
		initGitRepo(t, dir)
		if IsWorktree(dir) {
			t.Error("expected false for main repo")
		}
	})

	t.Run("Worktree", func(t *testing.T) {
		mainDir := t.TempDir()
		initGitRepo(t, mainDir)

		wtDir := filepath.Join(t.TempDir(), "my-worktree")
		cmd := exec.Command("git", "worktree", "add", wtDir, "-b", "test-branch")
		cmd.Dir = mainDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git worktree add failed: %s\n%s", err, out)
		}

		if !IsWorktree(wtDir) {
			t.Error("expected true for worktree directory")
		}
	})
}

func TestMainRepoPath(t *testing.T) {
	t.Run("NonGitDirectory", func(t *testing.T) {
		dir := t.TempDir()
		_, err := MainRepoPath(dir)
		if err == nil {
			t.Error("expected error for non-git directory")
		}
	})

	t.Run("MainRepo", func(t *testing.T) {
		dir := t.TempDir()
		initGitRepo(t, dir)

		result, err := MainRepoPath(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// For a main repo, MainRepoPath should return the repo dir itself
		// Resolve symlinks for macOS /tmp -> /private/tmp
		expected, _ := filepath.EvalSymlinks(dir)
		actual, _ := filepath.EvalSymlinks(result)
		if actual != expected {
			t.Errorf("MainRepoPath() = %q, want %q", actual, expected)
		}
	})

	t.Run("Worktree", func(t *testing.T) {
		mainDir := t.TempDir()
		initGitRepo(t, mainDir)

		wtDir := filepath.Join(t.TempDir(), "my-worktree")
		cmd := exec.Command("git", "worktree", "add", wtDir, "-b", "test-branch-2")
		cmd.Dir = mainDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git worktree add failed: %s\n%s", err, out)
		}

		result, err := MainRepoPath(wtDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected, _ := filepath.EvalSymlinks(mainDir)
		actual, _ := filepath.EvalSymlinks(result)
		if actual != expected {
			t.Errorf("MainRepoPath() = %q, want %q", actual, expected)
		}
	})
}
