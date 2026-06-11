package ui

import "strings"

const (
	placeholderColor = "\x1b[38;5;245m"
	selectionColor   = "\x1b[38;5;183m"
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
	inputLine := "> "
	if m.query == "" {
		inputLine += placeholderColor + "Search worktree paths" + resetColor
	} else {
		inputLine += m.query
	}

	header := []string{
		" grove",
		"",
		inputLine,
		"",
	}
	helpHeight := 1
	resultsHeight := max(1, contentHeight-len(header)-helpHeight)
	visibleRows := resultsHeight
	m.syncScroll(visibleRows)

	results := make([]string, 0, visibleRows)
	if len(m.filtered) == 0 {
		results = append(results, "No matches")
	} else {
		end := min(len(m.filtered), m.scroll+visibleRows)
		for i := m.scroll; i < end; i++ {
			if i == m.selected {
				results = append(results, selectionColor+m.filtered[i]+resetColor)
				continue
			}
			results = append(results, m.filtered[i])
		}
	}
	for len(results) < visibleRows {
		results = append(results, "")
	}

	panel := append(header, results...)
	panel = append(
		panel,
		fitLine(" tab/shift+tab or arrows to move, esc to quit", max(contentWidth, 0)),
	)

	for len(panel) < contentHeight {
		panel = append(panel, repeatLine(" ", max(contentWidth, 0)))
	}
	if len(panel) > contentHeight {
		panel = panel[:contentHeight]
	}

	return panel
}
