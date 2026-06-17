package app

func (a *App) RequestSubmitSelectedPath(path string) Command {
	if path == "" {
		a.appendStatus(StatusInfo, "no worktree selected")
		return nil
	}
	a.state.SubmittedPath = path
	return func() Message { return QuitRequested{} }
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
