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
	if m.add.field == branchField {
		return &m.add.branch
	}
	return &m.add.path
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
	m.mode = ModeChange
	m.add = AddState{field: pathField}
}

func (m *Model) startAddMode() {
	m.mode = ModeAdd
	m.add.confirmCreateBranch = false
	m.add.pending = false
	m.add.field = pathField
	m.status = StatusState{}
}

func (m *Model) cancelAddMode() {
	m.resetAddState()
	m.status = StatusState{}
}

func (m *Model) submitAdd() (string, string, bool) {
	path := strings.TrimSpace(m.add.path)
	branch := strings.TrimSpace(m.add.branch)
	if path == "" {
		m.clearError()
		m.setStatus("worktree path is required")
		return "", "", false
	}
	if branch == "" {
		m.clearError()
		m.setStatus("branch name is required")
		return "", "", false
	}
	m.clearError()
	m.clearStatus()
	return path, branch, true
}
