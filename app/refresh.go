package app

// RefreshWorktrees reloads the worktree list on the user's request (the
// refresh key). It shows a loading entry like the initial load so the keypress
// has visible feedback, and supersedes any load already in flight. Like the
// initial load it is a foreground listing, so git may refresh the index on
// disk and make the polls that follow cheap.
func (a *App) RefreshWorktrees() Command {
	return a.startWorktreeLoad(a.setLoading("refreshing worktrees"), false)
}

// RefreshWorktreesInBackground reloads the worktree list to pick up changes
// made outside grove (another terminal switching a worktree's branch, adding
// or removing worktrees). It is issued on terminal focus and on a periodic
// tick, so it is silent — no loading entry — and polite: it does nothing while
// a load is already in flight or a blocking operation is running, since that
// operation reloads the list itself when it completes, and it lists without
// taking index locks so it can never make the user's own git command fail.
func (a *App) RefreshWorktreesInBackground() Command {
	if a.worktreeLoadPending || a.state.IsBusy() {
		return nil
	}
	return a.startWorktreeLoad("", true)
}
