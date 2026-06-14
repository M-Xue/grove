package loading

import "github.com/charmbracelet/lipgloss"

type Model struct{}

func New() Model { return Model{} }

func (Model) View(message string, width int) string {
	if message == "" {
		return ""
	}
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)
	return style.Render("... " + message)
}
