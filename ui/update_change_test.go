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

func TestUpdateChangeCtrlDStartsRemoveConfirmation(t *testing.T) {
	m := New(nil)
	m.change.worktrees = []worktree.WorktreeInfo{{Path: "/repo-feature", Branch: "feature/auth"}}
	m.change.filtered = []string{"/repo-feature"}
	m.change.selectedItem = "/repo-feature"

	updated, cmd := m.updateChange(tea.KeyMsg{Type: tea.KeyCtrlD})
	got := updated.(Model)

	if !got.change.confirmRemove {
		t.Fatal("expected remove confirmation to start")
	}
	if got.change.confirmPath != "/repo-feature" {
		t.Fatalf("unexpected confirm path: got %q", got.change.confirmPath)
	}
	if cmd != nil {
		t.Fatal("expected no command")
	}
}

func TestUpdateChangeConfirmRemoveNoCancelsConfirmation(t *testing.T) {
	m := New(nil)
	m.change.confirmRemove = true
	m.change.confirmPath = "/repo-feature"
	m.setStatus("pending")

	updated, cmd := m.updateChange(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	got := updated.(Model)

	if got.change.confirmRemove {
		t.Fatal("expected remove confirmation to be cleared")
	}
	if got.change.confirmPath != "" {
		t.Fatalf("expected confirm path to be cleared, got %q", got.change.confirmPath)
	}
	if got.status.message != "" {
		t.Fatalf("expected status to be cleared, got %q", got.status.message)
	}
	if cmd != nil {
		t.Fatal("expected no command")
	}
}

func TestUpdateChangeConfirmRemoveYesReturnsCommand(t *testing.T) {
	m := New(nil)
	m.change.confirmRemove = true
	m.change.confirmPath = "/repo-feature"

	updated, cmd := m.updateChange(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	got := updated.(Model)

	if got.change.confirmRemove {
		t.Fatal("expected remove confirmation to be cleared")
	}
	if got.status.message != "removing worktree" {
		t.Fatalf("unexpected status message: got %q", got.status.message)
	}
	if cmd == nil {
		t.Fatal("expected remove command")
	}
	if got.change.confirmPath != "" {
		t.Fatalf("expected confirm path to be cleared, got %q", got.change.confirmPath)
	}
}

func TestUpdateChangeQuestionMarkOpensDocs(t *testing.T) {
	m := New(nil)

	updated, cmd := m.updateChange(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	got := updated.(Model)

	if cmd == nil {
		t.Fatal("expected docs command")
	}
	if got.mode != ModeChange {
		t.Fatalf("expected mode to remain change, got %q", got.mode)
	}
}

func TestHandleWorktreeDocsLoadedEntersDocsMode(t *testing.T) {
	m := New(nil)
	updated, cmd := m.handleWorktreeDocsLoaded(worktreeDocsLoadedMsg{lines: []string{"line 1", "line 2"}})
	got := updated.(Model)

	if got.mode != ModeDocs {
		t.Fatalf("expected docs mode, got %q", got.mode)
	}
	if len(got.docs.lines) != 2 {
		t.Fatalf("expected docs lines to load, got %d", len(got.docs.lines))
	}
	if cmd != nil {
		t.Fatal("expected no command")
	}
}

func TestHandleWorktreeDocsLoadedSetsError(t *testing.T) {
	m := New(nil)
	wantErr := worktree.ErrNotGitRepo
	updated, cmd := m.handleWorktreeDocsLoaded(worktreeDocsLoadedMsg{err: wantErr})
	got := updated.(Model)

	if got.status.err != wantErr {
		t.Fatalf("unexpected error: got %v want %v", got.status.err, wantErr)
	}
	if got.mode != ModeChange {
		t.Fatalf("expected change mode, got %q", got.mode)
	}
	if cmd != nil {
		t.Fatal("expected no command")
	}
}
