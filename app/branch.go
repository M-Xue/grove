package app

import "strings"

const branchCommitPreviewLimit = 10

func (a *App) OpenBranch() Command {
	a.state.Screen = ScreenBranch
	a.clearDialog()
	if len(a.state.Loading) > 0 {
		a.clearLoading()
	}
	return a.loadBranches()
}

func (a *App) CloseBranch() Command {
	a.state.Screen = ScreenChange
	a.clearDialog()
	a.clearLoading()
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

func (a *App) RequestDeleteAllBranches() Command {
	if a.state.BranchScope != "local" {
		a.appendStatus(StatusInfo, "switch to local branches before deleting")
		return nil
	}
	if len(a.state.Branches) == 0 {
		a.appendStatus(StatusInfo, "no local branches available")
		return nil
	}
	branchNames := make([]string, 0, len(a.state.Branches))
	for _, br := range a.state.Branches {
		if br.Name == "" {
			continue
		}
		branchNames = append(branchNames, br.Name)
	}
	description := "Delete all local branches except ones currently checked out here or in another worktree?"
	if len(branchNames) > 0 {
		description += "\n\n" + strings.Join(branchNames, "\n")
	}
	a.state.Dialog = DialogState{
		Active:      true,
		Title:       "Delete all local branches?",
		Description: description,
		Buttons:     []DialogButton{{ID: "confirm", Label: "Delete"}, {ID: "cancel", Label: "Cancel"}},
		FocusedID:   "cancel",
		Kind:        DialogConfirmDeleteAllBranches,
	}
	return nil
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

func (a *App) RequestDeleteBranch(name string) Command {
	if name == "" {
		a.appendStatus(StatusInfo, "no branch selected")
		return nil
	}
	a.state.Dialog = DialogState{
		Active:      true,
		Title:       "Delete branch?",
		Description: name,
		Buttons:     []DialogButton{{ID: "confirm", Label: "Delete"}, {ID: "cancel", Label: "Cancel"}},
		FocusedID:   "cancel",
		Kind:        DialogConfirmDeleteBranch,
		Branch:      name,
	}
	return nil
}

func (a *App) RequestFetchBranches() Command {
	return a.fetchBranches()
}
