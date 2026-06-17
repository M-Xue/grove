package worktree

import "strings"

// validateAddInput trims the path and branch and ensures both are present,
// returning the cleaned values or a required-field error.
func validateAddInput(path, branch string) (string, string, error) {
	path = strings.TrimSpace(path)
	branch = strings.TrimSpace(branch)

	if path == "" {
		return "", "", ErrWorktreePathRequired
	}
	if branch == "" {
		return "", "", ErrBranchNameRequired
	}

	return path, branch, nil
}
