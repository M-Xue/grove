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

// refreshTickMsg fires on the periodic worktree-refresh timer so changes made
// outside grove show up without a keypress.
type refreshTickMsg struct{}

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
	// minRefreshInterval is the floor on how often the worktree list is
	// re-read from git in the background. Each refresh costs a few git
	// subprocesses per worktree, so it is deliberately slow; focus changes
	// and ctrl+r cover the moments a user actually expects an instant update.
	// The actual interval stretches with how long the last listing took (see
	// refreshIntervalFor), so a repository whose git status is expensive is
	// not kept permanently busy by grove's polling.
	minRefreshInterval = 3 * time.Second
	// maxRefreshInterval caps that stretch so even a very slow repository is
	// still polled now and then.
	maxRefreshInterval = 30 * time.Second
	// refreshIntervalFactor is how many listing-durations of idle time
	// separate background refreshes once the floor is exceeded: a 1s listing
	// is polled every 10s, a 3s one every 30s.
	refreshIntervalFactor = 10
)

type Model struct {
	app      *app.App
	width    int
	height   int
	screen   app.ScreenID
	loading  loading.Model
	spinning bool
	status   status.Model
	change   *screens.ChangeScreen
}

func New(application *app.App) *Model {
	return &Model{
		app:     application,
		screen:  application.State().Screen,
		loading: loading.New(),
		status:  status.New(),
		change:  screens.NewChangeScreen(application),
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.run(m.app.Init()), m.initialDetails(), m.ensureSpinner(), m.refreshTick())
}

// initialDetails fetches the hovered worktree's details (its commits, and
// from those its PR) when the list was seeded from the disk cache. Without
// it the details panel would spin until the live list lands — on a large
// repository the slowest part of startup — even though the hovered branch is
// already known. When the live list does land the screen refetches anyway,
// since the branch tip may have moved, so this is purely a head start. It is
// nil when nothing was seeded.
func (m *Model) initialDetails() tea.Cmd {
	m.syncScreens()
	return m.run(m.change.Init())
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
	case tea.FocusMsg:
		// The user has come back from another window — the most likely moment
		// for the worktrees to have changed under us.
		return m, m.backgroundRefresh()
	case refreshTickMsg:
		return m, tea.Batch(m.backgroundRefresh(), m.refreshTick())
	case spinnerTickMsg:
		if !m.hasSpinnerWork() {
			m.spinning = false
			return m, nil
		}
		m.loading.Tick()
		m.change.TickSpinner()
		return m, spinnerTick()
	case tea.KeyMsg:
		m.app.DismissCompletedLoading()
		if len(m.app.State().Statuses) > 0 {
			m.app.ClearStatus()
		}
		m.syncScreens()
		if inputFrozen(m.app.State(), msg) {
			// A blocking operation is in flight: ignore all input but quit.
			return m, m.ensureSpinner()
		}
		ctx := m.screenContext()
		var cmd tea.Cmd
		// Route by screen ID; ScreenChange is the only screen today, but the
		// dispatch point stays so a future screen slots in as a new case.
		switch m.app.State().Screen {
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

// inputFrozen reports whether a keystroke should be ignored because a blocking
// operation is in flight. Quit (ctrl+c) always passes through so the user is
// never trapped.
func inputFrozen(state app.State, msg tea.KeyMsg) bool {
	return state.IsBusy() && msg.Type != tea.KeyCtrlC
}

// refreshTick schedules the next background worktree refresh. Init starts the
// loop and every tick re-arms it, so exactly one loop runs for the session.
// The delay is recomputed on every re-arm from the latest listing's duration.
func (m *Model) refreshTick() tea.Cmd {
	return tea.Tick(refreshIntervalFor(m.app.State().WorktreeLoadDuration), func(time.Time) tea.Msg {
		return refreshTickMsg{}
	})
}

// refreshIntervalFor is the delay before the next background refresh given
// how long the last worktree listing took: refreshIntervalFactor times that
// duration, never below minRefreshInterval nor above maxRefreshInterval. A
// listing of 300ms keeps the 3s floor; one of 9s is polled every 30s.
func refreshIntervalFor(lastLoad time.Duration) time.Duration {
	return max(minRefreshInterval, min(lastLoad*refreshIntervalFactor, maxRefreshInterval))
}

// backgroundRefresh issues a silent worktree reload; app declines it while a
// load or a blocking operation is already in flight, in which case this is nil.
func (m *Model) backgroundRefresh() tea.Cmd {
	return m.run(m.app.RefreshWorktreesInBackground())
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

// hasSpinnerWork reports whether anything on screen still animates a spinner:
// an active loading entry, or a details-panel fetch (commits/PR) in flight.
func (m *Model) hasSpinnerWork() bool {
	return m.hasActiveLoading() || m.change.DetailsPending(m.app.State())
}

// ensureSpinner starts the spinner tick loop when there is spinner work and
// the loop is not already running, guaranteeing a single loop at a time.
func (m *Model) ensureSpinner() tea.Cmd {
	if m.spinning || !m.hasSpinnerWork() {
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
	// Reserve three rows beneath the body: a notice row (status/loading lines
	// bottom-align there instead of over the panels' bottom border), a blank
	// spacer, and the footer.
	bodyHeight := max(0, contentHeight-3)
	var body string
	switch state.Screen {
	default:
		body = m.change.View(contentWidth, bodyHeight, state)
	}
	bodyLines := fitBody(body, contentWidth, bodyHeight)
	lines := make([]string, 0, contentHeight)
	lines = append(lines, bodyLines...)
	lines = append(lines, fitBottomLine("", contentWidth)) // notice row
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
	m.screen = state.Screen
}

func (m *Model) syncScreenTransitions(previous, current app.ScreenID) {
	if previous == current {
		return
	}
	if previous == app.ScreenChange {
		m.change.Reset()
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
	if state.IsBusy() {
		return "working… ctrl+c quit"
	}
	switch state.Screen {
	default:
		return m.change.Footer(contentWidth)
	}
}
