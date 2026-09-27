package app

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/pr"
)

type stubPRService struct {
	info   pr.Info
	found  bool
	err    error
	called []string
}

func (s *stubPRService) ForBranch(name string) (pr.Info, bool, error) {
	s.called = append(s.called, name)
	return s.info, s.found, s.err
}

func TestLoadBranchPRStoresResultByBranch(t *testing.T) {
	info := pr.Info{Number: 42, Title: "Add auth flow", Status: pr.StatusOpen, Author: "max", URL: "https://example.com/pull/42"}
	service := &stubPRService{info: info, found: true}
	a := New(Services{PR: service})

	cmd := a.LoadBranchPR("feature")
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
	got := a.State().BranchPRs["feature"]
	if !got.Found || !reflect.DeepEqual(got.Info, info) {
		t.Fatalf("unexpected stored PR: %#v", got)
	}
}

func TestLoadBranchPRCachesNoPROutcome(t *testing.T) {
	service := &stubPRService{found: false}
	a := New(Services{PR: service})

	a.HandleMessage(a.LoadBranchPR("feature")())
	entry, cached := a.State().BranchPRs["feature"]
	if !cached || entry.Found {
		t.Fatalf("expected a cached no-PR entry, got %#v cached=%v", entry, cached)
	}

	// The definitive "no PR" answer is cached: no second lookup is issued.
	if cmd := a.LoadBranchPR("feature"); cmd != nil {
		t.Fatal("expected nil command for an already-fetched branch")
	}
}

func TestLoadBranchPRSkipsBlankBranchAndMissingService(t *testing.T) {
	a := New(Services{PR: &stubPRService{}})
	if cmd := a.LoadBranchPR(""); cmd != nil {
		t.Fatal("expected nil command for a blank branch name")
	}
	if cmd := New(Services{}).LoadBranchPR("main"); cmd != nil {
		t.Fatal("expected nil command without a PR service")
	}
}

func TestLoadBranchPRUnavailableLatchesKillSwitch(t *testing.T) {
	service := &stubPRService{err: fmt.Errorf("%w: gh not found", pr.ErrUnavailable)}
	a := New(Services{PR: service})

	if next := a.HandleMessage(a.LoadBranchPR("feature")()); next != nil {
		t.Fatal("expected no chained command")
	}
	if !a.State().PRLookupUnavailable {
		t.Fatal("expected PR lookups to be marked unavailable")
	}
	if len(a.State().Statuses) != 0 {
		t.Fatalf("expected no status entries, got %#v", a.State().Statuses)
	}
	if cmd := a.LoadBranchPR("other"); cmd != nil {
		t.Fatal("expected no further lookups once unavailable")
	}
}

func TestLoadBranchPRTransientErrorsAreRetried(t *testing.T) {
	service := &stubPRService{err: errors.New("network is unreachable")}
	a := New(Services{PR: service})

	a.HandleMessage(a.LoadBranchPR("feature")())
	if a.State().PRLookupUnavailable {
		t.Fatal("transient errors must not latch the kill switch")
	}
	// The failure is recorded — so the panel can resolve its pending state —
	// but marked Failed, which LoadBranchPR treats as uncached.
	entry, resolved := a.State().BranchPRs["feature"]
	if !resolved || !entry.Failed || entry.Found {
		t.Fatalf("expected a failed entry, got %#v resolved=%v", entry, resolved)
	}
	// The next hover retries, and a success replaces the failed entry.
	cmd := a.LoadBranchPR("feature")
	if cmd == nil {
		t.Fatal("expected a retry command after a transient failure")
	}
	service.err, service.found, service.info = nil, true, pr.Info{Number: 9}
	a.HandleMessage(cmd())
	if got := a.State().BranchPRs["feature"]; got.Failed || !got.Found || got.Info.Number != 9 {
		t.Fatalf("expected the retry to replace the failed entry, got %#v", got)
	}
}

func TestBranchCommitsLoadChainsPRLookup(t *testing.T) {
	prService := &stubPRService{found: true, info: pr.Info{Number: 7}}
	a := New(Services{Branch: &stubBranchService{}, PR: prService})

	next := a.HandleMessage(BranchCommitsLoadedMessage{
		Branch:  "feature",
		Commits: []branch.CommitInfo{{Hash: "abc123"}},
	})
	if next == nil {
		t.Fatal("expected the commits load to chain a PR lookup")
	}
	a.HandleMessage(next())
	if got := a.State().BranchPRs["feature"]; !got.Found || got.Info.Number != 7 {
		t.Fatalf("unexpected stored PR after chain: %#v", got)
	}
	if len(prService.called) != 1 || prService.called[0] != "feature" {
		t.Fatalf("unexpected PR service calls: %v", prService.called)
	}

	// A repeat commits load (e.g. after a worktree reload) must not refetch
	// the PR — the chain is cut by LoadBranchPR's cache.
	if next := a.HandleMessage(BranchCommitsLoadedMessage{Branch: "feature"}); next != nil {
		t.Fatal("expected no PR chain for an already-fetched branch")
	}
}
