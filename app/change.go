package app

func (a *App) RequestSubmitSelectedPath(path string) Command {
	if path == "" {
		a.appendStatus(StatusInfo, "no worktree selected")
		return nil
	}
	a.state.SubmittedPath = path
	return func() Message { return QuitRequested{} }
}

// RemoveWorktree removes the worktree at path. An empty path reports "no
// worktree selected" and does nothing.
func (a *App) RemoveWorktree(path string) Command {
	if path == "" {
		a.appendStatus(StatusInfo, "no worktree selected")
		return nil
	}
	return a.removeWorktree(path)
}
