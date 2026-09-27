// Package theme centralizes the colors shared by grove's chrome (panels and
// dialog frames). Component-local accents still live in their components; the
// full migration into this package is tracked in architecture-review.md.
package theme

import "github.com/charmbracelet/lipgloss"

var (
	// BorderActive outlines the box that owns the keyboard: an open dialog,
	// otherwise the focused panel.
	BorderActive = lipgloss.Color("111")
	// BorderInactive is the muted slate blue for every other section border.
	BorderInactive = lipgloss.Color("60")
	// TitleFg/TitleBg render section titles as a light periwinkle tab with
	// dark text.
	TitleFg = lipgloss.Color("235")
	TitleBg = lipgloss.Color("146")
)
