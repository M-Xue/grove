package ui

import "github.com/charmbracelet/bubbles/key"

func docsHeaderLines() []string {
	return []string{"Git worktree docs", "", "Press esc, q, or ? to return.", ""}
}

func docsBodyLines(m Model, availableHeight int) []string {
	visibleRows := max(1, availableHeight)
	m.syncDocsScroll(visibleRows)

	results := make([]string, 0, visibleRows)
	end := min(len(m.docs.lines), m.docs.scroll+visibleRows)
	for i := m.docs.scroll; i < end; i++ {
		results = append(results, m.docs.lines[i])
	}
	for len(results) < visibleRows {
		results = append(results, "")
	}
	return results
}

func docsFooterLines(m Model, contentWidth int) []string {
	helper := m.help
	helper.Width = max(contentWidth, 0)
	return []string{helper.ShortHelpView([]key.Binding{m.keys.docs, m.keys.moveUp, m.keys.moveDown, m.keys.changeQuit})}
}
