package utils

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

func GetWorkingDirPath(cmd *cli.Command) (string, error) {
	if cmd.Args().Present() && cmd.Args().First() != "" {
		return cmd.Args().First(), nil
	}

	// Home dir serves as a fallback if no argument is provided
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %v", err)
	}

	return homeDir, nil
}
