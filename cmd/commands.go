package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/vbrdnk/tmx/pkg/config"
	"github.com/vbrdnk/tmx/pkg/fzf"
	"github.com/vbrdnk/tmx/pkg/search"
	"github.com/vbrdnk/tmx/pkg/tmux"

	"github.com/urfave/cli/v3"
)

func DefaultAction(targetDir string, config *config.Config, cliDepth int) error {
	workDir, err := selectWorkDir(targetDir, config, cliDepth)
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

func selectWorkDir(path string, config *config.Config, cliDepth int) (string, error) {
	dirList, err := buildDirectoryList(path, config, cliDepth)
	if err != nil {
		return "", fmt.Errorf("error building directory list: %v", err)
	}

	if cwd, err := fzf.FuzzyFind(dirList); err != nil {
		if errors.Is(err, fzf.ErrNoSelection) {
			color.Yellow("No folder selected, exiting.")
			os.Exit(0)
		}
		return "", err
	} else {
		// Strip the frecency indicator if present
		cwd = strings.TrimPrefix(cwd, "★ ")
		return cwd, nil
	}
}

// buildDirectoryList constructs a list of directories combining zoxide frecency and find/fd results
func buildDirectoryList(path string, config *config.Config, cliDepth int) ([]byte, error) {
	searchDepth := config.GetSearchDepth(cliDepth)
	useZoxide := config.GetUseZoxide()

	var directories []string
	seenPaths := make(map[string]bool)

	// 1. Get zoxide results if enabled
	if useZoxide {
		zoxideResults, err := search.GetZoxideResults(path)
		if err == nil && len(zoxideResults) > 0 {
			for _, dir := range zoxideResults {
				if !seenPaths[dir] {
					directories = append(directories, "★ "+dir)
					seenPaths[dir] = true
				}
			}
		}
		// Silently ignore zoxide errors (not installed, no results, etc.)
	}

	// 2. Get find/fd results
	findResults, err := search.GetFindResults(path, searchDepth)
	if err != nil {
		return nil, err
	}

	for _, dir := range findResults {
		if !seenPaths[dir] {
			directories = append(directories, dir)
			seenPaths[dir] = true
		}
	}

	return []byte(strings.Join(directories, "\n")), nil
}
