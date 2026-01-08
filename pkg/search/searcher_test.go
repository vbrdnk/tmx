package search

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewDirectorySearcher(t *testing.T) {
	searcher := NewDirectorySearcher()

	if searcher == nil {
		t.Fatal("NewDirectorySearcher() returned nil")
	}

	// Should cache fd availability
	if !searcher.fdAvailable && isFdAvailable() {
		t.Error("Expected fdAvailable to be true when fd is installed")
	}
}

func TestSearchWithFind(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir, err := os.MkdirTemp("", "tmx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test directory structure
	testDirs := []string{
		"dir1",
		"dir2",
		"dir1/subdir1",
	}

	for _, dir := range testDirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0o755); err != nil {
			t.Fatalf("Failed to create test dir %s: %v", dir, err)
		}
	}

	searcher := NewDirectorySearcher()

	t.Run("SearchWithDepth1", func(t *testing.T) {
		results, err := searcher.Search(tmpDir, 1)
		if err != nil {
			t.Fatalf("SearchWithFind() error = %v", err)
		}

		// Should find dir1 and dir2, but not dir1/subdir1
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d: %v", len(results), results)
		}
	})

	t.Run("SearchWithDepth2", func(t *testing.T) {
		results, err := searcher.Search(tmpDir, 2)
		if err != nil {
			t.Fatalf("SearchWithFind() error = %v", err)
		}

		// Should find all 3 directories
		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d: %v", len(results), results)
		}
	})

	t.Run("SearchWithUnlimitedDepth", func(t *testing.T) {
		results, err := searcher.Search(tmpDir, 0)
		if err != nil {
			t.Fatalf("SearchWithFind() error = %v", err)
		}

		// Should find all directories with unlimited depth
		if len(results) < 3 {
			t.Errorf("Expected at least 3 results with unlimited depth, got %d", len(results))
		}
	})
}

func TestSearchWithZoxide(t *testing.T) {
	searcher := NewDirectorySearcher()

	// This test will pass silently if zoxide is not installed
	// We're mainly testing that the method doesn't panic
	t.Run("SearchWithZoxide", func(t *testing.T) {
		// Use a path that's unlikely to have zoxide results
		results, err := searcher.QueryZoxideCache("/nonexistent/path/for/testing")

		// It's okay if zoxide returns an error (not installed) or empty results
		if err == nil && results != nil {
			// If zoxide is installed, results should be a valid slice
			if results == nil {
				t.Error("Expected non-nil results when no error")
			}
		}
	})
}

func TestParseDirectoryOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected int
	}{
		{
			name:     "Empty output",
			input:    []byte(""),
			expected: 0,
		},
		{
			name:     "Single directory",
			input:    []byte("/path/to/dir\n"),
			expected: 1,
		},
		{
			name:     "Multiple directories",
			input:    []byte("/path/to/dir1\n/path/to/dir2\n/path/to/dir3\n"),
			expected: 3,
		},
		{
			name:     "With trailing whitespace",
			input:    []byte("/path/to/dir1  \n/path/to/dir2\t\n"),
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := parseDirectoryOutput(tt.input)
			if len(results) != tt.expected {
				t.Errorf("parseDirectoryOutput() = %d results, want %d", len(results), tt.expected)
			}
		})
	}
}

func TestIsFdAvailable(t *testing.T) {
	// This test just ensures the function doesn't panic
	// The actual result depends on the system
	result := isFdAvailable()

	// Result should be a boolean
	_ = result
}
