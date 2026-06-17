package main

import (
	"fmt"
	"os"

	"github.com/M-Xue/grove/app"
	branchService "github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/cli"
	"github.com/M-Xue/grove/shellinit"
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
	if cmd.Kind == cli.KindShellInit {
		execPath, err := os.Executable()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error running grove: %v\n", err)
			os.Exit(1)
		}
		script, err := shellinit.Script(cmd.Shell, execPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error running grove: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(script)
		return
	}

	worktreeService := worktree.NewService()
	if err := worktreeService.InRepo(); err != nil {
		fmt.Fprintf(os.Stderr, "error running grove: %v\n", err)
		os.Exit(1)
	}

	application := app.New(app.Services{
		Worktree: worktreeService,
		Branch:   branchService.NewService(),
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
