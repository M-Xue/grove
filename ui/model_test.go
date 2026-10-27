package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/branch"
	"github.com/M-Xue/grove/worktree"
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

// refreshStubWorktrees lets the model's refresh paths be executed in tests.
type refreshStubWorktrees struct {
	worktree.Service
	list []worktree.Info
}

func (s refreshStubWorktrees) List(worktree.ListOptions) ([]worktree.Info, error) { return s.list, nil }

func TestFocusTriggersBackgroundRefresh(t *testing.T) {
	application := app.New(app.Services{Worktree: refreshStubWorktrees{list: []worktree.Info{{Path: "/repo", Branch: "renamed"}}}})
	m := New(application)
	_, cmd := m.Update(tea.FocusMsg{})
	if cmd == nil {
		t.Fatal("expected a focus event to issue a refresh")
	}
	m.Update(cmd())
	if got := application.State().Worktrees; len(got) != 1 || got[0].Branch != "renamed" {
		t.Fatalf("expected the focus refresh to apply the live list, got %#v", got)
	}
	if len(application.State().Loading) != 0 {
		t.Fatalf("expected a silent refresh, got loading entries %#v", application.State().Loading)
	}
}

func TestRefreshTickRefreshesOnlyWhenIdle(t *testing.T) {
	application := app.New(app.Services{Worktree: refreshStubWorktrees{}})
	m := New(application)
	if m.backgroundRefresh() == nil {
		t.Fatal("expected a background refresh when idle")
	}
	// A tick while that refresh is still in flight issues no second load but
	// still re-arms the timer.
	if m.backgroundRefresh() != nil {
		t.Fatal("expected no second refresh while one is in flight")
	}
	if _, cmd := m.Update(refreshTickMsg{}); cmd == nil {
		t.Fatal("expected the refresh tick to re-arm itself")
	}
}

func TestRefreshIntervalScalesWithListingDuration(t *testing.T) {
	cases := []struct {
		lastLoad time.Duration
		want     time.Duration
	}{
		{0, minRefreshInterval},                      // nothing measured yet
		{300 * time.Millisecond, minRefreshInterval}, // cheap repo keeps the floor
		{600 * time.Millisecond, 6 * time.Second},    // stretches with cost
		{9 * time.Second, maxRefreshInterval},        // but is capped
	}
	for _, c := range cases {
		if got := refreshIntervalFor(c.lastLoad); got != c.want {
			t.Fatalf("refreshIntervalFor(%v) = %v, want %v", c.lastLoad, got, c.want)
		}
	}
}

// stubBranches lets details fetches be issued in tests.
type stubBranches struct{}

func (stubBranches) RecentCommits(string, int) ([]branch.CommitInfo, error) { return nil, nil }

func TestInitialDetailsFetchUsesCachedSeed(t *testing.T) {
	seeded := app.New(app.Services{Branch: stubBranches{}},
		app.WithCachedWorktrees([]worktree.Info{{Path: "/repo", Branch: "main"}}))
	if New(seeded).initialDetails() == nil {
		t.Fatal("expected a cached seed to start the hovered branch's details fetch at startup")
	}
	unseeded := app.New(app.Services{Branch: stubBranches{}})
	if New(unseeded).initialDetails() != nil {
		t.Fatal("expected no details fetch when nothing is seeded")
	}
}
