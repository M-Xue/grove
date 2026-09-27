package screens

import (
	"github.com/M-Xue/grove/pr"
	"github.com/charmbracelet/lipgloss"
)

// prDetailLabelWidth is the column the PR section's labels are padded to,
// sized to the widest label, "CI Checks:".
const prDetailLabelWidth = 10

// prHeaderIcon is the nerd-font source-pull glyph shown beside the PR header.
const prHeaderIcon = "\U000F04C2"

// Check and pill accents. Styles are built per call so they bind to the
// default renderer main installs at startup (see detailLabel).
var (
	pillTextColor = lipgloss.Color("235")
	passedColor   = lipgloss.Color("114")
	pendingColor  = lipgloss.Color("179")
	failedColor   = lipgloss.Color("174")
	mutedColor    = lipgloss.Color("245")
	mergedColor   = lipgloss.Color("183")
)

// prRows renders a pull request's details (the screen renders the section
// heading): the status pill, the labelled title/author/URL (long values
// wrapping in the label column), then the CI rollup with one line per check.
func prRows(info pr.Info, width int) []string {
	lines := []string{fitLine(detailLabel("Status:"), prDetailLabelWidth) + " " + statusPill(info.Status)}
	lines = append(lines, labeledRows("Title:", info.Title, prDetailLabelWidth, width)...)
	lines = append(lines, labeledRows("Author:", info.Author, prDetailLabelWidth, width)...)
	lines = append(lines, labeledRows("URL:", info.URL, prDetailLabelWidth, width)...)
	// A PR with no CI still shows the rollup line — as "None", muted — so the
	// row set stays stable rather than the section silently shrinking.
	summary := checkSummary(pr.SummarizeChecks(info.Checks))
	lines = append(lines, fitLine(detailLabel("CI Checks:"), prDetailLabelWidth)+" "+summary)
	for _, check := range info.Checks {
		lines = append(lines, "  "+checkIcon(check.State)+" "+check.Name)
	}
	return lines
}

// statusPill renders the PR's state as a colored pill, lazyworktree-style:
// dark bold text on a state-colored background.
func statusPill(status pr.Status) string {
	label, background := "Closed", failedColor
	switch status {
	case pr.StatusOpen:
		label, background = "Open", passedColor
	case pr.StatusDraft:
		label, background = "Draft", mutedColor
	case pr.StatusMerged:
		label, background = "Merged", mergedColor
	}
	return lipgloss.NewStyle().Foreground(pillTextColor).Background(background).Bold(true).Render(" " + label + " ")
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
