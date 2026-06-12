package ui

import (
	"errors"
	"testing"
)

func TestStatusLinesRendersStyledStatusMessage(t *testing.T) {
	m := New(nil)
	m.setStatus("worktree added")

	lines := statusLines(m, 80)
	if len(lines) != 1 {
		t.Fatalf("expected 1 status line, got %d", len(lines))
	}

	want := statusColor + "[status]" + resetColor + " " + labelColor + ">" + resetColor + " worktree added"
	if lines[0][:len(want)] != want {
		t.Fatalf("unexpected status line prefix: got %q want prefix %q", lines[0], want)
	}
}

func TestStatusLinesRendersStyledErrorMessage(t *testing.T) {
	m := New(nil)
	m.setError(errors.New("git failed"))

	lines := statusLines(m, 80)
	if len(lines) != 1 {
		t.Fatalf("expected 1 status line, got %d", len(lines))
	}

	want := errorColor + "[error]" + resetColor + " " + labelColor + ">" + resetColor + " git failed"
	if lines[0][:len(want)] != want {
		t.Fatalf("unexpected error line prefix: got %q want prefix %q", lines[0], want)
	}
}
