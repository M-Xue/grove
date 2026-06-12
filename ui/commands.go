package ui

import (
	"github.com/M-Xue/grove/worktree"
	tea "github.com/charmbracelet/bubbletea"
)

func loadWorktreesCmd(manager worktree.Manager) tea.Cmd {
	return func() tea.Msg {
		worktrees, err := manager.List()
		return worktreesLoadedMsg{worktrees: worktrees, err: err}
	}
}

func checkBranchExistsCmd(manager worktree.Manager, path, branch string) tea.Cmd {
	return func() tea.Msg {
		exists, err := manager.BranchExists(branch)
		return branchCheckedMsg{path: path, branch: branch, exists: exists, err: err}
	}
}

func addWorktreeCmd(manager worktree.Manager, path, branch string, createBranch bool) tea.Cmd {
	return func() tea.Msg {
		var err error
		if createBranch {
			err = manager.AddNewBranch(path, branch)
		} else {
			err = manager.Add(path, branch)
		}
		return worktreeAddedMsg{err: err}
	}
}
