package app

import (
	"strings"
	"testing"

	branchsvc "github.com/M-Xue/grove/branch"
)

func TestRequestSubmitSelectedPathSetsOutput(t *testing.T) {
	a := New(Services{})
	effect := a.RequestSubmitSelectedPath("/repo")
	if _, ok := effect.(QuitEffect); !ok {
		t.Fatalf("expected QuitEffect, got %#v", effect)
	}
	if a.SubmittedPath() != "/repo" {
		t.Fatalf("unexpected submitted path: %q", a.SubmittedPath())
	}
}

func TestRequestAddWorktreeRequiresPathAndBranch(t *testing.T) {
	a := New(Services{})
	if effect := a.RequestAddWorktree("", "branch"); effect != nil {
		t.Fatal("expected nil effect for invalid path")
	}
	if len(a.State().Statuses) != 1 {
		t.Fatalf("expected one status, got %d", len(a.State().Statuses))
	}
}

func TestDialogChooseCancelClearsDialog(t *testing.T) {
	a := New(Services{})
	a.RequestRemoveWorktree("/repo")
	a.DialogChoose("cancel")
	if a.State().Dialog.Active {
		t.Fatal("expected dialog to clear")
	}
}

func TestHandleResultMarksLoadingDoneOnSuccess(t *testing.T) {
	a := New(Services{})
	a.Init()
	a.HandleResult(WorktreesLoadedResult{})
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
	a.setLoading("checking branch")
	a.setLoading("adding worktree")
	if len(a.State().Loading) != 2 {
		t.Fatalf("expected 2 loading entries, got %d", len(a.State().Loading))
	}
	a.markLoadingDone()
	a.markLoadingDone()
	state := a.State()
	if !state.Loading[0].Completed || !state.Loading[1].Completed {
		t.Fatalf("expected both loading entries completed, got %#v", state.Loading)
	}
}

func TestBranchCheckedPreservesCompletedCheckingPhase(t *testing.T) {
	a := New(Services{})
	a.setLoading("checking branch")
	effect := a.HandleResult(BranchCheckedResult{Path: "../repo", Branch: "feature", Exists: true})
	if _, ok := effect.(AddWorktreeEffect); !ok {
		t.Fatalf("expected AddWorktreeEffect, got %#v", effect)
	}
	state := a.State()
	if len(state.Loading) != 2 {
		t.Fatalf("expected 2 loading entries, got %d", len(state.Loading))
	}
	if state.Loading[0].Message != "checking branch" || !state.Loading[0].Completed {
		t.Fatalf("expected completed checking branch entry, got %#v", state.Loading[0])
	}
	if state.Loading[1].Message != "adding worktree" || state.Loading[1].Completed {
		t.Fatalf("expected active adding worktree entry, got %#v", state.Loading[1])
	}
}

func TestRequestCheckoutBranchRequiresSelection(t *testing.T) {
	a := New(Services{})
	if effect := a.RequestCheckoutBranch(""); effect != nil {
		t.Fatalf("expected nil effect, got %#v", effect)
	}
	if len(a.State().Statuses) != 1 {
		t.Fatalf("expected one status, got %d", len(a.State().Statuses))
	}
}

func TestRequestDeleteBranchRequiresSelection(t *testing.T) {
	a := New(Services{})
	if effect := a.RequestDeleteBranch(""); effect != nil {
		t.Fatalf("expected nil effect, got %#v", effect)
	}
	if len(a.State().Statuses) != 1 {
		t.Fatalf("expected one status, got %d", len(a.State().Statuses))
	}
}

func TestRequestDeleteBranchOpensConfirmationDialog(t *testing.T) {
	a := New(Services{})
	if effect := a.RequestDeleteBranch("feature/a"); effect != nil {
		t.Fatalf("expected nil effect, got %#v", effect)
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

func TestDialogChooseDeleteBranchReturnsDeleteEffect(t *testing.T) {
	a := New(Services{})
	a.RequestDeleteBranch("feature/a")
	effect := a.DialogChoose("confirm")
	deleteEffect, ok := effect.(DeleteBranchEffect)
	if !ok {
		t.Fatalf("expected DeleteBranchEffect, got %#v", effect)
	}
	if deleteEffect.Name != "feature/a" {
		t.Fatalf("unexpected branch name: %q", deleteEffect.Name)
	}
}

func TestRequestDeleteAllBranchesRequiresLocalScope(t *testing.T) {
	a := New(Services{})
	a.state.BranchScope = branchsvc.ScopeRemoteTracking
	if effect := a.RequestDeleteAllBranches(); effect != nil {
		t.Fatalf("expected nil effect, got %#v", effect)
	}
	if len(a.State().Statuses) != 1 {
		t.Fatalf("expected one status, got %d", len(a.State().Statuses))
	}
}

func TestRequestDeleteAllBranchesRequiresLoadedBranches(t *testing.T) {
	a := New(Services{})
	a.state.BranchScope = branchsvc.ScopeLocal
	if effect := a.RequestDeleteAllBranches(); effect != nil {
		t.Fatalf("expected nil effect, got %#v", effect)
	}
	if len(a.State().Statuses) != 1 {
		t.Fatalf("expected one status, got %d", len(a.State().Statuses))
	}
}

func TestRequestDeleteAllBranchesOpensConfirmationDialog(t *testing.T) {
	a := New(Services{})
	a.state.BranchScope = branchsvc.ScopeLocal
	a.state.Branches = []branchsvc.Info{{Name: "feature/a"}, {Name: "main"}}
	if effect := a.RequestDeleteAllBranches(); effect != nil {
		t.Fatalf("expected nil effect, got %#v", effect)
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

func TestDialogChooseDeleteAllBranchesReturnsDeleteEffect(t *testing.T) {
	a := New(Services{})
	a.state.BranchScope = branchsvc.ScopeLocal
	a.state.Branches = []branchsvc.Info{{Name: "feature/a"}}
	a.RequestDeleteAllBranches()
	effect := a.DialogChoose("confirm")
	_, ok := effect.(DeleteAllBranchesEffect)
	if !ok {
		t.Fatalf("expected DeleteAllBranchesEffect, got %#v", effect)
	}
}
