package keys

import tea "github.com/charmbracelet/bubbletea"

type Key string

const (
	KeyUnknown   Key = ""
	KeyEnter     Key = "enter"
	KeyEsc       Key = "esc"
	KeyCtrlC     Key = "ctrl+c"
	KeyCtrlA     Key = "ctrl+a"
	KeyCtrlB     Key = "ctrl+b"
	KeyCtrlD     Key = "ctrl+d"
	KeyCtrlShiftD Key = "ctrl+shift+d"
	KeyCtrlF     Key = "ctrl+f"
	KeyCtrlO     Key = "ctrl+o"
	KeyCtrlP     Key = "ctrl+p"
	KeyUp        Key = "up"
	KeyDown      Key = "down"
	KeyLeft      Key = "left"
	KeyRight     Key = "right"
	KeyTab       Key = "tab"
	KeyShiftTab  Key = "shift+tab"
	KeyBackspace Key = "backspace"
	KeyJ         Key = "j"
	KeyK         Key = "k"
	KeyQ         Key = "q"
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
	case "ctrl+b":
		return KeyCtrlB
	case "ctrl+d":
		return KeyCtrlD
	case "ctrl+shift+d":
		return KeyCtrlShiftD
	case "ctrl+f":
		return KeyCtrlF
	case "ctrl+o":
		return KeyCtrlO
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
