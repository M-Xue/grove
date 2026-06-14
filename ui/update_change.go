package ui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) updateChange(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.change.confirmRemove {
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.cancelRemoveWorktree()
			return m, nil
		case "y", "Y":
			path := m.change.confirmPath
			m.change.confirmRemove = false
			m.change.confirmPath = ""
			m.setStatus("removing worktree")
			m.clearError()
			return m, removeWorktreeCmd(m.manager, path)
		case "n", "N":
			m.cancelRemoveWorktree()
			return m, nil
		default:
			return m, nil
		}
	}

	switch msg.String() {
	case "ctrl+c", "esc":
		return m, tea.Quit
	case "ctrl+a":
		m.startAddMode()
		return m, nil
	case "?":
		return m, openWorktreeDocsCmd()
	case "ctrl+d":
		path, ok := m.selectedChangePath()
		if !ok {
			m.setStatus("no worktree selected")
			return m, nil
		}
		m.change.confirmRemove = true
		m.change.confirmPath = path
		m.clearStatus()
		m.clearError()
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
	case "enter":
		path, ok := m.selectedChangePath()
		if !ok {
			m.setStatus("no worktree selected")
			return m, nil
		}
		m.change.submittedPath = path
		return m, tea.Quit
	case "left", "right":
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
	m.change.submittedPath = ""
	m.clearError()
	return m, nil
}

func (m Model) handleWorktreeRemoved(msg worktreeRemovedMsg) (tea.Model, tea.Cmd) {
	m.change.confirmRemove = false
	m.change.confirmPath = ""
	if msg.err != nil {
		m.setError(msg.err)
		return m, nil
	}
	if msg.path != "" {
		m.change.selectedItem = msg.path
	}
	m.setStatus("worktree removed")
	m.clearError()
	return m, loadWorktreesCmd(m.manager)
}
