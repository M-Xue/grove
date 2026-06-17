package cli

import (
	"testing"

	"github.com/M-Xue/grove/app"
)

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

func TestParseDefaultsToRun(t *testing.T) {
	cmd, err := Parse(nil)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if cmd.Kind != KindRun {
		t.Fatalf("unexpected command kind: %v", cmd.Kind)
	}
	if cmd.Screen != app.ScreenChange {
		t.Fatalf("unexpected screen: %q", cmd.Screen)
	}
}

func TestParseSupportsShellInit(t *testing.T) {
	cmd, err := Parse([]string{"shell-init", "zsh"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if cmd.Kind != KindShellInit {
		t.Fatalf("unexpected command kind: %v", cmd.Kind)
	}
	if cmd.Shell != "zsh" {
		t.Fatalf("unexpected shell: %q", cmd.Shell)
	}
}

func TestParseRejectsUnsupportedShell(t *testing.T) {
	if _, err := Parse([]string{"shell-init", "tcsh"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseSupportsPowershellShellInit(t *testing.T) {
	cmd, err := Parse([]string{"shell-init", "powershell"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if cmd.Kind != KindShellInit {
		t.Fatalf("unexpected command kind: %v", cmd.Kind)
	}
	if cmd.Shell != "powershell" {
		t.Fatalf("unexpected shell: %q", cmd.Shell)
	}
}
