package ui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) updateChange(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		return m, tea.Quit
	case "ctrl+a":
		m.startAddMode()
		return m, nil
	case "up", "shift+tab":
		m.moveChangeSelection(-1)
		return m, nil
	case "down", "tab":
		m.moveChangeSelection(1)
		return m, nil
	case "backspace":
		if len(m.change.query) > 0 {
			m.change.query = m.change.query[:len(m.change.query)-1]
			m.refreshChangeFiltered()
		}
		return m, nil
	case "enter", "left", "right":
		return m, nil
	default:
		if msg.Type == tea.KeyRunes {
			m.change.query += msg.String()
			m.refreshChangeFiltered()
			m.clearStatus()
			m.clearError()
		}
		return m, nil
	}
}

func (m Model) handleWorktreesLoaded(msg worktreesLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.setError(msg.err)
		return m, nil
	}
	m.change.worktrees = msg.worktrees
	m.syncChangeItemsFromWorktrees()
	m.clearError()
	return m, nil
}
