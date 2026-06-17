package app

import "fmt"

// setLoading appends a new active loading entry and returns its ID so the
// completing Command can later mark or clear exactly that entry.
func (a *App) setLoading(message string) string {
	a.loadingCounter++
	id := nextLoadingID(a.loadingCounter)
	a.state.Loading = append(a.state.Loading, LoadingEntry{
		ID:        id,
		Active:    true,
		Completed: false,
		Message:   message,
	})
	return id
}

// markLoadingDone marks the loading entry with the given ID as completed,
// leaving any other (still-pending) entries untouched.
func (a *App) markLoadingDone(id string) {
	for i := range a.state.Loading {
		if a.state.Loading[i].ID == id {
			a.state.Loading[i].Completed = true
			return
		}
	}
}

// clearLoadingEntry removes the loading entry with the given ID, leaving any
// other entries untouched. Used when a single task fails or is superseded.
func (a *App) clearLoadingEntry(id string) {
	filtered := a.state.Loading[:0]
	for _, entry := range a.state.Loading {
		if entry.ID == id {
			continue
		}
		filtered = append(filtered, entry)
	}
	a.state.Loading = filtered
}

// ClearLoading removes every loading entry. Used on screen transitions.
func (a *App) ClearLoading() {
	a.state.Loading = nil
}

func (a *App) clearDialog() {
	a.state.Dialog = DialogState{}
}

func nextLoadingID(counter int) string {
	return fmt.Sprintf("loading-%d", counter)
}
