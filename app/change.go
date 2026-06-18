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

// PruneWorktrees removes every stale worktree via git's global prune. When no
// worktree is stale it reports "no stale worktrees" and does nothing.
func (a *App) PruneWorktrees() Command {
	if !a.hasStaleWorktree() {
		a.appendStatus(StatusInfo, "no stale worktrees")
		return nil
	}
	return a.pruneWorktrees()
}

func (a *App) hasStaleWorktree() bool {
	for _, worktree := range a.state.Worktrees {
		if worktree.Stale {
			return true
		}
	}
	return false
}
