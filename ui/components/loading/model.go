package loading

import (
	"github.com/M-Xue/grove/app"
	"github.com/charmbracelet/lipgloss"
)

type Model struct{}

func New() Model { return Model{} }

func (Model) View(entries []app.LoadingEntry, width int) []string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)
	lines := make([]string, 0, len(entries))
	for _, entry := range entries {
		message := entry.Message
		if message == "" {
			continue
		}
		if entry.Completed {
			message += " [DONE]"
		}
		lines = append(lines, style.Render("... "+message))
	}
	return lines
}
