package app

import (
	branchsvc "github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/worktree"
)

type Effect interface{}

type LoadWorktreesEffect struct{}

type LoadBranchesEffect struct{}

type LoadBranchCommitsEffect struct {
	Name  string
	Limit int
}

type ToggleBranchScopeEffect struct{}

type CheckoutBranchEffect struct {
	Name string
}

type DeleteBranchEffect struct {
	Name string
}

type DeleteAllBranchesEffect struct{}

type FetchBranchesEffect struct{}

type LoadDocsEffect struct{}

type CheckBranchExistsEffect struct {
	Path   string
	Branch string
}

type AddWorktreeEffect struct {
	Path         string
	Branch       string
	CreateBranch bool
}

type RemoveWorktreeEffect struct {
	Path string
}

type QuitEffect struct{}

type Result interface{}

type WorktreesLoadedResult struct {
	Worktrees []worktree.WorktreeInfo
	Err       error
}

type BranchesLoadedResult struct {
	Branches []branchsvc.Info
	Scope    branchsvc.Scope
	Err      error
}

type BranchCommitsLoadedResult struct {
	Name    string
	Commits []branchsvc.CommitInfo
	Err     error
}

type BranchCheckedOutResult struct {
	Err error
}

type BranchDeletedResult struct {
	Err error
}

type AllBranchesDeletedResult struct {
	Deleted []string
	Skipped []string
	Err     error
}

type BranchesFetchedResult struct {
	Err error
}

type BranchCheckedResult struct {
	Path   string
	Branch string
	Exists bool
	Err    error
}

type WorktreeAddedResult struct {
	Err error
}

type WorktreeRemovedResult struct {
	Path string
	Err  error
}

type DocsLoadedResult struct {
	Lines []string
	Err   error
}
