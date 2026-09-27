package app

import (
	"errors"
	"reflect"
	"testing"

	"github.com/M-Xue/grove/branch"
)

type stubBranchService struct {
	commits []branch.CommitInfo
	err     error
	called  []string
}

func (s *stubBranchService) RecentCommits(name string, limit int) ([]branch.CommitInfo, error) {
	s.called = append(s.called, name)
	if limit != RecentCommitLimit {
		return nil, errors.New("unexpected limit")
	}
	return s.commits, s.err
}

func TestLoadBranchCommitsStoresCommitsByBranch(t *testing.T) {
	commits := []branch.CommitInfo{{Hash: "abc123", Author: "Max", Subject: "Add auth flow"}}
	service := &stubBranchService{commits: commits}
	a := New(Services{Branch: service})

	cmd := a.LoadBranchCommits("feature")
	if cmd == nil {
		t.Fatal("expected a command")
	}
	// The fetch is silent: no loading entry accompanies it.
	if len(a.State().Loading) != 0 {
		t.Fatalf("expected no loading entries, got %#v", a.State().Loading)
	}

	if next := a.HandleMessage(cmd()); next != nil {
		t.Fatal("expected no chained command")
	}
	if got := a.State().BranchCommits["feature"]; !reflect.DeepEqual(got, commits) {
		t.Fatalf("unexpected stored commits: %#v", got)
	}
}

func TestLoadBranchCommitsSkipsBlankBranchAndMissingService(t *testing.T) {
	a := New(Services{Branch: &stubBranchService{}})
	if cmd := a.LoadBranchCommits(""); cmd != nil {
		t.Fatal("expected nil command for a blank branch name")
	}
	if cmd := New(Services{}).LoadBranchCommits("main"); cmd != nil {
		t.Fatal("expected nil command without a branch service")
	}
}

func TestLoadBranchCommitsErrorResolvesQuietlyAsEmpty(t *testing.T) {
	service := &stubBranchService{err: errors.New("boom")}
	a := New(Services{Branch: service})

	cmd := a.LoadBranchCommits("feature")
	if next := a.HandleMessage(cmd()); next != nil {
		t.Fatal("expected no chained command")
	}
	if len(a.State().Statuses) != 0 {
		t.Fatalf("expected no status entries, got %#v", a.State().Statuses)
	}
	// The failure resolves as an empty entry (so the panel's pending state
	// clears) rather than being surfaced or left unresolved.
	commits, resolved := a.State().BranchCommits["feature"]
	if !resolved || len(commits) != 0 {
		t.Fatalf("expected an empty resolved entry, got %#v resolved=%v", commits, resolved)
	}
}

func TestLoadBranchCommitsFailedRefreshKeepsOldCommits(t *testing.T) {
	commits := []branch.CommitInfo{{Hash: "abc123", Author: "Max", Subject: "Add auth flow"}}
	service := &stubBranchService{commits: commits}
	a := New(Services{Branch: service})
	a.HandleMessage(a.LoadBranchCommits("feature")())

	service.commits, service.err = nil, errors.New("boom")
	a.HandleMessage(a.LoadBranchCommits("feature")())
	if got := a.State().BranchCommits["feature"]; !reflect.DeepEqual(got, commits) {
		t.Fatalf("expected the failed refresh to keep old commits, got %#v", got)
	}
}
