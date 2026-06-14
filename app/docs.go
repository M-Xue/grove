package app

import (
	branchsvc "github.com/M-Xue/grove/branch"
	"fmt"
	"strings"
)

func (a *App) OpenDocs() Effect {
	a.state.Screen = ScreenDocs
	a.setLoading("loading docs")
	return LoadDocsEffect{}
}

func (a *App) CloseDocs() {
	a.state.Screen = ScreenChange
	a.clearDialog()
	a.clearLoading()
}

func (a *App) DialogChoose(buttonID string) Effect {
	if !a.state.Dialog.Active {
		return nil
	}
	if buttonID == "cancel" {
		a.clearDialog()
		return nil
	}
	switch a.state.Dialog.Kind {
	case DialogConfirmRemove:
		path := a.state.Dialog.Path
		a.clearDialog()
		a.setLoading("removing worktree")
		return RemoveWorktreeEffect{Path: path}
	case DialogConfirmCreateBranch:
		path := a.state.Dialog.Path
		branch := a.state.Dialog.Branch
		a.clearDialog()
		a.setLoading("creating branch and worktree")
		return AddWorktreeEffect{Path: path, Branch: branch, CreateBranch: true}
	case DialogConfirmDeleteBranch:
		branch := a.state.Dialog.Branch
		a.clearDialog()
		a.setLoading("deleting branch")
		return DeleteBranchEffect{Name: branch}
	case DialogConfirmDeleteAllBranches:
		a.clearDialog()
		a.setLoading("deleting local branches")
		return DeleteAllBranchesEffect{}
	default:
		return nil
	}
}

func (a *App) HandleResult(result Result) Effect {
	switch msg := result.(type) {
	case WorktreesLoadedResult:
		if msg.Err != nil {
			a.clearLoading()
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.state.Worktrees = msg.Worktrees
		a.state.SubmittedPath = ""
		a.markLoadingDone()
		return nil
	case BranchesLoadedResult:
		if msg.Err != nil {
			a.clearLoading()
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.state.Branches = msg.Branches
		a.state.BranchScope = msg.Scope
		a.markLoadingDone()
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
		a.setLoading("loading branch commits")
		return LoadBranchCommitsEffect{Name: selected, Limit: branchCommitPreviewLimit}
	case BranchCommitsLoadedResult:
		if msg.Err != nil {
			a.clearLoading()
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.state.Branch.SelectedName = msg.Name
		a.state.Branch.Commits = msg.Commits
		a.markLoadingDone()
		return nil
	case BranchCheckedOutResult:
		if msg.Err != nil {
			a.clearLoading()
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.markLoadingDone()
		a.appendStatus(StatusSuccess, "branch switched")
		a.setLoading("loading branches")
		return LoadBranchesEffect{}
	case BranchDeletedResult:
		if msg.Err != nil {
			a.clearLoading()
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.markLoadingDone()
		a.appendStatus(StatusSuccess, "branch deleted")
		a.setLoading("loading branches")
		return LoadBranchesEffect{}
	case AllBranchesDeletedResult:
		if msg.Err != nil {
			a.clearLoading()
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.markLoadingDone()
		if len(msg.Deleted) == 0 {
			a.appendStatus(StatusInfo, "no local branches deleted")
		} else {
			a.appendStatus(StatusSuccess, fmt.Sprintf("deleted %d local branches", len(msg.Deleted)))
		}
		if len(msg.Skipped) > 0 {
			a.appendStatus(StatusInfo, fmt.Sprintf("skipped checked out branches: %s", strings.Join(msg.Skipped, ", ")))
		}
		a.setLoading("loading branches")
		return LoadBranchesEffect{}
	case BranchesFetchedResult:
		if msg.Err != nil {
			a.clearLoading()
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.markLoadingDone()
		a.appendStatus(StatusSuccess, "fetch complete")
		a.setLoading("loading branches")
		return LoadBranchesEffect{}
	case BranchCheckedResult:
		if msg.Err != nil {
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.markLoadingDone()
		if msg.Exists {
			a.setLoading("adding worktree")
			return AddWorktreeEffect{Path: msg.Path, Branch: msg.Branch, CreateBranch: false}
		}
		a.state.Dialog = DialogState{
			Active:      true,
			Title:       "Branch does not exist",
			Description: fmt.Sprintf("Create a new branch named %q?", msg.Branch),
			Buttons:     []DialogButton{{ID: "confirm", Label: "Create"}, {ID: "cancel", Label: "Cancel"}},
			FocusedID:   "confirm",
			Kind:        DialogConfirmCreateBranch,
			Path:        msg.Path,
			Branch:      msg.Branch,
		}
		return nil
	case WorktreeAddedResult:
		if msg.Err != nil {
			a.clearLoading()
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.markLoadingDone()
		a.state.Screen = ScreenChange
		a.appendStatus(StatusSuccess, "worktree added")
		a.setLoading("loading worktrees")
		return LoadWorktreesEffect{}
	case WorktreeRemovedResult:
		if msg.Err != nil {
			a.clearLoading()
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.markLoadingDone()
		a.state.Screen = ScreenChange
		a.appendStatus(StatusSuccess, "worktree removed")
		a.setLoading("loading worktrees")
		return LoadWorktreesEffect{}
	case DocsLoadedResult:
		if msg.Err != nil {
			a.clearLoading()
			a.state.Screen = ScreenChange
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.state.DocsLines = msg.Lines
		a.markLoadingDone()
		return nil
	default:
		return nil
	}
}

func hasBranch(branches []branchsvc.Info, name string) bool {
	for _, branch := range branches {
		if branch.Name == name {
			return true
		}
	}
	return false
}
