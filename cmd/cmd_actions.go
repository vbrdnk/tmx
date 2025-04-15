package cmd

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strings"
	"tmx/fzf"
	"tmx/tmux"

	"github.com/urfave/cli/v3"
)

func DefaultAction(_ctx context.Context, cmd *cli.Command) error {
	workDir, err := selectWorkDir(cmd)
	if err != nil {
		return nil
	}

	tmux.ManageSession(workDir)
	return nil
}

func ListSessionsAction(_ctx context.Context, _cmd *cli.Command) error {
	listSessionsCmd := exec.Command("tmux", "list-sessions")
	listSessionsCmd.Stdout = os.Stdout
	listSessionsCmd.Run()
	return nil
}

func AttachToSessionAction(_ctx context.Context, _cmd *cli.Command) error {
	session, err := getActiveSessionName()
	if err != nil {
		return nil
	}

	tmux.AttachToSession(session)
	return nil
}

func getActiveSessionName() (string, error) {
	listSessionsCmd := exec.Command("tmux", "list-sessions")
	listSessionsCmdOutput, err := listSessionsCmd.Output()
	if err != nil {
		return "", err
	}

	fullSessionName, err := fzf.RunFzf(listSessionsCmdOutput)
	if err != nil {
		return "", err
	}
	return strings.Split(fullSessionName, ":")[0], nil
}

func selectWorkDir(cmd *cli.Command) (string, error) {
	searchDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if cmd.Args().Present() {
		searchDir = cmd.Args().First()
	}

	// Run find command first and collect its output
	findCmd := exec.Command("find", searchDir, "-mindepth", "1", "-maxdepth", "1", "-type", "d")
	findOutput, err := findCmd.Output()
	if err != nil {
		log.Println("Error executing find command:", err)
		return "", err
	}

	return fzf.RunFzf(findOutput)
}
