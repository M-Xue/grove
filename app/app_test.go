package app

import (
	"strings"
	"testing"

	"github.com/M-Xue/grove/branch"
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

func TestInitUsesInitialScreenEffect(t *testing.T) {
	tests := []struct {
		name   string
		screen ScreenID
		want   any
	}{
		{name: "change", screen: ScreenChange, want: LoadWorktreesEffect{}},
		{name: "add", screen: ScreenAdd, want: nil},
		{name: "branch", screen: ScreenBranch, want: LoadWorktreesEffect{}},
	}

	for _, test := range tests {
		a := New(Services{}, WithInitialScreen(test.screen))
		got := a.Init()
		switch want := test.want.(type) {
		case nil:
			if got != nil {
				t.Fatalf("%s: expected nil effect, got %#v", test.name, got)
			}
		default:
			if got != want {
				t.Fatalf("%s: expected %#v, got %#v", test.name, want, got)
			}
		}
	}
}

func TestSelectBranchLoadsCommitsForSelection(t *testing.T) {
	a := New(Services{})
	effect := a.SelectBranch("feature/a")
	loadEffect, ok := effect.(LoadBranchCommitsEffect)
	if !ok {
		t.Fatalf("expected LoadBranchCommitsEffect, got %#v", effect)
	}
	if loadEffect.Name != "feature/a" || loadEffect.Limit != branchCommitPreviewLimit {
		t.Fatalf("unexpected load effect: %#v", loadEffect)
	}
}

func TestHandleBranchesLoadedRequestsCommitPreviewForSelection(t *testing.T) {
	a := New(Services{})
	effect := a.HandleResult(BranchesLoadedResult{Branches: []branch.Info{{Name: "feature/a"}}})
	loadEffect, ok := effect.(LoadBranchCommitsEffect)
	if !ok {
		t.Fatalf("expected LoadBranchCommitsEffect, got %#v", effect)
	}
	if loadEffect.Name != "feature/a" {
		t.Fatalf("unexpected branch name: %q", loadEffect.Name)
	}
}

func TestBranchInitLoadsBranchesAfterWorktrees(t *testing.T) {
	a := New(Services{}, WithInitialScreen(ScreenBranch))

	effect := a.Init()
	if effect != (LoadWorktreesEffect{}) {
		t.Fatalf("expected LoadWorktreesEffect, got %#v", effect)
	}

	effect = a.HandleResult(WorktreesLoadedResult{})
	if effect != (LoadBranchesEffect{}) {
		t.Fatalf("expected LoadBranchesEffect after worktrees, got %#v", effect)
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
	a.state.BranchScope = branch.ScopeRemoteTracking
	if effect := a.RequestDeleteAllBranches(); effect != nil {
		t.Fatalf("expected nil effect, got %#v", effect)
	}
	if len(a.State().Statuses) != 1 {
		t.Fatalf("expected one status, got %d", len(a.State().Statuses))
	}
}

func TestRequestDeleteAllBranchesRequiresLoadedBranches(t *testing.T) {
	a := New(Services{})
	a.state.BranchScope = branch.ScopeLocal
	if effect := a.RequestDeleteAllBranches(); effect != nil {
		t.Fatalf("expected nil effect, got %#v", effect)
	}
	if len(a.State().Statuses) != 1 {
		t.Fatalf("expected one status, got %d", len(a.State().Statuses))
	}
}

func TestRequestDeleteAllBranchesOpensConfirmationDialog(t *testing.T) {
	a := New(Services{})
	a.state.BranchScope = branch.ScopeLocal
	a.state.Branches = []branch.Info{{Name: "feature/a"}, {Name: "main"}}
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
	a.state.BranchScope = branch.ScopeLocal
	a.state.Branches = []branch.Info{{Name: "feature/a"}}
	a.RequestDeleteAllBranches()
	effect := a.DialogChoose("confirm")
	_, ok := effect.(DeleteAllBranchesEffect)
	if !ok {
		t.Fatalf("expected DeleteAllBranchesEffect, got %#v", effect)
	}
}
