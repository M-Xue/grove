package keys

import tea "github.com/charmbracelet/bubbletea"

type Key string

const (
	KeyUnknown   Key = ""
	KeyEnter     Key = "enter"
	KeyEsc       Key = "esc"
	KeyCtrlC     Key = "ctrl+c"
	KeyCtrlA     Key = "ctrl+a"
	KeyCtrlD     Key = "ctrl+d"
	KeyQuestion  Key = "?"
	KeyUp        Key = "up"
	KeyDown      Key = "down"
	KeyLeft      Key = "left"
	KeyRight     Key = "right"
	KeyTab       Key = "tab"
	KeyShiftTab  Key = "shift+tab"
	KeyBackspace Key = "backspace"
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
	case "ctrl+d":
		return KeyCtrlD
	case "?":
		return KeyQuestion
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
	case "q":
		return KeyQ
	default:
		return KeyUnknown
	}
}
