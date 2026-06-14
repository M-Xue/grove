package dialog

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	dialogTitleColor    = "\x1b[38;5;230m"
	dialogTextColor     = "\x1b[38;5;252m"
	dialogBorderColor   = "245"
	buttonSelectedColor = "\x1b[38;5;183m"
	buttonMutedColor    = "\x1b[38;5;245m"
	buttonResetColor    = "\x1b[0m"
)

type Button struct {
	ID    string
	Label string
}

type Model struct {
	title       string
	description string
	buttons     []Button
	focusedID   string
}

func New() Model                             { return Model{} }
func (m *Model) SetTitle(value string)       { m.title = value }
func (m *Model) SetDescription(value string) { m.description = value }
func (m *Model) SetButtons(buttons []Button) {
	m.buttons = append([]Button(nil), buttons...)
	m.syncFocus()
}
func (m *Model) SetFocusedID(id string) { m.focusedID = id; m.syncFocus() }
func (m Model) FocusedID() (string, bool) {
	for _, button := range m.buttons {
		if button.ID == m.focusedID {
			return button.ID, true
		}
	}
	return "", false
}

func (m *Model) Update(msg tea.KeyMsg) (bool, tea.Cmd) {
	if len(m.buttons) == 0 {
		return false, nil
	}
	switch msg.String() {
	case "left", "shift+tab":
		m.move(-1)
		return true, nil
	case "right", "tab":
		m.move(1)
		return true, nil
	default:
		return false, nil
	}
}

func (m Model) View(width, height int) string {
	buttonLabels := make([]string, 0, len(m.buttons))
	for _, button := range m.buttons {
		label := buttonMutedColor + "[ ] " + button.Label + buttonResetColor
		if button.ID == m.focusedID {
			label = buttonSelectedColor + "[x] " + button.Label + buttonResetColor
		}
		buttonLabels = append(buttonLabels, label)
	}
	body := []string{dialogTitleColor + m.title + buttonResetColor, "", dialogTextColor + m.description + buttonResetColor, "", strings.Join(buttonLabels, "  ")}
	content := strings.Join(body, "\n")
	style := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color(dialogBorderColor)).Padding(1, 2).MaxWidth(max(28, width-10))
	return style.Render(content)
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
