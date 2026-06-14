package worktree

import (
	"errors"
	"strings"
)

var (
	ErrWorktreePathRequired = errors.New("worktree path is required")
	ErrBranchNameRequired   = errors.New("branch name is required")
	ErrNotGitRepo           = errors.New("current directory is not a git repository")
)

type WorktreeInfo struct {
	Path                  string
	Branch                string
	CommitLabel           string
	CommitHash            string
	HasUncommittedChanges bool
}

type Runner interface {
	CombinedOutput(name string, args ...string) ([]byte, error)
}

type Service interface {
	Add(path, branch string) error
	AddNewBranch(path, branch string) error
	BranchExists(branch string) (bool, error)
	InRepo() error
	List() ([]WorktreeInfo, error)
	Remove(path string) error
}

type manager struct {
	runner Runner
}

func NewService() Service {
	return manager{runner: commandRunner{}}
}

func NewServiceWithRunner(runner Runner) Service {
	if runner == nil {
		return NewService()
	}
	return manager{runner: runner}
}

func (s manager) Add(path, branch string) error {
	path, branch, err := validateAddInput(path, branch)
	if err != nil {
		return err
	}

	_, err = s.runner.CombinedOutput("git", "worktree", "add", path, branch)
	return err
}

func (s manager) AddNewBranch(path, branch string) error {
	path, branch, err := validateAddInput(path, branch)
	if err != nil {
		return err
	}

	_, err = s.runner.CombinedOutput("git", "worktree", "add", "-b", branch, path)
	return err
}

func (s manager) Remove(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return ErrWorktreePathRequired
	}

	_, err := s.runner.CombinedOutput("git", "worktree", "remove", path)
	return err
}

func (s manager) BranchExists(branch string) (bool, error) {
	branch = strings.TrimSpace(branch)
	if branch == "" {
		return false, ErrBranchNameRequired
	}

	output, err := s.runner.CombinedOutput(
		"git",
		"for-each-ref",
		"--format=%(refname:short)",
		"refs/heads/"+branch,
	)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(output)) == branch, nil
}

func (s manager) InRepo() error {
	output, err := s.runner.CombinedOutput("git", "rev-parse", "--is-inside-work-tree")
	if err != nil {
		return ErrNotGitRepo
	}
	if strings.TrimSpace(string(output)) != "true" {
		return ErrNotGitRepo
	}
	return nil
}

func (s manager) List() ([]WorktreeInfo, error) {
	output, err := s.runner.CombinedOutput("git", "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	entries, err := parseWorktreeList(output)
	if err != nil {
		return nil, err
	}

	worktrees := make([]WorktreeInfo, 0, len(entries))
	for _, entry := range entries {
		commitLabel, err := s.commitLabel(entry.path)
		if err != nil {
			return nil, err
		}

		dirty, err := s.hasUncommittedChanges(entry.path)
		if err != nil {
			return nil, err
		}

		worktrees = append(worktrees, WorktreeInfo{
			Path:                  entry.path,
			Branch:                normalizeBranch(entry.branch),
			CommitLabel:           commitLabel,
			CommitHash:            entry.commitHash,
			HasUncommittedChanges: dirty,
		})
	}

	return worktrees, nil
}

func (s manager) commitLabel(path string) (string, error) {
	output, err := s.runner.CombinedOutput("git", "-C", path, "log", "-1", "--pretty=%s")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (s manager) hasUncommittedChanges(path string) (bool, error) {
	output, err := s.runner.CombinedOutput("git", "-C", path, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(output)) != "", nil
}

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
