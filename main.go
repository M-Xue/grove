package main

import (
	"fmt"
	"os"

	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/cache"
	"github.com/M-Xue/grove/cli"
	"github.com/M-Xue/grove/command"
	"github.com/M-Xue/grove/repo"
	"github.com/M-Xue/grove/ui"
	"github.com/M-Xue/grove/ui/components/dialog"
	"github.com/M-Xue/grove/worktree"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
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

	options := []app.Option{app.WithInitialScreen(cmd.Screen)}
	options = append(options, worktreeCacheOptions(runner)...)

	application := app.New(app.Services{
		Worktree: worktree.NewService(runner),
		Branch:   branch.NewService(runner),
	}, options...)

	// The TUI renders to stderr, but lipgloss's global default renderer detects
	// its color profile from stdout. When grove runs inside the shell wrapper, its
	// stdout is captured (a pipe, not a tty), so lipgloss would detect no color and
	// strip the styling from anything rendered through the default renderer (e.g.
	// the bubbles/help footer hints). Point the default renderer at stderr — the
	// real tty the TUI writes to — so colors survive.
	lipgloss.SetDefaultRenderer(lipgloss.NewRenderer(os.Stderr))

	// Detect the terminal background before the TUI takes over stdin, and derive
	// the dialog panel color from it. The query goes to stderr (where the TUI
	// renders) so it never pollutes stdout, which carries the selected path.
	dialog.SetTerminalBackground(terminalBackgroundHex())

	p := tea.NewProgram(ui.New(application), tea.WithAltScreen(), tea.WithOutput(os.Stderr))
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

// worktreeCacheOptions wires the persistent worktree cache into the app so the
// change screen paints instantly on launch (stale-while-revalidate). The cache
// is a pure optimization: if its directory or repo key can't be resolved, this
// returns no options and grove falls back to a live git listing. The saver is
// best-effort — a write failure is intentionally ignored so caching never
// surfaces an error to the user.
func worktreeCacheOptions(runner command.Runner) []app.Option {
	store, ok := cache.New()
	if !ok {
		return nil
	}
	key, err := repo.WorktreeCacheKey(runner)
	if err != nil {
		return nil
	}

	var options []app.Option
	if list, hit := store.Load(key); hit {
		options = append(options, app.WithCachedWorktrees(list))
	}
	options = append(options, app.WithWorktreeCacheSaver(func(list []worktree.Info) {
		_ = store.Save(key, list)
	}))
	return options
}

// terminalBackgroundHex returns the terminal's background color as a "#rrggbb"
// hex string, or "" if it cannot be detected. It queries via stderr — the fd the
// TUI renders to and a real tty — so it neither pollutes stdout nor competes with
// the running program for input.
func terminalBackgroundHex() string {
	output := termenv.NewOutput(os.Stderr)
	bg := output.BackgroundColor()
	if _, ok := bg.(termenv.NoColor); ok {
		return ""
	}
	return termenv.ConvertToRGB(bg).Hex()
}
