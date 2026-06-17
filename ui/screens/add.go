package screens

import (
	"strings"

	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/ui/components/dialog"
	"github.com/M-Xue/grove/ui/components/textinput"
	"github.com/M-Xue/grove/ui/keys"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AddScreen struct {
	dialog          dialog.Model
	path            textinput.Model
	branch          textinput.Model
	focusedPath     bool
	defaultHandlers BindingSet
	confirmHandlers BindingSet
	dialogSignature string
}

func NewAddScreen() *AddScreen {
	s := &AddScreen{
		dialog:      dialog.New(),
		path:        textinput.New("Relative path"),
		branch:      textinput.New("Branch name"),
		focusedPath: true,
	}
	s.path.Focus()
	s.initHandlers()
	return s
}

func (s *AddScreen) Sync(state app.State) {
	if state.Dialog.Active {
		signature := addDialogSignature(state.Dialog)
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

func (s *AddScreen) Update(ctx *ScreenContext, msg tea.KeyMsg, state app.State) tea.Cmd {
	if state.Dialog.Active {
		if consumed, cmd := s.dialog.Update(msg); consumed {
			return cmd
		}
		if handler, ok := s.confirmHandlers.HandlerFor(keys.Normalize(msg)); ok && handler != nil {
			return handler(ctx, msg)
		}
		return nil
	}
	active := &s.path
	other := &s.branch
	if !s.focusedPath {
		active = &s.branch
		other = &s.path
	}
	if consumed, cmd := active.Update(msg); consumed {
		return cmd
	}
	if handler, ok := s.defaultHandlers.HandlerFor(keys.Normalize(msg)); ok && handler != nil {
		return handler(ctx, msg)
	}
	other.Blur()
	active.Focus()
	return nil
}

func (s *AddScreen) View(width, height int, state app.State) string {
	header := []string{"grove", "", lipgloss.NewStyle().Bold(true).Render("Add worktree"), "", s.path.View(), s.branch.View()}
	content := strings.Join(header, "\n")
	if state.Dialog.Active {
		return overlayDialog(content, s.dialog.View(width, height), width, height)
	}
	return content
}

func (s *AddScreen) Footer(helpWidth int, dialogActive bool) string {
	model := NewHelpModel()
	model.Width = max(helpWidth, 0)
	if dialogActive {
		order := []keys.Key{keys.KeyEnter, keys.KeyTab, keys.KeyEsc, keys.KeyCtrlC}
		return model.ShortHelpView(s.confirmHandlers.HelpBindings(order))
	}
	order := []keys.Key{keys.KeyEnter, keys.KeyTab, keys.KeyEsc, keys.KeyCtrlC}
	return model.ShortHelpView(s.defaultHandlers.HelpBindings(order))
}

func (s *AddScreen) initHandlers() {
	s.defaultHandlers = BindingSet{
		keys.KeyEnter:    {Key: keys.KeyEnter, Handler: s.handleSubmit, Help: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit"))},
		keys.KeyEsc:      {Key: keys.KeyEsc, Handler: s.handleClose, Help: key.NewBinding(key.WithKeys("esc", "ctrl+a"), key.WithHelp("esc", "close"))},
		keys.KeyCtrlA:    {Key: keys.KeyCtrlA, Handler: s.handleClose, Help: key.NewBinding(key.WithKeys("esc", "ctrl+a"), key.WithHelp("esc", "close"))},
		keys.KeyCtrlC:    {Key: keys.KeyCtrlC, Handler: s.handleQuit, Help: key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit"))},
		keys.KeyTab:      {Key: keys.KeyTab, Handler: s.handleSwitchFocus, Help: key.NewBinding(key.WithKeys("tab", "shift+tab", "up", "down"), key.WithHelp("tab", "switch field"))},
		keys.KeyShiftTab: {Key: keys.KeyShiftTab, Handler: s.handleSwitchFocus, Help: key.NewBinding(key.WithKeys("tab", "shift+tab", "up", "down"), key.WithHelp("tab", "switch field"))},
		keys.KeyUp:       {Key: keys.KeyUp, Handler: s.handleSwitchFocus, Help: key.NewBinding(key.WithKeys("tab", "shift+tab", "up", "down"), key.WithHelp("↑", "switch field"))},
		keys.KeyDown:     {Key: keys.KeyDown, Handler: s.handleSwitchFocus, Help: key.NewBinding(key.WithKeys("tab", "shift+tab", "up", "down"), key.WithHelp("↓", "switch field"))},
	}
	s.confirmHandlers = BindingSet{
		keys.KeyEnter:    {Key: keys.KeyEnter, Handler: s.handleConfirmDialog, Help: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm"))},
		keys.KeyTab:      {Key: keys.KeyTab, Handler: nil, Help: key.NewBinding(key.WithKeys("tab", "shift+tab"), key.WithHelp("tab", "move"))},
		keys.KeyShiftTab: {Key: keys.KeyShiftTab, Handler: nil, Help: key.NewBinding(key.WithKeys("tab", "shift+tab"), key.WithHelp("tab", "move"))},
		keys.KeyEsc:      {Key: keys.KeyEsc, Handler: s.handleCancelDialog, Help: key.NewBinding(key.WithKeys("esc", "ctrl+a"), key.WithHelp("esc", "cancel"))},
		keys.KeyCtrlA:    {Key: keys.KeyCtrlA, Handler: s.handleCancelDialog, Help: key.NewBinding(key.WithKeys("esc", "ctrl+a"), key.WithHelp("esc", "cancel"))},
		keys.KeyCtrlC:    {Key: keys.KeyCtrlC, Handler: s.handleQuit, Help: key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit"))},
	}
}

func (s *AddScreen) handleSubmit(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	return ctx.Run(ctx.App.RequestAddWorktree(strings.TrimSpace(s.path.Value()), strings.TrimSpace(s.branch.Value())))
}

func (s *AddScreen) handleClose(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	ctx.App.CloseAdd()
	s.path.Clear()
	s.branch.Clear()
	s.focusedPath = true
	s.path.Focus()
	s.branch.Blur()
	return nil
}

func (s *AddScreen) handleQuit(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	return ctx.Quit()
}

func (s *AddScreen) handleSwitchFocus(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	s.focusedPath = !s.focusedPath
	if s.focusedPath {
		s.path.Focus()
		s.branch.Blur()
	} else {
		s.path.Blur()
		s.branch.Focus()
	}
	return nil
}

func (s *AddScreen) handleConfirmDialog(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	buttonID, _ := s.dialog.FocusedID()
	return ctx.Run(ctx.App.DialogChoose(buttonID))
}

func (s *AddScreen) handleCancelDialog(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	ctx.App.DismissDialog()
	return nil
}

func (s *AddScreen) Reset() {
	s.path.Clear()
	s.branch.Clear()
	s.focusedPath = true
	s.dialogSignature = ""
	s.path.Focus()
	s.branch.Blur()
}

func addDialogSignature(state app.DialogState) string {
	parts := []string{string(state.Kind), state.Title, state.Description, state.FocusedID}
	for _, button := range state.Buttons {
		parts = append(parts, button.ID, button.Label)
	}
	return strings.Join(parts, "\x00")
}
