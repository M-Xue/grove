package keys

import tea "github.com/charmbracelet/bubbletea"

type Key string

const (
	KeyUnknown    Key = ""
	KeyEnter      Key = "enter"
	KeyEsc        Key = "esc"
	KeyCtrlC      Key = "ctrl+c"
	KeyCtrlA      Key = "ctrl+a"
	KeyCtrlD      Key = "ctrl+d"
	KeyCtrlP      Key = "ctrl+p"
	KeyUp         Key = "up"
	KeyDown       Key = "down"
	KeyLeft       Key = "left"
	KeyRight      Key = "right"
	KeyTab        Key = "tab"
	KeyShiftTab   Key = "shift+tab"
	KeyShiftUp    Key = "shift+up"
	KeyShiftDown  Key = "shift+down"
	KeyShiftLeft  Key = "shift+left"
	KeyShiftRight Key = "shift+right"
	KeyBackspace  Key = "backspace"
	KeyJ          Key = "j"
	KeyK          Key = "k"
	KeyQ          Key = "q"
)

func Normalize(msg tea.KeyMsg) Key {
	switch msg.String() {
	case "enter":
		return KeyEnter
	case "esc":
		return KeyEsc
	case "ctrl+c":
		return KeyCtrlC
	case "ctrl+a":
		return KeyCtrlA
	case "ctrl+d":
		return KeyCtrlD
	case "ctrl+p":
		return KeyCtrlP
	case "up":
		return KeyUp
	case "down":
		return KeyDown
	case "left":
		return KeyLeft
	case "right":
		return KeyRight
	case "tab":
		return KeyTab
	case "shift+tab":
		return KeyShiftTab
	case "shift+up":
		return KeyShiftUp
	case "shift+down":
		return KeyShiftDown
	case "shift+left":
		return KeyShiftLeft
	case "shift+right":
		return KeyShiftRight
	case "backspace":
		return KeyBackspace
	case "j":
		return KeyJ
	case "k":
		return KeyK
	case "q":
		return KeyQ
	default:
		return KeyUnknown
	}
}
