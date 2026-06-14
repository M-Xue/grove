package app

import (
	"github.com/M-Xue/grove/docs"
	"github.com/M-Xue/grove/worktree"
)

type ScreenID string

const (
	ScreenChange ScreenID = "change"
	ScreenAdd    ScreenID = "add"
	ScreenDocs   ScreenID = "docs"
)

type StatusKind string

const (
	StatusInfo    StatusKind = "info"
	StatusSuccess StatusKind = "success"
	StatusError   StatusKind = "error"
)

type DialogKind string

const (
	DialogNone                DialogKind = ""
	DialogConfirmRemove       DialogKind = "confirm-remove"
	DialogConfirmCreateBranch DialogKind = "confirm-create-branch"
)

type Services struct {
	Worktree worktree.Service
	Docs     docs.Service
}

type LoadingState struct {
	Active  bool
	Message string
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

type StatusEntry struct {
	ID      string
	Kind    StatusKind
	Message string
}

type ChangeState struct{}

type AddState struct{}

type DocsState struct{}

type State struct {
	Screen        ScreenID
	SubmittedPath string
	Worktrees     []worktree.WorktreeInfo
	DocsLines     []string
	Loading       LoadingState
	Dialog        DialogState
	Statuses      []StatusEntry

	Change ChangeState
	Add    AddState
	Docs   DocsState
}
