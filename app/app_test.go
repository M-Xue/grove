package app

import (
	"strings"
	"testing"

	"github.com/M-Xue/grove/branch"
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

func TestDialogChooseCancelClearsDialog(t *testing.T) {
	a := New(Services{})
	a.RequestRemoveWorktree("/repo")
	if cmd := a.DialogChoose("cancel"); cmd != nil {
		t.Fatal("expected nil command on cancel")
	}
	if a.State().Dialog.Active {
		t.Fatal("expected dialog to clear")
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

func TestBranchCheckedPreservesCompletedCheckingPhase(t *testing.T) {
	a := New(Services{})
	id := a.setLoading("checking branch")
	cmd := a.HandleMessage(BranchCheckedMessage{LoadingID: id, Path: "../repo", Branch: "feature", Exists: true})
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
	if state.Loading[1].Message != "adding worktree" || state.Loading[1].Completed {
		t.Fatalf("expected active adding worktree entry, got %#v", state.Loading[1])
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

func TestRequestDeleteBranchRequiresSelection(t *testing.T) {
	a := New(Services{})
	if cmd := a.RequestDeleteBranch(""); cmd != nil {
		t.Fatalf("expected nil command, got %#v", cmd)
	}
	if len(a.State().Statuses) != 1 {
		t.Fatalf("expected one status, got %d", len(a.State().Statuses))
	}
}

func TestRequestDeleteBranchOpensConfirmationDialog(t *testing.T) {
	a := New(Services{})
	if cmd := a.RequestDeleteBranch("feature/a"); cmd != nil {
		t.Fatalf("expected nil command, got %#v", cmd)
	}
	state := a.State()
	if !state.Dialog.Active {
		t.Fatal("expected dialog to be active")
	}
	if state.Dialog.Kind != DialogConfirmDeleteBranch {
		t.Fatalf("unexpected dialog kind: %q", state.Dialog.Kind)
	}
	if state.Dialog.Branch != "feature/a" {
		t.Fatalf("unexpected dialog branch: %q", state.Dialog.Branch)
	}
}

func TestDialogChooseDeleteBranchReturnsCommand(t *testing.T) {
	a := New(Services{})
	a.RequestDeleteBranch("feature/a")
	cmd := a.DialogChoose("confirm")
	if cmd == nil {
		t.Fatal("expected a delete command")
	}
	if len(a.State().Loading) != 1 || a.State().Loading[0].Message != "deleting branch" {
		t.Fatalf("expected delete-branch loading entry, got %#v", a.State().Loading)
	}
}

func TestRequestDeleteAllBranchesRequiresLocalScope(t *testing.T) {
	a := New(Services{})
	a.state.BranchScope = branch.ScopeRemoteTracking
	if cmd := a.RequestDeleteAllBranches(); cmd != nil {
		t.Fatalf("expected nil command, got %#v", cmd)
	}
	if len(a.State().Statuses) != 1 {
		t.Fatalf("expected one status, got %d", len(a.State().Statuses))
	}
}

func TestRequestDeleteAllBranchesRequiresLoadedBranches(t *testing.T) {
	a := New(Services{})
	a.state.BranchScope = branch.ScopeLocal
	if cmd := a.RequestDeleteAllBranches(); cmd != nil {
		t.Fatalf("expected nil command, got %#v", cmd)
	}
	if len(a.State().Statuses) != 1 {
		t.Fatalf("expected one status, got %d", len(a.State().Statuses))
	}
}

func TestRequestDeleteAllBranchesOpensConfirmationDialog(t *testing.T) {
	a := New(Services{})
	a.state.BranchScope = branch.ScopeLocal
	a.state.Branches = []branch.Info{{Name: "feature/a"}, {Name: "main"}}
	if cmd := a.RequestDeleteAllBranches(); cmd != nil {
		t.Fatalf("expected nil command, got %#v", cmd)
	}
	state := a.State()
	if !state.Dialog.Active {
		t.Fatal("expected dialog to be active")
	}
	if state.Dialog.Kind != DialogConfirmDeleteAllBranches {
		t.Fatalf("unexpected dialog kind: %q", state.Dialog.Kind)
	}
	if !strings.Contains(state.Dialog.Description, "feature/a") || !strings.Contains(state.Dialog.Description, "main") {
		t.Fatalf("unexpected dialog description: %q", state.Dialog.Description)
	}
}

func TestDialogChooseDeleteAllBranchesReturnsCommand(t *testing.T) {
	a := New(Services{})
	a.state.BranchScope = branch.ScopeLocal
	a.state.Branches = []branch.Info{{Name: "feature/a"}}
	a.RequestDeleteAllBranches()
	cmd := a.DialogChoose("confirm")
	if cmd == nil {
		t.Fatal("expected a delete-all command")
	}
	if len(a.State().Loading) != 1 || a.State().Loading[0].Message != "deleting local branches" {
		t.Fatalf("expected delete-all loading entry, got %#v", a.State().Loading)
	}
}
