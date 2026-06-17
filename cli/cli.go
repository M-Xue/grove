// Package cli parses grove's command-line invocation into a structured
// Command. It owns the CLI grammar (the shell-init subcommand and the
// initial-screen flags) so that main.go is left with process wiring only.
//
// cli depends on app solely for the app.ScreenID type it returns.
package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/shellinit"
)

// Kind identifies which top-level command grove was invoked with.
type Kind int

const (
	// KindRun launches the interactive TUI.
	KindRun Kind = iota
	// KindShellInit emits a shell integration script.
	KindShellInit
)

// Command is the parsed result of a grove invocation.
type Command struct {
	Kind   Kind
	Screen app.ScreenID
	Shell  string
}

// Parse interprets the process arguments (excluding the program name) into a
// Command, validating the shell-init subcommand and the initial-screen flags.
func Parse(args []string) (Command, error) {
	if len(args) >= 1 && args[0] == "shell-init" {
		if len(args) < 2 {
			return Command{}, fmt.Errorf("shell-init requires a shell name")
		}
		shell := args[1]
		if len(args) > 2 {
			return Command{}, fmt.Errorf("unexpected arguments: %s", strings.Join(args[2:], " "))
		}
		if !shellinit.IsSupported(shell) {
			return Command{}, fmt.Errorf("unsupported shell %q", shell)
		}
		return Command{Kind: KindShellInit, Shell: shell}, nil
	}

	screen, err := parseInitialScreen(args)
	if err != nil {
		return Command{}, err
	}
	return Command{Kind: KindRun, Screen: screen}, nil
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
