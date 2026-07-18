package dialog

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

// The dialog renders directly over the terminal's own background so it blends
// with the surroundings, and is set apart only by a border drawn in the same
// color as its text. SetTerminalBackground derives the text/border color from the
// detected background; until it is called (or if detection fails) these light/dark
// adaptive values are used as a fallback. The button accent colors are ANSI base
// colors (0-15), so they follow the user's colorscheme palette.
var (
	dialogTextColor     lipgloss.TerminalColor = lipgloss.AdaptiveColor{Light: "235", Dark: "15"}
	buttonSelectedColor                        = lipgloss.Color("183")
	buttonMutedColor                           = lipgloss.Color("245")
)

// SetTerminalBackground chooses the dialog's text and border color for contrast
// against the terminal's actual background color, given as a "#rrggbb" hex string:
// white on a dark background, black on a light one. An empty or unparseable hex is
// ignored, leaving the adaptive fallback in place. Call this once at startup,
// before the TUI begins reading input.
func SetTerminalBackground(hex string) {
	base, err := colorful.Hex(hex)
	if err != nil {
		return
	}
	_, _, l := base.Hsl()
	if l < 0.5 {
		dialogTextColor = lipgloss.Color("#ffffff")
	} else {
		dialogTextColor = lipgloss.Color("#000000")
	}
}

type Button struct {
	ID    string
	Label string
}

type Model struct {
	title       string
	description string
	buttons     []Button
	focusedID   string
}

func New() Model                             { return Model{} }
func (m *Model) SetTitle(value string)       { m.title = value }
func (m *Model) SetDescription(value string) { m.description = value }
func (m *Model) SetButtons(buttons []Button) {
	m.buttons = append([]Button(nil), buttons...)
	m.syncFocus()
}
func (m *Model) SetFocusedID(id string) { m.focusedID = id; m.syncFocus() }
func (m Model) FocusedID() (string, bool) {
	for _, button := range m.buttons {
		if button.ID == m.focusedID {
			return button.ID, true
		}
	}
	return "", false
}

func (m *Model) Update(msg tea.KeyMsg) (bool, tea.Cmd) {
	if len(m.buttons) == 0 {
		return false, nil
	}
	switch msg.String() {
	case "shift+tab":
		m.move(-1)
		return true, nil
	case "tab":
		m.move(1)
		return true, nil
	default:
		return false, nil
	}
}

func (m Model) View(width, height int) string {
	textStyle := lipgloss.NewStyle().Foreground(dialogTextColor)
	selectedStyle := lipgloss.NewStyle().Foreground(buttonSelectedColor)
	mutedStyle := lipgloss.NewStyle().Foreground(buttonMutedColor)

	buttonLabels := make([]string, 0, len(m.buttons))
	for _, button := range m.buttons {
		label := mutedStyle.Render("[ ] " + button.Label)
		if button.ID == m.focusedID {
			label = selectedStyle.Render("[x] " + button.Label)
		}
		buttonLabels = append(buttonLabels, label)
	}
	body := []string{textStyle.Render(m.title), "", textStyle.Render(m.description), "", strings.Join(buttonLabels, "  ")}
	content := strings.Join(body, "\n")
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(dialogTextColor).
		Padding(1, 2).
		MaxWidth(max(28, width-10))
	return style.Render(content)
}

func (m *Model) syncFocus() {
	if len(m.buttons) == 0 {
		m.focusedID = ""
		return
	}
	for _, button := range m.buttons {
		if button.ID == m.focusedID {
			return
		}
	}
	m.focusedID = m.buttons[0].ID
}

func (m *Model) move(delta int) {
	if len(m.buttons) == 0 {
		return
	}
	index := 0
	for i, button := range m.buttons {
		if button.ID == m.focusedID {
			index = i
			break
		}
	}
	index += delta
	if index < 0 {
		index = len(m.buttons) - 1
	}
	if index >= len(m.buttons) {
		index = 0
	}
	m.focusedID = m.buttons[index].ID
}
