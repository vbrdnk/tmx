package cmd

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/vbrdnk/tmx/fzf"
	"github.com/vbrdnk/tmx/tmux"

	"github.com/urfave/cli/v3"
)

func DefaultAction(_ctx context.Context, cmd *cli.Command) error {
	workDir, err := selectWorkDir(cmd)
	if err != nil {
		return nil
	}

	tmux.ResolveSession(workDir)
	return nil
}

func ListSessionsAction(_ctx context.Context, _cmd *cli.Command) error {
	cmd := tmux.NewTmuxCommand("list-sessions")
	return cmd.ExecuteWithIO()
}

func AttachToSessionAction(_ctx context.Context, _cmd *cli.Command) error {
	session, err := selectFromActiveSessions()
	if err != nil {
		log.Println("Error getting active session name:", err)
		return nil
	}

	tmux.AttachToSession(session)
	return nil
}

func KillSessionAction(_ctx context.Context, _cmd *cli.Command) error {
	session, err := selectFromActiveSessions()
	if err != nil {
		log.Println("Error selecting active session:", err)
		return nil
	}

	tmux.KillSession(session)
	return nil
}

func selectFromActiveSessions() (string, error) {
	cmd := exec.Command("tmux", "list-sessions")
	cmdOutput, err := cmd.Output()
	if err != nil {
		return "", err
	}

	fullSessionName, err := fzf.RunFzf(cmdOutput)
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
