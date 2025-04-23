package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseConfigAt(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "tmx.toml")
	t.Logf("TempDir: %s", t.TempDir())
	tomlData := `
		[[workspace]]
		directory = "/tmp"
		name = "test"
		windows = ["a", "b"]
	`

	if err := os.WriteFile(tmpFile, []byte(tomlData), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := parseConfigFile(tmpFile)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg == nil {
		t.Fatalf("expected non-nil config, got nil")
	}
	if len(cfg.Workspace) != 1 {
		t.Fatalf("expected 1 workspace, got %d", len(cfg.Workspace))
	}
	ws := cfg.Workspace[0]
	if ws.Directory != "/tmp" {
		t.Errorf("expected Directory '/tmp', got '%s'", ws.Directory)
	}
	if ws.Name != "test" {
		t.Errorf("expected Name 'test', got '%s'", ws.Name)
	}
	expectedWindows := []string{"a", "b"}
	if len(ws.Windows) != len(expectedWindows) {
		t.Errorf("expected %d windows, got %d", len(expectedWindows), len(ws.Windows))
	}
	for i, win := range expectedWindows {
		if ws.Windows[i] != win {
			t.Errorf("expected window %d to be '%s', got '%s'", i, win, ws.Windows[i])
		}
	}
}
