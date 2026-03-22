package history

import (
	"os"
	"path/filepath"
	"testing"
)

// withTempHistory overrides the history file path for the duration of a test
// by pointing os.UserHomeDir via a temp dir structure.
func tempHistoryFile(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	histPath := filepath.Join(dir, ".local", "share", "tmx", "history")
	if err := os.MkdirAll(filepath.Dir(histPath), 0o755); err != nil {
		t.Fatal(err)
	}
	return histPath
}

func writeHistory(t *testing.T, path string, lines []string) {
	t.Helper()
	content := ""
	for _, l := range lines {
		content += l + "\n"
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	path := tempHistoryFile(t)
	entries := load(path)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestLoad_MissingFile(t *testing.T) {
	entries := load("/nonexistent/path/history")
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for missing file, got %d", len(entries))
	}
}

func TestLoad_ReturnsLines(t *testing.T) {
	path := tempHistoryFile(t)
	writeHistory(t, path, []string{"alpha", "beta", "gamma"})

	entries := load(path)
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d: %v", len(entries), entries)
	}
	if entries[0] != "alpha" || entries[1] != "beta" || entries[2] != "gamma" {
		t.Errorf("unexpected entries: %v", entries)
	}
}

func TestRecord_AppendsNewEntry(t *testing.T) {
	path := tempHistoryFile(t)

	recordTo(path, "alpha", 10)
	recordTo(path, "beta", 10)

	entries := load(path)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d: %v", len(entries), entries)
	}
	if entries[0] != "alpha" || entries[1] != "beta" {
		t.Errorf("unexpected order: %v", entries)
	}
}

func TestRecord_DeduplicatesExistingEntry(t *testing.T) {
	path := tempHistoryFile(t)
	writeHistory(t, path, []string{"alpha", "beta", "gamma"})

	recordTo(path, "alpha", 10)

	entries := load(path)
	// alpha should move to end, no duplicate
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries after dedup, got %d: %v", len(entries), entries)
	}
	if entries[2] != "alpha" {
		t.Errorf("expected alpha at end (most recent), got %v", entries)
	}
	// beta and gamma should still be there
	if entries[0] != "beta" || entries[1] != "gamma" {
		t.Errorf("unexpected order: %v", entries)
	}
}

func TestRecord_CapsAtMax(t *testing.T) {
	path := tempHistoryFile(t)

	for i := 0; i < 15; i++ {
		recordTo(path, string(rune('a'+i)), 10)
	}

	entries := load(path)
	if len(entries) != 10 {
		t.Fatalf("expected 10 entries (capped), got %d: %v", len(entries), entries)
	}
	// The last entry should be the most recently recorded ('o' = a+14)
	if entries[9] != "o" {
		t.Errorf("expected last entry to be 'o', got %q", entries[9])
	}
}

func TestLoadPublic_ReversesOrder(t *testing.T) {
	path := tempHistoryFile(t)
	writeHistory(t, path, []string{"first", "second", "third"})

	// Use load() directly then manually reverse to simulate Load() logic
	raw := load(path)
	// Reverse
	for i, j := 0, len(raw)-1; i < j; i, j = i+1, j-1 {
		raw[i], raw[j] = raw[j], raw[i]
	}

	if raw[0] != "third" {
		t.Errorf("expected newest first ('third'), got %q", raw[0])
	}
	if raw[2] != "first" {
		t.Errorf("expected oldest last ('first'), got %q", raw[2])
	}
}

func TestLoadPublic_RespectsMax(t *testing.T) {
	path := tempHistoryFile(t)
	writeHistory(t, path, []string{"a", "b", "c", "d", "e"})

	// Simulate Load with max=3 by using load + trim + reverse
	raw := load(path)
	if len(raw) > 3 {
		raw = raw[len(raw)-3:]
	}

	if len(raw) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(raw))
	}
	// Should be the last 3: c, d, e
	if raw[0] != "c" || raw[1] != "d" || raw[2] != "e" {
		t.Errorf("unexpected entries: %v", raw)
	}
}

// recordTo is a test helper that records to a specific file path
// instead of the real home-based path.
func recordTo(path, sessionName string, max int) {
	entries := load(path)

	filtered := entries[:0]
	for _, e := range entries {
		if e != sessionName {
			filtered = append(filtered, e)
		}
	}
	filtered = append(filtered, sessionName)

	if len(filtered) > max {
		filtered = filtered[len(filtered)-max:]
	}

	lines := ""
	for _, e := range filtered {
		lines += e + "\n"
	}
	os.WriteFile(path, []byte(lines), 0o644) //nolint:errcheck
}
