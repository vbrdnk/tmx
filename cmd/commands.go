package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/vbrdnk/tmx/pkg/config"
	"github.com/vbrdnk/tmx/pkg/fzf"
	"github.com/vbrdnk/tmx/pkg/tmux"

	"github.com/urfave/cli/v3"
)

func DefaultAction(targetDir string, config *config.Config) error {
	workDir, err := selectWorkDir(targetDir)
	if err != nil {
		color.Red(err.Error())
		return nil
	}

	tmux.ResolveSession(workDir, config)
	return nil
}

func ListSessionsAction(_ctx context.Context, _cmd *cli.Command) error {
	if err := tmux.ListSessions(); err != nil {
		color.Red("Error getting sessions list")
	}

	return nil
}

func AttachToSessionAction(_ctx context.Context, _cmd *cli.Command) error {
	session, err := selectFromActiveSessions()
	if err != nil {
		color.Red("Error selecting active session: %v", err)
		return nil
	}

	if err := tmux.AttachToSession(session); err != nil {
		color.Red("Error connecting to %s tmux session: %v", session, err)
	}

	return nil
}

func KillSessionAction(_ctx context.Context, _cmd *cli.Command) error {
	session, err := selectFromActiveSessions()
	if err != nil {
		color.Red("Error selecting active session: %v", err)
		return nil
	}

	if err := tmux.KillSession(session); err != nil {
		color.Red("Error killing %s tmux session: %v", session, err)
	}

	return nil
}

func selectFromActiveSessions() (string, error) {
	cmd := tmux.NewTmuxCommand("list-sessions")
	cmdOutput, err := cmd.Output()
	if err != nil {
		return "", errors.New("no active tmux sessions")
	}

	if fullSessionName, err := fzf.FuzzyFind(cmdOutput); err != nil {
		if errors.Is(err, fzf.ErrNoSelection) {
			color.Yellow("No sesison selected, exiting.")
			os.Exit(0)
		}
		return "", err
	} else {
		// Full session name is in the format "session_name:session_index"
		return strings.Split(fullSessionName, ":")[0], nil
	}
}

func selectWorkDir(path string) (string, error) {
	// Run find command first and collect its output
	findCmd := exec.Command("find", path, "-mindepth", "1", "-maxdepth", "1", "-type", "d")
	findOutput, err := findCmd.Output()
	if err != nil {
		return "", fmt.Errorf("error executing find command: %v", err)
	}

	if cwd, err := fzf.FuzzyFind(findOutput); err != nil {
		if errors.Is(err, fzf.ErrNoSelection) {
			color.Yellow("No folder selected, exiting.")
			os.Exit(0)
		}
		return "", err
	} else {
		// Full session name is in the format "session_name:session_index"
		return cwd, nil
	}
}
