package textinput

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdateAddsASCIIInput(t *testing.T) {
	m := New("placeholder")
	m.Focus()
	consumed, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a', 'b'}})
	if !consumed || m.Value() != "ab" {
		t.Fatalf("unexpected value: %q", m.Value())
	}
}

func TestUpdateIgnoresAltModifiedRunes(t *testing.T) {
	m := New("placeholder")
	m.Focus()
	consumed, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}, Alt: true})
	if consumed || m.Value() != "" {
		t.Fatalf("expected alt+rune to be ignored, got consumed=%v value=%q", consumed, m.Value())
	}
}

func TestUpdateBackspaceRemovesCharacter(t *testing.T) {
	m := New("placeholder")
	m.Focus()
	m.SetValue("ab")
	m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if m.Value() != "a" {
		t.Fatalf("unexpected value: %q", m.Value())
	}
}

func TestDisabledUpdateIgnoresInput(t *testing.T) {
	m := New("placeholder")
	m.Focus()
	m.SetValue("ab")
	m.SetDisabled(true)
	consumed, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if consumed || cmd != nil {
		t.Fatalf("expected disabled input to ignore keys, got consumed=%v cmd=%v", consumed, cmd)
	}
	if m.Value() != "ab" {
		t.Fatalf("expected value unchanged, got %q", m.Value())
	}
}

func TestDisabledViewIsInert(t *testing.T) {
	m := New("placeholder")
	m.Focus()
	m.SetValue("branch-name")
	m.SetDisabled(true)
	view := m.View()
	if strings.Contains(view, focusColor) || strings.Contains(view, "> ") {
		t.Fatalf("expected no focus prefix in disabled view, got %q", view)
	}
	if strings.Contains(view, cursorColor) {
		t.Fatalf("expected no caret in disabled view, got %q", view)
	}
	if !strings.Contains(view, "branch-name") {
		t.Fatalf("expected value to still render, got %q", view)
	}
}

func TestClearingDisabledRestoresCaret(t *testing.T) {
	m := New("placeholder")
	m.Focus()
	m.SetValue("branch-name")
	m.SetDisabled(true)
	m.SetDisabled(false)
	if !strings.Contains(m.View(), cursorColor) {
		t.Fatalf("expected caret to return once re-enabled, got %q", m.View())
	}
}
