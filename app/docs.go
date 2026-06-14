package app

import "fmt"

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
	default:
		return nil
	}
}

func (a *App) HandleResult(result Result) Effect {
	switch msg := result.(type) {
	case WorktreesLoadedResult:
		a.clearLoading()
		if msg.Err != nil {
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.state.Worktrees = msg.Worktrees
		a.state.SubmittedPath = ""
		return nil
	case BranchCheckedResult:
		a.clearLoading()
		if msg.Err != nil {
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
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
		a.clearLoading()
		if msg.Err != nil {
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.state.Screen = ScreenChange
		a.appendStatus(StatusSuccess, "worktree added")
		a.setLoading("loading worktrees")
		return LoadWorktreesEffect{}
	case WorktreeRemovedResult:
		a.clearLoading()
		if msg.Err != nil {
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.state.Screen = ScreenChange
		a.appendStatus(StatusSuccess, "worktree removed")
		a.setLoading("loading worktrees")
		return LoadWorktreesEffect{}
	case DocsLoadedResult:
		a.clearLoading()
		if msg.Err != nil {
			a.state.Screen = ScreenChange
			a.appendStatus(StatusError, msg.Err.Error())
			return nil
		}
		a.state.DocsLines = msg.Lines
		return nil
	default:
		return nil
	}
}
