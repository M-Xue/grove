package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func repeatLine(fill string, width int) string {
	if width <= 0 {
		return ""
	}
	return strings.Repeat(fill, width)
}

func fitLine(content string, width int) string {
	if width <= 0 {
		return ""
	}
	contentWidth := lipgloss.Width(content)
	if contentWidth > width {
		return lipgloss.NewStyle().MaxWidth(width).Render(content)
	}
	return content + strings.Repeat(" ", width-contentWidth)
}

func (m Model) displayItem(path string) string {
	for _, worktree := range m.change.worktrees {
		if worktree.Path == path {
			return path + " [" + worktree.Branch + "]"
		}
	}
	return path
}

func (m Model) selectedChangePath() (string, bool) {
	if len(m.change.filtered) == 0 {
		return "", false
	}
	if m.change.selected < 0 || m.change.selected >= len(m.change.filtered) {
		return "", false
	}
	return m.change.filtered[m.change.selected], true
}

func (m *Model) setError(err error) {
	m.status.err = err
	m.status.message = ""
}

func (m *Model) clearError() {
	m.status.err = nil
}

func (m *Model) setStatus(message string) {
	m.status.message = message
	m.status.err = nil
}

func (m *Model) clearStatus() {
	m.status.message = ""
}
