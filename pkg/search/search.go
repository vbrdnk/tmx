package search

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetFindResults uses fd (with fallback to find) to discover directories
func GetFindResults(path string, depth int) ([]string, error) {
	// Try fd first
	if isFdAvailable() {
		return getFdResults(path, depth)
	}

	// Fallback to find
	return getClassicFindResults(path, depth)
}

// isFdAvailable checks if fd is installed and available in PATH
func isFdAvailable() bool {
	_, err := exec.LookPath("fd")
	return err == nil
}

// GetZoxideResults queries zoxide for frecent directories under the given path
func GetZoxideResults(path string) ([]string, error) {
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

// getFdResults uses fd to find directories
func getFdResults(path string, depth int) ([]string, error) {
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

// getClassicFindResults uses classic find command
func getClassicFindResults(path string, depth int) ([]string, error) {
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
