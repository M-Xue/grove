package screens

import (
	"github.com/M-Xue/grove/pr"
	"github.com/charmbracelet/lipgloss"
)

// prDetailLabelWidth is the column the PR section's labels are padded to,
// sized to the widest label, "CI Checks:".
const prDetailLabelWidth = 10

// Nerd-font octicon glyphs shown in the status pill, one per PR state.
const (
	iconPROpen   = ""
	iconPRDraft  = ""
	iconPRMerged = ""
	iconPRClosed = ""
)

// Check and pill accents. Styles are built per call so they bind to the
// default renderer main installs at startup (see detailLabel).
var (
	pillTextColor = lipgloss.Color("235")
	passedColor   = lipgloss.Color("114")
	pendingColor  = lipgloss.Color("179")
	failedColor   = lipgloss.Color("174")
	mutedColor    = lipgloss.Color("245")

	// Status pill backgrounds follow GitHub's PR state colors, in Catppuccin
	// Macchiato: green for open, grey (overlay1) for draft, blue for merged,
	// red for closed.
	openColor   = lipgloss.Color("#a6da95")
	draftColor  = lipgloss.Color("#8087a2")
	mergedColor = lipgloss.Color("#8aadf4")
	closedColor = lipgloss.Color("#ed8796")
)

// prFieldRows renders a pull request's labelled fields (the screen renders the
// section heading): the status pill, then title/author/URL with long values
// wrapping in the label column.
func prFieldRows(info pr.Info, width int) []string {
	lines := []string{fitLine(detailLabel("Status:"), prDetailLabelWidth) + " " + statusPill(info.Status)}
	lines = append(lines, labeledRows("Title:", info.Title, prDetailLabelWidth, width)...)
	lines = append(lines, labeledRows("Author:", info.Author, prDetailLabelWidth, width)...)
	lines = append(lines, labeledRows("URL:", info.URL, prDetailLabelWidth, width)...)
	return lines
}

// prChecksRows renders the CI rollup line, then one row per visible check.
// The rollup folds every check — not just the visible page — so the verdict
// stays honest while the list is paged; rangeStatus (when non-empty) marks
// which slice is showing. A PR with no CI still shows the rollup line — as
// "None", muted — so the row set stays stable rather than the section silently
// shrinking.
func prChecksRows(checks, visible []pr.Check, rangeStatus string) []string {
	summary := checkSummary(pr.SummarizeChecks(checks))
	rollup := fitLine(detailLabel("CI Checks:"), prDetailLabelWidth) + " " + summary
	if rangeStatus != "" {
		rollup += " " + mutedText(rangeStatus)
	}
	lines := []string{rollup}
	for _, check := range visible {
		lines = append(lines, "  "+checkIcon(check.State)+" "+check.Name)
	}
	return lines
}

// statusPill renders the PR's state as a colored pill, lazyworktree-style:
// dark bold text on a state-colored background, led by the state's glyph.
func statusPill(status pr.Status) string {
	icon, label, background := iconPRClosed, "Closed", closedColor
	switch status {
	case pr.StatusOpen:
		icon, label, background = iconPROpen, "Open", openColor
	case pr.StatusDraft:
		icon, label, background = iconPRDraft, "Draft", draftColor
	case pr.StatusMerged:
		icon, label, background = iconPRMerged, "Merged", mergedColor
	}
	return lipgloss.NewStyle().Foreground(pillTextColor).Background(background).Bold(true).Render(" " + icon + " " + label + " ")
}

// checkSummary renders the rollup word for the CI Checks line.
func checkSummary(state pr.CheckState) string {
	word, color := "None", mutedColor
	switch state {
	case pr.CheckPassed:
		word, color = "Passed", passedColor
	case pr.CheckPending:
		word, color = "Pending", pendingColor
	case pr.CheckFailed:
		word, color = "Failed", failedColor
	}
	return lipgloss.NewStyle().Foreground(color).Bold(true).Render(word)
}

// checkIcon renders one check's state as a colored circle glyph: a filled
// check for passed, a cross for failed or cancelled, and a hollow circle for
// anything inconsequential (pending keeps its own tint).
func checkIcon(state pr.CheckState) string {
	glyph, color := "", mutedColor
	switch state {
	case pr.CheckPassed:
		glyph, color = "", passedColor
	case pr.CheckFailed, pr.CheckCancelled:
		glyph, color = "", failedColor
	case pr.CheckPending:
		color = pendingColor
	}
	return lipgloss.NewStyle().Foreground(color).Render(glyph)
}
