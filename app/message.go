package app

import (
	"github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/worktree"
)

// Command is a unit of async work. Its thunk runs on a goroutine and may only
// read the args and services it closes over, returning a Message — it must
// never touch App.State. All state mutation happens in intents and
// HandleMessage on the main loop.
type Command func() Message

// Message is an inspectable result event produced by a Command and consumed by
// HandleMessage. Concrete message types are plain structs so app tests can
// pattern-match on them.
type Message interface{}

// QuitRequested signals that grove should exit. The app cannot control the
// terminal directly, so ui translates this into a quit.
type QuitRequested struct{}

type WorktreesLoadedMessage struct {
	LoadingID string
	Worktrees []worktree.Info
	Err       error
}

type BranchesLoadedMessage struct {
	LoadingID string
	Branches  []branch.Info
	Scope     branch.Scope
	Err       error
}

type BranchCommitsLoadedMessage struct {
	LoadingID string
	Seq       int
	Name      string
	Commits   []branch.CommitInfo
	Err       error
}

type BranchCheckedOutMessage struct {
	LoadingID string
	Err       error
}

type BranchDeletedMessage struct {
	LoadingID string
	Err       error
}

type AllBranchesDeletedMessage struct {
	LoadingID string
	Deleted   []string
	Skipped   []string
	Err       error
}

type BranchesFetchedMessage struct {
	LoadingID string
	Err       error
}

type BranchCheckedMessage struct {
	LoadingID string
	Path      string
	Branch    string
	Exists    bool
	Err       error
}

type WorktreeAddedMessage struct {
	LoadingID string
	Err       error
}

type WorktreeRemovedMessage struct {
	LoadingID string
	Path      string
	Err       error
}
