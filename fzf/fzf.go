package fzf

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
)

func RunFzf(input []byte) (string, error) {
	// Then run fzf with the find output as input
	fzfCmd := exec.Command("fzf")

	// Create the stdin pipe for feeding data to fzf
	fzfStdin, err := fzfCmd.StdinPipe()
	if err != nil {
		log.Println("Error creating stdin pipe:", err)
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
		log.Println("Error starting fzf:", err)
		return "", err
	}

	// Write the find output to fzf's stdin
	_, err = fzfStdin.Write(input)
	if err != nil {
		log.Println("Error writing to fzf stdin:", err)
		return "", err
	}
	fzfStdin.Close()

	// Wait for fzf to complete
	err = fzfCmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 || exitErr.ExitCode() == 130 {
				log.Println("Nothing selected, exiting.")
				os.Exit(0)
			}
		}

		return "", err
	}

	// Get user's selection
	selection := strings.TrimSpace(outputBuf.String())

	if selection == "" {
		log.Println("Nothing selected, exiting.")
		os.Exit(0)
	}

	return selection, nil
}
