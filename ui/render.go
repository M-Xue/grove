package ui

import "strings"

const (
	placeholderColor = "\x1b[38;5;245m"
	selectionColor   = "\x1b[38;5;183m"
	statusColor      = "\x1b[38;5;117m"
	errorColor       = "\x1b[38;5;203m"
	labelColor       = "\x1b[38;5;240m"
	resetColor       = "\x1b[0m"

	horizontalPadding = 4
	verticalPadding   = 2
)

func renderView(m Model) string {
	contentWidth := max(0, m.width-horizontalPadding*2)
	contentHeight := max(0, m.height-verticalPadding*2)
	content := panelLines(m, contentWidth, contentHeight)

	padded := make([]string, 0, max(m.height, 0))
	for i := 0; i < verticalPadding; i++ {
		padded = append(padded, repeatLine(" ", max(m.width, 0)))
	}
	for _, line := range content {
		padded = append(padded, strings.Repeat(" ", horizontalPadding)+fitLine(line, contentWidth))
	}
	for len(padded) < max(m.height-verticalPadding, 0) {
		padded = append(padded, repeatLine(" ", max(m.width, 0)))
	}
	for len(padded) < max(m.height, 0) {
		padded = append(padded, repeatLine(" ", max(m.width, 0)))
	}
	if m.height > 0 && len(padded) > m.height {
		padded = padded[:m.height]
	}

	return strings.Join(padded, "\n")
}

func panelLines(m Model, contentWidth, contentHeight int) []string {
	header := headerLines(m)
	footer := footerLine(m, contentWidth)
	status := statusLines(m, contentWidth)
	bottom := status
	bottom = append(bottom, footer...)

	body := bodyLines(m, contentHeight-len(header)-len(bottom))
	panel := header
	panel = append(panel, body...)
	for len(panel)+len(bottom) < contentHeight {
		panel = append(panel, repeatLine(" ", max(contentWidth, 0)))
	}
	panel = append(panel, bottom...)

	for len(panel) < contentHeight {
		panel = append(panel, repeatLine(" ", max(contentWidth, 0)))
	}
	if len(panel) > contentHeight {
		panel = panel[:contentHeight]
	}

	return panel
}

func headerLines(m Model) []string {
	modeLabel := "Change worktree"
	if m.mode == ModeAdd {
		modeLabel = "Add worktree"
	} else if m.mode == ModeDocs {
		modeLabel = "Docs"
	}

	lines := []string{"grove", "", modeLabel, ""}
	if m.mode == ModeAdd {
		return append(lines, addHeaderLines(m)...)
	}
	if m.mode == ModeDocs {
		return append(lines, docsHeaderLines()...)
	}
	return append(lines, changeHeaderLines(m)...)
}

func statusLines(m Model, contentWidth int) []string {
	if m.status.err != nil {
		return []string{renderStatusLine("error", errorColor, m.status.err.Error(), contentWidth)}
	}
	if m.status.message != "" {
		return []string{renderStatusLine("status", statusColor, m.status.message, contentWidth)}
	}
	return nil
}

func renderStatusLine(label, color, message string, contentWidth int) string {
	return fitLine(color+"["+label+"]"+resetColor+" "+labelColor+">"+resetColor+" "+message, max(contentWidth, 0))
}
