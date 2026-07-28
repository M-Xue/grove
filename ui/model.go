package ui

import (
	"strings"
	"time"

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

// spinnerTickMsg advances the loading spinner one frame.
type spinnerTickMsg struct{}

// streamNextMsg carries one app.Message drained from a streaming operation's
// channel, along with the channel so the reader can re-subscribe for the next
// message until the channel is closed.
type streamNextMsg struct {
	msg app.Message
	ch  <-chan app.Message
}

// readStream returns a command that reads the next message from ch. It returns
// nil once the channel is closed, which ends the drain loop.
func readStream(ch <-chan app.Message) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return streamNextMsg{msg: msg, ch: ch}
	}
}

const (
	horizontalPadding = 4
	verticalPadding   = 2
	// spinnerInterval is how often the Braille spinner advances a frame.
	spinnerInterval = 80 * time.Millisecond
)

type Model struct {
	app     *app.App
	width   int
	height  int
	screen  app.ScreenID
	loading  loading.Model
	spinning bool
	status   status.Model
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
		change:  screens.NewChangeScreen(application),
		add:     screens.NewAddScreen(application),
		branch:  screens.NewBranchScreen(application),
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.run(m.app.Init()), m.ensureSpinner())
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
		// A streaming operation returns a channel of messages rather than a
		// single result; begin draining it instead of handling it as state.
		if start, ok := msg.msg.(app.WorktreeAddStartedMessage); ok {
			return m, readStream(start.Updates)
		}
		previousScreen := m.screen
		next := m.app.HandleMessage(msg.msg)
		m.syncScreenTransitions(previousScreen, m.app.State().Screen)
		m.syncScreens()
		// Deliver the message to the active screen for UI reactions (e.g.
		// opening a dialog) in addition to app's state update above.
		reaction := m.deliverMessage(msg.msg)
		return m, tea.Batch(m.run(next), reaction, m.ensureSpinner())
	case streamNextMsg:
		previousScreen := m.screen
		next := m.app.HandleMessage(msg.msg)
		m.syncScreenTransitions(previousScreen, m.app.State().Screen)
		m.syncScreens()
		reaction := m.deliverMessage(msg.msg)
		// Re-subscribe for the next message; readStream stops when the channel
		// closes after the terminal message.
		return m, tea.Batch(m.run(next), reaction, readStream(msg.ch), m.ensureSpinner())
	case spinnerTickMsg:
		if !m.hasActiveLoading() {
			m.spinning = false
			return m, nil
		}
		m.loading.Tick()
		return m, spinnerTick()
	case tea.KeyMsg:
		m.app.DismissCompletedLoading()
		if len(m.app.State().Statuses) > 0 {
			m.app.ClearStatus()
		}
		m.syncScreens()
		ctx := m.screenContext()
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
		return m, tea.Batch(cmd, m.ensureSpinner())
	default:
		return m, nil
	}
}

// spinnerTick schedules the next spinner frame.
func spinnerTick() tea.Cmd {
	return tea.Tick(spinnerInterval, func(time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

// hasActiveLoading reports whether any loading entry is still in progress.
func (m *Model) hasActiveLoading() bool {
	for _, entry := range m.app.State().Loading {
		if !entry.Completed {
			return true
		}
	}
	return false
}

// ensureSpinner starts the spinner tick loop when there is active loading work
// and the loop is not already running, guaranteeing a single loop at a time.
func (m *Model) ensureSpinner() tea.Cmd {
	if m.spinning || !m.hasActiveLoading() {
		return nil
	}
	m.spinning = true
	return spinnerTick()
}

func (m *Model) screenContext() *screens.ScreenContext {
	return &screens.ScreenContext{Run: m.run}
}

// deliverMessage hands a completed message to the active screen so it can react
// in the UI (open a dialog, clear search). The screen owns this presentation
// state; app owns the domain state updated by HandleMessage.
func (m *Model) deliverMessage(msg app.Message) tea.Cmd {
	ctx := m.screenContext()
	switch m.app.State().Screen {
	case app.ScreenAdd:
		return m.add.OnMessage(ctx, msg)
	case app.ScreenBranch:
		return m.branch.OnMessage(ctx, msg)
	default:
		return m.change.OnMessage(ctx, msg)
	}
}

func (m *Model) View() string {
	state := m.app.State()
	m.syncScreens()
	contentWidth := max(0, m.width-horizontalPadding*2)
	footer := m.footer(contentWidth, state)
	statusLines := expandRenderedLines(m.status.View(state.Statuses, contentWidth))
	loadingLines := expandRenderedLines(m.loading.View(state.Loading, contentWidth))
	noticeLines := composeNotices(statusLines, loadingLines, contentWidth)
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

// composeNotices lays status messages along the left and loading messages along
// the right of the same bottom band, each column bottom-aligned so both grow
// upward from the footer.
func composeNotices(statusLines, loadingLines []string, width int) []string {
	height := max(len(statusLines), len(loadingLines))
	if height == 0 {
		return nil
	}
	statusOffset := height - len(statusLines)
	loadingOffset := height - len(loadingLines)
	rows := make([]string, 0, height)
	for i := 0; i < height; i++ {
		left := ""
		if i-statusOffset >= 0 {
			left = statusLines[i-statusOffset]
		}
		right := ""
		if i-loadingOffset >= 0 {
			right = loadingLines[i-loadingOffset]
		}
		rows = append(rows, placeLeftRight(left, right, width))
	}
	return rows
}

// placeLeftRight renders left flush-left and right flush against the right edge
// of width, padding the gap between them. If they would collide, the left text
// is truncated so the right text stays fully visible.
func placeLeftRight(left, right string, width int) string {
	if width <= 0 {
		return ""
	}
	rightWidth := lipgloss.Width(right)
	if rightWidth >= width {
		return lipgloss.NewStyle().MaxWidth(width).Render(right)
	}
	leftMax := width - rightWidth
	if lipgloss.Width(left) > leftMax {
		left = lipgloss.NewStyle().MaxWidth(leftMax).Render(left)
	}
	gap := width - lipgloss.Width(left) - rightWidth
	if gap < 0 {
		gap = 0
	}
	return left + strings.Repeat(" ", gap) + right
}

func (m *Model) footer(contentWidth int, state app.State) string {
	switch state.Screen {
	case app.ScreenAdd:
		return m.add.Footer(contentWidth)
	case app.ScreenBranch:
		return m.branch.Footer(contentWidth)
	default:
		return m.change.Footer(contentWidth)
	}
}
