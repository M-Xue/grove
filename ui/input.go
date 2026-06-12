package ui

import "strings"

type addField int

const (
	pathField addField = iota
	branchField
)

func placeholder(value, text string) string {
	if value != "" {
		return value
	}
	return placeholderColor + text + resetColor
}

func (m *Model) activeAddValue() *string {
	if m.addField == branchField {
		return &m.addBranch
	}
	return &m.addPath
}

func (m *Model) appendToActiveInput(value string) {
	field := m.activeAddValue()
	*field += value
}

func (m *Model) backspaceActiveInput() {
	field := m.activeAddValue()
	if len(*field) > 0 {
		*field = (*field)[:len(*field)-1]
	}
}

func (m *Model) resetAddState() {
	m.addMode = false
	m.confirmCreateBranch = false
	m.pendingAdd = false
	m.addField = pathField
	m.addPath = ""
	m.addBranch = ""
	m.confirmPath = ""
	m.confirmBranch = ""
}

func (m *Model) startAddMode() {
	m.addMode = true
	m.confirmCreateBranch = false
	m.pendingAdd = false
	m.addField = pathField
	m.errorMessage = ""
	m.statusMessage = ""
	if m.addPath == "" {
		m.addPath = ""
	}
	if m.addBranch == "" {
		m.addBranch = ""
	}
}

func (m *Model) cancelAddMode() {
	m.resetAddState()
	m.errorMessage = ""
	m.statusMessage = ""
}

func (m *Model) submitAdd() (string, string, bool) {
	path := strings.TrimSpace(m.addPath)
	branch := strings.TrimSpace(m.addBranch)
	if path == "" {
		m.errorMessage = "worktree path is required"
		return "", "", false
	}
	if branch == "" {
		m.errorMessage = "branch name is required"
		return "", "", false
	}
	m.errorMessage = ""
	return path, branch, true
}

func (m *Model) syncItemsFromWorktrees() {
	m.items = make([]string, 0, len(m.worktrees))
	for _, item := range m.worktrees {
		m.items = append(m.items, item.Path)
	}
	m.refreshFiltered()
}
