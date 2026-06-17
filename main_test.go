package main

import (
	"testing"

	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/command"
	"github.com/M-Xue/grove/ui"
	"github.com/M-Xue/grove/worktree"
)

func TestSelectedPathOutputReturnsSubmittedPath(t *testing.T) {
	runner := command.New()
	m := ui.New(app.New(app.Services{
		Worktree: worktree.NewService(runner),
		Branch:   branch.NewService(runner),
	}))

	if got := selectedPathOutput(m); got != "" {
		t.Fatalf("expected empty path, got %q", got)
	}
}
