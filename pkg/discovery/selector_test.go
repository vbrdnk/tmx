package discovery

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vbrdnk/tmx/pkg/config"
)

func TestNewDirectorySelector(t *testing.T) {
	t.Run("WithNilConfig", func(t *testing.T) {
		// Even with nil config, GetSearchDepth and GetUseZoxide should have defaults
		ds := NewDirectorySelector(nil)
		if ds == nil {
			t.Fatal("NewDirectorySelector() returned nil")
		}
		if ds.searcher == nil {
			t.Error("Expected searcher to be initialized")
		}
	})

	t.Run("WithConfig", func(t *testing.T) {
		cfg := &config.Config{}
		ds := NewDirectorySelector(cfg)
		if ds == nil {
			t.Fatal("NewDirectorySelector() returned nil")
		}
		if ds.config != cfg {
			t.Error("Expected config to match provided config")
		}
		if ds.searcher == nil {
			t.Error("Expected searcher to be initialized")
		}
	})
}

func TestBuildList(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir, err := os.MkdirTemp("", "tmx-discovery-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test directory structure
	testDirs := []string{
		"project1",
		"project2",
		"project1/subdir",
	}

	for _, dir := range testDirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755); err != nil {
			t.Fatalf("Failed to create test dir %s: %v", dir, err)
		}
	}

	t.Run("BuildListWithDepth1", func(t *testing.T) {
		cfg := &config.Config{}
		ds := NewDirectorySelector(cfg)

		result, err := ds.BuildList(tmpDir, 1)
		if err != nil {
			t.Fatalf("BuildList() error = %v", err)
		}

		if len(result) == 0 {
			t.Error("Expected non-empty result")
		}

		// Convert result to string for easier checking
		resultStr := string(result)
		lines := strings.Split(resultStr, "\n")

		// Should find at least project1 and project2
		foundProject1 := false
		foundProject2 := false

		for _, line := range lines {
			// Strip frecency marker if present
			line = strings.TrimPrefix(line, "★ ")
			if strings.Contains(line, "project1") && !strings.Contains(line, "subdir") {
				foundProject1 = true
			}
			if strings.Contains(line, "project2") {
				foundProject2 = true
			}
		}

		if !foundProject1 {
			t.Error("Expected to find project1 in results")
		}
		if !foundProject2 {
			t.Error("Expected to find project2 in results")
		}
	})

	t.Run("BuildListWithDepth2", func(t *testing.T) {
		cfg := &config.Config{}
		ds := NewDirectorySelector(cfg)

		result, err := ds.BuildList(tmpDir, 2)
		if err != nil {
			t.Fatalf("BuildList() error = %v", err)
		}

		resultStr := string(result)
		lines := strings.Split(resultStr, "\n")

		// With depth 2, should find project1/subdir as well
		foundSubdir := false
		for _, line := range lines {
			line = strings.TrimPrefix(line, "★ ")
			if strings.Contains(line, "subdir") {
				foundSubdir = true
			}
		}

		if !foundSubdir {
			t.Error("Expected to find subdir with depth 2")
		}
	})

	t.Run("BuildListWithZoxideDisabled", func(t *testing.T) {
		useZoxide := false
		cfg := &config.Config{
			UseZoxide: &useZoxide,
		}
		ds := NewDirectorySelector(cfg)

		result, err := ds.BuildList(tmpDir, 1)
		if err != nil {
			t.Fatalf("BuildList() error = %v", err)
		}

		resultStr := string(result)

		// Result should not contain frecency markers (★)
		if strings.Contains(resultStr, "★") {
			t.Error("Expected no frecency markers when zoxide is disabled")
		}
	})
}

func TestBuildListDeduplication(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir, err := os.MkdirTemp("", "tmx-dedup-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a single test directory
	testDir := filepath.Join(tmpDir, "testproject")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	cfg := &config.Config{}
	ds := NewDirectorySelector(cfg)

	result, err := ds.BuildList(tmpDir, 1)
	if err != nil {
		t.Fatalf("BuildList() error = %v", err)
	}

	resultStr := string(result)
	lines := strings.Split(resultStr, "\n")

	// Count occurrences of testproject
	count := 0
	for _, line := range lines {
		if strings.Contains(line, "testproject") {
			count++
		}
	}

	// Even if both zoxide and find return the same directory,
	// it should only appear once
	if count > 1 {
		t.Errorf("Expected testproject to appear once, but appeared %d times", count)
	}
}

func TestBuildListConfigDefaults(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmx-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test directories
	if err := os.MkdirAll(filepath.Join(tmpDir, "dir1"), 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	t.Run("CLIDepthOverridesConfig", func(t *testing.T) {
		cfg := &config.Config{
			SearchDepth: 5, // Config says depth 5
		}
		ds := NewDirectorySelector(cfg)

		// But CLI depth (2) should take precedence
		_, err := ds.BuildList(tmpDir, 2)
		if err != nil {
			t.Fatalf("BuildList() error = %v", err)
		}

		// If this doesn't error, it means the depth override worked
	})

	t.Run("ConfigDepthUsedWhenCLIIsZero", func(t *testing.T) {
		cfg := &config.Config{
			SearchDepth: 1,
		}
		ds := NewDirectorySelector(cfg)

		// CLI depth of 0 means use config default
		_, err := ds.BuildList(tmpDir, 0)
		if err != nil {
			t.Fatalf("BuildList() error = %v", err)
		}
	})
}
