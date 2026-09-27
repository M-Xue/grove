// Package dialog provides the shared chrome for grove's dialog overlays: a
// bordered, title-tabbed box sized to its content. It owns no dialog content
// or behavior — screens compose panels from Frame and whatever widgets the
// dialog needs (buttongroup, textinput, …). Dialog text carries no styling of
// its own, so it matches the rest of the app's default-foreground text.
package dialog

import (
	"strings"

	"github.com/M-Xue/grove/ui/components/panel"
	"github.com/charmbracelet/lipgloss"
)

// Frame wraps content in the dialog's bordered box with title embedded as a
// tab in the top border, panel-style. The box is sized to the content and
// clamped against the available width; the border is always the active color
// because an open dialog owns the keyboard.
func Frame(title, content string, width int) string {
	lines := strings.Split(content, "\n")
	maxTotal := max(28, width-10)
	contentMax := max(1, maxTotal-4)
	contentWidth := 0
	for _, line := range lines {
		contentWidth = max(contentWidth, lipgloss.Width(line))
	}
	contentWidth = min(contentWidth, contentMax)

	// One blank row above and below the content; horizontally the content sits
	// a single space from each border (the panel's own padding column).
	inner := make([]string, 0, len(lines)+2)
	inner = append(inner, "")
	for _, line := range lines {
		if lipgloss.Width(line) > contentWidth {
			line = lipgloss.NewStyle().MaxWidth(contentWidth).Render(line)
		}
		inner = append(inner, line)
	}
	inner = append(inner, "")

	boxWidth := min(maxTotal, max(contentWidth+4, lipgloss.Width(panel.Tab(title))+4))
	return panel.Render(title, strings.Join(inner, "\n"), boxWidth, len(inner)+2, true)
}
