package ui

import "strings"

func filterItems(items []string, query string) []string {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return append([]string(nil), items...)
	}

	filtered := make([]string, 0, len(items))
	for _, item := range items {
		if strings.Contains(strings.ToLower(item), query) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}
