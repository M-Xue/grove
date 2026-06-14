package main

import (
	"testing"

	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/docs"
	"github.com/M-Xue/grove/ui"
	"github.com/M-Xue/grove/worktree"
)

func TestSelectedPathOutputReturnsSubmittedPath(t *testing.T) {
	m := ui.New(app.New(app.Services{
		Worktree: worktree.NewServiceWithRunner(nil),
		Docs:     docs.NewService(),
	}))

	if got := selectedPathOutput(m); got != "" {
		t.Fatalf("expected empty path, got %q", got)
	}
}
