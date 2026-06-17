package app

import (
	branchsvc "github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/worktree"
)

type ScreenID string

const (
	ScreenChange ScreenID = "change"
	ScreenAdd    ScreenID = "add"
	ScreenBranch ScreenID = "branch"
)

type DialogKind string

const (
	DialogNone                DialogKind = ""
	DialogConfirmRemove       DialogKind = "confirm-remove"
	DialogConfirmCreateBranch DialogKind = "confirm-create-branch"
	DialogConfirmDeleteBranch DialogKind = "confirm-delete-branch"
	DialogConfirmDeleteAllBranches DialogKind = "confirm-delete-all-branches"
)

type Services struct {
	Worktree worktree.Service
	Branch   branchsvc.Service
}

type LoadingEntry struct {
	ID        string
	Active    bool
	Completed bool
	Message   string
}

type DialogButton struct {
	ID    string
	Label string
}

type DialogState struct {
	Active      bool
	Title       string
	Description string
	Buttons     []DialogButton
	FocusedID   string
	Kind        DialogKind
	Path        string
	Branch      string
}

type ChangeState struct{}

type AddState struct{}

type BranchState struct {
	SelectedName string
	Commits      []branchsvc.CommitInfo
}

type State struct {
	Screen        ScreenID
	SubmittedPath string
	Worktrees     []worktree.Info
	Branches      []branchsvc.Info
	BranchScope   branchsvc.Scope
	Loading       []LoadingEntry
	Dialog        DialogState
	Statuses      []StatusEntry

	Change ChangeState
	Add    AddState
	Branch BranchState
}
