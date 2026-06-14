package ui

import (
	"errors"
	"strings"
	"testing"

	"github.com/M-Xue/grove/worktree"
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

func TestChangeHeaderLinesRenderRemoveConfirmation(t *testing.T) {
	m := New(nil)
	m.change.worktrees = []worktree.WorktreeInfo{{Path: "/repo-feature", Branch: "feature/auth"}}
	m.change.confirmRemove = true
	m.change.confirmPath = "/repo-feature"

	lines := changeHeaderLines(m)
	if len(lines) < 3 {
		t.Fatalf("expected confirmation lines, got %d", len(lines))
	}
	if lines[0] != "Delete worktree?" {
		t.Fatalf("unexpected first line: got %q", lines[0])
	}
	if !strings.Contains(lines[1], "/repo-feature [feature/auth]") {
		t.Fatalf("unexpected second line: got %q", lines[1])
	}
	if lines[2] != "This cannot be undone. [y/n]" {
		t.Fatalf("unexpected third line: got %q", lines[2])
	}
}
