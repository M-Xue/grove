package app

import (
	"fmt"
	"strings"

	"github.com/M-Xue/grove/branch"
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
		if a.state.Screen == ScreenBranch && len(a.state.Branches) == 0 {
			return a.loadBranches()
		}
		return nil
	case BranchesLoadedMessage:
		if msg.Err != nil {
			a.clearLoadingEntry(msg.LoadingID)
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.state.Branches = msg.Branches
		a.state.BranchScope = msg.Scope
		a.markLoadingDone(msg.LoadingID)
		if len(msg.Branches) == 0 {
			a.state.Branch.SelectedName = ""
			a.state.Branch.Commits = nil
			return nil
		}
		selected := a.state.Branch.SelectedName
		if selected == "" || !hasBranch(msg.Branches, selected) {
			selected = msg.Branches[0].Name
		}
		a.state.Branch.SelectedName = selected
		return a.loadBranchCommits(selected)
	case BranchCommitsLoadedMessage:
		if msg.Seq != a.branchCommitSeq {
			// A newer selection superseded this request; drop the stale
			// result and remove only its loading entry.
			a.clearLoadingEntry(msg.LoadingID)
			return nil
		}
		if msg.Err != nil {
			a.clearLoadingEntry(msg.LoadingID)
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.state.Branch.SelectedName = msg.Name
		a.state.Branch.Commits = msg.Commits
		a.markLoadingDone(msg.LoadingID)
		return nil
	case BranchCheckedOutMessage:
		if msg.Err != nil {
			a.clearLoadingEntry(msg.LoadingID)
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.markLoadingDone(msg.LoadingID)
		a.appendStatus(StatusSuccess, "branch switched")
		return a.loadBranches()
	case BranchDeletedMessage:
		if msg.Err != nil {
			a.clearLoadingEntry(msg.LoadingID)
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.markLoadingDone(msg.LoadingID)
		a.appendStatus(StatusSuccess, "branch deleted")
		return a.loadBranches()
	case AllBranchesDeletedMessage:
		if msg.Err != nil {
			a.clearLoadingEntry(msg.LoadingID)
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.markLoadingDone(msg.LoadingID)
		if len(msg.Deleted) == 0 {
			a.appendStatus(StatusInfo, "no local branches deleted")
		} else {
			a.appendStatus(StatusSuccess, fmt.Sprintf("deleted %d local branches", len(msg.Deleted)))
		}
		if len(msg.Skipped) > 0 {
			a.appendStatus(StatusInfo, fmt.Sprintf("skipped checked out branches: %s", strings.Join(msg.Skipped, ", ")))
		}
		return a.loadBranches()
	case BranchesFetchedMessage:
		if msg.Err != nil {
			a.clearLoadingEntry(msg.LoadingID)
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.markLoadingDone(msg.LoadingID)
		a.appendStatus(StatusSuccess, "fetch complete")
		return a.loadBranches()
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
	case WorktreeAddedMessage:
		if msg.Err != nil {
			a.clearLoadingEntry(msg.LoadingID)
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.markLoadingDone(msg.LoadingID)
		a.state.Screen = ScreenChange
		a.appendStatus(StatusSuccess, "worktree added")
		return a.loadWorktrees()
	case WorktreeRemovedMessage:
		if msg.Err != nil {
			a.clearLoadingEntry(msg.LoadingID)
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.markLoadingDone(msg.LoadingID)
		a.state.Screen = ScreenChange
		a.appendStatus(StatusSuccess, "worktree removed")
		return a.loadWorktrees()
	case WorktreePrunedMessage:
		if msg.Err != nil {
			a.clearLoadingEntry(msg.LoadingID)
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.markLoadingDone(msg.LoadingID)
		a.state.Screen = ScreenChange
		a.appendStatus(StatusSuccess, "worktree pruned")
		return a.loadWorktrees()
	default:
		return nil
	}
}

func hasBranch(branches []branch.Info, name string) bool {
	for _, br := range branches {
		if br.Name == name {
			return true
		}
	}
	return false
}
