package ui

import (
	"github.com/M-Xue/grove/worktree"
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
)

type Mode string

const (
	ModeChange Mode = "change"
	ModeAdd    Mode = "add"
)

type StatusState struct {
	message string
	err     error
}

type ChangeState struct {
	query         string
	selectedItem  string
	selected      int
	scroll        int
	items         []string
	filtered      []string
	worktrees     []worktree.WorktreeInfo
	submittedPath string
}

type AddState struct {
	field               addField
	path                string
	branch              string
	confirmCreateBranch bool
	confirmPath         string
	confirmBranch       string
	pending             bool
}

type Model struct {
	manager worktree.Manager
	help    help.Model
	keys    keyMap
	width   int
	height  int
	mode    Mode
	status  StatusState
	change  ChangeState
	add     AddState
}

func New(manager worktree.Manager) Model {
	return Model{
		manager: manager,
		help:    help.New(),
		keys:    newKeyMap(),
		mode:    ModeChange,
		add: AddState{
			field: pathField,
		},
	}
}

func (m Model) Init() tea.Cmd {
	return loadWorktreesCmd(m.manager)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case worktreesLoadedMsg:
		return m.handleWorktreesLoaded(msg)
	case branchCheckedMsg:
		return m.handleBranchChecked(msg)
	case worktreeAddedMsg:
		return m.handleWorktreeAdded(msg)
	case worktreeRemovedMsg:
		return m.handleWorktreeRemoved(msg)
	case tea.KeyMsg:
		switch m.mode {
		case ModeAdd:
			return m.updateAdd(msg)
		default:
			return m.updateChange(msg)
		}
	}

	return m, nil
}

func (m Model) View() string {
	return renderView(m)
}
