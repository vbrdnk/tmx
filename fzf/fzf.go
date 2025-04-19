package fzf

import (
	"bytes"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

func RunFzf(input []byte) (string, error) {
	// Then run fzf with the find output as input
	fzfCmd := exec.Command("fzf")

	// Create the stdin pipe for feeding data to fzf
	fzfStdin, err := fzfCmd.StdinPipe()
	if err != nil {
		color.Red("Error creating stdin pipe: %v", err)
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
		color.Red("Error starting fzf: %v", err)
		return "", err
	}

	// Write the find output to fzf's stdin
	_, err = fzfStdin.Write(input)
	if err != nil {
		color.Red("Error writing to fzf stdin: %v", err)
		return "", err
	}
	fzfStdin.Close()

	// Wait for fzf to complete
	err = fzfCmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 || exitErr.ExitCode() == 130 {
				color.Yellow("Nothing selected, exiting.")
				os.Exit(0)
			}
		}

		return "", err
	}

	// Get user's selection
	selection := strings.TrimSpace(outputBuf.String())

	if selection == "" {
		color.Yellow("Nothing selected, exiting.")
		os.Exit(0)
	}

	return selection, nil
}
