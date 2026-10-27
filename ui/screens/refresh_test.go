package screens

import (
	"fmt"
	"testing"

	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/ui/keys"
	"github.com/M-Xue/grove/worktree"
	tea "github.com/charmbracelet/bubbletea"
)

type refreshRecorderApp struct {
	fakeApp
	refreshes int
}

func (a *refreshRecorderApp) RefreshWorktrees() app.Command {
	a.refreshes++
	return func() app.Message { return nil }
}

func TestChangeScreenCtrlRRequestsRefresh(t *testing.T) {
	recorder := &refreshRecorderApp{}
	s := NewChangeScreen(recorder)
	binding, ok := s.registry[ModeDefault].lookup(keys.KeyCtrlR)
	if !ok {
		t.Fatal("expected ctrl+r to be bound in the default mode")
	}
	if cmd := binding.Action(&ActionCtx{Key: tea.KeyMsg{Type: tea.KeyCtrlR}}); cmd == nil {
		t.Fatal("expected ctrl+r to return the refresh command")
	}
	if recorder.refreshes != 1 {
		t.Fatalf("expected one refresh request, got %d", recorder.refreshes)
	}
}

func TestChangeScreenReloadRefetchesCommitsWithoutResettingScroll(t *testing.T) {
	recorder := &commitsRecorderApp{}
	s := NewChangeScreen(recorder)
	state := app.State{Worktrees: []worktree.Info{{Path: "/repo", Branch: "main"}}}
	s.Sync(state)
	if s.fetchHoveredCommits() == nil {
		t.Fatal("expected the initial hover to fetch commits")
	}
	commits := make([]branch.CommitInfo, 12)
	for i := range commits {
		commits[i] = branch.CommitInfo{Hash: fmt.Sprintf("c%02d", i+1)}
	}
	s.commitsPager.SetItems(commits)
	s.commitsPager.Page(5)
	s.commitsPager.Next()

	ctx := &ScreenContext{Run: func(cmd app.Command) tea.Cmd {
		if cmd == nil {
			return nil
		}
		return func() tea.Msg { return cmd() }
	}}
	if s.OnMessage(ctx, app.WorktreesLoadedMessage{}) == nil {
		t.Fatal("expected a worktree reload to refetch the hovered branch's commits")
	}
	if len(recorder.branches) != 2 || recorder.branches[1] != "main" {
		t.Fatalf("expected a second fetch for main, got %v", recorder.branches)
	}
	if page := s.commitsPager.Page(5); len(page) == 0 || page[0].Hash != "c06" {
		t.Fatalf("expected the commits pager to keep its page across a reload, got %v", page)
	}
}

func TestChangeScreenInitFetchesHoveredCommitsFromSeed(t *testing.T) {
	recorder := &commitsRecorderApp{}
	s := NewChangeScreen(recorder)
	if s.Init() != nil {
		t.Fatal("expected no fetch before any worktree is listed")
	}
	s.Sync(app.State{Worktrees: []worktree.Info{{Path: "/repo", Branch: "main"}}})
	if s.Init() == nil {
		t.Fatal("expected the seeded hover to fetch commits")
	}
	if len(recorder.branches) != 1 || recorder.branches[0] != "main" {
		t.Fatalf("expected one fetch for main, got %v", recorder.branches)
	}
}
