package config

import (
	"errors"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Workspace []WorkspaceConfig `toml:"workspace"`
}

type WorkspaceConfig struct {
	Directory string   `toml:"directory"`
	Name      string   `toml:"name"`
	Windows   []string `toml:"windows"`
}

func ParseConfig() (*Config, error) {
	path, err := getPath()
	if err != nil {
		return &Config{}, err
	}

	return parseConfigFile(path)
}

func parseConfigFile(path string) (*Config, error) {
	config := &Config{}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return config, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}

	if _, err := toml.Decode(string(content), &config); err != nil {
		return config, err
	}

	return config, nil
}

func getPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return homeDir + "/.config/tmx.toml", nil
}
