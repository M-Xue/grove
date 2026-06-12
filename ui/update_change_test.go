package ui

import (
	"testing"

	"github.com/M-Xue/grove/worktree"
	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdateChangeEnterSubmitsSelectedPath(t *testing.T) {
	m := New(nil)
	m.change.filtered = []string{"/repo", "/repo-feature"}
	m.change.selected = 1
	m.change.selectedItem = "/repo-feature"

	updated, cmd := m.updateChange(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(Model)

	if got.ChangeSubmittedPath() != "/repo-feature" {
		t.Fatalf("unexpected submitted path: got %q", got.ChangeSubmittedPath())
	}
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}

func TestUpdateChangeEnterWithoutSelectionSetsStatus(t *testing.T) {
	m := New(nil)
	m.change.filtered = nil

	updated, cmd := m.updateChange(tea.KeyMsg{Type: tea.KeyEnter})
	got := updated.(Model)

	if got.status.message != "no worktree selected" {
		t.Fatalf("unexpected status message: got %q", got.status.message)
	}
	if got.ChangeSubmittedPath() != "" {
		t.Fatalf("expected no submitted path, got %q", got.ChangeSubmittedPath())
	}
	if cmd != nil {
		t.Fatal("expected no command")
	}
}

func TestHandleWorktreesLoadedClearsSubmittedPath(t *testing.T) {
	m := New(nil)
	m.change.submittedPath = "/repo-feature"

	updated, cmd := m.handleWorktreesLoaded(worktreesLoadedMsg{
		worktrees: []worktree.WorktreeInfo{{Path: "/repo", Branch: "main"}},
	})
	got := updated.(Model)

	if got.ChangeSubmittedPath() != "" {
		t.Fatalf("expected submitted path to be cleared, got %q", got.ChangeSubmittedPath())
	}
	if cmd != nil {
		t.Fatal("expected no command")
	}
}
