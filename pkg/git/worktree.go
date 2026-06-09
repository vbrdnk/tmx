package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// IsWorktree returns true if dir is a git worktree (not the main working tree).
func IsWorktree(dir string) bool {
	gitDir := gitRevParse(dir, "--git-dir")
	commonDir := gitRevParse(dir, "--git-common-dir")
	if gitDir == "" || commonDir == "" {
		return false
	}

	// Resolve to absolute paths for comparison
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(dir, gitDir)
	}
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(dir, commonDir)
	}

	gitDir = filepath.Clean(gitDir)
	commonDir = filepath.Clean(commonDir)

	return gitDir != commonDir
}

// MainRepoPath returns the working directory of the main repository for a worktree.
// For a main repo, it returns the repo's own directory.
func MainRepoPath(dir string) (string, error) {
	commonDir := gitRevParse(dir, "--git-common-dir")
	if commonDir == "" {
		return "", fmt.Errorf("not a git repository: %s", dir)
	}

	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(dir, commonDir)
	}
	commonDir = filepath.Clean(commonDir)

	// commonDir points to the .git directory of the main repo.
	// The repo root is its parent.
	return filepath.Dir(commonDir), nil
}

func gitRevParse(dir string, arg string) string {
	cmd := exec.Command("git", "-C", dir, "rev-parse", arg)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
