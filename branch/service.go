// Package branch provides read-only branch information backed by git. It was
// once a full branch-management service; after the branch screen was removed,
// only the recent-commits lookup (used by the change screen's details panel)
// survives.
package branch

import (
	"fmt"
	"strings"
)

// CommitInfo is one commit on a branch: its abbreviated hash, author name and
// subject line.
type CommitInfo struct {
	Hash    string
	Author  string
	Subject string
}

type Runner interface {
	CombinedOutput(name string, args ...string) ([]byte, error)
}

type Service interface {
	RecentCommits(name string, limit int) ([]CommitInfo, error)
}

type service struct {
	runner Runner
}

// NewService returns a branch Service backed by the injected command runner.
func NewService(runner Runner) Service {
	return service{runner: runner}
}

// RecentCommits returns the most recent commits reachable from the named
// branch, newest first, capped at limit (defaulting to 10 when non-positive).
// Fields are separated by the ASCII unit separator so subjects containing any
// printable character parse cleanly.
func (s service) RecentCommits(name string, limit int) ([]CommitInfo, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("branch name is required")
	}
	if limit <= 0 {
		limit = 10
	}

	output, err := s.runner.CombinedOutput(
		"git",
		"log",
		fmt.Sprintf("-n%d", limit),
		"--format=%h%x1f%an%x1f%s",
		name,
	)
	if err != nil {
		return nil, err
	}

	trimmed := strings.TrimSpace(string(output))
	if trimmed == "" {
		return nil, nil
	}

	lines := strings.Split(trimmed, "\n")
	commits := make([]CommitInfo, 0, len(lines))
	for _, line := range lines {
		parts := strings.SplitN(line, "\x1f", 3)
		if len(parts) != 3 {
			continue
		}
		commits = append(commits, CommitInfo{
			Hash:    strings.TrimSpace(parts[0]),
			Author:  strings.TrimSpace(parts[1]),
			Subject: strings.TrimSpace(parts[2]),
		})
	}

	return commits, nil
}
