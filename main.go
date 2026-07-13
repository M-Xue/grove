package main

import (
	"fmt"
	"os"

	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/cli"
	"github.com/M-Xue/grove/command"
	"github.com/M-Xue/grove/repo"
	"github.com/M-Xue/grove/ui"
	"github.com/M-Xue/grove/worktree"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	cmd, err := cli.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running grove: %v\n", err)
		os.Exit(1)
	}
	runner := command.New()
	if err := repo.EnsureInRepo(runner); err != nil {
		fmt.Fprintf(os.Stderr, "error running grove: %v\n", err)
		os.Exit(1)
	}

	application := app.New(app.Services{
		Worktree: worktree.NewService(runner),
		Branch:   branch.NewService(runner),
	}, app.WithInitialScreen(cmd.Screen))

	p := tea.NewProgram(ui.New(application), tea.WithAltScreen())
	model, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running grove: %v\n", err)
		os.Exit(1)
	}

	finalModel, ok := model.(*ui.Model)
	if !ok {
		fmt.Fprintln(os.Stderr, "error running grove: unexpected final model type")
		os.Exit(1)
	}

	if path := selectedPathOutput(finalModel); path != "" {
		fmt.Println(path)
	}
}

func selectedPathOutput(model *ui.Model) string {
	return model.SubmittedPath()
}
