package ui

func (m *Model) moveDocsScroll(delta int) {
	if len(m.docs.lines) == 0 {
		m.docs.scroll = 0
		return
	}
	m.docs.scroll = max(0, m.docs.scroll+delta)
}

func (m *Model) syncDocsScroll(visibleRows int) {
	if visibleRows <= 0 || len(m.docs.lines) == 0 {
		m.docs.scroll = 0
		return
	}
	maxScroll := max(0, len(m.docs.lines)-visibleRows)
	m.docs.scroll = clamp(m.docs.scroll, 0, maxScroll)
}
