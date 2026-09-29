// Package buttongroup renders a horizontal row of focusable buttons and owns
// the focus-cycling between them. It is a plain widget like selectlist and
// textinput: it knows nothing about dialogs, screens, or what a button does —
// the container reads FocusedID and decides what confirming means.
package buttongroup

import (
	"strings"

	"github.com/M-Xue/grove/ui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	selectedColor = lipgloss.Color("183")
	// Unfocused buttons dim to the shared inactive tone so they read as idle
	// alongside the rest of the muted chrome.
	mutedColor = theme.TextInactive
)

type Button struct {
	ID    string
	Label string
}

type Model struct {
	buttons   []Button
	focusedID string
}

// New returns a button group focused on the first button.
func New(buttons ...Button) Model {
	m := Model{buttons: append([]Button(nil), buttons...)}
	m.syncFocus()
	return m
}

// SetFocusedID focuses the button with the given ID, falling back to the first
// button when the ID is unknown.
func (m *Model) SetFocusedID(id string) {
	m.focusedID = id
	m.syncFocus()
}

func (m Model) FocusedID() (string, bool) {
	for _, button := range m.buttons {
		if button.ID == m.focusedID {
			return button.ID, true
		}
	}
	return "", false
}

// Update moves focus on tab/shift+tab and the left/right arrows (wrapping at
// either end), reporting whether the key was consumed.
func (m *Model) Update(msg tea.KeyMsg) (bool, tea.Cmd) {
	if len(m.buttons) == 0 {
		return false, nil
	}
	switch msg.String() {
	case "shift+tab", "left":
		m.move(-1)
		return true, nil
	case "tab", "right":
		m.move(1)
		return true, nil
	default:
		return false, nil
	}
}

func (m Model) View() string {
	selectedStyle := lipgloss.NewStyle().Foreground(selectedColor)
	mutedStyle := lipgloss.NewStyle().Foreground(mutedColor)

	labels := make([]string, 0, len(m.buttons))
	for _, button := range m.buttons {
		label := mutedStyle.Render("[ ] " + button.Label)
		if button.ID == m.focusedID {
			label = selectedStyle.Render("[x] " + button.Label)
		}
		labels = append(labels, label)
	}
	return strings.Join(labels, "  ")
}

func (m *Model) syncFocus() {
	if len(m.buttons) == 0 {
		m.focusedID = ""
		return
	}
	for _, button := range m.buttons {
		if button.ID == m.focusedID {
			return
		}
	}
	m.focusedID = m.buttons[0].ID
}

func (m *Model) move(delta int) {
	if len(m.buttons) == 0 {
		return
	}
	index := 0
	for i, button := range m.buttons {
		if button.ID == m.focusedID {
			index = i
			break
		}
	}
	index += delta
	if index < 0 {
		index = len(m.buttons) - 1
	}
	if index >= len(m.buttons) {
		index = 0
	}
	m.focusedID = m.buttons[index].ID
}
