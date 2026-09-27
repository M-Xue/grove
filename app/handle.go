package app

import (
	"errors"

	"github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/pr"
)

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
		if a.state.BranchCommits == nil {
			a.state.BranchCommits = make(map[string][]branch.CommitInfo)
		}
		if msg.Err != nil {
			// Hover-driven lookup: a failure (e.g. the branch was just
			// deleted) is not worth a status-line entry. An empty result is
			// recorded when nothing is loaded yet so the details panel’s
			// spinner resolves; a failed refresh keeps the old commits.
			if _, loaded := a.state.BranchCommits[msg.Branch]; !loaded {
				a.state.BranchCommits[msg.Branch] = nil
			}
		} else {
			a.state.BranchCommits[msg.Branch] = msg.Commits
		}
		// Chain the branch’s PR lookup off its commits load: every hover that
		// fetches commits thereby fetches the PR too, without the UI issuing a
		// second command. LoadBranchPR’s own caching keeps this from repeating.
		return a.LoadBranchPR(msg.Branch)
	case BranchPRLoadedMessage:
		if a.state.BranchPRs == nil {
			a.state.BranchPRs = make(map[string]BranchPR)
		}
		if msg.Err != nil {
			// An environmental failure (gh missing, unauthenticated, no GitHub
			// remote) latches the kill switch so grove stops asking. A
			// transient failure is recorded as a failed entry — the details
			// panel resolves its spinner and says so — but stays uncached in
			// LoadBranchPR’s eyes, so a later hover of the branch retries.
			if errors.Is(msg.Err, pr.ErrUnavailable) {
				a.state.PRLookupUnavailable = true
				return nil
			}
			a.state.BranchPRs[msg.Branch] = BranchPR{Failed: true}
			return nil
		}
		a.state.BranchPRs[msg.Branch] = BranchPR{Found: msg.Found, Info: msg.Info}
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
