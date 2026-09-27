package app

import (
	"github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/pr"
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

// BranchExistsMessage and BranchAbsentMessage are the semantic outcomes of
// checking whether a branch exists before adding a worktree. The UI reacts to
// BranchAbsentMessage by offering to create the branch; HandleMessage reacts to
// BranchExistsMessage by adding the worktree directly.
type BranchExistsMessage struct {
	LoadingID string
	Path      string
	Branch    string
}

type BranchAbsentMessage struct {
	LoadingID string
	Path      string
	Branch    string
}

type BranchCheckFailedMessage struct {
	LoadingID string
	Err       error
}

// WorktreeAddStartedMessage carries the channel of progress and completion
// messages emitted while a worktree is being added. The UI drains Updates,
// feeding each message through HandleMessage, until the channel is closed. It
// is handled entirely in the UI layer and never reaches HandleMessage.
type WorktreeAddStartedMessage struct {
	Updates <-chan Message
}

// WorktreeProgressMessage reports checkout progress for the in-flight worktree
// add identified by LoadingID.
type WorktreeProgressMessage struct {
	LoadingID string
	Done      int
	Total     int
}

type WorktreeAddedMessage struct {
	LoadingID string
	Err       error
}

// BranchCommitsLoadedMessage carries the recent commits fetched for one
// branch's details panel. It has no LoadingID because the fetch is a silent,
// hover-driven background lookup with no loading entry.
type BranchCommitsLoadedMessage struct {
	Branch  string
	Commits []branch.CommitInfo
	Err     error
}

// BranchPRLoadedMessage carries the pull-request lookup result for one branch.
// Like the commits fetch it is a silent background lookup with no LoadingID.
// Found=false with a nil Err is the ordinary "no PR for this branch" outcome.
type BranchPRLoadedMessage struct {
	Branch string
	Info   pr.Info
	Found  bool
	Err    error
}

type WorktreeRemovedMessage struct {
	LoadingID string
	Path      string
	Err       error
}

type WorktreesPrunedMessage struct {
	LoadingID string
	Err       error
}
