package progressbar

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// renderedWidth is the visible width: 10 bar cells + 1 space + 4-char label.
const renderedWidth = cells + 1 + 4

func TestViewPercentageIsExactAndLeftAligned(t *testing.T) {
	cases := []struct {
		done, total int
		label       string
	}{
		{0, 0, "0%  "},
		{0, 57, "0%  "},
		{27, 57, "47% "},
		{57, 57, "100%"},
	}
	for _, c := range cases {
		got := New().View(c.done, c.total)
		if !strings.HasSuffix(got, c.label) {
			t.Fatalf("View(%d,%d) = %q; want suffix %q", c.done, c.total, got, c.label)
		}
	}
}

func TestViewWidthIsStable(t *testing.T) {
	for _, tc := range []struct{ done, total int }{{0, 0}, {1, 57}, {27, 57}, {57, 57}} {
		if w := lipgloss.Width(New().View(tc.done, tc.total)); w != renderedWidth {
			t.Fatalf("View(%d,%d) width = %d; want %d", tc.done, tc.total, w, renderedWidth)
		}
	}
}
