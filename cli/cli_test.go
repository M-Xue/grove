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

func TestParseInitialScreenRejectsUnknownFlags(t *testing.T) {
	if _, err := parseInitialScreen([]string{"-a"}); err == nil {
		t.Fatal("expected error for a removed flag")
	}
}

func TestParseInitialScreenRejectsArguments(t *testing.T) {
	if _, err := parseInitialScreen([]string{"extra"}); err == nil {
		t.Fatal("expected error for unexpected arguments")
	}
}

func TestParseDefaultsToChangeScreen(t *testing.T) {
	cmd, err := Parse(nil)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if cmd.Screen != app.ScreenChange {
		t.Fatalf("unexpected screen: %q", cmd.Screen)
	}
}
