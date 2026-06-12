package ui

import (
	"github.com/charmbracelet/bubbles/key"
)

func changeHeaderLines(m Model) []string {
	inputLine := "> "
	if m.change.query == "" {
		inputLine += placeholderColor + "Search worktree paths" + resetColor
	} else {
		inputLine += m.change.query
	}
	return []string{inputLine, ""}
}

func changeBodyLines(m Model, availableHeight int) []string {
	visibleRows := max(1, availableHeight)
	m.syncChangeScroll(visibleRows)

	results := make([]string, 0, visibleRows)
	if len(m.change.filtered) == 0 {
		results = append(results, "No matches")
	} else {
		end := min(len(m.change.filtered), m.change.scroll+visibleRows)
		for i := m.change.scroll; i < end; i++ {
			label := m.displayItem(m.change.filtered[i])
			if i == m.change.selected {
				results = append(results, selectionColor+label+resetColor)
				continue
			}
			results = append(results, label)
		}
	}
	for len(results) < visibleRows {
		results = append(results, "")
	}
	return results
}

func changeFooterLines(m Model, contentWidth int) []string {
	helper := m.help
	helper.Width = max(contentWidth, 0)
	return []string{helper.ShortHelpView([]key.Binding{m.keys.submit, m.keys.addMode, m.keys.remove, m.keys.moveUp, m.keys.moveDown, m.keys.changeQuit})}
}
