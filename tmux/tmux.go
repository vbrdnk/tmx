package tmux

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"tmx/config"
)

func AttachToSession(sessionName string) {
	_, tmuxRunning := os.LookupEnv("TMUX")

	var cmd *exec.Cmd

	if !tmuxRunning {
		cmd = exec.Command("tmux", "attach-session", "-t", sessionName)
	} else {
		cmd = exec.Command("tmux", "switch-client", "-t", sessionName)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error attaching to %s tmux session: %v\n", sessionName, err)
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
	var sessionName string
	config, err := config.ParseConfig()
	if err != nil {
		sessionName = createSessionName(filepath.Base(dir))
	} else {
		for _, ws := range config.Workspace {
			if strings.Contains(dir, ws.Directory) {
				sessionName = createSessionName(ws.Name)
			}
		}
	}

	if sessionName == "" {
		sessionName = createSessionName(filepath.Base(dir))
	}

	// Extract the base directory name to use as session name

	if exists := checkIfSessionExists(sessionName); !exists {
		fmt.Printf("Session %s does not exist. Creating a new session...\n", sessionName)
		newSession(sessionName, dir)
	}
	AttachToSession(sessionName)
}

func newSession(sessionName string, dir string) {
	fmt.Printf("Create new session: %s in directory: %s\n", sessionName, dir)

	config, err := config.ParseConfig()
	cmds := []*exec.Cmd{}

	if err != nil {
		fmt.Printf("Using default configuration (no config file found): %v\n", err)

		defaultCmd := exec.Command("tmux", "new-session", "-ds", sessionName, "-c", dir)
		cmds = append(cmds, defaultCmd)
	} else {
		matchFound := false

		for _, ws := range config.Workspace {
			fmt.Printf("Checking workspace - Name: %s, Directory: %s\n", ws.Name, ws.Directory)
			log.Printf("Checking workspace - Name: %s, Directory: %s\n", ws.Name, ws.Directory)

			fmt.Printf("Comparing %s with %s\n", ws.Directory, dir)
			if strings.Contains(dir, ws.Directory) {
				matchFound = true
				sessionName = createSessionName(ws.Name)

				fmt.Printf("Found matching workspace: %s\n", ws.Name)
				firstWindow := true

				for _, window := range ws.Windows {
					if firstWindow {
						cmds = append(cmds, exec.Command("tmux", "new-session", "-ds", sessionName, "-c", dir, "-n", window))
						firstWindow = false
					} else {
						cmds = append(cmds, createTmuxWindowCmd(sessionName, dir, window))
					}
				}

				// No need to loop through the rest of the workspaces
				break
			}
		}

		if !matchFound {
			fmt.Printf("No matching workspace found. Creating a new session...\n")

			cmds = append(cmds, exec.Command("tmux", "new-session", "-ds", sessionName, "-c", dir))
		}
	}

	if len(cmds) == 0 {
		fmt.Println("No commands to execute.")
		return
	}

	fmt.Printf("Executing %d tmux commands\n", len(cmds))

	for _, cmd := range cmds {
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			fmt.Printf("Stderr: %s\n", stderr.String())
		} else {
			fmt.Printf("Command executed successfully: %s\n", cmd.String())
		}
	}

	fmt.Printf("Successfully started tmux session: %s\n", sessionName)
}
