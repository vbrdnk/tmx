package cmd

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/vbrdnk/tmx/pkg/config"
	"github.com/vbrdnk/tmx/pkg/discovery"
	"github.com/vbrdnk/tmx/pkg/session"
	"github.com/vbrdnk/tmx/pkg/ui"

	"github.com/urfave/cli/v3"
)

func DefaultAction(targetDir string, config *config.Config, cliDepth int) error {
	// Use DirectorySelector to find and select directory
	selector := discovery.NewDirectorySelector(config)
	workDir, err := selector.SelectDirectory(targetDir, cliDepth)
	if err != nil {
		color.Red(err.Error())
		return nil
	}

	// Use SessionManager to resolve and attach to session
	sessionManager := session.NewSessionManager(config)
	if err := sessionManager.ResolveSession(workDir); err != nil {
		color.Red("Error resolving session: %v", err)
	}
	return nil
}

func ListSessionsAction(_ctx context.Context, _cmd *cli.Command) error {
	if err := session.ListSessions(); err != nil {
		color.Red("Error getting sessions list")
	}

	return nil
}

func AttachToSessionAction(_ctx context.Context, _cmd *cli.Command) error {
	sess, err := selectFromActiveSessions()
	if err != nil {
		color.Red("Error selecting active session: %v", err)
		return nil
	}

	if err := session.AttachToSession(sess); err != nil {
		color.Red("Error connecting to %s tmux session: %v", sess, err)
	}

	return nil
}

func KillSessionAction(_ctx context.Context, _cmd *cli.Command) error {
	sess, err := selectFromActiveSessions()
	if err != nil {
		color.Red("Error selecting active session: %v", err)
		return nil
	}

	if err := session.KillSession(sess); err != nil {
		color.Red("Error killing %s tmux session: %v", sess, err)
	}

	return nil
}

func selectFromActiveSessions() (string, error) {
	cmd := session.NewTmuxCommand("list-sessions")
	cmdOutput, err := cmd.Output()
	if err != nil {
		return "", errors.New("no active tmux sessions")
	}

	if fullSessionName, err := ui.FuzzyFind(cmdOutput); err != nil {
		if errors.Is(err, ui.ErrNoSelection) {
			color.Yellow("No sesison selected, exiting.")
			os.Exit(0)
		}
		return "", err
	} else {
		// Full session name is in the format "session_name:session_index"
		return strings.Split(fullSessionName, ":")[0], nil
	}
}
