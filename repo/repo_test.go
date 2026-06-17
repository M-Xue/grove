package repo

import (
	"errors"
	"testing"
)

type stubRunner struct {
	output []byte
	err    error
}

func (s stubRunner) CombinedOutput(name string, args ...string) ([]byte, error) {
	return s.output, s.err
}

func TestEnsureInRepoReturnsNilInsideRepo(t *testing.T) {
	if err := EnsureInRepo(stubRunner{output: []byte("true\n")}); err != nil {
		t.Fatalf("EnsureInRepo returned error: %v", err)
	}
}

func TestEnsureInRepoReturnsNotGitRepoWhenCommandFails(t *testing.T) {
	err := EnsureInRepo(stubRunner{err: errors.New("git failed")})
	if !errors.Is(err, ErrNotGitRepo) {
		t.Fatalf("expected ErrNotGitRepo, got %v", err)
	}
}

func TestEnsureInRepoReturnsNotGitRepoWhenFalse(t *testing.T) {
	err := EnsureInRepo(stubRunner{output: []byte("false\n")})
	if !errors.Is(err, ErrNotGitRepo) {
		t.Fatalf("expected ErrNotGitRepo, got %v", err)
	}
}
