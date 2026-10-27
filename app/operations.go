package app

import (
	"time"

	"github.com/M-Xue/grove/worktree"
)

// This file defines the async operations grove can perform. Each helper runs
// on the main loop: it synchronously sets a loading entry (capturing its ID)
// and returns a Command whose thunk performs the git work off the main loop,
// reading only the args and services it closes over. The returned Message
// carries the loading ID back so HandleMessage can resolve exactly that entry.

func (a *App) loadWorktrees() Command {
	return a.startWorktreeLoad(a.setLoading("loading worktrees"), false)
}

// startWorktreeLoad issues a worktree-list load under the given loading entry
// (empty for a silent background refresh). background selects the worktree
// service's lock-free listing, which silent refreshes use so they never hold
// a worktree's index.lock against the user (see worktree.ListOptions). Each
// load takes the next sequence number; a still-pending earlier load is
// superseded — its loading entry is dropped now and its result discarded when
// it lands — so the list on screen and in the disk cache always reflects the
// most recently issued load. The load is timed so the UI can pace background
// refreshes to what a listing actually costs in this repository.
func (a *App) startWorktreeLoad(loadingID string, background bool) Command {
	if a.worktreeLoadPending && a.worktreeLoadID != "" && a.worktreeLoadID != loadingID {
		a.clearLoadingEntry(a.worktreeLoadID)
	}
	a.worktreeLoadSeq++
	a.worktreeLoadPending = true
	a.worktreeLoadID = loadingID
	seq := a.worktreeLoadSeq
	worktrees := a.services.Worktree
	return func() Message {
		started := time.Now()
		list, err := worktrees.List(worktree.ListOptions{Background: background})
		return WorktreesLoadedMessage{
			LoadingID: loadingID,
			Seq:       seq,
			Worktrees: list,
			Err:       err,
			Elapsed:   time.Since(started),
		}
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
