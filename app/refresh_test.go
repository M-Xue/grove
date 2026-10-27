package app

import (
	"errors"
	"testing"
	"time"

	"github.com/M-Xue/grove/worktree"
)

// stubWorktreeService returns a fixed list (or error) from List and is
// otherwise unused; it lets refresh commands be executed in tests.
type stubWorktreeService struct {
	worktree.Service
	list []worktree.Info
	err  error
}

func (s stubWorktreeService) List(worktree.ListOptions) ([]worktree.Info, error) {
	return s.list, s.err
}

func TestRefreshWorktreesShowsLoadingAndReloads(t *testing.T) {
	a := New(Services{Worktree: stubWorktreeService{list: []worktree.Info{{Path: "/repo", Branch: "main"}}}})
	cmd := a.RefreshWorktrees()
	if cmd == nil {
		t.Fatal("expected a refresh command")
	}
	if len(a.State().Loading) != 1 || a.State().Loading[0].Message != "refreshing worktrees" {
		t.Fatalf("expected a refreshing loading entry, got %#v", a.State().Loading)
	}
	a.HandleMessage(cmd())
	if got := a.State().Worktrees; len(got) != 1 || got[0].Branch != "main" {
		t.Fatalf("expected the refreshed list applied, got %#v", got)
	}
}

func TestBackgroundRefreshIsSilent(t *testing.T) {
	a := New(Services{Worktree: stubWorktreeService{list: []worktree.Info{{Path: "/repo", Branch: "feature"}}}})
	cmd := a.RefreshWorktreesInBackground()
	if cmd == nil {
		t.Fatal("expected a background refresh command")
	}
	if len(a.State().Loading) != 0 {
		t.Fatalf("expected no loading entry for a background refresh, got %#v", a.State().Loading)
	}
	a.HandleMessage(cmd())
	if got := a.State().Worktrees; len(got) != 1 || got[0].Branch != "feature" {
		t.Fatalf("expected the refreshed list applied, got %#v", got)
	}
	if len(a.State().Loading) != 0 {
		t.Fatalf("expected no loading entry after a background refresh, got %#v", a.State().Loading)
	}
}

func TestBackgroundRefreshSkipsWhileLoadPending(t *testing.T) {
	a := New(Services{Worktree: stubWorktreeService{}})
	first := a.RefreshWorktreesInBackground()
	if first == nil {
		t.Fatal("expected the first refresh to be issued")
	}
	if a.RefreshWorktreesInBackground() != nil {
		t.Fatal("expected no second refresh while the first is in flight")
	}
	a.HandleMessage(first())
	if a.RefreshWorktreesInBackground() == nil {
		t.Fatal("expected refreshes to resume once the load lands")
	}
}

func TestBackgroundRefreshSkipsWhileBusy(t *testing.T) {
	a := New(Services{Worktree: stubWorktreeService{}})
	a.setBlockingLoading("removing worktree")
	if a.RefreshWorktreesInBackground() != nil {
		t.Fatal("expected no background refresh during a blocking operation")
	}
}

func TestSupersededLoadIsDiscarded(t *testing.T) {
	var saved [][]worktree.Info
	a := New(Services{}, WithWorktreeCacheSaver(func(list []worktree.Info) { saved = append(saved, list) }))
	a.Init()
	oldSeq := a.worktreeLoadSeq
	oldID := a.State().Loading[0].ID
	a.RefreshWorktrees()
	newSeq := a.worktreeLoadSeq
	newID := a.worktreeLoadID

	// Starting the newer load drops the older load's loading entry.
	if entries := a.State().Loading; len(entries) != 1 || entries[0].ID != newID {
		t.Fatalf("expected only the newer load's entry, got %#v", entries)
	}

	fresh := []worktree.Info{{Path: "/repo", Branch: "new"}}
	stale := []worktree.Info{{Path: "/repo", Branch: "old"}}
	a.HandleMessage(WorktreesLoadedMessage{LoadingID: newID, Seq: newSeq, Worktrees: fresh})
	a.HandleMessage(WorktreesLoadedMessage{LoadingID: oldID, Seq: oldSeq, Worktrees: stale})

	if got := a.State().Worktrees; len(got) != 1 || got[0].Branch != "new" {
		t.Fatalf("expected the newer load to win, got %#v", got)
	}
	if len(saved) != 1 || saved[0][0].Branch != "new" {
		t.Fatalf("expected only the newer list persisted, got %#v", saved)
	}
}

func TestSupersededLoadErrorIsNotReported(t *testing.T) {
	a := New(Services{})
	a.Init()
	oldSeq := a.worktreeLoadSeq
	a.RefreshWorktrees()
	a.HandleMessage(WorktreesLoadedMessage{Seq: oldSeq, Err: errors.New("stale failure")})
	if len(a.State().Statuses) != 0 {
		t.Fatalf("expected a superseded load's error to be dropped, got %#v", a.State().Statuses)
	}
}

func TestUnchangedListDoesNotRewriteCache(t *testing.T) {
	saves := 0
	a := New(Services{}, WithWorktreeCacheSaver(func([]worktree.Info) { saves++ }))
	list := []worktree.Info{{Path: "/repo", Branch: "main"}}
	a.HandleMessage(WorktreesLoadedMessage{Worktrees: list})
	a.HandleMessage(WorktreesLoadedMessage{Worktrees: []worktree.Info{{Path: "/repo", Branch: "main"}}})
	if saves != 1 {
		t.Fatalf("expected one cache write for two identical loads, got %d", saves)
	}
	a.HandleMessage(WorktreesLoadedMessage{Worktrees: []worktree.Info{{Path: "/repo", Branch: "other"}}})
	if saves != 2 {
		t.Fatalf("expected a cache write when the list changes, got %d", saves)
	}
}

func TestLoadDoesNotClearSubmittedPath(t *testing.T) {
	// A background refresh landing between the user's enter and the quit it
	// requests must not wipe the path grove is about to print.
	a := New(Services{})
	a.RequestSubmitSelectedPath("/repo")
	a.HandleMessage(WorktreesLoadedMessage{Worktrees: []worktree.Info{{Path: "/repo"}}})
	if a.SubmittedPath() != "/repo" {
		t.Fatalf("expected the submitted path to survive a reload, got %q", a.SubmittedPath())
	}
}

// recordingWorktreeService records the options of every List call.
type recordingWorktreeService struct {
	worktree.Service
	opts []worktree.ListOptions
}

func (s *recordingWorktreeService) List(opts worktree.ListOptions) ([]worktree.Info, error) {
	s.opts = append(s.opts, opts)
	return nil, nil
}

func TestBackgroundRefreshListsWithoutLocks(t *testing.T) {
	svc := &recordingWorktreeService{}
	a := New(Services{Worktree: svc})
	a.HandleMessage(a.RefreshWorktreesInBackground()())
	if len(svc.opts) != 1 || !svc.opts[0].Background {
		t.Fatalf("expected a background refresh to request the lock-free listing, got %#v", svc.opts)
	}
}

func TestForegroundLoadsListWithLocks(t *testing.T) {
	svc := &recordingWorktreeService{}
	a := New(Services{Worktree: svc})
	a.HandleMessage(a.Init()())
	a.HandleMessage(a.RefreshWorktrees()())
	if len(svc.opts) != 2 {
		t.Fatalf("expected two listings, got %d", len(svc.opts))
	}
	for i, opts := range svc.opts {
		if opts.Background {
			t.Fatalf("expected foreground load %d to let git status refresh the index, got %#v", i, opts)
		}
	}
}

func TestUnchangedListAcrossTimeLocationsDoesNotRewriteCache(t *testing.T) {
	// A list decoded from the JSON cache carries fixed-offset times; the live
	// list carries Local ones. Equal instants must not count as a change, or
	// every launch would rewrite the cache it just read.
	saves := 0
	instant := time.Date(2026, 8, 18, 10, 0, 0, 0, time.UTC)
	seed := []worktree.Info{{Path: "/repo", Branch: "main", CreatedAt: instant.In(time.FixedZone("AEDT", 11*3600))}}
	a := New(Services{}, WithCachedWorktrees(seed), WithWorktreeCacheSaver(func([]worktree.Info) { saves++ }))
	a.HandleMessage(WorktreesLoadedMessage{Worktrees: []worktree.Info{{Path: "/repo", Branch: "main", CreatedAt: instant.In(time.Local)}}})
	if saves != 0 {
		t.Fatalf("expected no cache write for a list equal to the cached seed, got %d", saves)
	}
}

func TestLoadDurationIsRecorded(t *testing.T) {
	a := New(Services{})
	a.HandleMessage(WorktreesLoadedMessage{Elapsed: 1500 * time.Millisecond})
	if got := a.State().WorktreeLoadDuration; got != 1500*time.Millisecond {
		t.Fatalf("expected the load duration recorded in state, got %v", got)
	}
	// A failed load leaves the last successful duration in place.
	a.HandleMessage(WorktreesLoadedMessage{Elapsed: time.Minute, Err: errors.New("boom")})
	if got := a.State().WorktreeLoadDuration; got != 1500*time.Millisecond {
		t.Fatalf("expected a failed load not to update the duration, got %v", got)
	}
}
