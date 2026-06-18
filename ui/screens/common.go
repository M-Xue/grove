package screens

import (
	"github.com/M-Xue/grove/app"
	"github.com/M-Xue/grove/ui/components/dialog"
	"github.com/M-Xue/grove/ui/keys"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// ScreenContext is the per-keypress context the model hands to a screen.
type ScreenContext struct {
	App *app.App
	Run func(app.Command) tea.Cmd
}

// ActionCtx is the context handed to a single Action when its key fires.
type ActionCtx struct {
	App *app.App
	Key tea.KeyMsg
}

// Action mutates screen-ephemeral state and/or returns an app.Command. It
// returns nil when nothing async is needed (a re-render still happens).
type Action func(*ActionCtx) app.Command

// Binding maps one or more keys to a single action and one footer hint. Keys is
// decoupled from Symbol so one entry can cover multiple triggers (e.g. up and
// shift+tab) yet render a single hint. An empty Label hides it from the footer.
type Binding struct {
	Keys   []keys.Key
	Symbol string
	Label  string
	Action Action
}

// Mode identifies which set of bindings is active on a screen.
type Mode int

const (
	ModeDefault Mode = iota
	ModeDialog
)

// ModeBindings is an ordered list of bindings plus a derived key→binding index.
// The ordered slice drives the footer; the index drives O(1) dispatch. Both
// come from one source, so they cannot drift.
type ModeBindings struct {
	bindings []Binding
	index    map[keys.Key]*Binding
}

func NewMode(bindings ...Binding) ModeBindings {
	mb := ModeBindings{bindings: bindings, index: make(map[keys.Key]*Binding, len(bindings))}
	for i := range mb.bindings {
		for _, k := range mb.bindings[i].Keys {
			mb.index[k] = &mb.bindings[i]
		}
	}
	return mb
}

func (mb ModeBindings) lookup(k keys.Key) (*Binding, bool) {
	b, ok := mb.index[k]
	return b, ok
}

// footer renders the mode's ordered bindings as a single short-help line,
// skipping entries with an empty label.
func (mb ModeBindings) footer(width int) string {
	model := NewHelpModel()
	model.Width = max(width, 0)
	bindings := make([]key.Binding, 0, len(mb.bindings))
	for _, b := range mb.bindings {
		if b.Label == "" {
			continue
		}
		keyStrings := make([]string, 0, len(b.Keys))
		for _, k := range b.Keys {
			keyStrings = append(keyStrings, string(k))
		}
		bindings = append(bindings, key.NewBinding(key.WithKeys(keyStrings...), key.WithHelp(b.Symbol, b.Label)))
	}
	return model.ShortHelpView(bindings)
}

// Registry is a screen's full set of modes.
type Registry map[Mode]ModeBindings

// confirmDialog is a screen-owned confirmation dialog. The screen authors the
// title, description, and button labels (presentation) and supplies the Action
// to run when the user confirms. app owns no dialog state.
type confirmDialog struct {
	model     dialog.Model
	active    bool
	onConfirm Action
}

// open configures and shows the dialog with a "confirm" and a "cancel" button.
// focusConfirm selects which button is focused initially.
func (d *confirmDialog) open(title, description, confirmLabel string, focusConfirm bool, onConfirm Action) {
	d.model = dialog.New()
	d.model.SetTitle(title)
	d.model.SetDescription(description)
	d.model.SetButtons([]dialog.Button{{ID: "confirm", Label: confirmLabel}, {ID: "cancel", Label: "Cancel"}})
	focus := "cancel"
	if focusConfirm {
		focus = "confirm"
	}
	d.model.SetFocusedID(focus)
	d.onConfirm = onConfirm
	d.active = true
}

func (d *confirmDialog) close() {
	d.active = false
	d.onConfirm = nil
}

// move forwards a focus-changing key (tab/shift+tab) to the dialog renderer.
func (d *confirmDialog) move(msg tea.KeyMsg) {
	d.model.Update(msg)
}

// confirm resolves the dialog: it runs the stored action when the confirm
// button is focused, or simply closes when cancel is focused.
func (d *confirmDialog) confirm(actx *ActionCtx) app.Command {
	id, _ := d.model.FocusedID()
	onConfirm := d.onConfirm
	d.close()
	if id == "cancel" || onConfirm == nil {
		return nil
	}
	return onConfirm(actx)
}

func (d *confirmDialog) view(width, height int) string {
	return d.model.View(width, height)
}

func NewHelpModel() help.Model {
	model := help.New()
	model.ShowAll = false
	return model
}
