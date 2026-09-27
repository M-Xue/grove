package screens

import (
	"fmt"
	"strings"

	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/ui/components/panel"
	"github.com/M-Xue/grove/ui/components/selectlist"
	"github.com/M-Xue/grove/ui/components/textinput"
	"github.com/M-Xue/grove/ui/keys"
	tea "github.com/charmbracelet/bubbletea"
)

// changeApp is the narrow view of app the change screen depends on.
type changeApp interface {
	RequestSubmitSelectedPath(path string) app.Command
	RequestAddWorktree(path, branch string) app.Command
	CreateBranchWorktree(path, branch string) app.Command
	RemoveWorktree(path string) app.Command
	ForceRemoveWorktree(path string) app.Command
	PruneWorktrees() app.Command
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
}

// staleColor is the ANSI escape used to dim stale worktrees in the list.
const staleColor = "\x1b[38;5;244m"

type appWorktree struct {
	id    string
	label string
	color string
	stale bool
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
		s.worktrees = append(s.worktrees, appWorktree{id: worktree.Path, label: label, color: color, stale: worktree.Stale, dirty: dirty})
	}
	items = filterItems(items, s.search.Value())
	s.list.SetItems(items)
}

// OnMessage reacts to the semantic outcome of the add dialog's branch check:
// when the branch is absent, it opens the confirm dialog offering to create it.
func (s *ChangeScreen) OnMessage(ctx *ScreenContext, msg app.Message) tea.Cmd {
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
			s.list.SetItems(filterItems(toItems(s.worktrees), s.search.Value()))
			return cmd
		}
	}
	return nil
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
	searchLabel := "Search "
	s.search.SetWidth(max(0, min(searchFieldWidth, leftWidth-4-len(searchLabel))))
	searchRow := searchLabel + s.search.View()
	interior := strings.Join(append([]string{"", searchRow, ""}, strings.Split(s.list.View(listHeight), "\n")...), "\n")
	// The worktrees panel is the active section whenever no dialog owns the
	// keyboard; dialogs render their own active border.
	worktreesActive := !s.confirm.active && !s.addDlg.active
	left := strings.Split(panel.Render("Worktrees", interior, leftWidth, panelHeight, worktreesActive), "\n")
	right := strings.Split(panel.Render("Branch details", "", rightWidth, panelHeight, false), "\n")
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
	return nil
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
