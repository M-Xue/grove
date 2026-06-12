package main

import (
	"fmt"
	"os"

	"github.com/M-Xue/grove/ui"
	"github.com/M-Xue/grove/worktree"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	manager := worktree.NewManager()
	if err := manager.InRepo(); err != nil {
		fmt.Fprintf(os.Stderr, "error running grove: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(ui.New(manager), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error running grove: %v\n", err)
		os.Exit(1)
	}
}
