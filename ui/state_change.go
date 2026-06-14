package ui

func (m *Model) syncChangeSelection() {
	if len(m.change.filtered) == 0 {
		m.change.selectedItem = ""
		m.change.selected = 0
		m.change.scroll = 0
		return
	}

	if m.change.selectedItem != "" {
		for i, item := range m.change.filtered {
			if item == m.change.selectedItem {
				m.change.selected = i
				m.change.selectedItem = item
				return
			}
		}
	}

	m.change.selected = 0
	m.change.selectedItem = m.change.filtered[0]
}

func (m *Model) refreshChangeFiltered() {
	m.change.filtered = filterItems(m.change.items, m.change.query)
	m.syncChangeSelection()
	if len(m.change.filtered) == 0 {
		m.change.scroll = 0
	}
}

func (m *Model) moveChangeSelection(delta int) {
	if len(m.change.filtered) == 0 {
		m.change.selected = 0
		m.change.scroll = 0
		m.change.selectedItem = ""
		return
	}

	next := m.change.selected + delta
	if next < 0 {
		next = len(m.change.filtered) - 1
	}
	if next >= len(m.change.filtered) {
		next = 0
	}

	m.change.selected = next
	m.change.selectedItem = m.change.filtered[m.change.selected]
}

func (m *Model) syncChangeScroll(visibleRows int) {
	if visibleRows <= 0 || len(m.change.filtered) == 0 {
		m.change.scroll = 0
		return
	}

	maxScroll := max(0, len(m.change.filtered)-visibleRows)
	if m.change.selected < m.change.scroll {
		m.change.scroll = m.change.selected
	}
	if m.change.selected >= m.change.scroll+visibleRows {
		m.change.scroll = m.change.selected - visibleRows + 1
	}
	m.change.scroll = clamp(m.change.scroll, 0, maxScroll)
}

func (m *Model) syncChangeItemsFromWorktrees() {
	m.change.items = make([]string, 0, len(m.change.worktrees))
	for _, item := range m.change.worktrees {
		m.change.items = append(m.change.items, item.Path)
	}
	m.refreshChangeFiltered()
}

func (m *Model) cancelRemoveWorktree() {
	m.change.confirmRemove = false
	m.change.confirmPath = ""
	m.clearStatus()
	m.clearError()
}
