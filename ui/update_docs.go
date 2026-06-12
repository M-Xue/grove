package ui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) updateDocs(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q", "?":
		m.mode = ModeChange
		m.docs.scroll = 0
		m.clearError()
		return m, nil
	case "up", "k", "shift+tab":
		m.moveDocsScroll(-1)
		return m, nil
	case "down", "j", "tab":
		m.moveDocsScroll(1)
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) handleWorktreeDocsLoaded(msg worktreeDocsLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.setError(msg.err)
		m.mode = ModeChange
		return m, nil
	}
	m.docs.lines = msg.lines
	m.docs.scroll = 0
	m.mode = ModeDocs
	m.clearError()
	m.clearStatus()
	return m, nil
}
