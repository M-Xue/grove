package ui

import "github.com/M-Xue/grove/worktree"

type worktreesLoadedMsg struct {
	worktrees []worktree.WorktreeInfo
	err       error
}

type branchCheckedMsg struct {
	path   string
	branch string
	exists bool
	err    error
}

type worktreeAddedMsg struct {
	err error
}

type worktreeRemovedMsg struct {
	path string
	err  error
}
