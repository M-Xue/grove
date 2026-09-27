// Package panel renders a titled, sharp-cornered box around a section of the
// app, in the style of lazygit panes: the title sits as a highlighted tab in
// the top border and the content is padded one column from each side. Like the
// other components it is pure presentation — it knows nothing about what a
// section contains.
package panel

import (
	"strings"

	"github.com/M-Xue/grove/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

// Tab renders a standalone title tab (e.g. the app name) in the same style as
// a panel's embedded title.
func Tab(label string) string {
	return tabWithBackground(label, theme.TitleBg)
}

// tabWithBackground renders a title tab on the given background, so a panel's
// tab always matches its border color.
func tabWithBackground(label string, background lipgloss.Color) string {
	return lipgloss.NewStyle().Foreground(theme.TitleFg).Background(background).Bold(true).Render(" " + label + " ")
}

// Render draws content inside a square-cornered box of exactly width×height
// cells, with title embedded as a tab in the top border. The border is drawn
// in the active color when active, muted otherwise. Content is clipped to the
// interior; missing lines are blank-filled so the box is always full height.
// Degenerate sizes return the content unboxed.
func Render(title, content string, width, height int, active bool) string {
	if width < 4 || height < 2 {
		return content
	}
	borderColor := theme.BorderInactive
	if active {
		borderColor = theme.BorderActive
	}
	border := lipgloss.NewStyle().Foreground(borderColor)

	inner := width - 2 // columns between the vertical borders
	tab := tabWithBackground(title, borderColor)
	if lipgloss.Width(tab) > inner-1 {
		tab = lipgloss.NewStyle().MaxWidth(inner - 1).Render(tab)
	}
	dashes := max(0, inner-1-lipgloss.Width(tab))
	top := border.Render("┌─") + tab + border.Render(strings.Repeat("─", dashes)+"┐")
	bottom := border.Render("└" + strings.Repeat("─", inner) + "┘")

	side := border.Render("│")
	lines := strings.Split(content, "\n")
	rows := make([]string, 0, height)
	rows = append(rows, top)
	for i := 0; i < height-2; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		rows = append(rows, side+" "+fit(line, inner-2)+" "+side)
	}
	rows = append(rows, bottom)
	return strings.Join(rows, "\n")
}

// fit pads or truncates line to exactly width display cells, ANSI-aware.
func fit(line string, width int) string {
	if width <= 0 {
		return ""
	}
	lineWidth := lipgloss.Width(line)
	if lineWidth > width {
		return lipgloss.NewStyle().MaxWidth(width).Render(line)
	}
	return line + strings.Repeat(" ", width-lineWidth)
}
