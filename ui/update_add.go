package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateAdd(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.add.confirmCreateBranch {
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+a", "esc":
			m.cancelAddMode()
			return m, nil
		case "y", "Y":
			m.add.confirmCreateBranch = false
			m.add.pending = true
			m.setStatus("creating branch and worktree")
			return m, addWorktreeCmd(m.manager, m.add.confirmPath, m.add.confirmBranch, true)
		case "n", "N":
			m.add.confirmCreateBranch = false
			m.add.confirmPath = ""
			m.add.confirmBranch = ""
			m.clearStatus()
			return m, nil
		default:
			return m, nil
		}
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "ctrl+a", "esc":
		m.cancelAddMode()
		return m, nil
	case "tab", "down":
		m.add.field = addField((int(m.add.field) + 1) % 2)
		return m, nil
	case "shift+tab", "up":
		if m.add.field == pathField {
			m.add.field = branchField
		} else {
			m.add.field = pathField
		}
		return m, nil
	case "backspace":
		m.backspaceActiveInput()
		return m, nil
	case "enter":
		path, branch, ok := m.submitAdd()
		if !ok {
			return m, nil
		}
		m.add.pending = true
		m.setStatus("checking branch")
		return m, checkBranchExistsCmd(m.manager, path, branch)
	default:
		if msg.Type == tea.KeyRunes {
			m.appendToActiveInput(msg.String())
		}
		return m, nil
	}
}

func (m Model) handleBranchChecked(msg branchCheckedMsg) (tea.Model, tea.Cmd) {
	m.add.pending = false
	if msg.err != nil {
		m.setError(msg.err)
		return m, nil
	}
	if msg.path != strings.TrimSpace(m.add.path) || msg.branch != strings.TrimSpace(m.add.branch) {
		return m, nil
	}
	if msg.exists {
		m.setStatus("adding worktree")
		m.add.pending = true
		return m, addWorktreeCmd(m.manager, msg.path, msg.branch, false)
	}
	m.add.confirmCreateBranch = true
	m.add.confirmPath = msg.path
	m.add.confirmBranch = msg.branch
	m.clearStatus()
	return m, nil
}

func (m Model) handleWorktreeAdded(msg worktreeAddedMsg) (tea.Model, tea.Cmd) {
	m.add.pending = false
	if msg.err != nil {
		m.setError(msg.err)
		return m, nil
	}
	m.resetAddState()
	m.clearError()
	m.setStatus("worktree added")
	return m, loadWorktreesCmd(m.manager)
}
