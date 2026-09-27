package app

import (
	"github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/pr"
	"github.com/M-Xue/grove/worktree"
)

type ScreenID string

const (
	ScreenChange ScreenID = "change"
)

type Services struct {
	Worktree worktree.Service
	Branch   branch.Service
	PR       pr.Service
}

// BranchPR is the cached result of one branch's pull-request lookup. Found
// distinguishes a definitive "this branch has no PR" from a branch that simply
// has no map entry yet (not fetched).
type BranchPR struct {
	Found bool
	Info  pr.Info
	// Failed records a transient lookup failure. The entry exists so the UI
	// can resolve its pending state (and say the lookup failed), but
	// LoadBranchPR treats it as uncached so a later hover retries.
	Failed bool
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
	// Blocking marks an entry whose in-flight operation should freeze the UI:
	// all input is ignored (except quit) until it completes. Passive background
	// loads (worktree/branch/commit lists) leave Blocking false.
	Blocking bool
}

type ChangeState struct{}

type State struct {
	Screen        ScreenID
	SubmittedPath string
	Worktrees     []worktree.Info
	// BranchCommits caches the recent commits fetched per branch, keyed by
	// short branch name. Entries are filled lazily as the change screen hovers
	// worktrees and overwritten wholesale when a fetch for that branch lands.
	BranchCommits map[string][]branch.CommitInfo
	// BranchPRs caches pull-request lookups per branch for the session: a PR
	// lookup is a network round-trip, so unlike commits it is fetched at most
	// once per branch and never refreshed on worktree reloads.
	BranchPRs map[string]BranchPR
	// PRLookupUnavailable is latched when a lookup reports pr.ErrUnavailable
	// (gh missing, unauthenticated, or no GitHub remote); no further lookups
	// are issued for the rest of the session.
	PRLookupUnavailable bool
	Loading             []LoadingEntry
	Statuses            []StatusEntry

	Change ChangeState
}
