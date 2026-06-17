package app

func (a *App) OpenAdd() {
	a.state.Screen = ScreenAdd
	if a.state.Dialog.Active {
		a.clearDialog()
	}
	if len(a.state.Loading) > 0 {
		a.ClearLoading()
	}
}

func (a *App) CloseAdd() {
	a.state.Screen = ScreenChange
	a.clearDialog()
	a.ClearLoading()
}

func (a *App) RequestAddWorktree(path, branch string) Command {
	if path == "" {
		a.appendStatus(StatusInfo, "worktree path is required")
		return nil
	}
	if branch == "" {
		a.appendStatus(StatusInfo, "branch name is required")
		return nil
	}
	return a.checkBranchExists(path, branch)
}
