package buttongroup

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func keyMsg(t tea.KeyType) tea.KeyMsg { return tea.KeyMsg{Type: t} }

func TestNewFocusesFirstButton(t *testing.T) {
	m := New(Button{ID: "confirm", Label: "Delete"}, Button{ID: "cancel", Label: "Cancel"})
	if id, ok := m.FocusedID(); !ok || id != "confirm" {
		t.Fatalf("expected first button focused, got %q ok=%v", id, ok)
	}
}

func TestSetFocusedIDFallsBackToFirstOnUnknown(t *testing.T) {
	m := New(Button{ID: "confirm", Label: "Delete"}, Button{ID: "cancel", Label: "Cancel"})
	m.SetFocusedID("cancel")
	if id, _ := m.FocusedID(); id != "cancel" {
		t.Fatalf("expected cancel focused, got %q", id)
	}
	m.SetFocusedID("nope")
	if id, _ := m.FocusedID(); id != "confirm" {
		t.Fatalf("expected fallback to first button, got %q", id)
	}
}

func TestUpdateCyclesFocusAndWraps(t *testing.T) {
	m := New(Button{ID: "a", Label: "A"}, Button{ID: "b", Label: "B"})

	if consumed, _ := m.Update(keyMsg(tea.KeyTab)); !consumed {
		t.Fatal("expected tab to be consumed")
	}
	if id, _ := m.FocusedID(); id != "b" {
		t.Fatalf("expected b after tab, got %q", id)
	}

	m.Update(keyMsg(tea.KeyTab))
	if id, _ := m.FocusedID(); id != "a" {
		t.Fatalf("expected wrap back to a, got %q", id)
	}

	m.Update(keyMsg(tea.KeyShiftTab))
	if id, _ := m.FocusedID(); id != "b" {
		t.Fatalf("expected shift+tab to wrap to b, got %q", id)
	}

	if consumed, _ := m.Update(keyMsg(tea.KeyEnter)); consumed {
		t.Fatal("expected enter to be left for the container")
	}
}

func TestViewMarksOnlyFocusedButton(t *testing.T) {
	m := New(Button{ID: "confirm", Label: "Delete"}, Button{ID: "cancel", Label: "Cancel"})
	view := m.View()
	if !strings.Contains(view, "[x] Delete") || !strings.Contains(view, "[ ] Cancel") {
		t.Fatalf("expected focused Delete and unfocused Cancel, got %q", view)
	}
}

func TestEmptyGroupIsInert(t *testing.T) {
	m := New()
	if _, ok := m.FocusedID(); ok {
		t.Fatal("expected no focus for an empty group")
	}
	if consumed, _ := m.Update(keyMsg(tea.KeyTab)); consumed {
		t.Fatal("expected keys unconsumed by an empty group")
	}
	if m.View() != "" {
		t.Fatalf("expected empty view, got %q", m.View())
	}
}
