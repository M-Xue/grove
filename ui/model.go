package ui

import (
	"strings"

	"github.com/M-Xue/grove/worktree"
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	manager             worktree.Manager
	help                help.Model
	keys                keyMap
	width               int
	height              int
	query               string
	selectedItem        string
	selected            int
	scroll              int
	items               []string
	filtered            []string
	worktrees           []worktree.WorktreeInfo
	addMode             bool
	addField            addField
	addPath             string
	addBranch           string
	confirmCreateBranch bool
	confirmPath         string
	confirmBranch       string
	pendingAdd          bool
	errorMessage        string
	statusMessage       string
}

func New(manager worktree.Manager) Model {
	m := Model{
		manager: manager,
		help:    help.New(),
		keys:    newKeyMap(),
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return loadWorktreesCmd(m.manager)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case worktreesLoadedMsg:
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			return m, nil
		}
		m.worktrees = msg.worktrees
		m.syncItemsFromWorktrees()
		m.errorMessage = ""
		return m, nil
	case branchCheckedMsg:
		m.pendingAdd = false
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			return m, nil
		}
		if msg.path != strings.TrimSpace(m.addPath) || msg.branch != strings.TrimSpace(m.addBranch) {
			return m, nil
		}
		if msg.exists {
			m.statusMessage = "adding worktree"
			m.errorMessage = ""
			m.pendingAdd = true
			return m, addWorktreeCmd(m.manager, msg.path, msg.branch, false)
		}
		m.confirmCreateBranch = true
		m.confirmPath = msg.path
		m.confirmBranch = msg.branch
		m.statusMessage = ""
		return m, nil
	case worktreeAddedMsg:
		m.pendingAdd = false
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			m.statusMessage = ""
			return m, nil
		}
		m.resetAddState()
		m.errorMessage = ""
		m.statusMessage = "worktree added"
		return m, loadWorktreesCmd(m.manager)
	case tea.KeyMsg:
		if m.confirmCreateBranch {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "ctrl+a", "esc":
				m.cancelAddMode()
				return m, nil
			case "y", "Y":
				m.confirmCreateBranch = false
				m.pendingAdd = true
				m.statusMessage = "creating branch and worktree"
				m.errorMessage = ""
				return m, addWorktreeCmd(m.manager, m.confirmPath, m.confirmBranch, true)
			case "n", "N":
				m.confirmCreateBranch = false
				m.confirmPath = ""
				m.confirmBranch = ""
				m.statusMessage = ""
				return m, nil
			}
		}

		if m.addMode {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "ctrl+a", "esc":
				m.cancelAddMode()
			case "tab", "down":
				m.addField = addField((int(m.addField) + 1) % 2)
			case "shift+tab", "up":
				if m.addField == pathField {
					m.addField = branchField
				} else {
					m.addField = pathField
				}
			case "backspace":
				m.backspaceActiveInput()
			case "enter":
				path, branch, ok := m.submitAdd()
				if !ok {
					return m, nil
				}
				m.pendingAdd = true
				m.statusMessage = "checking branch"
				return m, checkBranchExistsCmd(m.manager, path, branch)
			default:
				if msg.Type == tea.KeyRunes {
					m.appendToActiveInput(msg.String())
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m, tea.Quit
		case "ctrl+a":
			m.startAddMode()
		case "up", "shift+tab":
			m.moveSelection(-1)
		case "down", "tab":
			m.moveSelection(1)
		case "backspace":
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
				m.refreshFiltered()
			}
		case "enter", "left", "right":
			// Reserved for future navigation.
		default:
			if msg.Type == tea.KeyRunes {
				m.query += msg.String()
				m.refreshFiltered()
				m.statusMessage = ""
				m.errorMessage = ""
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m Model) View() string {
	return renderView(m)
}
