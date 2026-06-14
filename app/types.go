package app

import (
	branchsvc "github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/docs"
	"github.com/M-Xue/grove/worktree"
)

type ScreenID string

const (
	ScreenChange ScreenID = "change"
	ScreenAdd    ScreenID = "add"
	ScreenDocs   ScreenID = "docs"
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
	Docs     docs.Service
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

type DocsState struct{}

type BranchState struct {
	SelectedName string
	Commits      []branchsvc.CommitInfo
}

type State struct {
	Screen        ScreenID
	SubmittedPath string
	Worktrees     []worktree.WorktreeInfo
	Branches      []branchsvc.Info
	BranchScope   branchsvc.Scope
	DocsLines     []string
	Loading       []LoadingEntry
	Dialog        DialogState
	Statuses      []StatusEntry

	Change ChangeState
	Add    AddState
	Docs   DocsState
	Branch BranchState
}
