package tmux

import (
	"bytes"
	"log"
	"os"
	"os/exec"
)

// tmuxCommand represents a command to be executed with tmux
type TmuxCommand struct {
	args []string
}

// New creates a new TmuxCommand with the given arguments
func NewTmuxCommand(args ...string) *TmuxCommand {
	return &TmuxCommand{args: args}
}

// Execute runs the tmux command and returns any error
func (tc *TmuxCommand) Execute() error {
	cmd := exec.Command("tmux", tc.args...)
	return cmd.Run()
}

// ExecuteWithIO runs the tmux command with standard IO connected
func (tc *TmuxCommand) ExecuteWithIO() error {
	cmd := exec.Command("tmux", tc.args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ExecuteVerbose runs the command and prints detailed output
func (tc *TmuxCommand) ExecuteVerbose() error {
	cmd := exec.Command("tmux", tc.args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Printf("Stderr: %s\n", stderr.String())
	}
	return err
}

// ExecuteOutput runs the command and returns the output
func (tc *TmuxCommand) Output() ([]byte, error) {
	cmd := exec.Command("tmux", tc.args...)
	return cmd.Output()
}
