package app

import (
	"testing"

	"github.com/M-Xue/grove/branch"
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

func TestInitUsesInitialScreen(t *testing.T) {
	tests := []struct {
		name        string
		screen      ScreenID
		wantCommand bool
		wantLoading string
	}{
		{name: "change", screen: ScreenChange, wantCommand: true, wantLoading: "loading worktrees"},
		{name: "add", screen: ScreenAdd, wantCommand: false},
		{name: "branch", screen: ScreenBranch, wantCommand: true, wantLoading: "loading worktrees"},
	}

	for _, test := range tests {
		a := New(Services{}, WithInitialScreen(test.screen))
		cmd := a.Init()
		if test.wantCommand {
			if cmd == nil {
				t.Fatalf("%s: expected a command", test.name)
			}
			if len(a.State().Loading) != 1 || a.State().Loading[0].Message != test.wantLoading {
				t.Fatalf("%s: expected loading %q, got %#v", test.name, test.wantLoading, a.State().Loading)
			}
		} else if cmd != nil {
			t.Fatalf("%s: expected nil command, got %#v", test.name, cmd)
		}
	}
}

func TestSelectBranchLoadsCommitsForSelection(t *testing.T) {
	a := New(Services{})
	cmd := a.SelectBranch("feature/a")
	if cmd == nil {
		t.Fatal("expected a command to load commits")
	}
	if a.State().Branch.SelectedName != "feature/a" {
		t.Fatalf("unexpected selection: %q", a.State().Branch.SelectedName)
	}
	if len(a.State().Loading) != 1 || a.State().Loading[0].Message != "loading branch commits" {
		t.Fatalf("expected commit-loading entry, got %#v", a.State().Loading)
	}
}

func TestHandleBranchesLoadedRequestsCommitPreviewForSelection(t *testing.T) {
	a := New(Services{})
	cmd := a.HandleMessage(BranchesLoadedMessage{Branches: []branch.Info{{Name: "feature/a"}}})
	if cmd == nil {
		t.Fatal("expected a command to load commits")
	}
	if a.State().Branch.SelectedName != "feature/a" {
		t.Fatalf("unexpected branch name: %q", a.State().Branch.SelectedName)
	}
}

func TestBranchInitLoadsBranchesAfterWorktrees(t *testing.T) {
	a := New(Services{}, WithInitialScreen(ScreenBranch))

	if cmd := a.Init(); cmd == nil {
		t.Fatal("expected a command from Init")
	}

	worktreeID := a.State().Loading[0].ID
	next := a.HandleMessage(WorktreesLoadedMessage{LoadingID: worktreeID})
	if next == nil {
		t.Fatal("expected a command to load branches after worktrees")
	}

	state := a.State()
	if len(state.Loading) != 2 {
		t.Fatalf("expected two loading entries, got %#v", state.Loading)
	}
	if state.Loading[0].Message != "loading worktrees" || !state.Loading[0].Completed {
		t.Fatalf("unexpected first loading entry: %#v", state.Loading[0])
	}
	if state.Loading[1].Message != "loading branches" || state.Loading[1].Completed {
		t.Fatalf("unexpected second loading entry: %#v", state.Loading[1])
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

func TestStaleBranchCommitMessagesAreDropped(t *testing.T) {
	a := New(Services{})
	a.SelectBranch("feature-a") // seq 1
	a.SelectBranch("feature-b") // seq 2

	staleID := a.State().Loading[0].ID
	freshID := a.State().Loading[1].ID

	// A stale result (seq 1) must be dropped and only its loading entry removed.
	if cmd := a.HandleMessage(BranchCommitsLoadedMessage{
		LoadingID: staleID,
		Seq:       1,
		Name:      "feature-a",
		Commits:   []branch.CommitInfo{{Hash: "abc"}},
	}); cmd != nil {
		t.Fatal("expected nil command for stale message")
	}
	if len(a.State().Branch.Commits) != 0 {
		t.Fatalf("expected stale commits dropped, got %#v", a.State().Branch.Commits)
	}
	if a.State().Branch.SelectedName != "feature-b" {
		t.Fatalf("expected selection to remain feature-b, got %q", a.State().Branch.SelectedName)
	}

	// The fresh result (seq 2) is applied.
	a.HandleMessage(BranchCommitsLoadedMessage{
		LoadingID: freshID,
		Seq:       2,
		Name:      "feature-b",
		Commits:   []branch.CommitInfo{{Hash: "def"}},
	})
	if a.State().Branch.SelectedName != "feature-b" || len(a.State().Branch.Commits) != 1 {
		t.Fatalf("expected fresh commits applied, got %#v", a.State().Branch)
	}
}

func TestRequestCheckoutBranchRequiresSelection(t *testing.T) {
	a := New(Services{})
	if cmd := a.RequestCheckoutBranch(""); cmd != nil {
		t.Fatalf("expected nil command, got %#v", cmd)
	}
	if len(a.State().Statuses) != 1 {
		t.Fatalf("expected one status, got %d", len(a.State().Statuses))
	}
}

func TestDeleteBranchRequiresSelection(t *testing.T) {
	a := New(Services{})
	if cmd := a.DeleteBranch(""); cmd != nil {
		t.Fatalf("expected nil command, got %#v", cmd)
	}
	if len(a.State().Statuses) != 1 {
		t.Fatalf("expected one status, got %d", len(a.State().Statuses))
	}
}

func TestDeleteBranchReturnsCommand(t *testing.T) {
	a := New(Services{})
	cmd := a.DeleteBranch("feature/a")
	if cmd == nil {
		t.Fatal("expected a delete command")
	}
	if len(a.State().Loading) != 1 || a.State().Loading[0].Message != "deleting branch" {
		t.Fatalf("expected delete-branch loading entry, got %#v", a.State().Loading)
	}
}

func TestCanDeleteAllBranchesRequiresLocalScope(t *testing.T) {
	a := New(Services{})
	a.state.BranchScope = branch.ScopeRemoteTracking
	if a.CanDeleteAllBranches() {
		t.Fatal("expected false in remote-tracking scope")
	}
	if len(a.State().Statuses) != 1 {
		t.Fatalf("expected one status, got %d", len(a.State().Statuses))
	}
}

func TestCanDeleteAllBranchesRequiresLoadedBranches(t *testing.T) {
	a := New(Services{})
	a.state.BranchScope = branch.ScopeLocal
	if a.CanDeleteAllBranches() {
		t.Fatal("expected false with no branches")
	}
	if len(a.State().Statuses) != 1 {
		t.Fatalf("expected one status, got %d", len(a.State().Statuses))
	}
}

func TestCanDeleteAllBranchesAllowsLocalWithBranches(t *testing.T) {
	a := New(Services{})
	a.state.BranchScope = branch.ScopeLocal
	a.state.Branches = []branch.Info{{Name: "feature/a"}, {Name: "main"}}
	if !a.CanDeleteAllBranches() {
		t.Fatal("expected true for local scope with branches")
	}
	if len(a.State().Statuses) != 0 {
		t.Fatalf("expected no status, got %d", len(a.State().Statuses))
	}
}

func TestDeleteAllBranchesReturnsCommand(t *testing.T) {
	a := New(Services{})
	cmd := a.DeleteAllBranches()
	if cmd == nil {
		t.Fatal("expected a delete-all command")
	}
	if len(a.State().Loading) != 1 || a.State().Loading[0].Message != "deleting local branches" {
		t.Fatalf("expected delete-all loading entry, got %#v", a.State().Loading)
	}
}
