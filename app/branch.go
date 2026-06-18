package app

import (
	"strings"

	"github.com/M-Xue/grove/branch"
)

const branchCommitPreviewLimit = 10

func (a *App) OpenBranch() Command {
	a.state.Screen = ScreenBranch
	if len(a.state.Loading) > 0 {
		a.ClearLoading()
	}
	return a.loadBranches()
}

func (a *App) CloseBranch() Command {
	a.state.Screen = ScreenChange
	a.ClearLoading()
	return a.loadWorktrees()
}

func (a *App) RequestCheckoutBranch(name string) Command {
	if name == "" {
		a.appendStatus(StatusInfo, "no branch selected")
		return nil
	}
	return a.checkoutBranch(name)
}

func (a *App) RequestToggleBranchScope() Command {
	return a.toggleBranchScope()
}

func (a *App) SelectBranch(name string) Command {
	name = strings.TrimSpace(name)
	if name == "" {
		a.state.Branch.SelectedName = ""
		a.state.Branch.Commits = nil
		return nil
	}
	if a.state.Branch.SelectedName == name && len(a.state.Branch.Commits) > 0 {
		return nil
	}
	a.state.Branch.SelectedName = name
	return a.loadBranchCommits(name)
}

// DeleteBranch deletes the named branch. An empty name reports "no branch
// selected" and does nothing.
func (a *App) DeleteBranch(name string) Command {
	if name == "" {
		a.appendStatus(StatusInfo, "no branch selected")
		return nil
	}
	return a.deleteBranch(name)
}

// CanDeleteAllBranches reports whether deleting all local branches is currently
// possible, appending an explanatory status and returning false if not. The UI
// uses it to decide whether to open its confirmation dialog.
func (a *App) CanDeleteAllBranches() bool {
	if a.state.BranchScope != branch.ScopeLocal {
		a.appendStatus(StatusInfo, "switch to local branches before deleting")
		return false
	}
	if len(a.state.Branches) == 0 {
		a.appendStatus(StatusInfo, "no local branches available")
		return false
	}
	return true
}

// DeleteAllBranches deletes every local branch not currently checked out here
// or in another worktree.
func (a *App) DeleteAllBranches() Command {
	return a.deleteAllBranches()
}

func (a *App) RequestFetchBranches() Command {
	return a.fetchBranches()
}
