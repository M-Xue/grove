package textinput

import (
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
