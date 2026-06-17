package app

import "strings"

const branchCommitPreviewLimit = 10

func (a *App) OpenBranch() Effect {
	a.state.Screen = ScreenBranch
	a.clearDialog()
	if len(a.state.Loading) > 0 {
		a.clearLoading()
	}
	a.setLoading("loading branches")
	return LoadBranchesEffect{}
}

func (a *App) CloseBranch() Effect {
	a.state.Screen = ScreenChange
	a.clearDialog()
	a.clearLoading()
	a.setLoading("loading worktrees")
	return LoadWorktreesEffect{}
}

func (a *App) RequestCheckoutBranch(name string) Effect {
	if name == "" {
		a.appendStatus(StatusInfo, "no branch selected")
		return nil
	}
	a.setLoading("switching branch")
	return CheckoutBranchEffect{Name: name}
}

func (a *App) RequestToggleBranchScope() Effect {
	a.setLoading("loading branches")
	return ToggleBranchScopeEffect{}
}

func (a *App) RequestDeleteAllBranches() Effect {
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

func (a *App) SelectBranch(name string) Effect {
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
	a.setLoading("loading branch commits")
	return LoadBranchCommitsEffect{Name: name, Limit: branchCommitPreviewLimit}
}

func (a *App) RequestDeleteBranch(name string) Effect {
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

func (a *App) RequestFetchBranches() Effect {
	a.setLoading("fetching branches")
	return FetchBranchesEffect{}
}
