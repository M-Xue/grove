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
