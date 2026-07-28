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

// setBlockingLoading appends an active loading entry whose operation freezes the
// UI until it completes. It otherwise behaves like setLoading, returning the new
// entry's ID.
func (a *App) setBlockingLoading(message string) string {
	a.loadingCounter++
	id := nextLoadingID(a.loadingCounter)
	a.state.Loading = append(a.state.Loading, LoadingEntry{
		ID:        id,
		Active:    true,
		Completed: false,
		Message:   message,
		Blocking:  true,
	})
	return id
}

// setProgressLoading appends a loading entry that renders a checkout progress
// bar, starting at 0%. Worktree creation is inherently blocking, so it also
// freezes the UI. It otherwise behaves like setLoading, returning the new
// entry's ID.
func (a *App) setProgressLoading(message string) string {
	a.loadingCounter++
	id := nextLoadingID(a.loadingCounter)
	a.state.Loading = append(a.state.Loading, LoadingEntry{
		ID:        id,
		Active:    true,
		Completed: false,
		Message:   message,
		Progress:  true,
		Blocking:  true,
	})
	return id
}

// IsBusy reports whether any blocking operation is still in flight. The UI reads
// this to freeze input. Completed entries (awaiting dismissal) do not count, so
// the UI unfreezes the instant an operation finishes.
func (s State) IsBusy() bool {
	for _, entry := range s.Loading {
		if entry.Blocking && !entry.Completed {
			return true
		}
	}
	return false
}

// updateLoadingProgress records checkout progress against the loading entry
// with the given ID, leaving other entries untouched.
func (a *App) updateLoadingProgress(id string, done, total int) {
	for i := range a.state.Loading {
		if a.state.Loading[i].ID == id {
			a.state.Loading[i].Done = done
			a.state.Loading[i].Total = total
			return
		}
	}
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

func nextLoadingID(counter int) string {
	return fmt.Sprintf("loading-%d", counter)
}
