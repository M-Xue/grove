package ui

import (
	"strings"
	"testing"

	"github.com/M-Xue/grove/app"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func busyState() app.State {
	return app.State{Loading: []app.LoadingEntry{{Blocking: true}}}
}

func TestInputFrozenBlocksKeysWhileBusy(t *testing.T) {
	if !inputFrozen(busyState(), tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}) {
		t.Fatal("expected a rune key to be frozen while busy")
	}
	if !inputFrozen(busyState(), tea.KeyMsg{Type: tea.KeyEnter}) {
		t.Fatal("expected enter to be frozen while busy")
	}
}

func TestInputFrozenAllowsQuitWhileBusy(t *testing.T) {
	if inputFrozen(busyState(), tea.KeyMsg{Type: tea.KeyCtrlC}) {
		t.Fatal("expected ctrl+c to pass through while busy")
	}
}

func TestInputFrozenAllowsKeysWhenIdle(t *testing.T) {
	idle := app.State{Loading: []app.LoadingEntry{{Blocking: false}}}
	if inputFrozen(idle, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}) {
		t.Fatal("expected keys to pass through when no blocking op is in flight")
	}
}

func TestFooterShowsBusyHintWhileBusy(t *testing.T) {
	m := New(app.New(app.Services{}))
	got := m.footer(80, busyState())
	if !strings.Contains(got, "working") || !strings.Contains(got, "ctrl+c") {
		t.Fatalf("expected a busy footer hint, got %q", got)
	}
}

func TestFooterShowsScreenHintsWhenIdle(t *testing.T) {
	m := New(app.New(app.Services{}))
	got := m.footer(80, app.State{Screen: app.ScreenChange})
	if strings.Contains(got, "working") {
		t.Fatalf("expected the normal screen footer when idle, got %q", got)
	}
}

func TestPlaceLeftRightSplitsColumns(t *testing.T) {
	got := placeLeftRight("status", "loading", 20)
	if w := lipgloss.Width(got); w != 20 {
		t.Fatalf("expected width 20, got %d (%q)", w, got)
	}
	if !strings.HasPrefix(got, "status") {
		t.Fatalf("expected status flush left, got %q", got)
	}
	if !strings.HasSuffix(got, "loading") {
		t.Fatalf("expected loading flush right, got %q", got)
	}
}

func TestPlaceLeftRightRightOnlyIsRightAligned(t *testing.T) {
	got := placeLeftRight("", "loading", 20)
	if !strings.HasPrefix(got, " ") {
		t.Fatalf("expected leading padding, got %q", got)
	}
	if !strings.HasSuffix(got, "loading") {
		t.Fatalf("expected loading flush right, got %q", got)
	}
}

func TestPlaceLeftRightTruncatesLeftOnCollision(t *testing.T) {
	got := placeLeftRight("a very long status message", "load", 10)
	if w := lipgloss.Width(got); w != 10 {
		t.Fatalf("expected width 10, got %d (%q)", w, got)
	}
	if !strings.HasSuffix(got, "load") {
		t.Fatalf("expected loading to stay fully visible, got %q", got)
	}
}

func TestComposeNoticesBottomAlignsEachColumn(t *testing.T) {
	rows := composeNotices([]string{"s1", "s2"}, []string{"l1"}, 20)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	// Status occupies both rows (left), loading is bottom-aligned to the last row (right).
	if !strings.HasPrefix(rows[0], "s1") || strings.Contains(rows[0], "l1") {
		t.Fatalf("unexpected first row: %q", rows[0])
	}
	if !strings.HasPrefix(rows[1], "s2") || !strings.HasSuffix(rows[1], "l1") {
		t.Fatalf("unexpected second row: %q", rows[1])
	}
}

func TestComposeNoticesEmpty(t *testing.T) {
	if rows := composeNotices(nil, nil, 20); rows != nil {
		t.Fatalf("expected nil for no notices, got %#v", rows)
	}
}
