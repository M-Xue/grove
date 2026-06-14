package app

import "fmt"

func (a *App) setLoading(message string) {
	a.state.Loading = LoadingState{Active: true, Message: message}
}

func (a *App) clearLoading() {
	a.state.Loading = LoadingState{}
}

func (a *App) clearDialog() {
	a.state.Dialog = DialogState{}
}

func (a *App) appendStatus(kind StatusKind, message string) {
	a.statusCounter++
	a.state.Statuses = append(a.state.Statuses, StatusEntry{
		ID:      fmt.Sprintf("status-%d", a.statusCounter),
		Kind:    kind,
		Message: message,
	})
}
