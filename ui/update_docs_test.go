package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdateDocsQuestionMarkReturnsToChangeMode(t *testing.T) {
	m := New(nil)
	m.mode = ModeDocs
	m.docs.lines = []string{"line 1"}
	m.docs.scroll = 3

	updated, cmd := m.updateDocs(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	got := updated.(Model)

	if got.mode != ModeChange {
		t.Fatalf("expected change mode, got %q", got.mode)
	}
	if got.docs.scroll != 0 {
		t.Fatalf("expected docs scroll reset, got %d", got.docs.scroll)
	}
	if cmd != nil {
		t.Fatal("expected no command")
	}
}

func TestUpdateDocsMovesScrollDown(t *testing.T) {
	m := New(nil)
	m.mode = ModeDocs
	m.docs.lines = []string{"1", "2", "3"}

	updated, cmd := m.updateDocs(tea.KeyMsg{Type: tea.KeyDown})
	got := updated.(Model)

	if got.docs.scroll != 1 {
		t.Fatalf("expected scroll 1, got %d", got.docs.scroll)
	}
	if cmd != nil {
		t.Fatal("expected no command")
	}
}
