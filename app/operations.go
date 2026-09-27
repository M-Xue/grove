package app

import "github.com/M-Xue/grove/worktree"

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

func (a *App) checkBranchExists(path, branch string) Command {
	id := a.setBlockingLoading("checking branch")
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
	message := "creating worktree"
	if createBranch {
		message = "creating branch and worktree"
	}
	id := a.setProgressLoading(message)
	worktrees := a.services.Worktree
	return func() Message {
		// The git work runs on its own goroutine so the thunk can return the
		// channel immediately; progress and the terminal result are delivered
		// through it. The buffer keeps a fast checkout from blocking on the UI.
		updates := make(chan Message, 16)
		go func() {
			err := worktrees.AddWithProgress(path, branch, createBranch, func(p worktree.Progress) {
				updates <- WorktreeProgressMessage{LoadingID: id, Done: p.Done, Total: p.Total}
			})
			updates <- WorktreeAddedMessage{LoadingID: id, Err: err}
			close(updates)
		}()
		return WorktreeAddStartedMessage{Updates: updates}
	}
}

func (a *App) removeWorktree(path string, force bool) Command {
	id := a.setBlockingLoading("removing worktree")
	worktrees := a.services.Worktree
	return func() Message {
		err := worktrees.Remove(path, force)
		return WorktreeRemovedMessage{LoadingID: id, Path: path, Err: err}
	}
}

func (a *App) pruneWorktrees() Command {
	id := a.setLoading("pruning stale worktrees")
	worktrees := a.services.Worktree
	return func() Message {
		err := worktrees.Prune()
		return WorktreesPrunedMessage{LoadingID: id, Err: err}
	}
}
