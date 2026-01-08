package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseConfigAt(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tmx.toml")
	t.Logf("TempDir: %s", tmpDir)
	tomlData := `
		[[workspace]]
		directory = "/tmp"
		name = "test"
		windows = ["a", "b"]
	`

	if err := os.WriteFile(tmpFile, []byte(tomlData), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, errors := parseConfigFile(tmpDir)
	if len(errors) > 0 {
		t.Fatalf("expected no error, got: %v", errors)
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

func TestApplyDefaults(t *testing.T) {
	tests := []struct {
		name           string
		config         *Config
		expectedZoxide bool
	}{
		{
			name:           "Empty config gets default zoxide=true",
			config:         &Config{},
			expectedZoxide: true,
		},
		{
			name: "Disabled zoxide is preserved",
			config: func() *Config {
				disabled := false
				return &Config{UseZoxide: &disabled}
			}(),
			expectedZoxide: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyDefaults(tt.config)

			if tt.config.GetUseZoxide() != tt.expectedZoxide {
				t.Errorf("expected UseZoxide %v, got %v", tt.expectedZoxide, tt.config.GetUseZoxide())
			}
		})
	}
}

func TestGetSearchDepth(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		cliDepth    int
		expectedVal int
	}{
		{
			name:        "CLI flag takes precedence",
			config:      &Config{SearchDepth: 2},
			cliDepth:    5,
			expectedVal: 5,
		},
		{
			name:        "Config value used when CLI is 0",
			config:      &Config{SearchDepth: 3},
			cliDepth:    0,
			expectedVal: 3,
		},
		{
			name:        "Default to 1 when both are 0",
			config:      &Config{SearchDepth: 0},
			cliDepth:    0,
			expectedVal: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetSearchDepth(tt.cliDepth)
			if result != tt.expectedVal {
				t.Errorf("expected %d, got %d", tt.expectedVal, result)
			}
		})
	}
}

func TestGetUseZoxide(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectedVal bool
	}{
		{
			name:        "Nil pointer defaults to true",
			config:      &Config{UseZoxide: nil},
			expectedVal: true,
		},
		{
			name: "Explicit true is preserved",
			config: func() *Config {
				enabled := true
				return &Config{UseZoxide: &enabled}
			}(),
			expectedVal: true,
		},
		{
			name: "Explicit false is preserved",
			config: func() *Config {
				disabled := false
				return &Config{UseZoxide: &disabled}
			}(),
			expectedVal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetUseZoxide()
			if result != tt.expectedVal {
				t.Errorf("expected %v, got %v", tt.expectedVal, result)
			}
		})
	}
}

func TestParseConfigWithSearchOptions(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "tmx.toml")
	tomlData := `search_depth = 3
use_zoxide = false

[[workspace]]
directory = "/tmp"
name = "test"
windows = ["a", "b"]
`

	if err := os.WriteFile(tmpFile, []byte(tomlData), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, errors := parseConfigFile(tmpDir)
	if len(errors) > 0 {
		t.Fatalf("expected no errors, got: %v", errors)
	}

	// Manually verify TOML parsing
	if cfg.SearchDepth != 3 {
		t.Errorf("expected SearchDepth 3 from TOML, got %d", cfg.SearchDepth)
	}

	if cfg.UseZoxide == nil {
		t.Errorf("expected UseZoxide to be set from TOML, got nil")
	} else if *cfg.UseZoxide != false {
		t.Errorf("expected UseZoxide false from TOML, got true")
	}
}
