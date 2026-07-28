package loading

import (
	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/ui/components/progressbar"
	"github.com/charmbracelet/lipgloss"
)

// spinnerFrames are the Braille glyphs cycled while a loading entry is active.
// Completed entries drop the spinner entirely in favour of the [DONE] suffix.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Model struct {
	frame    int
	progress progressbar.Model
}

func New() Model { return Model{progress: progressbar.New()} }

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
		if entry.Progress {
			// Render the bar outside the loading style so its own cell colours
			// are preserved; a single space separates it from the message.
			bar := m.progress.View(entry.Done, entry.Total)
			lines = append(lines, style.Render(spinner+" "+message)+" "+bar)
			continue
		}
		lines = append(lines, style.Render(spinner+" "+message))
	}
	return lines
}
