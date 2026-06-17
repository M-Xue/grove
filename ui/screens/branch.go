package screens

import (
	"fmt"
	"strings"

	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/ui/components/dialog"
	"github.com/M-Xue/grove/ui/components/selectlist"
	"github.com/M-Xue/grove/ui/components/textinput"
	"github.com/M-Xue/grove/ui/keys"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const maxVisibleBranches = 10

const branchWarningPrefixColor = "\x1b[38;5;203m"
const branchWarningResetColor = "\x1b[0m"

type BranchScreen struct {
	dialog          dialog.Model
	search          textinput.Model
	list            selectlist.Model
	defaultHandlers BindingSet
	confirmHandlers BindingSet
	branches        []branchItem
	dialogSignature string
}

type branchItem struct {
	id    string
	label string
}

func NewBranchScreen() *BranchScreen {
	s := &BranchScreen{
		dialog: dialog.New(),
		search: textinput.New("Search branches"),
		list:   selectlist.New("No matches"),
	}
	s.search.Focus()
	s.initHandlers()
	return s
}

func (s *BranchScreen) Sync(state app.State) {
	selectedID, _ := s.list.SelectedID()
	s.branches = make([]branchItem, 0, len(state.Branches))
	items := make([]selectlist.Item, 0, len(state.Branches))
	for _, br := range state.Branches {
		label := br.Name
		if br.CheckedOutHere {
			label += " [current]"
		} else if br.CheckedOutElsewhere {
			label += " [worktree]"
		}
		items = append(items, selectlist.Item{ID: br.Name, Label: label})
		s.branches = append(s.branches, branchItem{id: br.Name, label: label})
	}
	filtered := filterItems(items, s.search.Value())
	s.list.SetItems(filtered)
	if selectedID != "" {
		s.list.SetSelectedID(selectedID)
	} else if state.Branch.SelectedName != "" {
		s.list.SetSelectedID(state.Branch.SelectedName)
	}
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

func (s *BranchScreen) Update(ctx *ScreenContext, msg tea.KeyMsg, state app.State) tea.Cmd {
	if state.Dialog.Active {
		if consumed, cmd := s.dialog.Update(msg); consumed {
			if cmd != nil {
				return cmd
			}
		}
		if handler, ok := s.confirmHandlers.HandlerFor(keys.Normalize(msg)); ok && handler != nil {
			return handler(ctx, msg)
		}
		return nil
	}
	if handler, ok := s.defaultHandlers.HandlerFor(keys.Normalize(msg)); ok && handler != nil {
		return handler(ctx, msg)
	}
	if consumed, cmd := s.search.Update(msg); consumed {
		s.list.SetItems(filterItems(toBranchSelectItems(s.branches), s.search.Value()))
		return tea.Batch(cmd, s.selectionCommand(ctx))
	}
	if consumed, cmd := s.list.Update(msg); consumed {
		return tea.Batch(cmd, s.selectionCommand(ctx))
	}
	return nil
}

func (s *BranchScreen) View(width, height int, state app.State) string {
	leftWidth := max(20, width/2)
	rightWidth := max(20, width-leftWidth-3)
	header := []string{
		"grove",
		"",
		fitLine(lipgloss.NewStyle().Bold(true).Render("Switch branch"), leftWidth) + "   " + fitLine(lipgloss.NewStyle().Bold(true).Render("Recent commits"), rightWidth),
		fitLine(fmt.Sprintf("Showing %s branches", scopeLabel(state.BranchScope)), leftWidth) + "   " + fitLine(branchLabel(state.Branch.SelectedName), rightWidth),
		fitLine(s.search.View(), leftWidth) + "   " + fitLine("", rightWidth),
		"",
	}
	if warning := branchSwitchWarning(state); warning != "" {
		header = append(header, warning, "")
	}
	listHeight := max(1, min(maxVisibleBranches, height-len(header)))
	leftBody := s.list.View(listHeight)
	rightBody := s.commitsView(rightWidth, listHeight, state)
	leftLines := strings.Split(leftBody, "\n")
	rightLines := strings.Split(rightBody, "\n")
	bodyLines := make([]string, 0, listHeight)
	for i := 0; i < listHeight; i++ {
		leftLine := ""
		if i < len(leftLines) {
			leftLine = fitLine(leftLines[i], leftWidth)
		} else {
			leftLine = fitLine("", leftWidth)
		}
		rightLine := ""
		if i < len(rightLines) {
			rightLine = fitLine(rightLines[i], rightWidth)
		} else {
			rightLine = fitLine("", rightWidth)
		}
		bodyLines = append(bodyLines, leftLine+"   "+rightLine)
	}
	content := strings.Join(append(header, bodyLines...), "\n")
	if state.Dialog.Active {
		return overlayDialog(content, s.dialog.View(width, height), width, height)
	}
	return content
}

func (s *BranchScreen) Footer(helpWidth int, scope branch.Scope) string {
	model := NewHelpModel()
	model.Width = max(helpWidth, 0)
	if s.dialogSignature != "" {
		bindings := s.confirmHandlers.HelpBindings([]keys.Key{keys.KeyEnter, keys.KeyTab, keys.KeyEsc, keys.KeyCtrlC})
		return model.ShortHelpView(bindings)
	}
	scopeHelp := "remote"
	if scope == branch.ScopeRemoteTracking {
		scopeHelp = "local"
	}
	bindings := s.defaultHandlers.HelpBindings([]keys.Key{keys.KeyEnter, keys.KeyCtrlD, keys.KeyCtrlShiftD, keys.KeyCtrlF, keys.KeyCtrlO, keys.KeyUp, keys.KeyDown, keys.KeyEsc})
	if len(bindings) >= 5 {
		bindings[4] = key.NewBinding(key.WithKeys("ctrl+o"), key.WithHelp("ctrl+o", scopeHelp))
	}
	return model.ShortHelpView(bindings)
}

func (s *BranchScreen) Reset() {
	s.search.Clear()
	s.search.Focus()
	s.dialogSignature = ""
	s.list.SetItems(toBranchSelectItems(s.branches))
}

func (s *BranchScreen) initHandlers() {
	s.defaultHandlers = BindingSet{
		keys.KeyEnter: {Key: keys.KeyEnter, Handler: s.handleCheckout, Help: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "switch"))},
		keys.KeyCtrlD: {Key: keys.KeyCtrlD, Handler: s.handleDelete, Help: key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "delete"))},
		keys.KeyCtrlShiftD: {Key: keys.KeyCtrlShiftD, Handler: s.handleDeleteAll, Help: key.NewBinding(key.WithKeys("ctrl+shift+d"), key.WithHelp("ctrl+shift+d", "delete all"))},
		keys.KeyCtrlF: {Key: keys.KeyCtrlF, Handler: s.handleFetch, Help: key.NewBinding(key.WithKeys("ctrl+f"), key.WithHelp("ctrl+f", "fetch"))},
		keys.KeyCtrlO: {Key: keys.KeyCtrlO, Handler: s.handleToggleScope, Help: key.NewBinding(key.WithKeys("ctrl+o"), key.WithHelp("ctrl+o", "remote"))},
		keys.KeyUp:    {Key: keys.KeyUp, Handler: nil, Help: key.NewBinding(key.WithKeys("up", "shift+tab"), key.WithHelp("↑/shift+tab", "move"))},
		keys.KeyDown:  {Key: keys.KeyDown, Handler: nil, Help: key.NewBinding(key.WithKeys("down", "tab"), key.WithHelp("↓/tab", "move"))},
		keys.KeyEsc:   {Key: keys.KeyEsc, Handler: s.handleClose, Help: key.NewBinding(key.WithKeys("esc", "ctrl+b"), key.WithHelp("esc", "close"))},
		keys.KeyCtrlB: {Key: keys.KeyCtrlB, Handler: s.handleClose, Help: key.NewBinding(key.WithKeys("esc", "ctrl+b"), key.WithHelp("esc", "close"))},
		keys.KeyCtrlC: {Key: keys.KeyCtrlC, Handler: s.handleQuit, Help: key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit"))},
	}
	s.confirmHandlers = BindingSet{
		keys.KeyEnter:    {Key: keys.KeyEnter, Handler: s.handleConfirmDialog, Help: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm"))},
		keys.KeyTab:      {Key: keys.KeyTab, Handler: nil, Help: key.NewBinding(key.WithKeys("tab", "shift+tab"), key.WithHelp("tab", "move"))},
		keys.KeyShiftTab: {Key: keys.KeyShiftTab, Handler: nil, Help: key.NewBinding(key.WithKeys("tab", "shift+tab"), key.WithHelp("tab", "move"))},
		keys.KeyEsc:      {Key: keys.KeyEsc, Handler: s.handleCancelDialog, Help: key.NewBinding(key.WithKeys("esc", "ctrl+b"), key.WithHelp("esc", "cancel"))},
		keys.KeyCtrlB:    {Key: keys.KeyCtrlB, Handler: s.handleCancelDialog, Help: key.NewBinding(key.WithKeys("esc", "ctrl+b"), key.WithHelp("esc", "cancel"))},
		keys.KeyCtrlC:    {Key: keys.KeyCtrlC, Handler: s.handleQuit, Help: key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit"))},
	}
}

func (s *BranchScreen) handleCheckout(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	item, ok := s.list.SelectedItem()
	if !ok {
		return ctx.Run(ctx.App.RequestCheckoutBranch(""))
	}
	return ctx.Run(ctx.App.RequestCheckoutBranch(item.ID))
}

func (s *BranchScreen) handleFetch(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	return ctx.Run(ctx.App.RequestFetchBranches())
}

func (s *BranchScreen) handleDelete(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	item, ok := s.list.SelectedItem()
	if !ok {
		return ctx.Run(ctx.App.RequestDeleteBranch(""))
	}
	return ctx.Run(ctx.App.RequestDeleteBranch(item.ID))
}

func (s *BranchScreen) handleDeleteAll(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	return ctx.Run(ctx.App.RequestDeleteAllBranches())
}

func (s *BranchScreen) handleConfirmDialog(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	buttonID, _ := s.dialog.FocusedID()
	return ctx.Run(ctx.App.DialogChoose(buttonID))
}

func (s *BranchScreen) handleCancelDialog(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	ctx.App.DismissDialog()
	return nil
}

func (s *BranchScreen) handleToggleScope(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	return ctx.Run(ctx.App.RequestToggleBranchScope())
}

func (s *BranchScreen) handleClose(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	return ctx.Run(ctx.App.CloseBranch())
}

func (s *BranchScreen) handleQuit(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	return ctx.Quit()
}

func (s *BranchScreen) selectionCommand(ctx *ScreenContext) tea.Cmd {
	item, ok := s.list.SelectedItem()
	if !ok {
		return ctx.Run(ctx.App.SelectBranch(""))
	}
	return ctx.Run(ctx.App.SelectBranch(item.ID))
}

func (s *BranchScreen) commitsView(width, height int, state app.State) string {
	lines := []string{}
	if len(state.Branch.Commits) == 0 {
		lines = append(lines, "No commits")
		return strings.Join(lines, "\n")
	}
	for _, commit := range state.Branch.Commits {
		author := truncateText(commit.Author, 8)
		subjectWidth := max(8, width-max(0, lipgloss.Width(commit.Hash)+lipgloss.Width(author)+6))
		subject := truncateText(commit.Subject, subjectWidth)
		lines = append(lines, fitLine(commit.Hash+"  "+author+"  "+subject, width))
	}
	return strings.Join(lines, "\n")
}

func toBranchSelectItems(branches []branchItem) []selectlist.Item {
	items := make([]selectlist.Item, 0, len(branches))
	for _, br := range branches {
		items = append(items, selectlist.Item{ID: br.id, Label: br.label})
	}
	return items
}

func scopeLabel(scope branch.Scope) string {
	if scope == branch.ScopeRemoteTracking {
		return "remote-tracking"
	}
	return "local"
}

func truncateText(value string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(value) <= width {
		return value
	}
	if width <= 3 {
		return value[:width]
	}
	trimmed := value
	for lipgloss.Width(trimmed) > width-3 && len(trimmed) > 0 {
		trimmed = trimmed[:len(trimmed)-1]
	}
	return trimmed + "..."
}

func branchLabel(name string) string {
	if name == "" {
		return ""
	}
	return "Branch: " + name
}

func branchSwitchWarning(state app.State) string {
	currentBranch := currentBranchName(state.Branches)
	if currentBranch == "" {
		return ""
	}
	for _, worktree := range state.Worktrees {
		if worktree.Branch == currentBranch && worktree.HasUncommittedChanges {
			return branchWarningPrefixColor + "warning" + branchWarningResetColor + " uncommitted files may prevent switching branches"
		}
	}
	return ""
}

func currentBranchName(branches []branch.Info) string {
	for _, br := range branches {
		if br.CheckedOutHere {
			return br.Name
		}
	}
	return ""
}
