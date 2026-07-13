// Package cli parses grove's command-line invocation into a structured
// Command. It owns the CLI grammar (the initial-screen flags) so that main.go
// is left with process wiring only.
//
// cli depends on app solely for the app.ScreenID type it returns.
package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/M-Xue/grove/app"
)

// Command is the parsed result of a grove invocation.
type Command struct {
	Screen app.ScreenID
}

// Parse interprets the process arguments (excluding the program name) into a
// Command, validating the initial-screen flags.
func Parse(args []string) (Command, error) {
	screen, err := parseInitialScreen(args)
	if err != nil {
		return Command{}, err
	}
	return Command{Screen: screen}, nil
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
