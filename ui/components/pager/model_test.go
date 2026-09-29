package pager

import (
	"reflect"
	"testing"
)

func items(n int) []int {
	list := make([]int, n)
	for i := range list {
		list[i] = i + 1
	}
	return list
}

func TestPageCutsFullAndRemainderPages(t *testing.T) {
	var m Model[int]
	m.SetItems(items(12))

	if got := m.Page(5); !reflect.DeepEqual(got, []int{1, 2, 3, 4, 5}) {
		t.Fatalf("expected the first page, got %v", got)
	}
	if got := m.Status(); got != "1-5/12" {
		t.Fatalf("expected status 1-5/12, got %q", got)
	}

	m.Next()
	if got := m.Page(5); !reflect.DeepEqual(got, []int{6, 7, 8, 9, 10}) {
		t.Fatalf("expected the second page, got %v", got)
	}

	// The last page holds just the remainder, not a full backfilled window.
	m.Next()
	if got := m.Page(5); !reflect.DeepEqual(got, []int{11, 12}) {
		t.Fatalf("expected the remainder page, got %v", got)
	}
	if got := m.Status(); got != "11-12/12" {
		t.Fatalf("expected status 11-12/12, got %q", got)
	}

	// Paging past the end stays on the last page.
	m.Next()
	if got := m.Page(5); !reflect.DeepEqual(got, []int{11, 12}) {
		t.Fatalf("expected to stay on the remainder page, got %v", got)
	}
}

func TestPrevStopsAtFirstPage(t *testing.T) {
	var m Model[int]
	m.SetItems(items(7))
	m.Page(3)
	m.Next()
	m.Prev()
	m.Prev()
	if got := m.Page(3); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("expected the first page after over-scrolling back, got %v", got)
	}
}

func TestStatusEmptyWhenListFitsOnePage(t *testing.T) {
	var m Model[int]
	m.SetItems(items(3))
	m.Page(5)
	if got := m.Status(); got != "" {
		t.Fatalf("expected empty status when everything fits, got %q", got)
	}
}

func TestPageRealignsWhenSizeChanges(t *testing.T) {
	var m Model[int]
	m.SetItems(items(10))
	m.Page(4)
	m.Next() // offset 4
	// A smaller budget next render: the offset snaps to the new grid.
	if got := m.Page(3); !reflect.DeepEqual(got, []int{4, 5, 6}) {
		t.Fatalf("expected realignment to the new page grid, got %v", got)
	}
}

func TestPageClampsAfterItemsShrink(t *testing.T) {
	var m Model[int]
	m.SetItems(items(12))
	m.Page(5)
	m.Next()
	m.Next() // offset 10
	m.SetItems(items(4))
	if got := m.Page(5); !reflect.DeepEqual(got, []int{1, 2, 3, 4}) {
		t.Fatalf("expected clamp onto the shrunken list, got %v", got)
	}
}

func TestEmptyAndDegenerateSizes(t *testing.T) {
	var m Model[int]
	if got := m.Page(5); got != nil {
		t.Fatalf("expected nil page for an empty list, got %v", got)
	}
	if got := m.Status(); got != "" {
		t.Fatalf("expected empty status for an empty list, got %q", got)
	}
	m.SetItems(items(2))
	if got := m.Page(0); !reflect.DeepEqual(got, []int{1}) {
		t.Fatalf("expected a size below one to clamp to one, got %v", got)
	}
}
