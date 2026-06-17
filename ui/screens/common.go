package screens

import (
	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/ui/keys"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type ScreenContext struct {
	App  *app.App
	Run  func(app.Command) tea.Cmd
	Quit func() tea.Cmd
}

type Handler func(*ScreenContext, tea.KeyMsg) tea.Cmd
type KeyHandlers map[keys.Key]Handler

type Binding struct {
	Key     keys.Key
	Handler Handler
	Help    key.Binding
}

type BindingSet map[keys.Key]Binding

func (b BindingSet) HandlerFor(k keys.Key) (Handler, bool) {
	binding, ok := b[k]
	if !ok {
		return nil, false
	}
	return binding.Handler, true
}

func (b BindingSet) HelpBindings(order []keys.Key) []key.Binding {
	bindings := make([]key.Binding, 0, len(order))
	for _, item := range order {
		binding, ok := b[item]
		if !ok {
			continue
		}
		bindings = append(bindings, binding.Help)
	}
	return bindings
}

func NewHelpModel() help.Model {
	model := help.New()
	model.ShowAll = false
	return model
}
