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
	if limit != recentCommitLimit {
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

func TestLoadBranchCommitsIgnoresErrorsSilently(t *testing.T) {
	service := &stubBranchService{err: errors.New("boom")}
	a := New(Services{Branch: service})

	cmd := a.LoadBranchCommits("feature")
	if next := a.HandleMessage(cmd()); next != nil {
		t.Fatal("expected no chained command")
	}
	if len(a.State().Statuses) != 0 {
		t.Fatalf("expected no status entries, got %#v", a.State().Statuses)
	}
	if _, ok := a.State().BranchCommits["feature"]; ok {
		t.Fatal("expected no commits stored on error")
	}
}
