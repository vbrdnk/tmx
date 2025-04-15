package config

import (
	"errors"
	"log"
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
	config := &Config{}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return config, err
	}

	configPath := homeDir + "/.config/tmx.toml"
	if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
		log.Println("Config file not found.")
		return config, err
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal(err)
		return config, err
	}

	if _, err := toml.Decode(string(content), &config); err != nil {
		log.Fatalf("Error reading config file: %v", err)
		return config, err
	}

	return config, nil
}
