package ui

import "github.com/charmbracelet/bubbles/key"

func addHeaderLines(m Model) []string {
	pathPrefix := "  "
	branchPrefix := "  "
	if !m.add.confirmCreateBranch {
		if m.add.field == pathField {
			pathPrefix = "> "
		} else {
			branchPrefix = "> "
		}
	}

	lines := []string{
		pathPrefix + placeholder(m.add.path, "Relative path"),
		branchPrefix + placeholder(m.add.branch, "Branch name"),
		"",
	}

	if m.add.confirmCreateBranch {
		lines = append(lines,
			"Branch does not exist.",
			"Create a new branch? [y/n]",
			"",
		)
	}

	return lines
}

func addBodyLines(Model, int) []string {
	return nil
}

func addFooterLines(m Model, contentWidth int) []string {
	helper := m.help
	helper.Width = max(contentWidth, 0)
	if m.add.confirmCreateBranch {
		return []string{helper.ShortHelpView([]key.Binding{m.keys.confirmYes, m.keys.confirmNo, m.keys.close, m.keys.addQuit})}
	}
	return []string{helper.ShortHelpView([]key.Binding{m.keys.submit, m.keys.switchField, m.keys.close, m.keys.addQuit})}
}
