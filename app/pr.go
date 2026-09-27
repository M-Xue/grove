package app

// LoadBranchPR fetches the pull request for the named branch. Like the commits
// fetch it is hover-driven and creates no loading entry, but its caching is
// stricter because each lookup is a network round-trip against a rate-limited
// API: a branch is fetched at most once per session (hit or "no PR" alike),
// and nothing is fetched once a lookup has reported the environment can't do
// PR lookups at all. Transient failures are not cached, so the next hover of
// that branch retries.
func (a *App) LoadBranchPR(name string) Command {
	if name == "" || a.services.PR == nil || a.state.PRLookupUnavailable {
		return nil
	}
	if entry, cached := a.state.BranchPRs[name]; cached && !entry.Failed {
		return nil
	}
	prs := a.services.PR
	return func() Message {
		info, found, err := prs.ForBranch(name)
		return BranchPRLoadedMessage{Branch: name, Info: info, Found: found, Err: err}
	}
}
