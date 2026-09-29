// Package pager owns paged-window state over a list of items: a screen feeds
// it the full list and a page size, cuts the visible page at render time, and
// mutates the page position directly when it receives a paging keybind. Like
// the other components it is pure state — it renders nothing and knows nothing
// about what the items are; callers draw the page and the range indicator.
package pager

import "fmt"

// Model pages over items. The zero value is an empty pager on its first page.
// The page size is not fixed at construction: every Page call passes the size
// for that render, so callers whose space budget changes per frame (or per
// terminal resize) stay correct — Next and Prev step by whatever size the last
// Page call used.
type Model[T any] struct {
	items []T
	// offset is the index of the first visible item. Next/Prev move it by the
	// last page size; Page re-aligns it to the current size's page grid and
	// clamps it into range, so the last page shows just the remainder items.
	offset   int
	pageSize int
}

// SetItems replaces the list. The page position is kept (Page clamps it);
// callers switching to a logically different list should also call Reset.
func (m *Model[T]) SetItems(items []T) {
	m.items = items
}

// Reset returns to the first page.
func (m *Model[T]) Reset() {
	m.offset = 0
}

// Page records size as the current page size and returns the visible page:
// the offset is clamped into range and aligned to the size's page grid, so a
// list that doesn't divide evenly ends on a short remainder page. Sizes below
// one are treated as one.
func (m *Model[T]) Page(size int) []T {
	m.pageSize = max(1, size)
	if len(m.items) == 0 {
		m.offset = 0
		return nil
	}
	m.offset = min(m.offset, len(m.items)-1) / m.pageSize * m.pageSize
	return m.items[m.offset:min(len(m.items), m.offset+m.pageSize)]
}

// Next moves one page forward. Only the lower bound is owned here; the upper
// bound depends on the item count and page size at render time, which Page
// clamps against.
func (m *Model[T]) Next() {
	m.offset += max(1, m.pageSize)
}

// Prev moves one page back, stopping at the first page.
func (m *Model[T]) Prev() {
	m.offset = max(0, m.offset-max(1, m.pageSize))
}

// Status describes the page Page last cut as "first-last/total" (1-based), or
// "" when the whole list fits on one page — callers show it only when there is
// something to page to.
func (m Model[T]) Status() string {
	if len(m.items) <= m.pageSize {
		return ""
	}
	end := min(len(m.items), m.offset+m.pageSize)
	return fmt.Sprintf("%d-%d/%d", m.offset+1, end, len(m.items))
}
