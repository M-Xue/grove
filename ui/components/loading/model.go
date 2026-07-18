package loading

import (
	"github.com/M-Xue/grove/app"
	"github.com/charmbracelet/lipgloss"
)

// spinnerFrames are the Braille glyphs cycled while a loading entry is active.
// Completed entries drop the spinner entirely in favour of the [DONE] suffix.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Model struct {
	frame int
}

func New() Model { return Model{} }

// Tick advances the spinner to its next frame. The UI drives this on a timer
// while any loading entry is still active.
func (m *Model) Tick() {
	m.frame = (m.frame + 1) % len(spinnerFrames)
}

func (m Model) View(entries []app.LoadingEntry, width int) []string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("117")).Bold(true)
	lines := make([]string, 0, len(entries))
	for _, entry := range entries {
		message := entry.Message
		if message == "" {
			continue
		}
		if entry.Completed {
			lines = append(lines, style.Render(message+" [DONE]"))
			continue
		}
		spinner := spinnerFrames[m.frame%len(spinnerFrames)]
		lines = append(lines, style.Render(spinner+" "+message))
	}
	return lines
}
