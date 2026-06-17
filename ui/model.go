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

// appMsg is the private envelope that carries an app.Message produced by a
// Command across the Bubble Tea bus.
type appMsg struct{ msg app.Message }

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
		branch:  screens.NewBranchScreen(),
	}
}

func (m *Model) Init() tea.Cmd {
	return m.run(m.app.Init())
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case appMsg:
		if _, ok := msg.msg.(app.QuitRequested); ok {
			return m, tea.Quit
		}
		previousScreen := m.screen
		next := m.app.HandleMessage(msg.msg)
		m.syncScreenTransitions(previousScreen, m.app.State().Screen)
		m.syncScreens()
		return m, m.run(next)
	case tea.KeyMsg:
		m.app.DismissCompletedLoading()
		if len(m.app.State().Statuses) > 0 {
			m.app.DismissStatuses()
		}
		m.syncScreens()
		ctx := &screens.ScreenContext{App: m.app, Run: m.run, Quit: func() tea.Cmd { return tea.Quit }}
		var cmd tea.Cmd
		switch m.app.State().Screen {
		case app.ScreenAdd:
			cmd = m.add.Update(ctx, msg, m.app.State())
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

// run is the single bridge between app and Bubble Tea: it executes a Command's
// thunk on a goroutine and wraps the returned Message in an appMsg envelope.
func (m *Model) run(cmd app.Command) tea.Cmd {
	if cmd == nil {
		return nil
	}
	return func() tea.Msg {
		return appMsg{msg: cmd()}
	}
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
	case app.ScreenBranch:
		return m.branch.Footer(contentWidth, state.BranchScope)
	default:
		if state.Dialog.Active {
			return m.change.DialogFooter(contentWidth)
		}
		return m.change.Footer(contentWidth)
	}
}
