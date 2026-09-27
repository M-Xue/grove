package app

import "github.com/M-Xue/grove/branch"

// HandleMessage applies a completed Command's Message to state and may return
// the next Command to chain. It is an inspectable switch so app tests can drive
// it directly.
func (a *App) HandleMessage(message Message) Command {
	switch msg := message.(type) {
	case WorktreesLoadedMessage:
		if msg.Err != nil {
			a.clearLoadingEntry(msg.LoadingID)
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.state.Worktrees = msg.Worktrees
		a.state.SubmittedPath = ""
		a.markLoadingDone(msg.LoadingID)
		// Persist the fresh list for the next launch's instant paint. This is
		// the single choke point every add/remove/prune funnels through, so
		// the cache stays consistent with what the screen shows.
		if a.saveWorktrees != nil {
			a.saveWorktrees(msg.Worktrees)
		}
		return nil
	case BranchExistsMessage:
		a.markLoadingDone(msg.LoadingID)
		return a.addWorktree(msg.Path, msg.Branch, false)
	case BranchAbsentMessage:
		// The branch is missing; the active screen reacts (via OnMessage) by
		// offering to create it. No state change beyond resolving the check.
		a.markLoadingDone(msg.LoadingID)
		return nil
	case BranchCheckFailedMessage:
		a.clearLoadingEntry(msg.LoadingID)
		a.appendStatus(StatusError, msg.Err.Error())
		return nil
	case BranchCommitsLoadedMessage:
		// Hover-driven lookup: a failure (e.g. the branch was just deleted)
		// simply leaves the panel without commits rather than spamming the
		// status line on every selection move.
		if msg.Err != nil {
			return nil
		}
		if a.state.BranchCommits == nil {
			a.state.BranchCommits = make(map[string][]branch.CommitInfo)
		}
		a.state.BranchCommits[msg.Branch] = msg.Commits
		return nil
	case WorktreeProgressMessage:
		a.updateLoadingProgress(msg.LoadingID, msg.Done, msg.Total)
		return nil
	case WorktreeAddedMessage:
		if msg.Err != nil {
			a.clearLoadingEntry(msg.LoadingID)
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.markLoadingDone(msg.LoadingID)
		a.appendStatus(StatusSuccess, "worktree added")
		return a.loadWorktrees()
	case WorktreeRemovedMessage:
		if msg.Err != nil {
			a.clearLoadingEntry(msg.LoadingID)
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.markLoadingDone(msg.LoadingID)
		a.appendStatus(StatusSuccess, "worktree removed")
		return a.loadWorktrees()
	case WorktreesPrunedMessage:
		if msg.Err != nil {
			a.clearLoadingEntry(msg.LoadingID)
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.markLoadingDone(msg.LoadingID)
		a.appendStatus(StatusSuccess, "stale worktrees pruned")
		return a.loadWorktrees()
	default:
		return nil
	}
}
