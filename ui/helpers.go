package ui

import "strings"

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func (m *Model) syncSelection() {
	if len(m.filtered) == 0 {
		m.selectedItem = ""
		m.selected = 0
		m.scroll = 0
		return
	}

	if m.selectedItem != "" {
		for i, item := range m.filtered {
			if item == m.selectedItem {
				m.selected = i
				m.selectedItem = item
				return
			}
		}
	}

	m.selected = 0
	m.selectedItem = m.filtered[0]
}

func (m *Model) refreshFiltered() {
	m.filtered = filterItems(m.items, m.query)
	m.syncSelection()
	if len(m.filtered) == 0 {
		m.scroll = 0
	}
}

func (m *Model) moveSelection(delta int) {
	if len(m.filtered) == 0 {
		m.selected = 0
		m.scroll = 0
		m.selectedItem = ""
		return
	}

	next := m.selected + delta
	if next < 0 {
		next = len(m.filtered) - 1
	}
	if next >= len(m.filtered) {
		next = 0
	}

	m.selected = next
	m.selectedItem = m.filtered[m.selected]
}

func (m *Model) syncScroll(visibleRows int) {
	if visibleRows <= 0 || len(m.filtered) == 0 {
		m.scroll = 0
		return
	}

	maxScroll := max(0, len(m.filtered)-visibleRows)
	if m.selected < m.scroll {
		m.scroll = m.selected
	}
	if m.selected >= m.scroll+visibleRows {
		m.scroll = m.selected - visibleRows + 1
	}
	m.scroll = clamp(m.scroll, 0, maxScroll)
}

func repeatLine(fill string, width int) string {
	if width <= 0 {
		return ""
	}
	return strings.Repeat(fill, width)
}

func fitLine(content string, width int) string {
	if width <= 0 {
		return ""
	}
	if len(content) > width {
		return content[:width]
	}
	return content + strings.Repeat(" ", width-len(content))
}
