// Package command provides the single external-command runner shared by
// grove's git-facing services. A Runner is constructed once at startup and
// injected into each service, so production code uses real subprocesses while
// tests substitute a stub.
package command

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// DefaultTimeout bounds every external command so that a hung git process
// cannot block grove indefinitely.
const DefaultTimeout = 30 * time.Second

// Runner executes external commands using the local process environment,
// applying a timeout to each invocation.
type Runner struct {
	timeout time.Duration
}

// New returns a Runner that applies DefaultTimeout to each command.
func New() Runner {
	return Runner{timeout: DefaultTimeout}
}

// CombinedOutput runs name with args under a timeout and returns the combined
// stdout and stderr.
//
// If the command fails and produced output, that output is attached to the
// returned error to make subprocess failures easier to understand at the call
// site. This is the one place git invocations are made cancellable.
func (r Runner) CombinedOutput(name string, args ...string) ([]byte, error) {
	timeout := r.timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
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
