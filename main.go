package main

import (
	"fmt"
	"os"

	"github.com/M-Xue/grove/ui"
	tea "github.com/charmbracelet/bubbletea"
)

var worktrees = []string{
	"main",
	"feature/auth-refresh",
	"feature/worktree-prune",
	"bugfix/status-line-wrap",
	"release/v0.1.0",
	"spike/bubbletea-navigation",
}

func main() {
	p := tea.NewProgram(ui.New(worktrees), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error running grove: %v\n", err)
		os.Exit(1)
	}
}
