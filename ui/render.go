package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
)

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
	header := headerLines(m)
	footer := footerLine(m, contentWidth)
	status := statusLines(m, contentWidth)
	bottom := append(status, footer...)
	helpHeight := len(bottom)
	results := []string{}

	if !m.addMode {
		resultsHeight := max(1, contentHeight-len(header)-helpHeight)
		visibleRows := resultsHeight
		m.syncScroll(visibleRows)

		results = make([]string, 0, visibleRows)
		if len(m.filtered) == 0 {
			results = append(results, "No matches")
		} else {
			end := min(len(m.filtered), m.scroll+visibleRows)
			for i := m.scroll; i < end; i++ {
				label := m.displayItem(m.filtered[i])
				if i == m.selected {
					results = append(results, selectionColor+label+resetColor)
					continue
				}
				results = append(results, label)
			}
		}
		for len(results) < visibleRows {
			results = append(results, "")
		}
	}

	panel := append(header, results...)
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
	modeLabel := " Change worktree"
	if m.addMode {
		modeLabel = " Add worktree"
	}

	lines := []string{"grove", "", modeLabel[1:], ""}

	if m.addMode {
		pathPrefix := "  "
		branchPrefix := "  "
		if !m.confirmCreateBranch {
			if m.addField == pathField {
				pathPrefix = "> "
			} else {
				branchPrefix = "> "
			}
		}

		lines = append(lines,
			pathPrefix+placeholder(m.addPath, "Relative path"),
			branchPrefix+placeholder(m.addBranch, "Branch name"),
			"",
		)

		if m.confirmCreateBranch {
			lines = append(lines,
				"Branch does not exist.",
				"Create a new branch? [y/n]",
				"",
			)
		}

		return lines
	}

	inputLine := "> "
	if m.query == "" {
		inputLine += placeholderColor + "Search worktree paths" + resetColor
	} else {
		inputLine += m.query
	}

	lines = append(lines, inputLine, "")
	return lines
}

func statusLines(m Model, contentWidth int) []string {
	if m.errorMessage != "" {
		return []string{fitLine(" error: "+m.errorMessage, max(contentWidth, 0))}
	}
	if m.statusMessage != "" {
		return []string{fitLine(" "+m.statusMessage, max(contentWidth, 0))}
	}
	return nil
}

func footerLine(m Model, contentWidth int) []string {
	helper := m.help
	helper.Width = max(contentWidth, 0)

	if m.addMode {
		if m.confirmCreateBranch {
			return []string{helper.ShortHelpView([]key.Binding{m.keys.confirmYes, m.keys.confirmNo, m.keys.close, m.keys.addQuit})}
		} else {
			return []string{helper.ShortHelpView([]key.Binding{m.keys.submit, m.keys.switchField, m.keys.close, m.keys.addQuit})}
		}
	}

	return []string{helper.ShortHelpView([]key.Binding{m.keys.addMode, m.keys.moveUp, m.keys.moveDown, m.keys.changeQuit})}
}
