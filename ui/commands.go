package ui

import (
	"os/exec"
	"regexp"
	"strings"

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

func removeWorktreeCmd(manager worktree.Manager, path string) tea.Cmd {
	return func() tea.Msg {
		err := manager.Remove(path)
		return worktreeRemovedMsg{path: path, err: err}
	}
}

func openWorktreeDocsCmd() tea.Cmd {
	return func() tea.Msg {
		output, err := exec.Command("git", "--no-pager", "help", "worktree").CombinedOutput()
		if err != nil {
			return worktreeDocsLoadedMsg{err: err}
		}
		return worktreeDocsLoadedMsg{lines: formatWorktreeDocs(string(output))}
	}
}

var backspaceOverstrikePattern = regexp.MustCompile(`.`)

func formatWorktreeDocs(output string) []string {
	cleaned := strings.ReplaceAll(output, "\r\n", "\n")
	cleaned = strings.ReplaceAll(cleaned, "\r", "\n")
	for backspaceOverstrikePattern.MatchString(cleaned) {
		cleaned = backspaceOverstrikePattern.ReplaceAllString(cleaned, "")
	}
	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "" {
		return []string{"No documentation available."}
	}
	return strings.Split(cleaned, "\n")
}
