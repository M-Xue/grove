package textinput

import (
	"strings"

	"github.com/M-Xue/grove/ui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// The input renders on a darker field background so it reads as a textbox.
// When a width is set the background fills the whole field, not just the text.
var (
	fieldBg    = lipgloss.Color("#1e2030")
	valueColor = lipgloss.Color("252")
	// The placeholder shares the inactive-border color so hint text and idle
	// chrome read as one muted layer.
	placeholderColor = theme.BorderInactive
)

type Model struct {
	value       string
	placeholder string
	focused     bool
	disabled    bool
	cursor      int
	width       int
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

func (m *Model) Focus() { m.focused = true }

func (m *Model) Blur() { m.focused = false }

func (m Model) Focused() bool { return m.focused }

func (m *Model) SetDisabled(v bool) { m.disabled = v }

// SetWidth fixes the rendered field width in display cells; the background is
// padded out to it. Zero (the default) renders at the content's natural width.
func (m *Model) SetWidth(width int) { m.width = width }

func (m *Model) Update(msg tea.KeyMsg) (bool, tea.Cmd) {
	if !m.focused || m.disabled {
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
	base := lipgloss.NewStyle().Background(fieldBg)
	text := base.Foreground(valueColor)
	placeholder := base.Foreground(placeholderColor)
	// The caret is the value color and field background reversed, so it reads
	// as a light block regardless of the terminal's default colors. It is the
	// field's only focus indicator.
	cursor := base.Foreground(valueColor).Reverse(true)

	var parts []string
	if m.disabled {
		// Inert: no caret, value dimmed to the placeholder color so the field
		// visibly reads as read-only while an operation is in flight.
		content := m.value
		if content == "" {
			content = m.placeholder
		}
		parts = append(parts, placeholder.Render(content))
		return m.fill(strings.Join(parts, ""), base)
	}

	switch {
	case m.value == "" && m.focused:
		parts = append(parts, cursor.Render(" "), placeholder.Render(m.placeholder))
	case m.value == "":
		parts = append(parts, placeholder.Render(m.placeholder))
	case !m.focused:
		parts = append(parts, text.Render(m.value))
	case m.cursor >= len(m.value):
		parts = append(parts, text.Render(m.value), cursor.Render(" "))
	default:
		parts = append(parts,
			text.Render(m.value[:m.cursor]),
			cursor.Render(m.value[m.cursor:m.cursor+1]),
			text.Render(m.value[m.cursor+1:]),
		)
	}
	return m.fill(strings.Join(parts, ""), base)
}

// fill pads the rendered content out to the configured width with
// field-background spaces, so the textbox reads as one solid field.
func (m Model) fill(content string, base lipgloss.Style) string {
	if m.width <= 0 {
		return content
	}
	gap := m.width - lipgloss.Width(content)
	if gap <= 0 {
		return content
	}
	return content + base.Render(strings.Repeat(" ", gap))
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
