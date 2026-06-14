package app

import "testing"

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
