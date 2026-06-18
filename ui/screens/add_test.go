package screens

import (
	"testing"

	"github.com/M-Xue/grove/app"
)

func TestAddScreenOpensCreateDialogOnBranchAbsent(t *testing.T) {
	s := NewAddScreen()
	ctx := &ScreenContext{}
	if s.confirm.active {
		t.Fatal("dialog should start closed")
	}
	s.OnMessage(ctx, app.BranchAbsentMessage{Path: "../feature", Branch: "feature"})
	if !s.confirm.active {
		t.Fatal("expected the create-branch dialog to open on BranchAbsentMessage")
	}
}

func TestAddScreenIgnoresUnrelatedMessages(t *testing.T) {
	s := NewAddScreen()
	s.OnMessage(&ScreenContext{}, app.BranchExistsMessage{Path: "../feature", Branch: "feature"})
	if s.confirm.active {
		t.Fatal("did not expect a dialog for BranchExistsMessage")
	}
}
