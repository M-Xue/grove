package selectlist

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	selectionColor = "\x1b[38;5;183m"
	resetColor     = "\x1b[0m"
)

type Item struct {
	ID    string
	Label string
	// Color is an optional ANSI escape applied to the label when the item is
	// not the current selection. Empty means default terminal colour.
	Color string
}

type Model struct {
	items      []Item
	selectedID string
	selected   int
	scroll     int
	emptyLabel string
}

func New(emptyLabel string) Model {
	return Model{emptyLabel: emptyLabel}
}

func (m *Model) SetItems(items []Item) {
	m.items = append([]Item(nil), items...)
	m.syncSelection()
}

func (m *Model) SetSelectedID(id string) {
	m.selectedID = id
	m.syncSelection()
}

func (m Model) SelectedID() (string, bool) {
	if len(m.items) == 0 || m.selected < 0 || m.selected >= len(m.items) {
		return "", false
	}
	return m.items[m.selected].ID, true
}

func (m Model) SelectedItem() (Item, bool) {
	if len(m.items) == 0 || m.selected < 0 || m.selected >= len(m.items) {
		return Item{}, false
	}
	return m.items[m.selected], true
}

func (m *Model) Update(msg tea.KeyMsg) (bool, tea.Cmd) {
	if len(m.items) == 0 {
		return false, nil
	}
	switch msg.String() {
	case "up", "shift+tab":
		m.move(-1)
		return true, nil
	case "down", "tab":
		m.move(1)
		return true, nil
	default:
		return false, nil
	}
}

func (m *Model) View(height int) string {
	if height <= 0 {
		height = 1
	}
	m.syncScroll(height)
	if len(m.items) == 0 {
		return m.emptyLabel
	}
	var lines []string
	end := min(len(m.items), m.scroll+height)
	for i := m.scroll; i < end; i++ {
		line := m.items[i].Label
		if i == m.selected {
			line = selectionColor + line + resetColor
		} else if m.items[i].Color != "" {
			line = m.items[i].Color + line + resetColor
		}
		lines = append(lines, line)
	}
	for len(lines) < height {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

func (m *Model) VisibleItems(height int) []Item {
	if height <= 0 {
		height = 1
	}
	m.syncScroll(height)
	if len(m.items) == 0 {
		return nil
	}
	end := min(len(m.items), m.scroll+height)
	return append([]Item(nil), m.items[m.scroll:end]...)
}

func (m *Model) syncSelection() {
	if len(m.items) == 0 {
		m.selected = 0
		m.selectedID = ""
		m.scroll = 0
		return
	}
	if m.selectedID != "" {
		for i, item := range m.items {
			if item.ID == m.selectedID {
				m.selected = i
				return
			}
		}
	}
	m.selected = 0
	m.selectedID = m.items[0].ID
}

func (m *Model) move(delta int) {
	if len(m.items) == 0 {
		return
	}
	next := m.selected + delta
	if next < 0 {
		next = len(m.items) - 1
	}
	if next >= len(m.items) {
		next = 0
	}
	m.selected = next
	m.selectedID = m.items[next].ID
}

func (m *Model) syncScroll(height int) {
	if len(m.items) == 0 {
		m.scroll = 0
		return
	}
	maxScroll := max(0, len(m.items)-height)
	if m.selected < m.scroll {
		m.scroll = m.selected
	}
	if m.selected >= m.scroll+height {
		m.scroll = m.selected - height + 1
	}
	m.scroll = max(0, min(m.scroll, maxScroll))
}

