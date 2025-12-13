package search

import (
	"fmt"
	"os/exec"
	"strings"
)

// DirectorySearcher handles directory discovery using various tools
type DirectorySearcher struct {
	fdAvailable bool
}

// NewDirectorySearcher creates a new DirectorySearcher instance
func NewDirectorySearcher() *DirectorySearcher {
	return &DirectorySearcher{
		fdAvailable: isFdAvailable(),
	}
}

// Search uses fd (with fallback to find) to discover directories
func (ds *DirectorySearcher) Search(path string, depth int) ([]string, error) {
	if ds.fdAvailable {
		return ds.performFdSearch(path, depth)
	}
	return ds.performFindSearch(path, depth)
}

// QueryZoxideCache queries zoxide for frecent directories under the given path
func (ds *DirectorySearcher) QueryZoxideCache(path string) ([]string, error) {
	cmd := exec.Command("zoxide", "query", "--list")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var results []string
	lines := strings.SplitSeq(string(output), "\n")

	// Filter results to only include paths under the target path
	for line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && strings.HasPrefix(line, path) && line != path {
			results = append(results, line)
		}
	}

	// Limit to top 30 results (zoxide already returns them sorted by frecency)
	if len(results) > 30 {
		results = results[:30]
	}

	return results, nil
}

// performFdSearch uses fd to find directories
func (ds *DirectorySearcher) performFdSearch(path string, depth int) ([]string, error) {
	args := []string{
		"--type", "d",
		"--hidden",
		"--exclude", ".git",
		"--exclude", "node_modules",
		"--exclude", ".DS_Store",
	}

	// Add depth constraint if not unlimited
	if depth > 0 {
		args = append(args, "--max-depth", fmt.Sprintf("%d", depth))
	}

	args = append(args, ".", path)

	cmd := exec.Command("fd", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error executing fd command: %v", err)
	}

	return parseDirectoryOutput(output), nil
}

// performFindSearch uses classic find command
func (ds *DirectorySearcher) performFindSearch(path string, depth int) ([]string, error) {
	args := []string{
		path,
		"-mindepth", "1",
		"-type", "d",
		"-not", "-path", "*/.*",
		"-not", "-path", "*/node_modules/*",
	}

	// Add maxdepth constraint if not unlimited
	if depth > 0 {
		args = append(args, "-maxdepth", fmt.Sprintf("%d", depth))
	}

	cmd := exec.Command("find", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error executing find command: %v", err)
	}

	return parseDirectoryOutput(output), nil
}

// isFdAvailable checks if fd is installed and available in PATH
func isFdAvailable() bool {
	_, err := exec.LookPath("fd")
	return err == nil
}

// parseDirectoryOutput splits command output into directory paths
func parseDirectoryOutput(output []byte) []string {
	var results []string
	lines := strings.SplitSeq(string(output), "\n")

	for line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			results = append(results, line)
		}
	}

	return results
}
