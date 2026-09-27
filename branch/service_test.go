package branch

import (
	"errors"
	"reflect"
	"testing"
)

type commandResult struct {
	output []byte
	err    error
}

type stubRunner struct {
	results map[string]commandResult
}

func (s *stubRunner) CombinedOutput(name string, args ...string) ([]byte, error) {
	result, ok := s.results[commandKey(name, args...)]
	if !ok {
		return nil, errors.New("unexpected command")
	}
	return result.output, result.err
}

func commandKey(name string, args ...string) string {
	joined := name
	for _, arg := range args {
		joined += "\x00" + arg
	}
	return joined
}

func TestRecentCommitsParsesGitLogOutput(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "log", "-n10", "--format=%h%x1f%an%x1f%s", "feature/a"): {
				output: []byte("abc123\x1fMax\x1fAdd auth flow\ndef456\x1fSam\x1fFix tests\n"),
			},
		},
	}

	service := NewService(runner)
	commits, err := service.RecentCommits("feature/a", 10)
	if err != nil {
		t.Fatalf("RecentCommits returned error: %v", err)
	}
	want := []CommitInfo{
		{Hash: "abc123", Author: "Max", Subject: "Add auth flow"},
		{Hash: "def456", Author: "Sam", Subject: "Fix tests"},
	}
	if !reflect.DeepEqual(commits, want) {
		t.Fatalf("unexpected commits: got %#v want %#v", commits, want)
	}
}

func TestRecentCommitsUsesRequestedLimit(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "log", "-n5", "--format=%h%x1f%an%x1f%s", "main"): {
				output: []byte("abc123\x1fMax\x1fInitial commit\n"),
			},
		},
	}

	service := NewService(runner)
	commits, err := service.RecentCommits("main", 5)
	if err != nil {
		t.Fatalf("RecentCommits returned error: %v", err)
	}
	if len(commits) != 1 {
		t.Fatalf("unexpected commit count: %d", len(commits))
	}
}

func TestRecentCommitsRequiresBranchName(t *testing.T) {
	service := NewService(&stubRunner{})
	if _, err := service.RecentCommits("  ", 5); err == nil {
		t.Fatal("expected an error for a blank branch name")
	}
}

func TestRecentCommitsEmptyOutputYieldsNoCommits(t *testing.T) {
	runner := &stubRunner{
		results: map[string]commandResult{
			commandKey("git", "log", "-n5", "--format=%h%x1f%an%x1f%s", "empty"): {
				output: []byte("\n"),
			},
		},
	}

	service := NewService(runner)
	commits, err := service.RecentCommits("empty", 5)
	if err != nil {
		t.Fatalf("RecentCommits returned error: %v", err)
	}
	if len(commits) != 0 {
		t.Fatalf("expected no commits, got %#v", commits)
	}
}
