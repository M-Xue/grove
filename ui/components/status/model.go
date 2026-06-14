package status

import (
	"fmt"

	"github.com/M-Xue/grove/app"
	"github.com/charmbracelet/lipgloss"
)

const (
	infoColor    = "\x1b[38;5;117m"
	successColor = "\x1b[38;5;149m"
	errorColor   = "\x1b[38;5;203m"
	resetColor   = "\x1b[0m"
)

type Model struct{}

func New() Model { return Model{} }

func (Model) View(items []app.StatusEntry, width int) []string {
	lines := make([]string, 0, len(items))
	for _, item := range items {
		color := infoColor
		borderColor := lipgloss.Color("117")
		label := "status"
		switch item.Kind {
		case app.StatusSuccess:
			color = successColor
			borderColor = lipgloss.Color("149")
			label = "success"
		case app.StatusError:
			color = errorColor
			borderColor = lipgloss.Color("203")
			label = "error"
		}
		line := fmt.Sprintf("%s[%s]%s %s", color, label, resetColor, item.Message)
		if width > 0 {
			line = lipgloss.NewStyle().MaxWidth(width).Render(line)
		}
		lines = append(lines, lipgloss.NewStyle().BorderLeft(true).BorderForeground(borderColor).PaddingLeft(1).Render(line))
	}
	return lines
}
