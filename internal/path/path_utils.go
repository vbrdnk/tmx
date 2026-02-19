package path

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

func GetWorkingDirPath(cmd *cli.Command) (string, error) {
	if cmd.Args().Present() && cmd.Args().First() != "" {
		arg := cmd.Args().First()
		info, err := os.Stat(arg)
		if err != nil || !info.IsDir() {
			return "", fmt.Errorf("unknown command or invalid directory: %s", arg)
		}

		return arg, nil
	}

	// Home dir serves as a fallback if no argument is provided
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %v", err)
	}

	return homeDir, nil
}
