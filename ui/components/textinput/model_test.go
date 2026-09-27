package textinput

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// forceColorProfile makes lipgloss emit colors even though tests run without a
// tty, so view assertions can see the caret and field background.
func forceColorProfile(t *testing.T) {
	t.Helper()
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(previous) })
}

// caretMarker is the reverse-video attribute the caret renders with.
const caretMarker = "\x1b[7;"

func TestUpdateAddsASCIIInput(t *testing.T) {
	m := New("placeholder")
	m.Focus()
	consumed, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a', 'b'}})
	if !consumed || m.Value() != "ab" {
		t.Fatalf("unexpected value: %q", m.Value())
	}
}

func TestUpdateIgnoresAltModifiedRunes(t *testing.T) {
	m := New("placeholder")
	m.Focus()
	consumed, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}, Alt: true})
	if consumed || m.Value() != "" {
		t.Fatalf("expected alt+rune to be ignored, got consumed=%v value=%q", consumed, m.Value())
	}
}

func TestUpdateBackspaceRemovesCharacter(t *testing.T) {
	m := New("placeholder")
	m.Focus()
	m.SetValue("ab")
	m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if m.Value() != "a" {
		t.Fatalf("unexpected value: %q", m.Value())
	}
}

func TestDisabledUpdateIgnoresInput(t *testing.T) {
	m := New("placeholder")
	m.Focus()
	m.SetValue("ab")
	m.SetDisabled(true)
	consumed, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if consumed || cmd != nil {
		t.Fatalf("expected disabled input to ignore keys, got consumed=%v cmd=%v", consumed, cmd)
	}
	if m.Value() != "ab" {
		t.Fatalf("expected value unchanged, got %q", m.Value())
	}
}

func TestDisabledViewIsInert(t *testing.T) {
	forceColorProfile(t)
	m := New("placeholder")
	m.Focus()
	m.SetValue("branch-name")
	m.SetDisabled(true)
	view := m.View()
	if strings.Contains(view, "> ") {
		t.Fatalf("expected no focus prefix in disabled view, got %q", view)
	}
	if strings.Contains(view, caretMarker) {
		t.Fatalf("expected no caret in disabled view, got %q", view)
	}
	if !strings.Contains(view, "branch-name") {
		t.Fatalf("expected value to still render, got %q", view)
	}
}

func TestClearingDisabledRestoresCaret(t *testing.T) {
	forceColorProfile(t)
	m := New("placeholder")
	m.Focus()
	m.SetValue("branch-name")
	m.SetDisabled(true)
	m.SetDisabled(false)
	if !strings.Contains(m.View(), caretMarker) {
		t.Fatalf("expected caret to return once re-enabled, got %q", m.View())
	}
}

func TestViewRendersFieldBackground(t *testing.T) {
	forceColorProfile(t)
	m := New("placeholder")
	// 48;5;… is the SGR background attribute; the exact code depends on how
	// the profile downsamples fieldBg, so only the attribute is asserted.
	if !strings.Contains(m.View(), "48;5;") {
		t.Fatalf("expected the field background in the view, got %q", m.View())
	}
}

func TestSetWidthPadsFieldToWidth(t *testing.T) {
	m := New("placeholder")
	m.SetWidth(40)
	if got := lipgloss.Width(m.View()); got != 40 {
		t.Fatalf("expected a 40-cell field, got %d", got)
	}
	m.SetValue("hi")
	if got := lipgloss.Width(m.View()); got != 40 {
		t.Fatalf("expected a 40-cell field with a value, got %d", got)
	}
}

func TestViewScrollsWithCursorBeyondWidth(t *testing.T) {
	forceColorProfile(t)
	m := New("")
	m.Focus()
	m.SetWidth(5)
	m.SetValue("abcdefgh")

	// Cursor is at the end: the window shows the tail with the caret cell.
	view := stripAnsi(m.View())
	if view != "efgh " {
		t.Fatalf("expected tail window %q, got %q", "efgh ", view)
	}

	// Walking left keeps the window until the cursor hits its left edge, then
	// scrolls back to reveal earlier text.
	for i := 0; i < 6; i++ {
		m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	}
	view = stripAnsi(m.View())
	if !strings.HasPrefix(view, "cdefg") {
		t.Fatalf("expected window scrolled back to %q..., got %q", "cdefg", view)
	}

	// Home reveals the head of the value again.
	m.Update(tea.KeyMsg{Type: tea.KeyHome})
	view = stripAnsi(m.View())
	if !strings.HasPrefix(view, "abcde") {
		t.Fatalf("expected head window %q..., got %q", "abcde", view)
	}
}

func TestViewClipsBlurredAndDisabledValues(t *testing.T) {
	forceColorProfile(t)
	m := New("")
	m.Focus()
	m.SetWidth(5)
	m.SetValue("abcdefgh")

	m.Blur()
	if view := stripAnsi(m.View()); view != "abcde" {
		t.Fatalf("expected blurred field clipped to %q, got %q", "abcde", view)
	}

	m.SetDisabled(true)
	if view := stripAnsi(m.View()); view != "abcde" {
		t.Fatalf("expected disabled field clipped to %q, got %q", "abcde", view)
	}
}

var ansiSequence = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripAnsi(s string) string {
	return ansiSequence.ReplaceAllString(s, "")
}
