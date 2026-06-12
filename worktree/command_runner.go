package worktree

import (
	"fmt"
	"os/exec"
	"strings"
)

// commandRunner executes external commands using the local process environment.
//
// It adapts os/exec to the Runner interface so package logic can stay testable
// with stub runners while still using real subprocesses in production.
type commandRunner struct{}

// CombinedOutput runs a command and returns its combined stdout and stderr.
//
// If the command fails and produced output, that output is attached to the
// returned error to make subprocess failures easier to understand at the call site.
func (commandRunner) CombinedOutput(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			return nil, err
		}
		return nil, fmt.Errorf("%w: %s", err, message)
	}
	return output, nil
}
