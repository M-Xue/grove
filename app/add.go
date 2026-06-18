package app

func (a *App) OpenAdd() {
	a.state.Screen = ScreenAdd
	if len(a.state.Loading) > 0 {
		a.ClearLoading()
	}
}

func (a *App) CloseAdd() {
	a.state.Screen = ScreenChange
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

// CreateBranchWorktree creates a new branch and a worktree for it. It is the
// operation the UI invokes when the user confirms creating a missing branch.
func (a *App) CreateBranchWorktree(path, branch string) Command {
	return a.addWorktree(path, branch, true)
}
