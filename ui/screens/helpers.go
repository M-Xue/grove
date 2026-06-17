package screens

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func overlayDialog(base, overlay string, width, height int) string {
	if overlay == "" {
		return base
	}
	baseCanvas := fitCanvas(base, width, height)
	overlayLines := strings.Split(overlay, "\n")
	overlayHeight := len(overlayLines)
	overlayWidth := 0
	for _, line := range overlayLines {
		overlayWidth = max(overlayWidth, lipgloss.Width(line))
	}
	row := max(0, (height-overlayHeight)/2)
	col := max(0, (width-overlayWidth)/2)

	for i, line := range overlayLines {
		targetRow := row + i
		if targetRow < 0 || targetRow >= len(baseCanvas) {
			continue
		}
		baseCanvas[targetRow] = overlayLine(baseCanvas[targetRow], line, col, width)
	}

	result := strings.Join(baseCanvas, "\n")
	return result
}

func fitCanvas(content string, width, height int) []string {
	lines := strings.Split(content, "\n")
	canvas := make([]string, 0, max(height, 0))
	for i := 0; i < height; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		canvas = append(canvas, fitLine(line, width))
	}
	return canvas
}

func fitLine(content string, width int) string {
	if width <= 0 {
		return ""
	}
	contentWidth := lipgloss.Width(content)
	if contentWidth > width {
		return lipgloss.NewStyle().MaxWidth(width).Render(content)
	}
	return content + strings.Repeat(" ", width-contentWidth)
}

func overlayLine(base, overlay string, left, width int) string {
	if width <= 0 {
		return ""
	}
	left = max(0, left)
	base = fitLine(base, width)
	overlayWidth := lipgloss.Width(overlay)
	if overlayWidth <= 0 || left >= width {
		return base
	}
	if left+overlayWidth > width {
		overlay = ansi.Cut(overlay, 0, width-left)
		overlayWidth = lipgloss.Width(overlay)
	}
	prefix := ansi.Cut(base, 0, left)
	suffix := ansi.Cut(base, left+overlayWidth, width)
	return prefix + overlay + suffix
}
