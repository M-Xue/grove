package screens

import (
	"fmt"
	"strings"

	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/ui/components/loading"
	"github.com/M-Xue/grove/ui/components/panel"
	"github.com/M-Xue/grove/ui/components/selectlist"
	"github.com/M-Xue/grove/ui/components/textinput"
	"github.com/M-Xue/grove/ui/keys"
	"github.com/M-Xue/grove/ui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// changeApp is the narrow view of app the change screen depends on.
type changeApp interface {
	RequestSubmitSelectedPath(path string) app.Command
	RequestAddWorktree(path, branch string) app.Command
	CreateBranchWorktree(path, branch string) app.Command
	RemoveWorktree(path string) app.Command
	ForceRemoveWorktree(path string) app.Command
	PruneWorktrees() app.Command
	LoadBranchCommits(branch string) app.Command
	Quit() app.Command
}

type ChangeScreen struct {
	app       changeApp
	confirm   confirmDialog
	addDlg    addDialog
	search    textinput.Model
	list      selectlist.Model
	registry  Registry
	worktrees []appWorktree
	// commitsBranch is the branch whose recent commits were last requested, so
	// hover moves that land on the same branch do not refetch.
	commitsBranch string
	// spinnerFrame indexes the loading spinner shown next to a details-panel
	// heading while its fetch is in flight; the model ticks it on the shared
	// spinner timer.
	spinnerFrame int
}

// staleColor is the ANSI escape used to dim stale worktrees in the list.
const staleColor = "\x1b[38;5;244m"

type appWorktree struct {
	id     string
	label  string
	color  string
	branch string
	stale  bool
	// dirty reports that the worktree has uncommitted or untracked changes, so
	// removing it requires git's --force.
	dirty bool
}

func NewChangeScreen(application changeApp) *ChangeScreen {
	s := &ChangeScreen{
		app:    application,
		addDlg: newAddDialog(),
		search: textinput.New(""),
		list:   selectlist.New("No matches"),
	}
	s.search.Focus()
	s.registry = s.buildRegistry()
	return s
}

func (s *ChangeScreen) Sync(state app.State) {
	s.worktrees = make([]appWorktree, 0, len(state.Worktrees))
	items := make([]selectlist.Item, 0, len(state.Worktrees))
	for _, worktree := range state.Worktrees {
		label := worktree.Path + " [" + worktree.Branch + "]"
		color := ""
		if worktree.Stale {
			label += " [stale]"
			color = staleColor
		} else if worktree.Locked {
			// A locked worktree is a deliberate state and often still usable, so
			// it is labelled but kept at its normal colour rather than dimmed.
			label += " [locked]"
		}
		items = append(items, selectlist.Item{ID: worktree.Path, Label: label, Color: color})
		dirty := worktree.HasUncommittedChanges || worktree.HasUntrackedFiles
		s.worktrees = append(s.worktrees, appWorktree{id: worktree.Path, label: label, color: color, branch: worktree.Branch, stale: worktree.Stale, dirty: dirty})
	}
	items = filterItems(items, s.search.Value())
	s.list.SetItems(items)
}

// OnMessage reacts to app messages the screen presents on: a fresh worktree
// list triggers a (re)fetch of the hovered branch's commits, and an absent
// branch from the add dialog's check opens the confirm dialog offering to
// create it.
func (s *ChangeScreen) OnMessage(ctx *ScreenContext, msg app.Message) tea.Cmd {
	if _, ok := msg.(app.WorktreesLoadedMessage); ok {
		// Force a refetch even when the hover is unchanged: the reload may
		// follow an add/remove that moved the branch's tip.
		s.commitsBranch = ""
		return ctx.Run(s.fetchHoveredCommits())
	}
	absent, ok := msg.(app.BranchAbsentMessage)
	if !ok {
		return nil
	}
	path, branchName := absent.Path, absent.Branch
	s.confirm.open(
		"Branch does not exist",
		fmt.Sprintf("Create a new branch named %q?", branchName),
		"Create",
		true,
		func(actx *ActionCtx) app.Command {
			return s.app.CreateBranchWorktree(path, branchName)
		},
	)
	return nil
}

func (s *ChangeScreen) activeMode() Mode {
	if s.confirm.active {
		return ModeDialog
	}
	if s.addDlg.active {
		return ModeAdd
	}
	return ModeDefault
}

func (s *ChangeScreen) Update(ctx *ScreenContext, msg tea.KeyMsg, state app.State) tea.Cmd {
	mode := s.activeMode()
	if binding, ok := s.registry[mode].lookup(keys.Normalize(msg)); ok {
		return ctx.Run(binding.Action(&ActionCtx{Key: msg}))
	}
	if mode == ModeAdd {
		_, cmd := s.addDlg.handleKey(msg)
		return cmd
	}
	if mode == ModeDefault {
		if consumed, cmd := s.search.Update(msg); consumed {
			// Filtering can move the hover to a different worktree, so refresh
			// the details panel's commits alongside the list.
			s.list.SetItems(filterItems(toItems(s.worktrees), s.search.Value()))
			return tea.Batch(cmd, ctx.Run(s.fetchHoveredCommits()))
		}
	}
	return nil
}

// fetchHoveredCommits returns the command that loads recent commits for the
// hovered worktree's branch, or nil when the hover is unchanged (or empty) so
// selection moves within the same branch do not refetch.
func (s *ChangeScreen) fetchHoveredCommits() app.Command {
	worktree, ok := s.hoveredWorktree()
	if !ok || worktree.branch == "" || worktree.branch == s.commitsBranch {
		return nil
	}
	s.commitsBranch = worktree.branch
	return s.app.LoadBranchCommits(worktree.branch)
}

// TickSpinner advances the details-panel spinner one frame. The model drives
// it on the same timer as the loading area, so all spinners animate in step.
func (s *ChangeScreen) TickSpinner() {
	s.spinnerFrame++
}

// DetailsPending reports whether the hovered worktree still has a details
// fetch in flight (commits or PR), so the model knows to keep the spinner
// timer running. Every fetch outcome — success, no result, failure,
// unavailability — resolves its map entry, so pending states cannot linger.
func (s *ChangeScreen) DetailsPending(state app.State) bool {
	worktree, ok := s.hoveredWorktree()
	if !ok || worktree.branch == "" {
		return false
	}
	return commitsPending(state, worktree.branch) || prPending(state, worktree.branch)
}

func commitsPending(state app.State, branchName string) bool {
	_, loaded := state.BranchCommits[branchName]
	return !loaded
}

func prPending(state app.State, branchName string) bool {
	if state.PRLookupUnavailable {
		return false
	}
	_, resolved := state.BranchPRs[branchName]
	return !resolved
}

// spinner renders the current spinner frame in the loading area's accent, for
// details-panel headings whose fetch is still in flight.
func (s *ChangeScreen) spinner() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true).Render(loading.Frame(s.spinnerFrame))
}

// hoveredWorktree resolves the list's current selection to its worktree.
func (s *ChangeScreen) hoveredWorktree() (appWorktree, bool) {
	item, ok := s.list.SelectedItem()
	if !ok {
		return appWorktree{}, false
	}
	for _, worktree := range s.worktrees {
		if worktree.id == item.ID {
			return worktree, true
		}
	}
	return appWorktree{}, false
}

// panelGap is the number of blank columns between side-by-side panels.
const panelGap = 1

// searchFieldWidth caps the search field so it reads as a compact input rather
// than filling the whole panel row.
const searchFieldWidth = 30

func (s *ChangeScreen) View(width, height int, state app.State) string {
	// App tab, blank line, then the sections boxed lazygit-style: worktrees on
	// the left two thirds, branch details on the right third.
	panelHeight := max(2, height-2)
	leftWidth := max(4, width*2/3)
	rightWidth := max(0, width-leftWidth-panelGap)
	// Interior rows: a blank line under the title, the labelled search field,
	// and a blank line above the list, all inside the borders.
	listHeight := max(1, panelHeight-5)
	searchLabel := detailLabel("Search") + " "
	s.search.SetWidth(max(0, min(searchFieldWidth, leftWidth-4-lipgloss.Width(searchLabel))))
	searchRow := searchLabel + s.search.View()
	interior := strings.Join(append([]string{"", searchRow, ""}, strings.Split(s.list.View(listHeight), "\n")...), "\n")
	// The worktrees panel is the active section whenever no dialog owns the
	// keyboard; dialogs render their own active border.
	worktreesActive := !s.confirm.active && !s.addDlg.active
	left := strings.Split(panel.Render("Worktrees", interior, leftWidth, panelHeight, worktreesActive), "\n")
	right := strings.Split(panel.Render("Branch Details", s.detailsView(rightWidth-4, state), rightWidth, panelHeight, false), "\n")
	rows := make([]string, 0, panelHeight)
	for i := 0; i < panelHeight; i++ {
		leftRow, rightRow := "", ""
		if i < len(left) {
			leftRow = left[i]
		}
		if i < len(right) {
			rightRow = right[i]
		}
		rows = append(rows, fitLine(leftRow, leftWidth)+strings.Repeat(" ", panelGap)+rightRow)
	}
	content := panel.Tab("grove") + "\n\n" + strings.Join(rows, "\n")
	if s.confirm.active {
		return overlayDialog(content, s.confirm.view(width, height), width, height)
	}
	if s.addDlg.active {
		return overlayDialog(content, s.addDlg.view(width, height), width, height)
	}
	return content
}

// Detail styling: labels and section titles render bold in the panel accent
// color; values keep the default foreground. The author initials take the same
// tint the list uses for its selection highlight. The styles are built per call
// (not package vars) so they bind to the default renderer main installs at
// startup rather than the one active at package init.
func detailLabel(text string) string {
	return lipgloss.NewStyle().Foreground(theme.BorderActive).Bold(true).Render(text)
}

func commitAuthor(text string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("183")).Render(text)
}

// detailsView renders the Branch Details panel content for the worktree
// currently hovered in the list, mirroring the left panel's blank line under
// the title: the branch and path, the branch's pull request (when one is
// known), then its recent commits. Empty when the list has no selection (e.g.
// no search matches). width is the panel interior available to content; values
// wrap within it.
func (s *ChangeScreen) detailsView(width int, state app.State) string {
	worktree, ok := s.hoveredWorktree()
	if !ok {
		return ""
	}
	lines := []string{""}
	lines = append(lines, detailRow("Branch:", worktree.branch, width)...)
	lines = append(lines, detailRow("Path:", worktree.id, width)...)
	// Both section headings are always present so the panel's shape is stable;
	// a heading whose fetch is still in flight carries a spinner instead of
	// appearing only once its data lands.
	commitsHeading := detailLabel("Commits")
	if worktree.branch != "" && commitsPending(state, worktree.branch) {
		commitsHeading += " " + s.spinner()
	}
	lines = append(lines, "", commitsHeading)
	// The commits section always occupies its full row budget — blank rows
	// stand in while loading (or when the branch has fewer commits) — so the
	// PR section beneath it never shifts as data lands.
	rows := commitRows(state.BranchCommits[worktree.branch])
	for len(rows) < app.RecentCommitLimit {
		rows = append(rows, "")
	}
	lines = append(lines, rows...)
	lines = append(lines, "")
	lines = append(lines, s.prSection(worktree.branch, state, width)...)
	return strings.Join(lines, "\n")
}

// prSection renders the PR part of the details panel. The heading is always
// shown; beneath it comes whichever the lookup has produced — a spinner while
// in flight, the PR's details, or a plain note for the no-PR, failed, and
// unavailable outcomes.
func (s *ChangeScreen) prSection(branchName string, state app.State, width int) []string {
	heading := detailLabel(prHeaderIcon + " PR")
	if branchName == "" {
		return []string{heading, "No PR available"}
	}
	if state.PRLookupUnavailable {
		return []string{heading, "PR lookup unavailable"}
	}
	entry, resolved := state.BranchPRs[branchName]
	switch {
	case !resolved:
		return []string{heading + " " + s.spinner()}
	case entry.Failed:
		return []string{heading, "PR lookup failed"}
	case !entry.Found:
		return []string{heading, "No PR available"}
	default:
		return append([]string{heading}, prRows(entry.Info, width)...)
	}
}

// detailRowLabelWidth is the column detail labels are padded to, so the values
// start aligned.
const detailRowLabelWidth = 7

// detailRow renders a labelled value in the details panel's label column.
func detailRow(label, value string, width int) []string {
	return labeledRows(label, value, detailRowLabelWidth, width)
}

// labeledRows renders a styled label padded to labelWidth followed by a value,
// wrapping the value across as many lines as needed within width; continuation
// lines are indented to the value column so the wrapped text stays aligned with
// the first line.
func labeledRows(label, value string, labelWidth, width int) []string {
	styled := detailLabel(label)
	valueWidth := width - labelWidth - 1
	if valueWidth < 1 {
		// Too narrow to wrap sensibly; emit one line and let it clip.
		return []string{fitLine(styled, labelWidth) + " " + value}
	}
	indent := strings.Repeat(" ", labelWidth)
	runes := []rune(value)
	lines := make([]string, 0, 1)
	prefix := fitLine(styled, labelWidth)
	for {
		chunk := runes
		if len(chunk) > valueWidth {
			chunk = chunk[:valueWidth]
		}
		lines = append(lines, prefix+" "+string(chunk))
		runes = runes[len(chunk):]
		if len(runes) == 0 {
			return lines
		}
		prefix = indent
	}
}

// commitRows renders one line per commit — aligned short hash, the author's
// initials in the accent tint, then the subject. Long subjects clip at the
// panel border rather than wrapping, keeping one row per commit.
func commitRows(commits []branch.CommitInfo) []string {
	hashWidth := 0
	for _, commit := range commits {
		hashWidth = max(hashWidth, len(commit.Hash))
	}
	rows := make([]string, 0, len(commits))
	for _, commit := range commits {
		author := commitAuthor(fitLine(authorInitials(commit.Author), 2))
		rows = append(rows, fitLine(commit.Hash, hashWidth)+" "+author+" "+commit.Subject)
	}
	return rows
}

// authorInitials abbreviates an author name to two characters: the first rune
// of the first and last words, or the first two runes of a single-word name.
func authorInitials(name string) string {
	fields := strings.Fields(name)
	switch len(fields) {
	case 0:
		return ""
	case 1:
		runes := []rune(fields[0])
		if len(runes) > 2 {
			runes = runes[:2]
		}
		return string(runes)
	default:
		return string([]rune(fields[0])[:1]) + string([]rune(fields[len(fields)-1])[:1])
	}
}

func (s *ChangeScreen) Footer(helpWidth int) string {
	return s.registry[s.activeMode()].footer(helpWidth)
}

func (s *ChangeScreen) buildRegistry() Registry {
	return Registry{
		ModeDefault: NewMode(
			Binding{Keys: []keys.Key{keys.KeyEnter}, Symbol: "enter", Label: "open", Action: s.actionSubmit},
			Binding{Keys: []keys.Key{keys.KeyCtrlA}, Symbol: "ctrl+a", Label: "add", Action: s.actionOpenAdd},
			Binding{Keys: []keys.Key{keys.KeyCtrlD}, Symbol: "ctrl+d", Label: "remove", Action: s.actionStartRemove},
			Binding{Keys: []keys.Key{keys.KeyCtrlP}, Symbol: "ctrl+p", Label: "prune", Action: s.actionStartPrune},
			Binding{Keys: []keys.Key{keys.KeyUp, keys.KeyShiftTab}, Symbol: "↑/shift+tab", Label: "move", Action: s.actionMoveSelection},
			Binding{Keys: []keys.Key{keys.KeyDown, keys.KeyTab}, Symbol: "↓/tab", Label: "move", Action: s.actionMoveSelection},
			Binding{Keys: []keys.Key{keys.KeyEsc, keys.KeyCtrlC}, Symbol: "esc", Label: "quit", Action: s.actionQuit},
		),
		ModeDialog: NewMode(
			Binding{Keys: []keys.Key{keys.KeyEnter}, Symbol: "enter", Label: "confirm", Action: s.actionConfirmDialog},
			Binding{Keys: []keys.Key{keys.KeyTab, keys.KeyShiftTab}, Symbol: "tab", Label: "move", Action: s.actionDialogMove},
			Binding{Keys: []keys.Key{keys.KeyEsc}, Symbol: "esc", Label: "cancel", Action: s.actionCancelDialog},
			Binding{Keys: []keys.Key{keys.KeyCtrlC}, Symbol: "ctrl+c", Label: "quit", Action: s.actionQuit},
		),
		ModeAdd: NewMode(
			Binding{Keys: []keys.Key{keys.KeyEnter}, Symbol: "enter", Label: "submit", Action: s.actionSubmitAdd},
			Binding{Keys: []keys.Key{keys.KeyTab, keys.KeyShiftTab, keys.KeyUp, keys.KeyDown}, Symbol: "tab", Label: "switch field", Action: s.actionAddSwitchFocus},
			Binding{Keys: []keys.Key{keys.KeyEsc, keys.KeyCtrlA}, Symbol: "esc", Label: "cancel", Action: s.actionCancelAdd},
			Binding{Keys: []keys.Key{keys.KeyCtrlC}, Symbol: "ctrl+c", Label: "quit", Action: s.actionQuit},
		),
	}
}

func (s *ChangeScreen) actionSubmit(actx *ActionCtx) app.Command {
	item, ok := s.list.SelectedItem()
	if !ok {
		return s.app.RequestSubmitSelectedPath("")
	}
	return s.app.RequestSubmitSelectedPath(item.ID)
}

func (s *ChangeScreen) actionOpenAdd(actx *ActionCtx) app.Command {
	s.addDlg.open()
	return nil
}

// actionSubmitAdd submits the add dialog. The dialog closes only when the app
// accepts the request (returns a command); a validation failure leaves it open
// with the typed values intact so the user can correct them.
func (s *ChangeScreen) actionSubmitAdd(actx *ActionCtx) app.Command {
	path, branch := s.addDlg.values()
	cmd := s.app.RequestAddWorktree(path, branch)
	if cmd != nil {
		s.addDlg.close()
	}
	return cmd
}

func (s *ChangeScreen) actionCancelAdd(actx *ActionCtx) app.Command {
	s.addDlg.close()
	return nil
}

func (s *ChangeScreen) actionAddSwitchFocus(actx *ActionCtx) app.Command {
	s.addDlg.switchFocus()
	return nil
}

func (s *ChangeScreen) actionStartRemove(actx *ActionCtx) app.Command {
	item, ok := s.list.SelectedItem()
	if !ok {
		return s.app.RemoveWorktree("")
	}
	path := item.ID
	if s.isDirty(path) {
		s.confirm.open(
			"Force delete worktree?",
			path+" has uncommitted or untracked changes that will be permanently lost.",
			"Force delete",
			false,
			func(actx *ActionCtx) app.Command {
				return s.app.ForceRemoveWorktree(path)
			},
		)
		return nil
	}
	s.confirm.open("Delete worktree?", path, "Delete", false, func(actx *ActionCtx) app.Command {
		return s.app.RemoveWorktree(path)
	})
	return nil
}

// isDirty reports whether the worktree at path has uncommitted or untracked
// changes, so that removing it needs git's --force.
func (s *ChangeScreen) isDirty(path string) bool {
	for _, worktree := range s.worktrees {
		if worktree.id == path {
			return worktree.dirty
		}
	}
	return false
}

func (s *ChangeScreen) actionStartPrune(actx *ActionCtx) app.Command {
	if !s.hasStale() {
		// Nothing to prune; the app reports the no-op without a dialog.
		return s.app.PruneWorktrees()
	}
	s.confirm.open("Prune all stale worktrees?", "", "Prune", false, func(actx *ActionCtx) app.Command {
		return s.app.PruneWorktrees()
	})
	return nil
}

func (s *ChangeScreen) hasStale() bool {
	for _, worktree := range s.worktrees {
		if worktree.stale {
			return true
		}
	}
	return false
}

func (s *ChangeScreen) actionMoveSelection(actx *ActionCtx) app.Command {
	s.list.Update(actx.Key)
	return s.fetchHoveredCommits()
}

func (s *ChangeScreen) actionQuit(actx *ActionCtx) app.Command {
	return s.app.Quit()
}

func (s *ChangeScreen) actionConfirmDialog(actx *ActionCtx) app.Command {
	return s.confirm.confirm(actx)
}

func (s *ChangeScreen) actionCancelDialog(actx *ActionCtx) app.Command {
	s.confirm.close()
	return nil
}

func (s *ChangeScreen) actionDialogMove(actx *ActionCtx) app.Command {
	s.confirm.move(actx.Key)
	return nil
}

func (s *ChangeScreen) Reset() {
	s.search.Clear()
	s.search.Focus()
	s.commitsBranch = ""
	s.confirm.close()
	s.addDlg.close()
	s.list.SetItems(toItems(s.worktrees))
}

func filterItems(items []selectlist.Item, query string) []selectlist.Item {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return append([]selectlist.Item(nil), items...)
	}
	filtered := make([]selectlist.Item, 0, len(items))
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Label), query) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func toItems(worktrees []appWorktree) []selectlist.Item {
	items := make([]selectlist.Item, 0, len(worktrees))
	for _, worktree := range worktrees {
		items = append(items, selectlist.Item{ID: worktree.id, Label: worktree.label, Color: worktree.color})
	}
	return items
}
