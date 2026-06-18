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

// PruneWorktree force-removes the stale worktree at path. An empty path reports
// "no worktree selected"; a path that is not stale reports "worktree is not
// stale". Either case does nothing, guarding a healthy worktree from a forced
// removal.
func (a *App) PruneWorktree(path string) Command {
	if path == "" {
		a.appendStatus(StatusInfo, "no worktree selected")
		return nil
	}
	if !a.isStaleWorktree(path) {
		a.appendStatus(StatusInfo, "worktree is not stale")
		return nil
	}
	return a.pruneWorktree(path)
}

func (a *App) isStaleWorktree(path string) bool {
	for _, worktree := range a.state.Worktrees {
		if worktree.Path == path {
			return worktree.Stale
		}
	}
	return false
}
