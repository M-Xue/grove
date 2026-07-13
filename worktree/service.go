package worktree

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var (
	ErrWorktreePathRequired = errors.New("worktree path is required")
	ErrBranchNameRequired   = errors.New("branch name is required")
)

type Info struct {
	Path                  string
	Branch                string
	CommitLabel           string
	CommitHash            string
	HasUncommittedChanges bool
	// Stale reports that git considers the worktree prunable, typically
	// because its working directory no longer exists on disk.
	Stale bool
	// Locked reports that the worktree is locked (git worktree lock). A locked
	// worktree is protected from pruning, so it is surfaced distinctly from a
	// stale one even when its working directory is gone.
	Locked bool
	// CreatedAt is when the worktree was added, derived from its git
	// administrative files. It is the zero time when it cannot be determined
	// (e.g. for stale worktrees, whose files no longer exist on disk).
	CreatedAt time.Time
}

type Runner interface {
	CombinedOutput(name string, args ...string) ([]byte, error)
}

type Service interface {
	Add(path, branch string) error
	AddWithNewBranch(path, branch string) error
	BranchExists(branch string) (bool, error)
	List() ([]Info, error)
	Remove(path string) error
	Prune() error
}

type service struct {
	runner Runner
	// stat resolves filesystem metadata; injectable so worktree creation
	// times can be exercised in tests without real git directories.
	stat func(name string) (os.FileInfo, error)
}

// NewService returns a worktree Service backed by the injected command runner.
func NewService(runner Runner) Service {
	return service{runner: runner, stat: os.Stat}
}

func (s service) Add(path, branch string) error {
	path, branch, err := validateAddInput(path, branch)
	if err != nil {
		return err
	}

	_, err = s.runner.CombinedOutput("git", "worktree", "add", path, branch)
	return err
}

func (s service) AddWithNewBranch(path, branch string) error {
	path, branch, err := validateAddInput(path, branch)
	if err != nil {
		return err
	}

	_, err = s.runner.CombinedOutput("git", "worktree", "add", "-b", branch, path)
	return err
}

func (s service) Remove(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return ErrWorktreePathRequired
	}

	_, err := s.runner.CombinedOutput("git", "worktree", "remove", path)
	return err
}

// Prune removes the administrative files of every stale worktree, i.e. those
// whose working directory no longer exists on disk.
func (s service) Prune() error {
	_, err := s.runner.CombinedOutput("git", "worktree", "prune")
	return err
}

func (s service) BranchExists(branch string) (bool, error) {
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

func (s service) List() ([]Info, error) {
	output, err := s.runner.CombinedOutput("git", "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	entries, err := parseWorktreeList(output)
	if err != nil {
		return nil, err
	}

	worktrees := make([]Info, 0, len(entries))
	for _, entry := range entries {
		// A worktree whose working directory is gone can no longer be inspected
		// with git -C, which would fail with exit status 128 and abort the whole
		// list. git usually flags these "prunable", but not always: a locked
		// worktree stays unprunable even once deleted, and git versions before
		// 2.36 omit the prunable annotation entirely. So we independently
		// confirm the directory exists on disk before running git inside it. A
		// missing worktree is surfaced as locked when git has it locked (it
		// cannot be pruned), and as stale otherwise.
		if entry.prunable || !s.pathExists(entry.path) {
			worktrees = append(worktrees, Info{
				Path:       entry.path,
				Branch:     normalizeBranch(entry.branch),
				CommitHash: entry.commitHash,
				Stale:      !entry.locked,
				Locked:     entry.locked,
			})
			continue
		}

		commitLabel, err := s.commitLabel(entry.path)
		if err != nil {
			return nil, err
		}

		dirty, err := s.hasUncommittedChanges(entry.path)
		if err != nil {
			return nil, err
		}

		worktrees = append(worktrees, Info{
			Path:                  entry.path,
			Branch:                normalizeBranch(entry.branch),
			CommitLabel:           commitLabel,
			CommitHash:            entry.commitHash,
			HasUncommittedChanges: dirty,
			Locked:                entry.locked,
			CreatedAt:             s.createdAt(entry.path),
		})
	}

	// git always lists the main worktree first, so worktrees[0] is the main
	// worktree; pin it to the top and order only the linked worktrees beneath
	// it, oldest first. The sort is stable so entries that compare equal keep
	// git's ordering. Worktrees with an undeterminable creation time (zero
	// value, e.g. stale ones) sort below the dated entries rather than above
	// them, since the zero time would otherwise read as the oldest.
	if len(worktrees) > 1 {
		rest := worktrees[1:]
		sort.SliceStable(rest, func(i, j int) bool {
			a, b := rest[i].CreatedAt, rest[j].CreatedAt
			if a.IsZero() != b.IsZero() {
				return b.IsZero()
			}
			return a.Before(b)
		})
	}

	return worktrees, nil
}

// pathExists reports whether the worktree's working directory is present on
// disk. It is the guard that keeps List from running git -C against a worktree
// whose directory has been removed.
func (s service) pathExists(path string) bool {
	_, err := s.stat(path)
	return err == nil
}

// createdAt reports when the worktree at path was added. Git does not expose
// this directly, so it reads the mod time of the worktree's administrative
// `gitdir` file, which git writes once at `git worktree add` time and rarely
// rewrites afterward. The main worktree has no such file, so it falls back to
// the git directory itself. Any failure yields the zero time rather than an
// error, so an unknown creation time never aborts the whole listing.
func (s service) createdAt(path string) time.Time {
	output, err := s.runner.CombinedOutput("git", "-C", path, "rev-parse", "--absolute-git-dir")
	if err != nil {
		return time.Time{}
	}
	gitDir := strings.TrimSpace(string(output))
	if gitDir == "" {
		return time.Time{}
	}

	info, err := s.stat(filepath.Join(gitDir, "gitdir"))
	if err != nil {
		info, err = s.stat(gitDir)
		if err != nil {
			return time.Time{}
		}
	}
	return info.ModTime()
}

func (s service) commitLabel(path string) (string, error) {
	output, err := s.runner.CombinedOutput("git", "-C", path, "log", "-1", "--pretty=%s")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (s service) hasUncommittedChanges(path string) (bool, error) {
	output, err := s.runner.CombinedOutput("git", "-C", path, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "?? ") {
			continue
		}
		return true, nil
	}
	return false, nil
}
