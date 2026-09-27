package panel

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestRenderBoxesContentWithTitleTab(t *testing.T) {
	got := Render("Worktrees", "one\ntwo", 30, 5, false)
	lines := strings.Split(got, "\n")
	if len(lines) != 5 {
		t.Fatalf("expected exactly 5 rows, got %d:\n%s", len(lines), got)
	}
	for i, line := range lines {
		if lipgloss.Width(line) != 30 {
			t.Fatalf("expected row %d to be 30 cells wide, got %d: %q", i, lipgloss.Width(line), line)
		}
	}
	if !strings.Contains(lines[0], "Worktrees") || !strings.Contains(lines[0], "┌") || !strings.Contains(lines[0], "┐") {
		t.Fatalf("expected the title embedded in a square top border, got %q", lines[0])
	}
	if !strings.Contains(lines[1], "one") || !strings.Contains(lines[2], "two") {
		t.Fatalf("expected content rows inside the box, got:\n%s", got)
	}
	if !strings.HasSuffix(stripAnsi(lines[4]), "┘") || !strings.Contains(lines[4], "└") {
		t.Fatalf("expected a square bottom border, got %q", lines[4])
	}
}

func TestRenderActiveAndInactiveBordersDiffer(t *testing.T) {
	// Tests run without a tty, where lipgloss strips all color; force a
	// profile so the border colors are actually emitted and comparable.
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	defer lipgloss.SetColorProfile(previous)

	active := Render("T", "x", 20, 4, true)
	inactive := Render("T", "x", 20, 4, false)
	if active == inactive {
		t.Fatal("expected the active border to render differently from the inactive one")
	}
	if stripAnsi(active) != stripAnsi(inactive) {
		t.Fatal("expected active/inactive to differ only in color, not geometry")
	}
}

func TestRenderBlankFillsMissingContentRows(t *testing.T) {
	got := Render("T", "only", 20, 6, false)
	lines := strings.Split(got, "\n")
	if len(lines) != 6 {
		t.Fatalf("expected 6 rows, got %d", len(lines))
	}
	// Rows past the content are blank interiors, still bordered.
	if !strings.Contains(lines[4], "│") {
		t.Fatalf("expected bordered blank row, got %q", lines[4])
	}
}

func TestRenderDegenerateSizeReturnsContent(t *testing.T) {
	if got := Render("T", "raw", 3, 1, false); got != "raw" {
		t.Fatalf("expected unboxed content for degenerate size, got %q", got)
	}
}

func stripAnsi(s string) string {
	var b strings.Builder
	inEscape := false
	for _, r := range s {
		switch {
		case inEscape:
			if r == 'm' {
				inEscape = false
			}
		case r == '\x1b':
			inEscape = true
		default:
			b.WriteRune(r)
		}
	}
	return strings.TrimRight(b.String(), " ")
}
