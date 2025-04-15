package cmd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"tmx/tmux"

	"github.com/urfave/cli/v3"
)

func Run() {
	app := &cli.Command{
		Name:        "tmux sessionizer",
		Description: "Tmux session manager",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			workDir, err := selectWorkDir(cmd)
			if err != nil {
				return nil
			}

			tmux.ManageSession(workDir)
			return nil
		},
		Commands: []*cli.Command{},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
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
		fmt.Println("Error executing find command:", err)
		return "", err
	}

	// Then run fzf with the find output as input
	fzfCmd := exec.Command("fzf")

	// Create the stdin pipe for feeding data to fzf
	fzfStdin, err := fzfCmd.StdinPipe()
	if err != nil {
		fmt.Println("Error creating stdin pipe:", err)
		return "", err
	}

	// Connect to the terminal for interactive use
	fzfCmd.Stdout = os.Stdout
	fzfCmd.Stderr = os.Stderr

	// Create a buffer to capture the selected output
	var outputBuf bytes.Buffer
	fzfCmd.Stdout = &outputBuf

	// Start the fzf command
	err = fzfCmd.Start()
	if err != nil {
		fmt.Println("Error starting fzf:", err)
		return "", err
	}

	// Write the find output to fzf's stdin
	_, err = fzfStdin.Write(findOutput)
	if err != nil {
		fmt.Println("Error writing to fzf stdin:", err)
		return "", err
	}
	fzfStdin.Close()

	// Wait for fzf to complete
	err = fzfCmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 || exitErr.ExitCode() == 130 {
				fmt.Println("No directory selected, exiting.")
				os.Exit(0)
			}
		}

		return "", err
	}

	// Get the selected path
	selected := strings.TrimSpace(outputBuf.String())

	if selected == "" {
		fmt.Println("No directory selected, exiting.")
		os.Exit(0)
	}

	return selected, nil
}
