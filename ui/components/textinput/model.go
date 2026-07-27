package textinput

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	placeholderColor = "\x1b[38;5;245m"
	focusColor       = "\x1b[38;5;183m"
	cursorColor      = "\x1b[7m"
	resetColor       = "\x1b[0m"
)

type Model struct {
	value       string
	placeholder string
	focused     bool
	cursor      int
}

func New(placeholder string) Model {
	return Model{placeholder: placeholder}
}

func (m *Model) SetPlaceholder(value string) { m.placeholder = value }
func (m *Model) SetValue(value string) {
	m.value = filterASCII(value)
	m.cursor = len(m.value)
}
func (m Model) Value() string { return m.value }
func (m *Model) Clear() {
	m.value = ""
	m.cursor = 0
}
func (m *Model) Focus()       { m.focused = true }
func (m *Model) Blur()        { m.focused = false }
func (m Model) Focused() bool { return m.focused }

func (m *Model) Update(msg tea.KeyMsg) (bool, tea.Cmd) {
	if !m.focused {
		return false, nil
	}
	switch msg.Type {
	case tea.KeyLeft:
		if m.cursor > 0 {
			m.cursor--
		}
		return true, nil
	case tea.KeyRight:
		if m.cursor < len(m.value) {
			m.cursor++
		}
		return true, nil
	case tea.KeyHome, tea.KeyCtrlA:
		m.cursor = 0
		return true, nil
	case tea.KeyEnd, tea.KeyCtrlE:
		m.cursor = len(m.value)
		return true, nil
	case tea.KeyBackspace:
		if m.cursor == 0 {
			return true, nil
		}
		m.value = m.value[:m.cursor-1] + m.value[m.cursor:]
		m.cursor--
		return true, nil
	case tea.KeyDelete:
		if m.cursor >= len(m.value) {
			return true, nil
		}
		m.value = m.value[:m.cursor] + m.value[m.cursor+1:]
		return true, nil
	case tea.KeyRunes, tea.KeySpace:
		if msg.Alt {
			return false, nil
		}
		runes := filterASCII(msg.String())
		m.value = m.value[:m.cursor] + runes + m.value[m.cursor:]
		m.cursor += len(runes)
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
		if m.focused {
			return prefix + cursorColor + " " + resetColor + placeholderColor + m.placeholder + resetColor
		}
		return prefix + placeholderColor + m.placeholder + resetColor
	}
	if !m.focused {
		return prefix + m.value
	}
	if m.cursor >= len(m.value) {
		return prefix + m.value + cursorColor + " " + resetColor
	}
	before := m.value[:m.cursor]
	at := m.value[m.cursor : m.cursor+1]
	after := m.value[m.cursor+1:]
	return prefix + before + cursorColor + at + resetColor + after
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
