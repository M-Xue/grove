// Package command provides the single external-command runner shared by
// grove's git-facing services. A Runner is constructed once at startup and
// injected into each service, so production code uses real subprocesses while
// tests substitute a stub.
package command

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// DefaultTimeout bounds every external command so that a hung git process
// cannot block grove indefinitely.
const DefaultTimeout = 30 * time.Second

// StreamTimeout bounds streaming commands. It is far longer than DefaultTimeout
// because a worktree checkout of a large repository can legitimately run for
// minutes, and its progress is reported live rather than blocking silently.
const StreamTimeout = 10 * time.Minute

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

// StreamProgress runs name with args under StreamTimeout and invokes onLine for
// each progress token written to stderr. git separates in-place progress
// updates with carriage returns rather than newlines, so tokens are split on
// both "\r" and "\n". GIT_PROGRESS_DELAY=0 is set so git reports progress
// immediately instead of after its default ~2s delay.
//
// Like CombinedOutput, a failing command attaches its captured output to the
// returned error. stdout is discarded; git writes its progress to stderr.
func (r Runner) StreamProgress(onLine func(string), name string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), StreamTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = append(os.Environ(), "GIT_PROGRESS_DELAY=0")

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	var collected strings.Builder
	scanner := bufio.NewScanner(stderr)
	scanner.Split(scanProgressTokens)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		collected.WriteString(line)
		collected.WriteByte('\n')
		if onLine != nil {
			onLine(line)
		}
	}

	if err := cmd.Wait(); err != nil {
		message := strings.TrimSpace(collected.String())
		if message == "" {
			return err
		}
		return fmt.Errorf("%w: %s", err, message)
	}
	return nil
}

// scanProgressTokens is a bufio.SplitFunc that yields a token whenever it
// reaches a carriage return or newline, so git's carriage-return-delimited
// progress updates surface one at a time. The delimiter is dropped from the
// returned token.
func scanProgressTokens(data []byte, atEOF bool) (int, []byte, error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
		return i + 1, data[:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}
