package screens

import (
	"regexp"
	"strings"
	"testing"

	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/worktree"
	tea "github.com/charmbracelet/bubbletea"
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

var ansiSequence = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripAnsiSequences(s string) string {
	return ansiSequence.ReplaceAllString(s, "")
}

// commitsRecorderApp records LoadBranchCommits calls.
type commitsRecorderApp struct {
	fakeApp
	branches []string
}

func (a *commitsRecorderApp) LoadBranchCommits(branch string) app.Command {
	a.branches = append(a.branches, branch)
	return func() app.Message { return nil }
}

func TestChangeScreenFetchesCommitsWhenHoverChangesBranch(t *testing.T) {
	recorder := &commitsRecorderApp{}
	s := NewChangeScreen(recorder)
	s.Sync(app.State{Worktrees: []worktree.Info{
		{Path: "/repo", Branch: "main"},
		{Path: "/repo-feature", Branch: "feature"},
	}})

	if cmd := s.fetchHoveredCommits(); cmd == nil {
		t.Fatal("expected the initial hover to fetch commits")
	}
	if len(recorder.branches) != 1 || recorder.branches[0] != "main" {
		t.Fatalf("expected a fetch for main, got %v", recorder.branches)
	}

	// Re-fetching for an unchanged hover is skipped.
	if cmd := s.fetchHoveredCommits(); cmd != nil {
		t.Fatal("expected no refetch while the hovered branch is unchanged")
	}

	// Moving the hover to another worktree fetches its branch.
	if cmd := s.actionMoveSelection(&ActionCtx{Key: keyMsg(tea.KeyDown)}); cmd == nil {
		t.Fatal("expected the hover move to fetch commits")
	}
	if len(recorder.branches) != 2 || recorder.branches[1] != "feature" {
		t.Fatalf("expected a fetch for feature, got %v", recorder.branches)
	}
}

func TestChangeScreenDetailsPanelShowsHoveredWorktree(t *testing.T) {
	s := NewChangeScreen(fakeApp{})
	state := app.State{
		Worktrees: []worktree.Info{{Path: "/repo", Branch: "main"}},
		BranchCommits: map[string][]branch.CommitInfo{
			"main": {
				{Hash: "abc123", Author: "Max Xue", Subject: "Add auth flow"},
				{Hash: "def456", Author: "dependabot", Subject: "Bump deps"},
			},
		},
	}
	s.Sync(state)

	view := stripAnsiSequences(s.View(120, 40, state))

	for _, want := range []string{"Branch: main", "Path:   /repo", "Commits", "abc123 MX Add auth flow", "def456 de Bump deps"} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected details panel to contain %q, got:\n%s", want, view)
		}
	}
}
