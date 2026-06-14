package app

import "github.com/M-Xue/grove/worktree"

type Effect interface{}

type LoadWorktreesEffect struct{}

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
