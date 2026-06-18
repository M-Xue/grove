package screens

import (
	"fmt"
	"strings"

	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/ui/components/textinput"
	"github.com/M-Xue/grove/ui/keys"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// addApp is the narrow view of app the add screen depends on.
type addApp interface {
	RequestAddWorktree(path, branch string) app.Command
	CloseAdd()
	CreateBranchWorktree(path, branch string) app.Command
	Quit() app.Command
}

type AddScreen struct {
	app         addApp
	confirm     confirmDialog
	path        textinput.Model
	branch      textinput.Model
	focusedPath bool
	registry    Registry
}

func NewAddScreen(application addApp) *AddScreen {
	s := &AddScreen{
		app:         application,
		path:        textinput.New("Relative path"),
		branch:      textinput.New("Branch name"),
		focusedPath: true,
	}
	s.path.Focus()
	s.registry = s.buildRegistry()
	return s
}

func (s *AddScreen) Sync(state app.State) {}

// OnMessage reacts to the semantic outcome of checking a branch: when the
// branch is absent, it opens a screen-owned dialog offering to create it.
func (s *AddScreen) OnMessage(ctx *ScreenContext, msg app.Message) tea.Cmd {
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

func (s *AddScreen) activeMode() Mode {
	if s.confirm.active {
		return ModeDialog
	}
	return ModeDefault
}

func (s *AddScreen) Update(ctx *ScreenContext, msg tea.KeyMsg, state app.State) tea.Cmd {
	mode := s.activeMode()
	if binding, ok := s.registry[mode].lookup(keys.Normalize(msg)); ok {
		return ctx.Run(binding.Action(&ActionCtx{Key: msg}))
	}
	if mode == ModeDefault {
		active := &s.path
		other := &s.branch
		if !s.focusedPath {
			active = &s.branch
			other = &s.path
		}
		if consumed, cmd := active.Update(msg); consumed {
			return cmd
		}
		other.Blur()
		active.Focus()
	}
	return nil
}

func (s *AddScreen) View(width, height int, state app.State) string {
	header := []string{"grove", "", lipgloss.NewStyle().Bold(true).Render("Add worktree"), "", s.path.View(), s.branch.View()}
	content := strings.Join(header, "\n")
	if s.confirm.active {
		return overlayDialog(content, s.confirm.view(width, height), width, height)
	}
	return content
}

func (s *AddScreen) Footer(helpWidth int) string {
	return s.registry[s.activeMode()].footer(helpWidth)
}

func (s *AddScreen) buildRegistry() Registry {
	return Registry{
		ModeDefault: NewMode(
			Binding{Keys: []keys.Key{keys.KeyEnter}, Symbol: "enter", Label: "submit", Action: s.actionSubmit},
			Binding{Keys: []keys.Key{keys.KeyTab, keys.KeyShiftTab, keys.KeyUp, keys.KeyDown}, Symbol: "tab", Label: "switch field", Action: s.actionSwitchFocus},
			Binding{Keys: []keys.Key{keys.KeyEsc, keys.KeyCtrlA}, Symbol: "esc", Label: "close", Action: s.actionClose},
			Binding{Keys: []keys.Key{keys.KeyCtrlC}, Symbol: "ctrl+c", Label: "quit", Action: s.actionQuit},
		),
		ModeDialog: NewMode(
			Binding{Keys: []keys.Key{keys.KeyEnter}, Symbol: "enter", Label: "confirm", Action: s.actionConfirmDialog},
			Binding{Keys: []keys.Key{keys.KeyTab, keys.KeyShiftTab}, Symbol: "tab", Label: "move", Action: s.actionDialogMove},
			Binding{Keys: []keys.Key{keys.KeyEsc, keys.KeyCtrlA}, Symbol: "esc", Label: "cancel", Action: s.actionCancelDialog},
			Binding{Keys: []keys.Key{keys.KeyCtrlC}, Symbol: "ctrl+c", Label: "quit", Action: s.actionQuit},
		),
	}
}

func (s *AddScreen) actionSubmit(actx *ActionCtx) app.Command {
	return s.app.RequestAddWorktree(strings.TrimSpace(s.path.Value()), strings.TrimSpace(s.branch.Value()))
}

func (s *AddScreen) actionClose(actx *ActionCtx) app.Command {
	s.app.CloseAdd()
	s.path.Clear()
	s.branch.Clear()
	s.focusedPath = true
	s.path.Focus()
	s.branch.Blur()
	return nil
}

func (s *AddScreen) actionQuit(actx *ActionCtx) app.Command {
	return s.app.Quit()
}

func (s *AddScreen) actionSwitchFocus(actx *ActionCtx) app.Command {
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

func (s *AddScreen) actionConfirmDialog(actx *ActionCtx) app.Command {
	return s.confirm.confirm(actx)
}

func (s *AddScreen) actionCancelDialog(actx *ActionCtx) app.Command {
	s.confirm.close()
	return nil
}

func (s *AddScreen) actionDialogMove(actx *ActionCtx) app.Command {
	s.confirm.move(actx.Key)
	return nil
}

func (s *AddScreen) Reset() {
	s.path.Clear()
	s.branch.Clear()
	s.focusedPath = true
	s.confirm.close()
	s.path.Focus()
	s.branch.Blur()
}
