package cmd

import (
	"context"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/urfave/cli/v3"
	"github.com/vbrdnk/tmx/internal/path"
	config "github.com/vbrdnk/tmx/pkg/config"
	"github.com/vbrdnk/tmx/pkg/session"
)

var Version = "dev" // will be overridden at build time with ldflags

func Run() {
	config, configErrors := config.ParseConfig()
	if len(configErrors) > 0 {
		for _, err := range configErrors {
			log.Printf("Configuration error: %v", err)
		}
		// Continue execution even if there are config errors
	}

	// Create session manager instance
	sessionManager := session.NewSessionManager(config)

	app := &cli.Command{
		Name:                  "tmux sessionizer",
		Description:           "Tmux session manager",
		Version:               Version,
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "depth",
				Aliases: []string{"d"},
				Usage:   "search depth for nested directories (0 = unlimited)",
				Value:   0, // 0 means use config default
			},
		},
		Action: func(_ctx context.Context, cmd *cli.Command) error {
			targetDirPath, err := path.GetWorkingDirPath(cmd)
			if err != nil {
				return err
			}

			depth := int(cmd.Int("depth"))
			return DefaultAction(targetDirPath, config, depth, sessionManager)
		},
		Commands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"l", "ls"},
				Usage:   "list currently active tmux sessions",
				Action: func(_ctx context.Context, _cmd *cli.Command) error {
					return ListSessionsAction(_ctx, _cmd, sessionManager)
				},
			},
			{
				Name:    "connect",
				Aliases: []string{"c", "conn"},
				Usage:   "connect to a tmux session",
				Action: func(_ctx context.Context, _cmd *cli.Command) error {
					return AttachToSessionAction(_ctx, _cmd, sessionManager)
				},
			},
			{
				Name:    "kill",
				Aliases: []string{"k"},
				Usage:   "kill tmux session",
				Action: func(_ctx context.Context, _cmd *cli.Command) error {
					return KillSessionAction(_ctx, _cmd, sessionManager)
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		color.Red("%v", err)
		os.Exit(1)
	}
}
