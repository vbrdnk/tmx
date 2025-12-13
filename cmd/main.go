package cmd

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"
	utils "github.com/vbrdnk/tmx/internal/utils"
	config "github.com/vbrdnk/tmx/pkg/config"
)

func Run() {
	config, configErrors := config.ParseConfig()
	if len(configErrors) > 0 {
		for _, err := range configErrors {
			log.Printf("Configuration error: %v", err)
		}
		// Continue execution even if there are config errors
	}

	app := &cli.Command{
		Name:        "tmux sessionizer",
		Description: "Tmux session manager",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "depth",
				Aliases: []string{"d"},
				Usage:   "search depth for nested directories (0 = unlimited)",
				Value:   0, // 0 means use config default
			},
		},
		Action: func(_ctx context.Context, cmd *cli.Command) error {
			targetDirPath, err := utils.GetWorkingDirPath(cmd)
			if err != nil || targetDirPath == "" {
				log.Fatal(err)
			}

			depth := int(cmd.Int("depth"))
			return DefaultAction(targetDirPath, config, depth)
		},
		Commands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"l", "ls"},
				Usage:   "list currently active tmux sessions",
				Action:  ListSessionsAction,
			},
			{
				Name:    "connect",
				Aliases: []string{"c", "conn"},
				Usage:   "connect to a tmux session",
				Action:  AttachToSessionAction,
			},
			{
				Name:    "kill",
				Aliases: []string{"k"},
				Usage:   "kill tmux session",
				Action:  KillSessionAction,
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
