package screens

import (
	"strings"
	"testing"

	"github.com/M-Xue/grove/ui/keys"
)

func TestModeLookupCoversAllKeysOfABinding(t *testing.T) {
	mode := NewMode(
		Binding{Keys: []keys.Key{keys.KeyUp, keys.KeyShiftTab}, Symbol: "↑/shift+tab", Label: "move"},
		Binding{Keys: []keys.Key{keys.KeyEnter}, Symbol: "enter", Label: "open"},
	)
	if b, ok := mode.lookup(keys.KeyUp); !ok || b.Label != "move" {
		t.Fatalf("expected up to map to the move binding, got %#v ok=%v", b, ok)
	}
	if b, ok := mode.lookup(keys.KeyShiftTab); !ok || b.Label != "move" {
		t.Fatalf("expected shift+tab to map to the move binding, got %#v ok=%v", b, ok)
	}
	if _, ok := mode.lookup(keys.KeyCtrlD); ok {
		t.Fatal("did not expect ctrl+d to be mapped")
	}
}

func TestModeFooterDerivesFromBindingsAndSkipsEmptyLabels(t *testing.T) {
	mode := NewMode(
		Binding{Keys: []keys.Key{keys.KeyEnter}, Symbol: "enter", Label: "open"},
		Binding{Keys: []keys.Key{keys.KeyCtrlC}, Symbol: "", Label: ""}, // dispatch-only, hidden
	)
	footer := mode.footer(80)
	if !strings.Contains(footer, "open") || !strings.Contains(footer, "enter") {
		t.Fatalf("expected footer to render the visible binding, got %q", footer)
	}
}

func TestActiveModeSwitchesWithDialog(t *testing.T) {
	s := NewChangeScreen(fakeApp{})
	if s.activeMode() != ModeDefault {
		t.Fatal("expected ModeDefault initially")
	}
	s.confirm.active = true
	if s.activeMode() != ModeDialog {
		t.Fatal("expected ModeDialog when the dialog is active")
	}
}
