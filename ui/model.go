package ui

import (
	"strings"

	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/ui/components/loading"
	"github.com/M-Xue/grove/ui/components/status"
	"github.com/M-Xue/grove/ui/screens"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type effectMsg struct{ result app.Result }

const (
	horizontalPadding = 4
	verticalPadding   = 2
)

type Model struct {
	app     *app.App
	width   int
	height  int
	screen  app.ScreenID
	loading loading.Model
	status  status.Model
	change  *screens.ChangeScreen
	add     *screens.AddScreen
	docs    *screens.DocsScreen
	branch  *screens.BranchScreen
}

func New(application *app.App) *Model {
	return &Model{
		app:     application,
		screen:  application.State().Screen,
		loading: loading.New(),
		status:  status.New(),
		change:  screens.NewChangeScreen(),
		add:     screens.NewAddScreen(),
		docs:    screens.NewDocsScreen(),
		branch:  screens.NewBranchScreen(),
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
		previousScreen := m.screen
		effect := m.app.HandleResult(msg.result)
		m.syncScreenTransitions(previousScreen, m.app.State().Screen)
		m.syncScreens()
		return m, m.runEffect(effect)
	case tea.KeyMsg:
		m.app.DismissCompletedLoading()
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
		case app.ScreenBranch:
			cmd = m.branch.Update(ctx, msg, m.app.State())
		default:
			cmd = m.change.Update(ctx, msg, m.app.State())
		}
		m.syncScreenTransitions(m.screen, m.app.State().Screen)
		m.syncScreens()
		return m, cmd
	default:
		return m, nil
	}
}

func (m *Model) View() string {
	state := m.app.State()
	m.syncScreens()
	contentWidth := max(0, m.width-horizontalPadding*2)
	footer := m.footer(contentWidth, state)
	loadingLines := m.loading.View(state.Loading, contentWidth)
	statusLines := m.status.View(state.Statuses, contentWidth)
	noticeLines := expandRenderedLines(append(append([]string(nil), loadingLines...), statusLines...))
	contentHeight := max(0, m.height-verticalPadding*2)
	bodyHeight := max(0, contentHeight-2)
	var body string
	switch state.Screen {
	case app.ScreenAdd:
		body = m.add.View(contentWidth, bodyHeight, state)
	case app.ScreenDocs:
		body = m.docs.View(contentWidth, bodyHeight, state)
	case app.ScreenBranch:
		body = m.branch.View(contentWidth, bodyHeight, state)
	default:
		body = m.change.View(contentWidth, bodyHeight, state)
	}
	bodyLines := fitBody(body, contentWidth, bodyHeight)
	lines := make([]string, 0, contentHeight)
	lines = append(lines, bodyLines...)
	lines = append(lines, fitBottomLine("", contentWidth))
	lines = append(lines, fitBottomLine(footer, contentWidth))
	for len(lines) < contentHeight {
		lines = append(lines, fitBottomLine("", contentWidth))
	}
	if len(lines) > contentHeight {
		lines = lines[:contentHeight]
	}
	overlayNoticeLines(lines, noticeLines, contentWidth)

	padded := make([]string, 0, m.height)
	blankLine := strings.Repeat(" ", max(m.width, 0))
	for i := 0; i < verticalPadding; i++ {
		padded = append(padded, blankLine)
	}
	leftPad := strings.Repeat(" ", horizontalPadding)
	for _, line := range lines {
		padded = append(padded, leftPad+fitBottomLine(line, contentWidth))
	}
	for len(padded) < m.height {
		padded = append(padded, blankLine)
	}
	if len(padded) > m.height {
		padded = padded[:m.height]
	}

	return strings.Join(padded, "\n")
}

func (m *Model) SubmittedPath() string {
	return m.app.SubmittedPath()
}

func (m *Model) syncScreens() {
	state := m.app.State()
	m.change.Sync(state)
	m.add.Sync(state)
	m.docs.Sync(state)
	m.branch.Sync(state)
	m.screen = state.Screen
}

func (m *Model) syncScreenTransitions(previous, current app.ScreenID) {
	if previous == current {
		return
	}
	if previous == app.ScreenAdd {
		m.add.Reset()
	}
	if previous == app.ScreenChange {
		m.change.Reset()
	}
	if previous == app.ScreenBranch {
		m.branch.Reset()
	}
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
	case app.LoadBranchesEffect:
		return func() tea.Msg {
			branches, scope, err := services.Branch.List()
			return effectMsg{result: app.BranchesLoadedResult{Branches: branches, Scope: scope, Err: err}}
		}
	case app.LoadBranchCommitsEffect:
		return func() tea.Msg {
			commits, err := services.Branch.RecentCommits(effect.Name, effect.Limit)
			return effectMsg{result: app.BranchCommitsLoadedResult{Name: effect.Name, Commits: commits, Err: err}}
		}
	case app.ToggleBranchScopeEffect:
		return func() tea.Msg {
			services.Branch.ToggleScope()
			branches, scope, err := services.Branch.List()
			return effectMsg{result: app.BranchesLoadedResult{Branches: branches, Scope: scope, Err: err}}
		}
	case app.CheckoutBranchEffect:
		return func() tea.Msg {
			err := services.Branch.Checkout(effect.Name)
			return effectMsg{result: app.BranchCheckedOutResult{Err: err}}
		}
	case app.DeleteBranchEffect:
		return func() tea.Msg {
			err := services.Branch.Delete(effect.Name)
			return effectMsg{result: app.BranchDeletedResult{Err: err}}
		}
	case app.DeleteAllBranchesEffect:
		return func() tea.Msg {
			summary, err := services.Branch.DeleteAllLocal()
			return effectMsg{result: app.AllBranchesDeletedResult{Deleted: summary.Deleted, Skipped: summary.Skipped, Err: err}}
		}
	case app.FetchBranchesEffect:
		return func() tea.Msg {
			err := services.Branch.Fetch()
			return effectMsg{result: app.BranchesFetchedResult{Err: err}}
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

func fitBody(content string, width, height int) []string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, height)
	for i := 0; i < height; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		result = append(result, fitBottomLine(line, width))
	}
	return result
}

func fitBottomLine(content string, width int) string {
	if width <= 0 {
		return ""
	}
	contentWidth := lipgloss.Width(content)
	if contentWidth > width {
		return lipgloss.NewStyle().MaxWidth(width).Render(content)
	}
	return content + strings.Repeat(" ", width-contentWidth)
}

func overlayNoticeLines(canvas []string, noticeLines []string, width int) {
	if len(canvas) < 2 || len(noticeLines) == 0 {
		return
	}
	maxNoticeHeight := max(0, len(canvas)-2)
	if maxNoticeHeight == 0 {
		return
	}
	if len(noticeLines) > maxNoticeHeight {
		noticeLines = append([]string(nil), noticeLines[len(noticeLines)-maxNoticeHeight:]...)
	}
	start := len(canvas) - 2 - len(noticeLines)
	for i, line := range noticeLines {
		row := start + i
		if row < 0 || row >= len(canvas)-2 {
			continue
		}
		canvas[row] = fitBottomLine(line, width)
	}
}

func expandRenderedLines(lines []string) []string {
	expanded := make([]string, 0, len(lines))
	for _, line := range lines {
		parts := strings.Split(line, "\n")
		expanded = append(expanded, parts...)
	}
	return expanded
}

func (m *Model) footer(contentWidth int, state app.State) string {
	switch state.Screen {
	case app.ScreenAdd:
		return m.add.Footer(contentWidth, state.Dialog.Active)
	case app.ScreenDocs:
		return m.docs.Footer(contentWidth)
	case app.ScreenBranch:
		return m.branch.Footer(contentWidth, state.BranchScope)
	default:
		if state.Dialog.Active {
			return m.change.DialogFooter(contentWidth)
		}
		return m.change.Footer(contentWidth)
	}
}
