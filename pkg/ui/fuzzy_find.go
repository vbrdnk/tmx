package ui

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/vbrdnk/tmx/pkg/session"
)

var ErrNoSelection = errors.New("nothing selected")

func FuzzyFind(input []byte) (string, error) {
	var fzfCmd *exec.Cmd

	// Then run fzf with the find output as input
	if session.TmuxRunning() {
		fzfCmd = exec.Command("fzf", "--tmux", "70%")
	} else {
		fzfCmd = exec.Command("fzf", "--height=70%", "--border", "--margin=1", "--padding=1")
	}

	// Create the stdin pipe for feeding data to fzf
	fzfStdin, err := fzfCmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdin pipe: %v", err)
	}

	defer fzfStdin.Close()

	// Capture the output in a buffer
	var outputBuf bytes.Buffer
	fzfCmd.Stdout = &outputBuf

	// Connect stderr to the terminal for fzf's UI messages
	fzfCmd.Stderr = os.Stderr

	// Start the fzf command
	if err := fzfCmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start fzf: %v", err)
	}

	// Write the inpuit data to fzf
	if _, err := fzfStdin.Write(input); err != nil {
		return "", fmt.Errorf("failed to write to fzf stdin: %v", err)
	}

	// Close the stdin pipe to signal EOF
	if err := fzfStdin.Close(); err != nil {
		return "", fmt.Errorf("failed to close fzf stdin: %v", err)
	}

	// Wait for fzf to complete
	if err := fzfCmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exiting with code 1 or 130 indicates that the user canceled the selection
			if exitErr.ExitCode() == 1 || exitErr.ExitCode() == 130 {
				return "", ErrNoSelection
			}
		}

		return "", fmt.Errorf("fzf command failed: %v", err)
	}

	// Get and validate user's selection
	selection := strings.TrimSpace(outputBuf.String())

	if selection == "" {
		return "", ErrNoSelection
	}

	return selection, nil
}
