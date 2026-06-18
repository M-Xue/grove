package screens

import (
	"strings"

	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/ui/components/selectlist"
	"github.com/M-Xue/grove/ui/components/textinput"
	"github.com/M-Xue/grove/ui/keys"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// changeApp is the narrow view of app the change screen depends on.
type changeApp interface {
	RequestSubmitSelectedPath(path string) app.Command
	OpenAdd()
	OpenBranch() app.Command
	RemoveWorktree(path string) app.Command
	Quit() app.Command
}

type ChangeScreen struct {
	app       changeApp
	confirm   confirmDialog
	search    textinput.Model
	list      selectlist.Model
	registry  Registry
	worktrees []appWorktree
}

type appWorktree struct {
	id    string
	label string
}

func NewChangeScreen(application changeApp) *ChangeScreen {
	s := &ChangeScreen{
		app:    application,
		search: textinput.New("Search worktree paths"),
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
		items = append(items, selectlist.Item{ID: worktree.Path, Label: label})
		s.worktrees = append(s.worktrees, appWorktree{id: worktree.Path, label: label})
	}
	items = filterItems(items, s.search.Value())
	s.list.SetItems(items)
}

func (s *ChangeScreen) OnMessage(ctx *ScreenContext, msg app.Message) tea.Cmd {
	return nil
}

func (s *ChangeScreen) activeMode() Mode {
	if s.confirm.active {
		return ModeDialog
	}
	return ModeDefault
}

func (s *ChangeScreen) Update(ctx *ScreenContext, msg tea.KeyMsg, state app.State) tea.Cmd {
	mode := s.activeMode()
	if binding, ok := s.registry[mode].lookup(keys.Normalize(msg)); ok {
		return ctx.Run(binding.Action(&ActionCtx{Key: msg}))
	}
	if mode == ModeDefault {
		if consumed, cmd := s.search.Update(msg); consumed {
			s.list.SetItems(filterItems(toItems(s.worktrees), s.search.Value()))
			return cmd
		}
	}
	return nil
}

func (s *ChangeScreen) View(width, height int, state app.State) string {
	header := []string{"grove", "", lipgloss.NewStyle().Bold(true).Render("Change worktree"), "", s.search.View(), ""}
	body := s.list.View(max(1, height-len(header)))
	contentLines := append(header, strings.Split(body, "\n")...)
	content := strings.Join(contentLines, "\n")
	if s.confirm.active {
		return overlayDialog(content, s.confirm.view(width, height), width, height)
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
			Binding{Keys: []keys.Key{keys.KeyCtrlB}, Symbol: "ctrl+b", Label: "branches", Action: s.actionOpenBranches},
			Binding{Keys: []keys.Key{keys.KeyCtrlD}, Symbol: "ctrl+d", Label: "remove", Action: s.actionStartRemove},
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
	s.app.OpenAdd()
	return nil
}

func (s *ChangeScreen) actionOpenBranches(actx *ActionCtx) app.Command {
	return s.app.OpenBranch()
}

func (s *ChangeScreen) actionStartRemove(actx *ActionCtx) app.Command {
	item, ok := s.list.SelectedItem()
	if !ok {
		return s.app.RemoveWorktree("")
	}
	path := item.ID
	s.confirm.open("Delete worktree?", path, "Delete", false, func(actx *ActionCtx) app.Command {
		return s.app.RemoveWorktree(path)
	})
	return nil
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
		items = append(items, selectlist.Item{ID: worktree.id, Label: worktree.label})
	}
	return items
}
