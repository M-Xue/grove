package textinput

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	placeholderColor = "\x1b[38;5;245m"
	focusColor       = "\x1b[38;5;183m"
	resetColor       = "\x1b[0m"
)

type Model struct {
	value       string
	placeholder string
	focused     bool
}

func New(placeholder string) Model {
	return Model{placeholder: placeholder}
}

func (m *Model) SetPlaceholder(value string) { m.placeholder = value }
func (m *Model) SetValue(value string)       { m.value = filterASCII(value) }
func (m Model) Value() string                { return m.value }
func (m *Model) Clear()                      { m.value = "" }
func (m *Model) Focus()                      { m.focused = true }
func (m *Model) Blur()                       { m.focused = false }
func (m Model) Focused() bool                { return m.focused }

func (m *Model) Update(msg tea.KeyMsg) (bool, tea.Cmd) {
	if !m.focused {
		return false, nil
	}
	switch msg.Type {
	case tea.KeyBackspace:
		if len(m.value) == 0 {
			return true, nil
		}
		m.value = m.value[:len(m.value)-1]
		return true, nil
	case tea.KeyRunes:
		m.value += filterASCII(msg.String())
		return true, nil
	default:
		return false, nil
	}
}

func (m Model) View() string {
	prefix := "  "
	if m.focused {
		prefix = focusColor + "> " + resetColor
	}
	if m.value == "" {
		return prefix + placeholderColor + m.placeholder + resetColor
	}
	return prefix + m.value
}

func filterASCII(value string) string {
	var b strings.Builder
	for _, r := range value {
		if r >= 32 && r <= 126 {
			b.WriteRune(r)
		}
	}
	return b.String()
}
