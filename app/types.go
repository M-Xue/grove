package app

import (
	"github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/worktree"
)

type ScreenID string

const (
	ScreenChange ScreenID = "change"
	ScreenAdd    ScreenID = "add"
	ScreenBranch ScreenID = "branch"
)

type Services struct {
	Worktree worktree.Service
	Branch   branch.Service
}

type LoadingEntry struct {
	ID        string
	Active    bool
	Completed bool
	Message   string
	// Progress marks an entry that renders a checkout progress bar alongside
	// its spinner. Done/Total are the files written so far out of the total;
	// a zero Total renders as 0%.
	Progress bool
	Done     int
	Total    int
}

type ChangeState struct{}

type AddState struct{}

type BranchState struct {
	SelectedName string
	Commits      []branch.CommitInfo
}

type State struct {
	Screen        ScreenID
	SubmittedPath string
	Worktrees     []worktree.Info
	Branches      []branch.Info
	BranchScope   branch.Scope
	Loading       []LoadingEntry
	Statuses      []StatusEntry

	Change ChangeState
	Add    AddState
	Branch BranchState
}
