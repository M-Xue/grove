package app

// This file defines the async operations grove can perform. Each helper runs
// on the main loop: it synchronously sets a loading entry (capturing its ID)
// and returns a Command whose thunk performs the git work off the main loop,
// reading only the args and services it closes over. The returned Message
// carries the loading ID back so HandleMessage can resolve exactly that entry.

func (a *App) loadWorktrees() Command {
	id := a.setLoading("loading worktrees")
	worktrees := a.services.Worktree
	return func() Message {
		list, err := worktrees.List()
		return WorktreesLoadedMessage{LoadingID: id, Worktrees: list, Err: err}
	}
}

func (a *App) loadBranches() Command {
	id := a.setLoading("loading branches")
	branches := a.services.Branch
	return func() Message {
		list, scope, err := branches.List()
		return BranchesLoadedMessage{LoadingID: id, Branches: list, Scope: scope, Err: err}
	}
}

func (a *App) toggleBranchScope() Command {
	id := a.setLoading("loading branches")
	branches := a.services.Branch
	return func() Message {
		branches.ToggleScope()
		list, scope, err := branches.List()
		return BranchesLoadedMessage{LoadingID: id, Branches: list, Scope: scope, Err: err}
	}
}

func (a *App) loadBranchCommits(name string) Command {
	a.branchCommitSeq++
	seq := a.branchCommitSeq
	id := a.setLoading("loading branch commits")
	branches := a.services.Branch
	limit := branchCommitPreviewLimit
	return func() Message {
		commits, err := branches.RecentCommits(name, limit)
		return BranchCommitsLoadedMessage{LoadingID: id, Seq: seq, Name: name, Commits: commits, Err: err}
	}
}

func (a *App) checkoutBranch(name string) Command {
	id := a.setLoading("switching branch")
	branches := a.services.Branch
	return func() Message {
		err := branches.Checkout(name)
		return BranchCheckedOutMessage{LoadingID: id, Err: err}
	}
}

func (a *App) deleteBranch(name string) Command {
	id := a.setLoading("deleting branch")
	branches := a.services.Branch
	return func() Message {
		err := branches.Delete(name)
		return BranchDeletedMessage{LoadingID: id, Err: err}
	}
}

func (a *App) deleteAllBranches() Command {
	id := a.setLoading("deleting local branches")
	branches := a.services.Branch
	return func() Message {
		summary, err := branches.DeleteAllLocal()
		return AllBranchesDeletedMessage{LoadingID: id, Deleted: summary.Deleted, Skipped: summary.Skipped, Err: err}
	}
}

func (a *App) fetchBranches() Command {
	id := a.setLoading("fetching branches")
	branches := a.services.Branch
	return func() Message {
		err := branches.Fetch()
		return BranchesFetchedMessage{LoadingID: id, Err: err}
	}
}

func (a *App) checkBranchExists(path, branch string) Command {
	id := a.setLoading("checking branch")
	worktrees := a.services.Worktree
	return func() Message {
		exists, err := worktrees.BranchExists(branch)
		if err != nil {
			return BranchCheckFailedMessage{LoadingID: id, Err: err}
		}
		if exists {
			return BranchExistsMessage{LoadingID: id, Path: path, Branch: branch}
		}
		return BranchAbsentMessage{LoadingID: id, Path: path, Branch: branch}
	}
}

func (a *App) addWorktree(path, branch string, createBranch bool) Command {
	message := "adding worktree"
	if createBranch {
		message = "creating branch and worktree"
	}
	id := a.setLoading(message)
	worktrees := a.services.Worktree
	return func() Message {
		var err error
		if createBranch {
			err = worktrees.AddWithNewBranch(path, branch)
		} else {
			err = worktrees.Add(path, branch)
		}
		return WorktreeAddedMessage{LoadingID: id, Err: err}
	}
}

func (a *App) removeWorktree(path string) Command {
	id := a.setLoading("removing worktree")
	worktrees := a.services.Worktree
	return func() Message {
		err := worktrees.Remove(path)
		return WorktreeRemovedMessage{LoadingID: id, Path: path, Err: err}
	}
}

func (a *App) pruneWorktree(path string) Command {
	id := a.setLoading("pruning worktree")
	worktrees := a.services.Worktree
	return func() Message {
		err := worktrees.Prune(path)
		return WorktreePrunedMessage{LoadingID: id, Path: path, Err: err}
	}
}
