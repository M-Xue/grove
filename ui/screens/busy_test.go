package screens

import (
	"strings"
	"testing"

	"github.com/M-Xue/grove/app"
)

// The focused text input renders a "> " caret prefix; a disabled one does not.
// Its absence is a proxy for the input being disabled while an op is in flight.

func TestAddScreenDisablesInputsWhenBusy(t *testing.T) {
	s := NewAddScreen(fakeApp{})
	idle := app.State{}
	s.Sync(idle)
	if !strings.Contains(s.View(80, 24, idle), "> ") {
		t.Fatal("expected focus caret on the add screen when idle")
	}
	busy := app.State{Loading: []app.LoadingEntry{{Blocking: true}}}
	s.Sync(busy)
	if strings.Contains(s.View(80, 24, busy), "> ") {
		t.Fatal("expected add screen inputs disabled (no caret) while busy")
	}
}

func TestBranchScreenDisablesSearchWhenBusy(t *testing.T) {
	s := NewBranchScreen(fakeApp{})
	idle := app.State{}
	s.Sync(idle)
	if !strings.Contains(s.View(80, 24, idle), "> ") {
		t.Fatal("expected focus caret on the branch search when idle")
	}
	busy := app.State{Loading: []app.LoadingEntry{{Blocking: true}}}
	s.Sync(busy)
	if strings.Contains(s.View(80, 24, busy), "> ") {
		t.Fatal("expected branch search disabled (no caret) while busy")
	}
}
