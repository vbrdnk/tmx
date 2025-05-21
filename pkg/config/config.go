package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// ConfigError represents an error that occurred while processing a specific config file
type ConfigError struct {
	File  string
	Error error
}

// Config represents the application configuration
type Config struct {
	Workspace []WorkspaceConfig `toml:"workspace"`
}

// WorkspaceConfig represents a single workspace configuration
type WorkspaceConfig struct {
	Directory string   `toml:"directory"`
	Name      string   `toml:"name"`
	Windows   []string `toml:"windows"`
}

// ParseConfig reads and parses all configuration files
func ParseConfig() (*Config, []ConfigError) {
	path, err := getPath()
	if err != nil {
		return nil, []ConfigError{{File: "path", Error: err}}
	}

	return parseConfigFile(path)
}

// parseConfigFile reads and parses all TOML files in the given directory
func parseConfigFile(path string) (*Config, []ConfigError) {
	config := &Config{
		Workspace: []WorkspaceConfig{},
	}

	var errors []ConfigError

	// Ensure the config directory exists
	if err := ensureConfigDir(path); err != nil {
		return config, []ConfigError{{File: path, Error: err}}
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return config, []ConfigError{{File: path, Error: err}}
	}

	for _, file := range files {
		// Skip non-TOML files and hidden files
		if !strings.HasSuffix(file.Name(), ".toml") || strings.HasPrefix(file.Name(), ".") {
			continue
		}

		filePath := filepath.Join(path, file.Name())
		tempConfig, err := parseSingleConfigFile(filePath)
		if err != nil {
			errors = append(errors, ConfigError{File: file.Name(), Error: err})
			continue
		}

		// Validate the workspace configurations
		if err := validateWorkspaceConfigs(tempConfig.Workspace); err != nil {
			errors = append(errors, ConfigError{File: file.Name(), Error: err})
			continue
		}

		config.Workspace = append(config.Workspace, tempConfig.Workspace...)
	}

	return config, errors
}

// parseSingleConfigFile reads and parses a single TOML configuration file
func parseSingleConfigFile(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	config := &Config{}
	if _, err := toml.Decode(string(content), config); err != nil {
		return nil, fmt.Errorf("failed to decode TOML: %w", err)
	}

	return config, nil
}

// validateWorkspaceConfigs validates a list of workspace configurations
func validateWorkspaceConfigs(workspaces []WorkspaceConfig) error {
	seenNames := make(map[string]bool)

	for _, ws := range workspaces {
		if err := validateWorkspaceConfig(ws); err != nil {
			return err
		}

		if seenNames[ws.Name] {
			return fmt.Errorf("duplicate workspace name: %s", ws.Name)
		}
		seenNames[ws.Name] = true
	}

	return nil
}

// validateWorkspaceConfig validates a single workspace configuration
func validateWorkspaceConfig(ws WorkspaceConfig) error {
	if ws.Name == "" {
		return fmt.Errorf("workspace name cannot be empty")
	}

	if ws.Directory == "" {
		return fmt.Errorf("workspace directory cannot be empty")
	}

	return nil
}

// ensureConfigDir ensures that the configuration directory exists
func ensureConfigDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create directory with restrictive permissions
			return os.MkdirAll(path, 0755)
		}
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("config path exists but is not a directory: %s", path)
	}

	return nil
}

// getPath returns the path to the configuration directory
func getPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".config", "tmx"), nil
}
