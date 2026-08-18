package screens

import (
	"strings"
	"testing"

	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/worktree"
)

func TestChangeScreenLabelsLockedAndStaleWorktrees(t *testing.T) {
	s := NewChangeScreen(fakeApp{})
	s.Sync(app.State{Worktrees: []worktree.Info{
		{Path: "/repo", Branch: "main"},
		{Path: "/gone", Branch: "old", Stale: true},
		{Path: "/pinned", Branch: "pinned", Locked: true},
	}})

	view := s.View(120, 40, app.State{})

	if !strings.Contains(view, "/gone [old] [stale]") {
		t.Fatalf("expected stale worktree to be labelled [stale], got:\n%s", view)
	}
	if !strings.Contains(view, "/pinned [pinned] [locked]") {
		t.Fatalf("expected locked worktree to be labelled [locked], got:\n%s", view)
	}
	// A locked worktree is distinct from a stale one; it must not be mislabelled.
	if strings.Contains(view, "/pinned [pinned] [stale]") {
		t.Fatalf("locked worktree must not be labelled [stale], got:\n%s", view)
	}

	// Stale worktrees are dimmed, but locked ones keep their normal colour.
	if !strings.Contains(view, staleColor+"/gone") {
		t.Fatalf("expected stale worktree to be dimmed, got:\n%s", view)
	}
	if strings.Contains(view, staleColor+"/pinned") {
		t.Fatalf("locked worktree must not be dimmed, got:\n%s", view)
	}
}

func TestChangeScreenRemoveDialogWarnsWhenDirty(t *testing.T) {
	cases := []struct {
		name      string
		worktree  worktree.Info
		wantTitle string
	}{
		{"clean", worktree.Info{Path: "/repo", Branch: "main"}, "Delete worktree?"},
		{"uncommitted", worktree.Info{Path: "/repo", Branch: "main", HasUncommittedChanges: true}, "Force delete worktree?"},
		{"untracked", worktree.Info{Path: "/repo", Branch: "main", HasUntrackedFiles: true}, "Force delete worktree?"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewChangeScreen(fakeApp{})
			s.Sync(app.State{Worktrees: []worktree.Info{tc.worktree}})

			s.actionStartRemove(&ActionCtx{})
			if !s.confirm.active {
				t.Fatal("expected the confirm dialog to open")
			}

			view := s.View(120, 40, app.State{})
			if !strings.Contains(view, tc.wantTitle) {
				t.Fatalf("expected dialog titled %q, got:\n%s", tc.wantTitle, view)
			}
		})
	}
}
