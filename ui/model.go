package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	width        int
	height       int
	query        string
	selectedItem string
	selected     int
	scroll       int
	items        []string
	filtered     []string
}

func New(items []string) Model {
	m := Model{
		items: append([]string(nil), items...),
	}
	m.refreshFiltered()
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "up", "shift+tab":
			m.moveSelection(-1)
		case "down", "tab":
			m.moveSelection(1)
		case "backspace":
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
				m.refreshFiltered()
			}
		case "enter", "left", "right":
			// Reserved for future navigation.
		default:
			if msg.Type == tea.KeyRunes {
				m.query += msg.String()
				m.refreshFiltered()
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m Model) View() string {
	return renderView(m)
}
