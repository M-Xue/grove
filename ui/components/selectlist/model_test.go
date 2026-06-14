package selectlist

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSelectionPreservedByID(t *testing.T) {
	m := New("No matches")
	m.SetItems([]Item{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}})
	m.SetSelectedID("b")
	m.SetItems([]Item{{ID: "c", Label: "C"}, {ID: "b", Label: "B"}})
	if id, ok := m.SelectedID(); !ok || id != "b" {
		t.Fatalf("unexpected selection: %q", id)
	}
}

func TestMoveDownChangesSelection(t *testing.T) {
	m := New("No matches")
	m.SetItems([]Item{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}})
	m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if id, _ := m.SelectedID(); id != "b" {
		t.Fatalf("unexpected selection: %q", id)
	}
}
