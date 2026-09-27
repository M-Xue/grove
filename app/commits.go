package app

// RecentCommitLimit is how many commits are fetched per branch for the
// details panel; the panel shows a smaller scrollable window of them.
const RecentCommitLimit = 20

// LoadBranchCommits fetches the recent commits of the named branch for the
// details panel. It is hover-driven and fires on every selection change, so it
// deliberately creates no loading entry: a spinner flashing on each keystroke
// would be noise for a lookup this cheap. Returns nil when there is nothing to
// fetch (blank branch name, e.g. a detached-HEAD worktree, or no service).
func (a *App) LoadBranchCommits(name string) Command {
	if name == "" || a.services.Branch == nil {
		return nil
	}
	branches := a.services.Branch
	return func() Message {
		commits, err := branches.RecentCommits(name, RecentCommitLimit)
		return BranchCommitsLoadedMessage{Branch: name, Commits: commits, Err: err}
	}
}
