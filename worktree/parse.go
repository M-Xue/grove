package worktree

import (
	"fmt"
	"strings"
)

type listEntry struct {
	path       string
	branch     string
	commitHash string
	prunable   bool
}

func parseWorktreeList(output []byte) ([]listEntry, error) {
	blocks := strings.Split(strings.TrimSpace(string(output)), "\n\n")
	entries := make([]listEntry, 0, len(blocks))

	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		entry, err := parseWorktreeBlock(block)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func parseWorktreeBlock(block string) (listEntry, error) {
	var entry listEntry

	for _, line := range strings.Split(block, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			entry.path = strings.TrimSpace(strings.TrimPrefix(line, "worktree "))
		case strings.HasPrefix(line, "HEAD "):
			entry.commitHash = strings.TrimSpace(strings.TrimPrefix(line, "HEAD "))
		case strings.HasPrefix(line, "branch "):
			entry.branch = strings.TrimSpace(strings.TrimPrefix(line, "branch "))
		case line == "prunable" || strings.HasPrefix(line, "prunable "):
			entry.prunable = true
		}
	}

	if entry.path == "" {
		return listEntry{}, fmt.Errorf("worktree path missing from git output")
	}
	if entry.commitHash == "" {
		return listEntry{}, fmt.Errorf("worktree commit hash missing from git output")
	}

	return entry, nil
}

func normalizeBranch(branch string) string {
	branch = strings.TrimSpace(branch)
	if branch == "" {
		return "detached"
	}
	return strings.TrimPrefix(branch, "refs/heads/")
}
