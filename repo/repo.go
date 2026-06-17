// Package repo provides the git repository preconditions grove checks at
// startup. These are not worktree operations, so they live apart from the
// worktree service.
package repo

import (
	"errors"
	"strings"
)

// ErrNotGitRepo is returned when grove is run outside a git working tree.
var ErrNotGitRepo = errors.New("current directory is not a git repository")

// Runner executes external commands and returns their combined output.
type Runner interface {
	CombinedOutput(name string, args ...string) ([]byte, error)
}

// EnsureInRepo verifies that the current directory is inside a git working
// tree, returning ErrNotGitRepo otherwise.
func EnsureInRepo(runner Runner) error {
	output, err := runner.CombinedOutput("git", "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return ErrNotGitRepo
	}
	if strings.TrimSpace(string(output)) != "true" {
		return ErrNotGitRepo
	}
	return nil
}
