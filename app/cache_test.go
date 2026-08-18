package app

import (
	"errors"
	"testing"

	"github.com/M-Xue/grove/worktree"
)

func TestWithCachedWorktreesSeedsStateBeforeInit(t *testing.T) {
	seed := []worktree.Info{{Path: "/repo", Branch: "main"}}
	a := New(Services{}, WithCachedWorktrees(seed))

	got := a.State().Worktrees
	if len(got) != 1 || got[0].Path != "/repo" {
		t.Fatalf("expected seeded worktrees, got %#v", got)
	}
}

func TestWorktreesLoadedSavesFreshListToCache(t *testing.T) {
	var saved []worktree.Info
	saver := func(list []worktree.Info) { saved = list }
	a := New(Services{}, WithWorktreeCacheSaver(saver))

	fresh := []worktree.Info{{Path: "/repo", Branch: "main"}, {Path: "/repo-x", Branch: "x"}}
	a.HandleMessage(WorktreesLoadedMessage{Worktrees: fresh})

	if len(saved) != 2 || saved[0].Path != "/repo" || saved[1].Path != "/repo-x" {
		t.Fatalf("expected fresh list saved to cache, got %#v", saved)
	}
}

func TestWorktreesLoadedDoesNotSaveOnError(t *testing.T) {
	called := false
	saver := func(list []worktree.Info) { called = true }
	a := New(Services{}, WithWorktreeCacheSaver(saver))

	a.HandleMessage(WorktreesLoadedMessage{Err: errors.New("git blew up")})

	if called {
		t.Fatal("cache saver should not run when the load failed")
	}
}

func TestWorktreesLoadedWithoutSaverDoesNotPanic(t *testing.T) {
	a := New(Services{})
	// No saver injected; handling a successful load must not panic.
	a.HandleMessage(WorktreesLoadedMessage{Worktrees: []worktree.Info{{Path: "/repo"}}})
}
