package tmux

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func AttachToSession(sessionName string) {
	_, sessionExists := os.LookupEnv("TMUX")

	var cmd *exec.Cmd

	if !sessionExists {
		cmd = exec.Command("tmux", "attach-session", "-t", sessionName)
	} else {
		cmd = exec.Command("tmux", "switch-client", "-t", sessionName)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error attaching to tmux session: %v\n", err)
	}
}

func createSessionName(dirName string) string {
	// Replace characters that tmux doesn't like in session names
	name := strings.ReplaceAll(dirName, ".", "_")
	name = strings.ReplaceAll(name, ":", "_")

	// Additional cleaning if needed
	invalidChars := []string{" ", "/", "\\", "$", "#", "&", "*", "(", ")", "{", "}", "[", "]", "@", "!"}
	for _, char := range invalidChars {
		name = strings.ReplaceAll(name, char, "_")
	}

	return name
}

func checkIfSessionExists(sessionName string) bool {
	// Check if the session exists
	cmd := exec.Command("tmux", "has-session", "-t", sessionName)
	err := cmd.Run()
	return err == nil
}

func createTmuxWindowCmd(sessionName string, dir string, windowName string) *exec.Cmd {
	return exec.Command("tmux", "neww", "-t", sessionName, "-c", dir, "-n", windowName)
}

func ManageSession(dir string) {
	// Extract the base directory name to use as session name
	sessionName := createSessionName(filepath.Base(dir))

	if exists := checkIfSessionExists(sessionName); !exists {
		newSession(sessionName, dir)
	}

	AttachToSession(sessionName)
}

func newSession(sessionName string, dir string) {
	// Create the tmux command
	createSessionCmd := exec.Command("tmux", "new-session", "-ds", sessionName, "-c", dir, "-n", "editor")
	editorCmd := createTmuxWindowCmd(sessionName, dir, "editor")
	lazygitCmd := createTmuxWindowCmd(sessionName, dir, "lazygit")

	// Capture stderr to see detailed error message
	var stderr bytes.Buffer
	createSessionCmd.Stderr = &stderr

	// Run the command
	err := createSessionCmd.Run()
	if err != nil {
		fmt.Printf("Error starting tmux session: %v\n", err)
		fmt.Printf("Stderr: %s\n", stderr.String())
		return
	}

	err = editorCmd.Run()
	if err != nil {
		fmt.Printf("Stderr: %s\n", stderr.String())
		return
	}

	err = lazygitCmd.Run()
	if err != nil {
		fmt.Printf("Stderr: %s\n", stderr.String())
		return
	}

	fmt.Printf("Successfully started tmux session: %s\n", sessionName)
}
