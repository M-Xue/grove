// Package cli parses grove's command-line invocation into a structured
// Command. It owns the CLI grammar so that main.go is left with process
// wiring only.
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
// Command, rejecting flags and arguments grove does not accept.
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
	if err := fs.Parse(args); err != nil {
		return "", err
	}
	if fs.NArg() > 0 {
		return "", fmt.Errorf("unexpected arguments: %s", fs.Args())
	}
	return app.ScreenChange, nil
}
