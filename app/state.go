package app

import "fmt"

func (a *App) setLoading(message string) {
	a.loadingCounter++
	a.state.Loading = append(a.state.Loading, LoadingEntry{
		ID:        nextLoadingID(a.loadingCounter),
		Active:    true,
		Completed: false,
		Message:   message,
	})
}

func (a *App) markLoadingDone() {
	if len(a.state.Loading) == 0 {
		return
	}
	for i := len(a.state.Loading) - 1; i >= 0; i-- {
		if a.state.Loading[i].Active && !a.state.Loading[i].Completed {
			a.state.Loading[i].Completed = true
			return
		}
	}
}

func (a *App) clearLoading() {
	a.state.Loading = nil
}

func (a *App) clearDialog() {
	a.state.Dialog = DialogState{}
}

func nextLoadingID(counter int) string {
	return fmt.Sprintf("loading-%d", counter)
}
