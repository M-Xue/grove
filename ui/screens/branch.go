package screens

import (
	"fmt"
	"strings"

	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/ui/components/selectlist"
	"github.com/M-Xue/grove/ui/components/textinput"
	"github.com/M-Xue/grove/ui/keys"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const maxVisibleBranches = 10

const branchWarningPrefixColor = "\x1b[38;5;203m"
const branchWarningResetColor = "\x1b[0m"

type BranchScreen struct {
	confirm  confirmDialog
	search   textinput.Model
	list     selectlist.Model
	registry Registry
	branches []branchItem
	scope    branch.Scope
}

type branchItem struct {
	id    string
	label string
}

func NewBranchScreen() *BranchScreen {
	s := &BranchScreen{
		search: textinput.New("Search branches"),
		list:   selectlist.New("No matches"),
	}
	s.search.Focus()
	s.registry = s.buildRegistry()
	return s
}

func (s *BranchScreen) Sync(state app.State) {
	selectedID, _ := s.list.SelectedID()
	s.scope = state.BranchScope
	s.registry = s.buildRegistry()
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
}

func (s *BranchScreen) OnMessage(ctx *ScreenContext, msg app.Message) tea.Cmd {
	return nil
}

func (s *BranchScreen) activeMode() Mode {
	if s.confirm.active {
		return ModeDialog
	}
	return ModeDefault
}

func (s *BranchScreen) Update(ctx *ScreenContext, msg tea.KeyMsg, state app.State) tea.Cmd {
	mode := s.activeMode()
	if binding, ok := s.registry[mode].lookup(keys.Normalize(msg)); ok {
		return ctx.Run(binding.Action(&ActionCtx{App: ctx.App, Key: msg}))
	}
	if mode == ModeDefault {
		if consumed, cmd := s.search.Update(msg); consumed {
			s.list.SetItems(filterItems(toBranchSelectItems(s.branches), s.search.Value()))
			return tea.Batch(cmd, s.selectionCommand(ctx))
		}
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
	if s.confirm.active {
		return overlayDialog(content, s.confirm.view(width, height), width, height)
	}
	return content
}

func (s *BranchScreen) Footer(helpWidth int) string {
	return s.registry[s.activeMode()].footer(helpWidth)
}

func (s *BranchScreen) buildRegistry() Registry {
	scopeToggleLabel := "remote"
	if s.scope == branch.ScopeRemoteTracking {
		scopeToggleLabel = "local"
	}
	return Registry{
		ModeDefault: NewMode(
			Binding{Keys: []keys.Key{keys.KeyEnter}, Symbol: "enter", Label: "switch", Action: s.actionCheckout},
			Binding{Keys: []keys.Key{keys.KeyCtrlD}, Symbol: "ctrl+d", Label: "delete", Action: s.actionDelete},
			Binding{Keys: []keys.Key{keys.KeyCtrlShiftD}, Symbol: "ctrl+shift+d", Label: "delete all", Action: s.actionDeleteAll},
			Binding{Keys: []keys.Key{keys.KeyCtrlF}, Symbol: "ctrl+f", Label: "fetch", Action: s.actionFetch},
			Binding{Keys: []keys.Key{keys.KeyCtrlO}, Symbol: "ctrl+o", Label: scopeToggleLabel, Action: s.actionToggleScope},
			Binding{Keys: []keys.Key{keys.KeyUp, keys.KeyShiftTab}, Symbol: "↑/shift+tab", Label: "move", Action: s.actionMoveSelection},
			Binding{Keys: []keys.Key{keys.KeyDown, keys.KeyTab}, Symbol: "↓/tab", Label: "move", Action: s.actionMoveSelection},
			Binding{Keys: []keys.Key{keys.KeyEsc, keys.KeyCtrlB}, Symbol: "esc", Label: "close", Action: s.actionClose},
			Binding{Keys: []keys.Key{keys.KeyCtrlC}, Symbol: "", Label: "", Action: s.actionQuit},
		),
		ModeDialog: NewMode(
			Binding{Keys: []keys.Key{keys.KeyEnter}, Symbol: "enter", Label: "confirm", Action: s.actionConfirmDialog},
			Binding{Keys: []keys.Key{keys.KeyTab, keys.KeyShiftTab}, Symbol: "tab", Label: "move", Action: s.actionDialogMove},
			Binding{Keys: []keys.Key{keys.KeyEsc, keys.KeyCtrlB}, Symbol: "esc", Label: "cancel", Action: s.actionCancelDialog},
			Binding{Keys: []keys.Key{keys.KeyCtrlC}, Symbol: "ctrl+c", Label: "quit", Action: s.actionQuit},
		),
	}
}

func (s *BranchScreen) actionCheckout(actx *ActionCtx) app.Command {
	item, ok := s.list.SelectedItem()
	if !ok {
		return actx.App.RequestCheckoutBranch("")
	}
	return actx.App.RequestCheckoutBranch(item.ID)
}

func (s *BranchScreen) actionFetch(actx *ActionCtx) app.Command {
	return actx.App.RequestFetchBranches()
}

func (s *BranchScreen) actionToggleScope(actx *ActionCtx) app.Command {
	return actx.App.RequestToggleBranchScope()
}

func (s *BranchScreen) actionClose(actx *ActionCtx) app.Command {
	return actx.App.CloseBranch()
}

func (s *BranchScreen) actionQuit(actx *ActionCtx) app.Command {
	return actx.App.Quit()
}

func (s *BranchScreen) actionMoveSelection(actx *ActionCtx) app.Command {
	s.list.Update(actx.Key)
	item, ok := s.list.SelectedItem()
	if !ok {
		return actx.App.SelectBranch("")
	}
	return actx.App.SelectBranch(item.ID)
}

func (s *BranchScreen) actionDelete(actx *ActionCtx) app.Command {
	item, ok := s.list.SelectedItem()
	if !ok {
		return actx.App.DeleteBranch("")
	}
	name := item.ID
	s.confirm.open("Delete branch?", name, "Delete", false, func(actx *ActionCtx) app.Command {
		return actx.App.DeleteBranch(name)
	})
	return nil
}

func (s *BranchScreen) actionDeleteAll(actx *ActionCtx) app.Command {
	if !actx.App.CanDeleteAllBranches() {
		return nil
	}
	description := "Delete all local branches except ones currently checked out here or in another worktree?"
	if names := s.branchNames(); len(names) > 0 {
		description += "\n\n" + strings.Join(names, "\n")
	}
	s.confirm.open("Delete all local branches?", description, "Delete", false, func(actx *ActionCtx) app.Command {
		return actx.App.DeleteAllBranches()
	})
	return nil
}

func (s *BranchScreen) actionConfirmDialog(actx *ActionCtx) app.Command {
	return s.confirm.confirm(actx)
}

func (s *BranchScreen) actionCancelDialog(actx *ActionCtx) app.Command {
	s.confirm.close()
	return nil
}

func (s *BranchScreen) actionDialogMove(actx *ActionCtx) app.Command {
	s.confirm.move(actx.Key)
	return nil
}

func (s *BranchScreen) branchNames() []string {
	names := make([]string, 0, len(s.branches))
	for _, b := range s.branches {
		if b.id == "" {
			continue
		}
		names = append(names, b.id)
	}
	return names
}

func (s *BranchScreen) selectionCommand(ctx *ScreenContext) tea.Cmd {
	item, ok := s.list.SelectedItem()
	if !ok {
		return ctx.Run(ctx.App.SelectBranch(""))
	}
	return ctx.Run(ctx.App.SelectBranch(item.ID))
}

func (s *BranchScreen) Reset() {
	s.search.Clear()
	s.search.Focus()
	s.confirm.close()
	s.list.SetItems(toBranchSelectItems(s.branches))
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
