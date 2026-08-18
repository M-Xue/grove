package repo

import (
	"errors"
	"testing"
)

func TestWorktreeCacheKeyReturnsCommonDir(t *testing.T) {
	runner := stubRunner{output: []byte("/home/user/project/.git\n")}
	key, err := WorktreeCacheKey(runner)
	if err != nil {
		t.Fatalf("WorktreeCacheKey returned error: %v", err)
	}
	if key != "/home/user/project/.git" {
		t.Fatalf("expected trimmed common dir, got %q", key)
	}
}

func TestWorktreeCacheKeyIsStableForSameRepo(t *testing.T) {
	runner := stubRunner{output: []byte("/home/user/project/.git\n")}
	first, err := WorktreeCacheKey(runner)
	if err != nil {
		t.Fatalf("first call errored: %v", err)
	}
	second, err := WorktreeCacheKey(runner)
	if err != nil {
		t.Fatalf("second call errored: %v", err)
	}
	if first != second {
		t.Fatalf("key not stable: %q != %q", first, second)
	}
}

func TestWorktreeCacheKeyPropagatesRunnerError(t *testing.T) {
	runner := stubRunner{err: errors.New("git failed")}
	if _, err := WorktreeCacheKey(runner); err == nil {
		t.Fatal("expected an error when the runner fails")
	}
}

func TestWorktreeCacheKeyErrorsOnEmptyOutput(t *testing.T) {
	runner := stubRunner{output: []byte("  \n")}
	if _, err := WorktreeCacheKey(runner); err == nil {
		t.Fatal("expected an error when git returns no common dir")
	}
}
