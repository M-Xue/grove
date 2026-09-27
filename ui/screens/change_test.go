package screens

import (
	"regexp"
	"strings"
	"testing"

	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/pr"
	"github.com/M-Xue/grove/ui/components/loading"
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

func TestChangeScreenDetailsPanelShowsPRSection(t *testing.T) {
	s := NewChangeScreen(fakeApp{})
	state := app.State{
		Worktrees: []worktree.Info{{Path: "/repo", Branch: "feature"}},
		BranchPRs: map[string]app.BranchPR{
			"feature": {Found: true, Info: pr.Info{
				Number: 2745,
				Title:  "gate remote annotation URLs",
				Status: pr.StatusOpen,
				Author: "chmouel",
				URL:    "https://github.com/org/repo/pull/2745",
				Checks: []pr.Check{
					{Name: "linters", State: pr.CheckPassed},
					{Name: "e2e tests", State: pr.CheckPending},
				},
			}},
		},
	}
	s.Sync(state)

	view := stripAnsiSequences(s.View(140, 40, state))

	// The status pill pads its label with its own spaces, hence " Open ".
	for _, want := range []string{
		"PR", "Status:     Open ", "Title:     gate remote annotation URLs",
		"Author:    chmouel", "URL:       https://github.com/org/repo/pul",
		"CI Checks: Pending", "linters", "e2e tests",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected details panel to contain %q, got:\n%s", want, view)
		}
	}
}

func TestChangeScreenDetailsPanelNotesWhenBranchHasNoPR(t *testing.T) {
	s := NewChangeScreen(fakeApp{})
	state := app.State{
		Worktrees: []worktree.Info{{Path: "/repo", Branch: "feature"}},
		// A definitive "no PR" lookup result: header plus a plain note, no
		// PR fields.
		BranchPRs: map[string]app.BranchPR{"feature": {Found: false}},
	}
	s.Sync(state)

	view := stripAnsiSequences(s.View(140, 40, state))
	if !strings.Contains(view, "No PR available") {
		t.Fatalf("expected the no-PR note, got:\n%s", view)
	}
	if strings.Contains(view, "Status:") || strings.Contains(view, "CI Checks:") {
		t.Fatalf("expected no PR fields without a PR, got:\n%s", view)
	}
}

func TestChangeScreenDetailsPanelOmitsPRSectionBeforeLookup(t *testing.T) {
	s := NewChangeScreen(fakeApp{})
	state := app.State{Worktrees: []worktree.Info{{Path: "/repo", Branch: "feature"}}}
	s.Sync(state)

	view := stripAnsiSequences(s.View(140, 40, state))
	if strings.Contains(view, "No PR available") || strings.Contains(view, "Status:") {
		t.Fatalf("expected no PR section before the lookup resolves, got:\n%s", view)
	}
}

func TestChangeScreenDetailsPanelShowsNoneForEmptyChecks(t *testing.T) {
	s := NewChangeScreen(fakeApp{})
	state := app.State{
		Worktrees: []worktree.Info{{Path: "/repo", Branch: "feature"}},
		BranchPRs: map[string]app.BranchPR{
			"feature": {Found: true, Info: pr.Info{Number: 3, Title: "t", Status: pr.StatusOpen, Author: "max", URL: "u"}},
		},
	}
	s.Sync(state)

	view := stripAnsiSequences(s.View(140, 40, state))
	if !strings.Contains(view, "CI Checks: None") {
		t.Fatalf("expected the empty rollup to render as None, got:\n%s", view)
	}
}

func TestChangeScreenDetailsHeadingsSpinWhileFetching(t *testing.T) {
	s := NewChangeScreen(fakeApp{})
	// Worktrees are known but neither the commits nor the PR lookup has
	// resolved: both headings are present, each with the spinner frame.
	state := app.State{Worktrees: []worktree.Info{{Path: "/repo", Branch: "feature"}}}
	s.Sync(state)

	if !s.DetailsPending(state) {
		t.Fatal("expected details to be pending before any fetch resolves")
	}
	view := stripAnsiSequences(s.View(140, 40, state))
	frame := loading.Frame(0)
	for _, want := range []string{"Commits " + frame, "PR " + frame} {
		if !strings.Contains(view, want) {
			t.Fatalf("expected heading with spinner %q, got:\n%s", want, view)
		}
	}

	// Once both lookups resolve, the spinners disappear and pending clears.
	state.BranchCommits = map[string][]branch.CommitInfo{"feature": nil}
	state.BranchPRs = map[string]app.BranchPR{"feature": {Found: false}}
	s.Sync(state)
	if s.DetailsPending(state) {
		t.Fatal("expected no pending details once both lookups resolved")
	}
	view = stripAnsiSequences(s.View(140, 40, state))
	if strings.Contains(view, frame) {
		t.Fatalf("expected no spinner after resolution, got:\n%s", view)
	}
}

func TestChangeScreenPRSectionNotesFailureAndUnavailability(t *testing.T) {
	s := NewChangeScreen(fakeApp{})

	failed := app.State{
		Worktrees: []worktree.Info{{Path: "/repo", Branch: "feature"}},
		BranchPRs: map[string]app.BranchPR{"feature": {Failed: true}},
	}
	s.Sync(failed)
	if view := stripAnsiSequences(s.View(140, 40, failed)); !strings.Contains(view, "PR lookup failed") {
		t.Fatalf("expected the failed note, got:\n%s", view)
	}

	unavailable := app.State{
		Worktrees:           []worktree.Info{{Path: "/repo", Branch: "feature"}},
		PRLookupUnavailable: true,
	}
	s.Sync(unavailable)
	if s.DetailsPending(unavailable) && len(unavailable.BranchCommits) != 0 {
		t.Fatal("unexpected pending state")
	}
	if view := stripAnsiSequences(s.View(140, 40, unavailable)); !strings.Contains(view, "PR lookup unavailable") {
		t.Fatalf("expected the unavailable note, got:\n%s", view)
	}
}

func TestChangeScreenPRSectionPositionIsStableWhileCommitsLoad(t *testing.T) {
	s := NewChangeScreen(fakeApp{})
	pending := app.State{Worktrees: []worktree.Info{{Path: "/repo", Branch: "feature"}}}
	loaded := app.State{
		Worktrees: []worktree.Info{{Path: "/repo", Branch: "feature"}},
		BranchCommits: map[string][]branch.CommitInfo{
			"feature": {{Hash: "abc123", Author: "Max", Subject: "one"}, {Hash: "def456", Author: "Max", Subject: "two"}},
		},
	}

	prLine := func(state app.State) int {
		s.Sync(state)
		for i, line := range strings.Split(stripAnsiSequences(s.View(140, 40, state)), "\n") {
			if strings.Contains(line, " PR") {
				return i
			}
		}
		return -1
	}

	before, after := prLine(pending), prLine(loaded)
	if before == -1 || before != after {
		t.Fatalf("expected the PR heading to stay on the same row, got %d then %d", before, after)
	}
}
