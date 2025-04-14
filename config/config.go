package config

import (
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Workspace WorkspaceConfig `toml:"workspace"`
}

type WorkspaceConfig struct {
	Panes string `toml:"panes"`
}

func ParseConfig() {
	var config Config

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	file, err := os.Open(homeDir + "/.config/tmx.toml")
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	content, err := os.ReadFile(file.Name())
	if err != nil {
		log.Fatal(err)
	}

	if _, err := toml.Decode(string(content), &config); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	fmt.Println("Config: ", config.Workspace.Panes)
}
