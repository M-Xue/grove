package screens

import (
	"strings"

	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/ui/keys"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
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
		keys.KeyUp:       {Key: keys.KeyUp, Handler: s.handleUp, Help: key.NewBinding(key.WithKeys("up", "shift+tab"), key.WithHelp("up", "scroll"))},
		keys.KeyDown:     {Key: keys.KeyDown, Handler: s.handleDown, Help: key.NewBinding(key.WithKeys("down", "tab"), key.WithHelp("down", "scroll"))},
		keys.KeyTab:      {Key: keys.KeyTab, Handler: s.handleDown, Help: key.NewBinding(key.WithKeys("down", "tab"), key.WithHelp("down", "scroll"))},
		keys.KeyShiftTab: {Key: keys.KeyShiftTab, Handler: s.handleUp, Help: key.NewBinding(key.WithKeys("up", "shift+tab"), key.WithHelp("up", "scroll"))},
	}
	return s
}

func (s *DocsScreen) Sync(state app.State) {}

func (s *DocsScreen) Update(ctx *ScreenContext, msg tea.KeyMsg, state app.State) tea.Cmd {
	if handler, ok := s.handlers.HandlerFor(keys.Normalize(msg)); ok {
		return handler(ctx, msg)
	}
	return nil
}

func (s *DocsScreen) View(width, height int, footer string, state app.State) string {
	visible := screenMax(1, height-6)
	maxScroll := screenMax(0, len(state.DocsLines)-visible)
	if s.scroll > maxScroll {
		s.scroll = maxScroll
	}
	end := min(len(state.DocsLines), s.scroll+visible)
	lines := []string{"grove", "", "Docs", ""}
	lines = append(lines, state.DocsLines[s.scroll:end]...)
	if footer != "" {
		lines = append(lines, "", footer)
	}
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
