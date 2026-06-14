package ui

import (
	"strings"

	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/ui/components/loading"
	"github.com/M-Xue/grove/ui/components/status"
	"github.com/M-Xue/grove/ui/screens"
	tea "github.com/charmbracelet/bubbletea"
)

type effectMsg struct{ result app.Result }

type Model struct {
	app     *app.App
	width   int
	height  int
	loading loading.Model
	status  status.Model
	change  *screens.ChangeScreen
	add     *screens.AddScreen
	docs    *screens.DocsScreen
}

func New(application *app.App) *Model {
	return &Model{
		app:     application,
		loading: loading.New(),
		status:  status.New(),
		change:  screens.NewChangeScreen(),
		add:     screens.NewAddScreen(),
		docs:    screens.NewDocsScreen(),
	}
}

func (m *Model) Init() tea.Cmd {
	return m.runEffect(m.app.Init())
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case effectMsg:
		m.syncScreens()
		return m, m.runEffect(m.app.HandleResult(msg.result))
	case tea.KeyMsg:
		if len(m.app.State().Statuses) > 0 {
			m.app.DismissStatuses()
		}
		m.syncScreens()
		ctx := &screens.ScreenContext{App: m.app, RunEffect: m.runEffect, Quit: func() tea.Cmd { return tea.Quit }}
		var cmd tea.Cmd
		switch m.app.State().Screen {
		case app.ScreenAdd:
			cmd = m.add.Update(ctx, msg, m.app.State())
		case app.ScreenDocs:
			cmd = m.docs.Update(ctx, msg, m.app.State())
		default:
			cmd = m.change.Update(ctx, msg, m.app.State())
		}
		m.syncScreens()
		return m, cmd
	default:
		return m, nil
	}
}

func (m *Model) View() string {
	state := m.app.State()
	m.syncScreens()
	contentWidth := max(0, m.width-8)
	contentHeight := max(0, m.height-4)
	var body string
	var footer string
	switch state.Screen {
	case app.ScreenAdd:
		footer = m.add.Footer(contentWidth, state.Dialog.Active)
		body = m.add.View(contentWidth, contentHeight, footer, state)
	case app.ScreenDocs:
		footer = m.docs.Footer(contentWidth)
		body = m.docs.View(contentWidth, contentHeight, footer, state)
	default:
		if state.Dialog.Active {
			footer = m.change.DialogFooter(contentWidth)
		} else {
			footer = m.change.Footer(contentWidth)
		}
		body = m.change.View(contentWidth, contentHeight, contentHeight, footer, state)
	}
	lines := []string{body}
	if state.Loading.Active {
		lines = append(lines, "", m.loading.View(state.Loading.Message, contentWidth))
	}
	if len(state.Statuses) > 0 {
		lines = append(lines, "")
		lines = append(lines, m.status.View(state.Statuses, contentWidth)...)
	}
	return strings.Join(lines, "\n")
}

func (m *Model) SubmittedPath() string {
	return m.app.SubmittedPath()
}

func (m *Model) syncScreens() {
	state := m.app.State()
	m.change.Sync(state)
	m.add.Sync(state)
	m.docs.Sync(state)
}

func (m *Model) runEffect(effect app.Effect) tea.Cmd {
	services := m.app.Services()
	switch effect := effect.(type) {
	case nil:
		return nil
	case app.LoadWorktreesEffect:
		return func() tea.Msg {
			worktrees, err := services.Worktree.List()
			return effectMsg{result: app.WorktreesLoadedResult{Worktrees: worktrees, Err: err}}
		}
	case app.LoadDocsEffect:
		return func() tea.Msg {
			lines, err := services.Docs.WorktreeHelp()
			return effectMsg{result: app.DocsLoadedResult{Lines: lines, Err: err}}
		}
	case app.CheckBranchExistsEffect:
		return func() tea.Msg {
			exists, err := services.Worktree.BranchExists(effect.Branch)
			return effectMsg{result: app.BranchCheckedResult{Path: effect.Path, Branch: effect.Branch, Exists: exists, Err: err}}
		}
	case app.AddWorktreeEffect:
		return func() tea.Msg {
			var err error
			if effect.CreateBranch {
				err = services.Worktree.AddNewBranch(effect.Path, effect.Branch)
			} else {
				err = services.Worktree.Add(effect.Path, effect.Branch)
			}
			return effectMsg{result: app.WorktreeAddedResult{Err: err}}
		}
	case app.RemoveWorktreeEffect:
		return func() tea.Msg {
			err := services.Worktree.Remove(effect.Path)
			return effectMsg{result: app.WorktreeRemovedResult{Path: effect.Path, Err: err}}
		}
	case app.QuitEffect:
		return tea.Quit
	default:
		return nil
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
