package screens

import (
	"strings"

	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/ui/components/dialog"
	"github.com/M-Xue/grove/ui/components/selectlist"
	"github.com/M-Xue/grove/ui/components/textinput"
	"github.com/M-Xue/grove/ui/keys"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type ChangeScreen struct {
	dialog          dialog.Model
	search          textinput.Model
	list            selectlist.Model
	defaultHandlers BindingSet
	confirmHandlers BindingSet
	worktrees       []appWorktree
	dialogSignature string
}

type appWorktree struct {
	id    string
	label string
}

func NewChangeScreen() *ChangeScreen {
	s := &ChangeScreen{
		dialog: dialog.New(),
		search: textinput.New("Search worktree paths"),
		list:   selectlist.New("No matches"),
	}
	s.search.Focus()
	s.initHandlers()
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
	if state.Dialog.Active {
		signature := dialogSignature(state.Dialog)
		if signature != s.dialogSignature {
			s.dialog.SetTitle(state.Dialog.Title)
			s.dialog.SetDescription(state.Dialog.Description)
			buttons := make([]dialog.Button, 0, len(state.Dialog.Buttons))
			for _, button := range state.Dialog.Buttons {
				buttons = append(buttons, dialog.Button{ID: button.ID, Label: button.Label})
			}
			s.dialog.SetButtons(buttons)
			s.dialog.SetFocusedID(state.Dialog.FocusedID)
			s.dialogSignature = signature
		}
	} else {
		s.dialogSignature = ""
	}
}

func (s *ChangeScreen) Update(ctx *ScreenContext, msg tea.KeyMsg, state app.State) tea.Cmd {
	if state.Dialog.Active {
		if consumed, cmd := s.dialog.Update(msg); consumed {
			if cmd != nil {
				return cmd
			}
		}
		if handler, ok := s.confirmHandlers.HandlerFor(keys.Normalize(msg)); ok {
			return handler(ctx, msg)
		}
		return nil
	}
	if consumed, cmd := s.search.Update(msg); consumed {
		s.list.SetItems(filterItems(toItems(s.worktrees), s.search.Value()))
		return cmd
	}
	if consumed, cmd := s.list.Update(msg); consumed {
		return cmd
	}
	if handler, ok := s.defaultHandlers.HandlerFor(keys.Normalize(msg)); ok {
		return handler(ctx, msg)
	}
	return nil
}

func (s *ChangeScreen) View(width, height int, state app.State) string {
	header := []string{"grove", "", "Change worktree", "", s.search.View(), ""}
	body := s.list.View(screenMax(1, height-len(header)))
	contentLines := append(header, strings.Split(body, "\n")...)
	content := strings.Join(contentLines, "\n")
	if state.Dialog.Active {
		return overlayDialog(content, s.dialog.View(width, height), width, height)
	}
	return content
}

func (s *ChangeScreen) Footer(helpWidth int) string {
	model := NewHelpModel()
	model.Width = screenMax(helpWidth, 0)
	order := []keys.Key{keys.KeyEnter, keys.KeyQuestion, keys.KeyCtrlA, keys.KeyCtrlD, keys.KeyUp, keys.KeyDown, keys.KeyEsc}
	return model.ShortHelpView(s.defaultHandlers.HelpBindings(order))
}

func (s *ChangeScreen) DialogFooter(helpWidth int) string {
	model := NewHelpModel()
	model.Width = screenMax(helpWidth, 0)
	order := []keys.Key{keys.KeyEnter, keys.KeyEsc, keys.KeyCtrlC}
	return model.ShortHelpView(s.confirmHandlers.HelpBindings(order))
}

func (s *ChangeScreen) initHandlers() {
	s.defaultHandlers = BindingSet{
		keys.KeyEnter:    {Key: keys.KeyEnter, Handler: s.handleEnter, Help: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open"))},
		keys.KeyCtrlA:    {Key: keys.KeyCtrlA, Handler: s.handleOpenAdd, Help: key.NewBinding(key.WithKeys("ctrl+a"), key.WithHelp("ctrl+a", "add"))},
		keys.KeyCtrlD:    {Key: keys.KeyCtrlD, Handler: s.handleStartRemove, Help: key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "remove"))},
		keys.KeyQuestion: {Key: keys.KeyQuestion, Handler: s.handleOpenDocs, Help: key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "docs"))},
		keys.KeyUp:       {Key: keys.KeyUp, Handler: nil, Help: key.NewBinding(key.WithKeys("up", "shift+tab"), key.WithHelp("up", "move"))},
		keys.KeyDown:     {Key: keys.KeyDown, Handler: nil, Help: key.NewBinding(key.WithKeys("down", "tab"), key.WithHelp("down", "move"))},
		keys.KeyEsc:      {Key: keys.KeyEsc, Handler: s.handleQuit, Help: key.NewBinding(key.WithKeys("esc", "ctrl+c"), key.WithHelp("esc", "quit"))},
		keys.KeyCtrlC:    {Key: keys.KeyCtrlC, Handler: s.handleQuit, Help: key.NewBinding(key.WithKeys("esc", "ctrl+c"), key.WithHelp("esc", "quit"))},
	}
	s.confirmHandlers = BindingSet{
		keys.KeyEnter: {Key: keys.KeyEnter, Handler: s.handleConfirmDialog, Help: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm"))},
		keys.KeyEsc:   {Key: keys.KeyEsc, Handler: s.handleCancelDialog, Help: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))},
		keys.KeyCtrlC: {Key: keys.KeyCtrlC, Handler: s.handleQuit, Help: key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit"))},
	}
}

func (s *ChangeScreen) handleEnter(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	item, ok := s.list.SelectedItem()
	if !ok {
		return ctx.RunEffect(ctx.App.RequestSubmitSelectedPath(""))
	}
	return ctx.RunEffect(ctx.App.RequestSubmitSelectedPath(item.ID))
}

func (s *ChangeScreen) handleOpenAdd(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	ctx.App.OpenAdd()
	return nil
}

func (s *ChangeScreen) handleStartRemove(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	item, ok := s.list.SelectedItem()
	if !ok {
		ctx.App.RequestRemoveWorktree("")
		return nil
	}
	ctx.App.RequestRemoveWorktree(item.ID)
	return nil
}

func (s *ChangeScreen) handleOpenDocs(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	return ctx.RunEffect(ctx.App.OpenDocs())
}

func (s *ChangeScreen) handleQuit(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	return ctx.Quit()
}

func (s *ChangeScreen) handleConfirmDialog(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	buttonID, _ := s.dialog.FocusedID()
	return ctx.RunEffect(ctx.App.DialogChoose(buttonID))
}

func (s *ChangeScreen) handleCancelDialog(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	ctx.App.DismissDialog()
	return nil
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

func dialogSignature(state app.DialogState) string {
	parts := []string{string(state.Kind), state.Title, state.Description, state.FocusedID}
	for _, button := range state.Buttons {
		parts = append(parts, button.ID, button.Label)
	}
	return strings.Join(parts, "\x00")
}
