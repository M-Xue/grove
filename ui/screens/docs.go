package screens

import (
	"strings"

	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/ui/keys"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DocsScreen struct {
	scroll   int
	handlers BindingSet
}

func NewDocsScreen() *DocsScreen {
	s := &DocsScreen{}
	s.handlers = BindingSet{
		keys.KeyEsc:      {Key: keys.KeyEsc, Handler: s.handleClose, Help: key.NewBinding(key.WithKeys("esc", "q", "?"), key.WithHelp("esc", "close"))},
		keys.KeyQ:        {Key: keys.KeyQ, Handler: s.handleClose, Help: key.NewBinding(key.WithKeys("esc", "q", "?"), key.WithHelp("esc", "close"))},
		keys.KeyQuestion: {Key: keys.KeyQuestion, Handler: s.handleClose, Help: key.NewBinding(key.WithKeys("esc", "q", "?"), key.WithHelp("esc", "close"))},
		keys.KeyCtrlC:    {Key: keys.KeyCtrlC, Handler: s.handleQuit, Help: key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit"))},
		keys.KeyUp:       {Key: keys.KeyUp, Handler: s.handleUp, Help: key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "scroll"))},
		keys.KeyK:        {Key: keys.KeyK, Handler: s.handleUp, Help: key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "scroll"))},
		keys.KeyDown:     {Key: keys.KeyDown, Handler: s.handleDown, Help: key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "scroll"))},
		keys.KeyJ:        {Key: keys.KeyJ, Handler: s.handleDown, Help: key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "scroll"))},
	}
	return s
}

func (s *DocsScreen) Sync(state app.State) {}

func (s *DocsScreen) Update(ctx *ScreenContext, msg tea.KeyMsg, state app.State) tea.Cmd {
	if handler, ok := s.handlers.HandlerFor(keys.Normalize(msg)); ok && handler != nil {
		return handler(ctx, msg)
	}
	return nil
}

func (s *DocsScreen) View(width, height int, state app.State) string {
	visible := screenMax(1, height-4)
	maxScroll := screenMax(0, len(state.DocsLines)-visible)
	if s.scroll > maxScroll {
		s.scroll = maxScroll
	}
	end := min(len(state.DocsLines), s.scroll+visible)
	lines := []string{"grove", "", lipgloss.NewStyle().Bold(true).Render("Docs"), ""}
	lines = append(lines, state.DocsLines[s.scroll:end]...)
	return strings.Join(lines, "\n")
}

func (s *DocsScreen) Footer(helpWidth int) string {
	model := NewHelpModel()
	model.Width = screenMax(helpWidth, 0)
	order := []keys.Key{keys.KeyEsc, keys.KeyUp, keys.KeyDown, keys.KeyCtrlC}
	return model.ShortHelpView(s.handlers.HelpBindings(order))
}

func (s *DocsScreen) handleClose(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	ctx.App.CloseDocs()
	s.scroll = 0
	return nil
}

func (s *DocsScreen) handleQuit(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	return ctx.Quit()
}

func (s *DocsScreen) handleUp(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	if s.scroll > 0 {
		s.scroll--
	}
	return nil
}

func (s *DocsScreen) handleDown(ctx *ScreenContext, msg tea.KeyMsg) tea.Cmd {
	s.scroll++
	return nil
}
