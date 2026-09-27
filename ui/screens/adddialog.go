package screens

import (
	"os"
	"strings"

	"github.com/M-Xue/grove/ui/components/dialog"
	"github.com/M-Xue/grove/ui/components/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// addDialogInteriorWidth is the width the dialog's rows are padded to, so the
// panel keeps a stable size while the user types. With the frame's border and
// padding columns (4 in total) the dialog box renders 80 cells wide.
const addDialogInteriorWidth = 76

// addDialogLabelWidth is the column the field labels are padded to, so the
// input fields start aligned. Sized to the widest label, "Relative path".
const addDialogLabelWidth = 13

// addDialog is the change screen's worktree-creation dialog: two text inputs
// (path and branch) rendered in the shared dialog frame under a header showing
// the directory the path input is relative to. The screen owns when it opens
// and what happens on submit; the dialog owns only input state.
type addDialog struct {
	path        textinput.Model
	branch      textinput.Model
	focusedPath bool
	active      bool
	// currentPath is the working directory grove was launched from, shown so
	// the user knows what the relative path input resolves against. Empty when
	// it cannot be determined.
	currentPath string
}

func newAddDialog() addDialog {
	d := addDialog{
		path:   textinput.New(""),
		branch: textinput.New(""),
	}
	d.currentPath, _ = os.Getwd()
	fieldWidth := addDialogInteriorWidth - addDialogLabelWidth - 1
	d.path.SetWidth(fieldWidth)
	d.branch.SetWidth(fieldWidth)
	return d
}

// open resets both inputs and shows the dialog with the path field focused.
func (d *addDialog) open() {
	d.path.Clear()
	d.branch.Clear()
	d.focusedPath = true
	d.path.Focus()
	d.branch.Blur()
	d.active = true
}

func (d *addDialog) close() {
	d.active = false
}

// switchFocus moves focus between the path and branch fields.
func (d *addDialog) switchFocus() {
	d.focusedPath = !d.focusedPath
	if d.focusedPath {
		d.path.Focus()
		d.branch.Blur()
		return
	}
	d.path.Blur()
	d.branch.Focus()
}

// handleKey feeds a keystroke to the focused input, reporting whether the
// input consumed it.
func (d *addDialog) handleKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	if d.focusedPath {
		return d.path.Update(msg)
	}
	return d.branch.Update(msg)
}

// values returns the trimmed path and branch entries.
func (d *addDialog) values() (string, string) {
	return strings.TrimSpace(d.path.Value()), strings.TrimSpace(d.branch.Value())
}

func (d *addDialog) view(width, height int) string {
	lines := labeledRows("Current path", d.currentPath, addDialogLabelWidth, addDialogInteriorWidth)
	lines = append(lines,
		"",
		labeledField("Relative path", d.path.View()),
		"",
		labeledField("Branch", d.branch.View()),
	)
	return dialog.Frame("Add worktree", strings.Join(lines, "\n"), width)
}

// labeledField renders a label column followed by an input field, labels
// styled like the detail panel's titles and padded to a common width so the
// fields start aligned.
func labeledField(label, field string) string {
	return fitLine(detailLabel(label), addDialogLabelWidth) + " " + field
}
