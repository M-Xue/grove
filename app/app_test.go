package app

import (
	"testing"

	"github.com/M-Xue/grove/worktree"
)

func TestRequestSubmitSelectedPathRequestsQuit(t *testing.T) {
	a := New(Services{})
	cmd := a.RequestSubmitSelectedPath("/repo")
	if cmd == nil {
		t.Fatal("expected a command")
	}
	if a.SubmittedPath() != "/repo" {
		t.Fatalf("unexpected submitted path: %q", a.SubmittedPath())
	}
	if _, ok := cmd().(QuitRequested); !ok {
		t.Fatalf("expected QuitRequested message, got %#v", cmd())
	}
}

func TestInitLoadsWorktrees(t *testing.T) {
	a := New(Services{}, WithInitialScreen(ScreenChange))
	if cmd := a.Init(); cmd == nil {
		t.Fatal("expected a command")
	}
	if len(a.State().Loading) != 1 || a.State().Loading[0].Message != "loading worktrees" {
		t.Fatalf("expected loading worktrees entry, got %#v", a.State().Loading)
	}
}

func TestRequestAddWorktreeRequiresPathAndBranch(t *testing.T) {
	a := New(Services{})
	if cmd := a.RequestAddWorktree("", "branch"); cmd != nil {
		t.Fatal("expected nil command for invalid path")
	}
	if len(a.State().Statuses) != 1 {
		t.Fatalf("expected one status, got %d", len(a.State().Statuses))
	}
}

func TestRemoveWorktreeRequiresSelection(t *testing.T) {
	a := New(Services{})
	if cmd := a.RemoveWorktree(""); cmd != nil {
		t.Fatalf("expected nil command, got %#v", cmd)
	}
	if len(a.State().Statuses) != 1 {
		t.Fatalf("expected one status, got %d", len(a.State().Statuses))
	}
}

func TestRemoveWorktreeReturnsCommand(t *testing.T) {
	a := New(Services{})
	cmd := a.RemoveWorktree("/repo")
	if cmd == nil {
		t.Fatal("expected a remove command")
	}
	if len(a.State().Loading) != 1 || a.State().Loading[0].Message != "removing worktree" {
		t.Fatalf("expected remove-worktree loading entry, got %#v", a.State().Loading)
	}
}

func TestForceRemoveWorktreeRequiresSelection(t *testing.T) {
	a := New(Services{})
	if cmd := a.ForceRemoveWorktree(""); cmd != nil {
		t.Fatalf("expected nil command, got %#v", cmd)
	}
	if len(a.State().Statuses) != 1 {
		t.Fatalf("expected one status, got %d", len(a.State().Statuses))
	}
}

func TestForceRemoveWorktreeReturnsCommand(t *testing.T) {
	a := New(Services{})
	cmd := a.ForceRemoveWorktree("/repo")
	if cmd == nil {
		t.Fatal("expected a force-remove command")
	}
	if len(a.State().Loading) != 1 || a.State().Loading[0].Message != "removing worktree" {
		t.Fatalf("expected remove-worktree loading entry, got %#v", a.State().Loading)
	}
}

func TestPruneWorktreesReportsWhenNothingStale(t *testing.T) {
	a := New(Services{})
	a.HandleMessage(WorktreesLoadedMessage{Worktrees: []worktree.Info{{Path: "/repo", Stale: false}}})
	if cmd := a.PruneWorktrees(); cmd != nil {
		t.Fatalf("expected nil command when nothing is stale, got %#v", cmd)
	}
	if len(a.State().Loading) != 0 {
		t.Fatalf("expected no loading entry, got %#v", a.State().Loading)
	}
	if len(a.State().Statuses) != 1 {
		t.Fatalf("expected one status, got %d", len(a.State().Statuses))
	}
}

func TestPruneWorktreesReturnsCommandWhenStaleExists(t *testing.T) {
	a := New(Services{})
	a.HandleMessage(WorktreesLoadedMessage{Worktrees: []worktree.Info{
		{Path: "/repo", Stale: false},
		{Path: "/repo-gone", Stale: true},
	}})
	cmd := a.PruneWorktrees()
	if cmd == nil {
		t.Fatal("expected a prune command")
	}
	if len(a.State().Loading) != 1 || a.State().Loading[0].Message != "pruning stale worktrees" {
		t.Fatalf("expected prune loading entry, got %#v", a.State().Loading)
	}
}

func TestHandleMessageMarksLoadingDoneOnSuccess(t *testing.T) {
	a := New(Services{})
	a.Init()
	id := a.State().Loading[0].ID
	a.HandleMessage(WorktreesLoadedMessage{LoadingID: id})
	state := a.State()
	if len(state.Loading) != 1 || !state.Loading[0].Active || !state.Loading[0].Completed {
		t.Fatalf("expected completed loading state, got %#v", state.Loading)
	}
	if state.Loading[0].Message != "loading worktrees" {
		t.Fatalf("unexpected loading message: %q", state.Loading[0].Message)
	}
	a.DismissCompletedLoading()
	if len(a.State().Loading) != 0 {
		t.Fatal("expected completed loading to clear on dismiss")
	}
}

func TestHandleWorktreeProgressUpdatesOnlyMatchingEntry(t *testing.T) {
	a := New(Services{})
	other := a.setLoading("loading worktrees")
	id := a.setProgressLoading("creating branch and worktree")
	if !a.State().Loading[1].Progress {
		t.Fatalf("expected progress entry, got %#v", a.State().Loading[1])
	}

	a.HandleMessage(WorktreeProgressMessage{LoadingID: id, Done: 27, Total: 57})

	state := a.State()
	if state.Loading[1].Done != 27 || state.Loading[1].Total != 57 {
		t.Fatalf("expected progress 27/57, got %#v", state.Loading[1])
	}
	if state.Loading[0].Done != 0 || state.Loading[0].Total != 0 {
		t.Fatalf("progress leaked onto entry %q: %#v", other, state.Loading[0])
	}
}

func TestSequentialLoadingStatesArePreserved(t *testing.T) {
	a := New(Services{})
	id1 := a.setLoading("checking branch")
	id2 := a.setLoading("adding worktree")
	if len(a.State().Loading) != 2 {
		t.Fatalf("expected 2 loading entries, got %d", len(a.State().Loading))
	}
	a.markLoadingDone(id1)
	a.markLoadingDone(id2)
	state := a.State()
	if !state.Loading[0].Completed || !state.Loading[1].Completed {
		t.Fatalf("expected both loading entries completed, got %#v", state.Loading)
	}
}

func TestConcurrentLoadingClearsOnlyCompletedEntry(t *testing.T) {
	a := New(Services{})
	id1 := a.setLoading("first")
	id2 := a.setLoading("second")
	a.markLoadingDone(id1)
	state := a.State()
	if len(state.Loading) != 2 {
		t.Fatalf("expected both entries to remain, got %#v", state.Loading)
	}
	if !state.Loading[0].Completed {
		t.Fatalf("expected first entry completed, got %#v", state.Loading[0])
	}
	if state.Loading[1].Completed {
		t.Fatalf("expected second entry still pending, got %#v", state.Loading[1])
	}
	_ = id2
}

func TestBranchExistsPreservesCompletedCheckingPhase(t *testing.T) {
	a := New(Services{})
	id := a.setLoading("checking branch")
	cmd := a.HandleMessage(BranchExistsMessage{LoadingID: id, Path: "../repo", Branch: "feature"})
	if cmd == nil {
		t.Fatal("expected an add-worktree command")
	}
	state := a.State()
	if len(state.Loading) != 2 {
		t.Fatalf("expected 2 loading entries, got %#v", state.Loading)
	}
	if state.Loading[0].Message != "checking branch" || !state.Loading[0].Completed {
		t.Fatalf("expected completed checking branch entry, got %#v", state.Loading[0])
	}
	if state.Loading[1].Message != "creating worktree" || state.Loading[1].Completed {
		t.Fatalf("expected active creating worktree entry, got %#v", state.Loading[1])
	}
}

func TestBranchAbsentResolvesCheckWithoutChaining(t *testing.T) {
	a := New(Services{})
	id := a.setLoading("checking branch")
	cmd := a.HandleMessage(BranchAbsentMessage{LoadingID: id, Path: "../repo", Branch: "feature"})
	if cmd != nil {
		t.Fatalf("expected nil command (UI opens the dialog), got %#v", cmd)
	}
	state := a.State()
	if len(state.Loading) != 1 || !state.Loading[0].Completed {
		t.Fatalf("expected checking-branch entry marked done, got %#v", state.Loading)
	}
}
