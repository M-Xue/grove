package ui

func bodyLines(m Model, availableHeight int) []string {
	if m.mode == ModeAdd {
		return addBodyLines(m, availableHeight)
	}
	if m.mode == ModeDocs {
		return docsBodyLines(m, availableHeight)
	}
	return changeBodyLines(m, availableHeight)
}

func footerLine(m Model, contentWidth int) []string {
	if m.mode == ModeAdd {
		return addFooterLines(m, contentWidth)
	}
	if m.mode == ModeDocs {
		return docsFooterLines(m, contentWidth)
	}
	return changeFooterLines(m, contentWidth)
}
