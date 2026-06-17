package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/M-Xue/grove/app"
	branchService "github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/shellinit"
	"github.com/M-Xue/grove/ui"
	"github.com/M-Xue/grove/worktree"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	cmd, err := parseCommand(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running grove: %v\n", err)
		os.Exit(1)
	}
	if cmd.Kind == commandShellInit {
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

type commandKind int

const (
	commandRun commandKind = iota
	commandShellInit
)

type command struct {
	Kind   commandKind
	Screen app.ScreenID
	Shell  string
}

func parseCommand(args []string) (command, error) {
	if len(args) >= 1 && args[0] == "shell-init" {
		if len(args) < 2 {
			return command{}, fmt.Errorf("shell-init requires a shell name")
		}
		shell := args[1]
		if len(args) > 2 {
			return command{}, fmt.Errorf("unexpected arguments: %s", strings.Join(args[2:], " "))
		}
		if !shellinit.IsSupported(shell) {
			return command{}, fmt.Errorf("unsupported shell %q", shell)
		}
		return command{Kind: commandShellInit, Shell: shell}, nil
	}

	screen, err := parseInitialScreen(args)
	if err != nil {
		return command{}, err
	}
	return command{Kind: commandRun, Screen: screen}, nil
}

func parseInitialScreen(args []string) (app.ScreenID, error) {
	fs := flag.NewFlagSet("grove", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	branch := fs.Bool("b", false, "open branch screen")
	add := fs.Bool("a", false, "open add screen")
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
	if selectedCount > 1 {
		return "", fmt.Errorf("only one of -a or -b may be provided")
	}
	if fs.NArg() > 0 {
		return "", fmt.Errorf("unexpected arguments: %s", fs.Args())
	}
	return selected, nil
}
