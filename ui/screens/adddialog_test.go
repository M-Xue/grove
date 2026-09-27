package screens

import (
	"strings"
	"testing"

	"github.com/M-Xue/grove/app"
	tea "github.com/charmbracelet/bubbletea"
)

// addRecorderApp records RequestAddWorktree calls and lets a test choose
// whether the request is accepted (returns a command) or rejected (nil).
type addRecorderApp struct {
	fakeApp
	requested [][2]string
	accept    bool
}

func (a *addRecorderApp) RequestAddWorktree(path, branch string) app.Command {
	a.requested = append(a.requested, [2]string{path, branch})
	if !a.accept {
		return nil
	}
	return func() app.Message { return nil }
}

func testCtx() *ScreenContext {
	return &ScreenContext{Run: func(app.Command) tea.Cmd { return nil }}
}

func keyMsg(t tea.KeyType) tea.KeyMsg { return tea.KeyMsg{Type: t} }

func runeMsg(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestCtrlAOpensAddDialog(t *testing.T) {
	s := NewChangeScreen(fakeApp{})
	s.Update(testCtx(), keyMsg(tea.KeyCtrlA), app.State{})
	if !s.addDlg.active {
		t.Fatal("expected the add dialog to open on ctrl+a")
	}
	if s.activeMode() != ModeAdd {
		t.Fatal("expected ModeAdd while the add dialog is open")
	}

	view := s.View(120, 40, app.State{})
	if !strings.Contains(view, "Add worktree") {
		t.Fatalf("expected the add dialog title in the view, got:\n%s", view)
	}
	if !strings.Contains(view, "Current path") || !strings.Contains(view, "Relative path") || !strings.Contains(view, "Branch") {
		t.Fatalf("expected both field labels in the view, got:\n%s", view)
	}
}

func TestAddDialogFooterReflectsDialogKeymap(t *testing.T) {
	s := NewChangeScreen(fakeApp{})
	s.addDlg.open()
	footer := s.Footer(120)
	for _, hint := range []string{"submit", "switch field", "cancel", "quit"} {
		if !strings.Contains(footer, hint) {
			t.Fatalf("expected footer hint %q while the add dialog is open, got %q", hint, footer)
		}
	}
	if strings.Contains(footer, "remove") || strings.Contains(footer, "prune") {
		t.Fatalf("expected change-screen hints hidden while the dialog is open, got %q", footer)
	}
}

func TestAddDialogRoutesTypingAndFieldSwitch(t *testing.T) {
	s := NewChangeScreen(fakeApp{})
	s.addDlg.open()
	ctx := testCtx()

	s.Update(ctx, runeMsg("wt"), app.State{})
	s.Update(ctx, keyMsg(tea.KeyTab), app.State{})
	s.Update(ctx, runeMsg("feat"), app.State{})

	path, branch := s.addDlg.values()
	if path != "wt" || branch != "feat" {
		t.Fatalf("expected typing routed per focused field, got path=%q branch=%q", path, branch)
	}
}

func TestAddDialogTypingDoesNotLeakIntoSearch(t *testing.T) {
	s := NewChangeScreen(fakeApp{})
	s.addDlg.open()
	s.Update(testCtx(), runeMsg("x"), app.State{})
	if s.search.Value() != "" {
		t.Fatalf("expected search untouched while the dialog is open, got %q", s.search.Value())
	}
}

func TestAddDialogEscCancels(t *testing.T) {
	s := NewChangeScreen(fakeApp{})
	s.addDlg.open()
	s.Update(testCtx(), keyMsg(tea.KeyEsc), app.State{})
	if s.addDlg.active {
		t.Fatal("expected esc to close the add dialog")
	}
	if s.activeMode() != ModeDefault {
		t.Fatal("expected ModeDefault after cancelling")
	}
}

func TestAddDialogSubmitClosesOnlyWhenAccepted(t *testing.T) {
	rejecting := &addRecorderApp{accept: false}
	s := NewChangeScreen(rejecting)
	s.addDlg.open()
	s.addDlg.path.SetValue("  wt  ")
	if cmd := s.actionSubmitAdd(&ActionCtx{}); cmd != nil {
		t.Fatal("expected nil command from a rejected submit")
	}
	if !s.addDlg.active {
		t.Fatal("expected the dialog to stay open after a rejected submit")
	}
	if len(rejecting.requested) != 1 || rejecting.requested[0] != [2]string{"wt", ""} {
		t.Fatalf("expected trimmed values passed through, got %#v", rejecting.requested)
	}

	accepting := &addRecorderApp{accept: true}
	s = NewChangeScreen(accepting)
	s.addDlg.open()
	s.addDlg.path.SetValue("wt")
	s.addDlg.branch.SetValue("feature")
	if cmd := s.actionSubmitAdd(&ActionCtx{}); cmd == nil {
		t.Fatal("expected a command from an accepted submit")
	}
	if s.addDlg.active {
		t.Fatal("expected the dialog to close on an accepted submit")
	}
}

func TestAddDialogReopensCleared(t *testing.T) {
	s := NewChangeScreen(fakeApp{})
	s.addDlg.open()
	s.addDlg.path.SetValue("wt")
	s.addDlg.branch.SetValue("feature")
	s.addDlg.close()
	s.addDlg.open()
	path, branch := s.addDlg.values()
	if path != "" || branch != "" {
		t.Fatalf("expected cleared inputs on reopen, got path=%q branch=%q", path, branch)
	}
	if !s.addDlg.focusedPath {
		t.Fatal("expected the path field focused on reopen")
	}
}

func TestChangeScreenOpensCreateDialogOnBranchAbsent(t *testing.T) {
	s := NewChangeScreen(fakeApp{})
	if s.confirm.active {
		t.Fatal("dialog should start closed")
	}
	s.OnMessage(testCtx(), app.BranchAbsentMessage{Path: "../feature", Branch: "feature"})
	if !s.confirm.active {
		t.Fatal("expected the create-branch dialog to open on BranchAbsentMessage")
	}
	if s.activeMode() != ModeDialog {
		t.Fatal("expected the confirm dialog's mode to take precedence")
	}
}

func TestChangeScreenIgnoresUnrelatedMessages(t *testing.T) {
	s := NewChangeScreen(fakeApp{})
	s.OnMessage(testCtx(), app.BranchExistsMessage{Path: "../feature", Branch: "feature"})
	if s.confirm.active {
		t.Fatal("did not expect a dialog for BranchExistsMessage")
	}
}
