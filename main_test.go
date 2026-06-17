package main

import (
	"testing"

	"github.com/M-Xue/grove/app"
	branchsvc "github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/ui"
	"github.com/M-Xue/grove/worktree"
)

func TestSelectedPathOutputReturnsSubmittedPath(t *testing.T) {
	m := ui.New(app.New(app.Services{
		Worktree: worktree.NewServiceWithRunner(nil),
		Branch:   branchsvc.NewServiceWithRunner(nil),
	}))

	if got := selectedPathOutput(m); got != "" {
		t.Fatalf("expected empty path, got %q", got)
	}
}

func TestParseInitialScreenDefaultsToChange(t *testing.T) {
	screen, err := parseInitialScreen(nil)
	if err != nil {
		t.Fatalf("parseInitialScreen returned error: %v", err)
	}
	if screen != app.ScreenChange {
		t.Fatalf("unexpected screen: %q", screen)
	}
}

func TestParseInitialScreenSupportsFlags(t *testing.T) {
	tests := []struct {
		args []string
		want app.ScreenID
	}{
		{args: []string{"-a"}, want: app.ScreenAdd},
		{args: []string{"-b"}, want: app.ScreenBranch},
	}

	for _, test := range tests {
		screen, err := parseInitialScreen(test.args)
		if err != nil {
			t.Fatalf("parseInitialScreen(%v) returned error: %v", test.args, err)
		}
		if screen != test.want {
			t.Fatalf("parseInitialScreen(%v) = %q, want %q", test.args, screen, test.want)
		}
	}
}

func TestParseInitialScreenRejectsMultipleFlags(t *testing.T) {
	if _, err := parseInitialScreen([]string{"-a", "-b"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCommandDefaultsToRun(t *testing.T) {
	cmd, err := parseCommand(nil)
	if err != nil {
		t.Fatalf("parseCommand returned error: %v", err)
	}
	if cmd.Kind != commandRun {
		t.Fatalf("unexpected command kind: %v", cmd.Kind)
	}
	if cmd.Screen != app.ScreenChange {
		t.Fatalf("unexpected screen: %q", cmd.Screen)
	}
}

func TestParseCommandSupportsShellInit(t *testing.T) {
	cmd, err := parseCommand([]string{"shell-init", "zsh"})
	if err != nil {
		t.Fatalf("parseCommand returned error: %v", err)
	}
	if cmd.Kind != commandShellInit {
		t.Fatalf("unexpected command kind: %v", cmd.Kind)
	}
	if cmd.Shell != "zsh" {
		t.Fatalf("unexpected shell: %q", cmd.Shell)
	}
}

func TestParseCommandRejectsUnsupportedShell(t *testing.T) {
	if _, err := parseCommand([]string{"shell-init", "tcsh"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCommandSupportsPowershellShellInit(t *testing.T) {
	cmd, err := parseCommand([]string{"shell-init", "powershell"})
	if err != nil {
		t.Fatalf("parseCommand returned error: %v", err)
	}
	if cmd.Kind != commandShellInit {
		t.Fatalf("unexpected command kind: %v", cmd.Kind)
	}
	if cmd.Shell != "powershell" {
		t.Fatalf("unexpected shell: %q", cmd.Shell)
	}
}
