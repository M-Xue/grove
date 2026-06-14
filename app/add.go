package app

func (a *App) OpenAdd() {
	a.state.Screen = ScreenAdd
	if a.state.Dialog.Active {
		a.clearDialog()
	}
	if a.state.Loading.Active {
		a.clearLoading()
	}
}

func (a *App) CloseAdd() {
	a.state.Screen = ScreenChange
	a.clearDialog()
	a.clearLoading()
}

func (a *App) RequestAddWorktree(path, branch string) Effect {
	if path == "" {
		a.appendStatus(StatusInfo, "worktree path is required")
		return nil
	}
	if branch == "" {
		a.appendStatus(StatusInfo, "branch name is required")
		return nil
	}
	a.setLoading("checking branch")
	return CheckBranchExistsEffect{Path: path, Branch: branch}
}
