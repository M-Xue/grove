package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/M-Xue/grove/app"
	branchsvc "github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/docs"
	"github.com/M-Xue/grove/ui"
	"github.com/M-Xue/grove/worktree"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	initialScreen, err := parseInitialScreen(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running grove: %v\n", err)
		os.Exit(1)
	}

	worktreeService := worktree.NewService()
	if err := worktreeService.InRepo(); err != nil {
		fmt.Fprintf(os.Stderr, "error running grove: %v\n", err)
		os.Exit(1)
	}

	application := app.New(app.Services{
		Worktree: worktreeService,
		Branch:   branchsvc.NewService(),
		Docs:     docs.NewService(),
	}, app.WithInitialScreen(initialScreen))

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

func parseInitialScreen(args []string) (app.ScreenID, error) {
	fs := flag.NewFlagSet("grove", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	branch := fs.Bool("b", false, "open branch screen")
	add := fs.Bool("a", false, "open add screen")
	docsFlag := fs.Bool("d", false, "open docs screen")
	if err := fs.Parse(args); err != nil {
		return "", err
	}
	selected := app.ScreenChange
	selectedCount := 0
	if *branch {
		selected = app.ScreenBranch
		selectedCount++
	}
	if *add {
		selected = app.ScreenAdd
		selectedCount++
	}
	if *docsFlag {
		selected = app.ScreenDocs
		selectedCount++
	}
	if selectedCount > 1 {
		return "", fmt.Errorf("only one of -a, -b, or -d may be provided")
	}
	if fs.NArg() > 0 {
		return "", fmt.Errorf("unexpected arguments: %s", fs.Args())
	}
	return selected, nil
}
