package app

func (a *App) RequestSubmitSelectedPath(path string) Effect {
	if path == "" {
		a.appendStatus(StatusInfo, "no worktree selected")
		return nil
	}
	a.state.SubmittedPath = path
	return QuitEffect{}
}

func (a *App) RequestRemoveWorktree(path string) {
	if path == "" {
		a.appendStatus(StatusInfo, "no worktree selected")
		return
	}
	a.state.Dialog = DialogState{
		Active:      true,
		Title:       "Delete worktree?",
		Description: path,
		Buttons:     []DialogButton{{ID: "confirm", Label: "Delete"}, {ID: "cancel", Label: "Cancel"}},
		FocusedID:   "cancel",
		Kind:        DialogConfirmRemove,
		Path:        path,
	}
}
