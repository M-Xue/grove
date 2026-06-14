package worktree

import (
	"errors"
	"testing"
)

func TestManagerInRepoReturnsNilInsideRepo(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "rev-parse", "--is-inside-work-tree"): {
				output: []byte("true\n"),
			},
		},
	}

	manager := NewServiceWithRunner(runner)
	if err := manager.InRepo(); err != nil {
		t.Fatalf("InRepo returned error: %v", err)
	}
}

func TestManagerInRepoReturnsNotGitRepoWhenCommandFails(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "rev-parse", "--is-inside-work-tree"): {
				err: errors.New("git failed"),
			},
		},
	}

	manager := NewServiceWithRunner(runner)
	err := manager.InRepo()
	if !errors.Is(err, ErrNotGitRepo) {
		t.Fatalf("expected ErrNotGitRepo, got %v", err)
	}
}

func TestManagerInRepoReturnsNotGitRepoWhenFalse(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "rev-parse", "--is-inside-work-tree"): {
				output: []byte("false\n"),
			},
		},
	}

	manager := NewServiceWithRunner(runner)
	err := manager.InRepo()
	if !errors.Is(err, ErrNotGitRepo) {
		t.Fatalf("expected ErrNotGitRepo, got %v", err)
	}
}
