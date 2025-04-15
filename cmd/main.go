package cmd

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func Run() {
	app := &cli.Command{
		Name:        "tmux sessionizer",
		Description: "Tmux session manager",
		Action:      DefaultAction,
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
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
