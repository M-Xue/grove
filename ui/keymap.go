package ui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	addMode     key.Binding
	docs        key.Binding
	remove      key.Binding
	moveUp      key.Binding
	moveDown    key.Binding
	close       key.Binding
	changeQuit  key.Binding
	addQuit     key.Binding
	submit      key.Binding
	switchField key.Binding
	confirmYes  key.Binding
	confirmNo   key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		addMode: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "add worktree"),
		),
		docs: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "docs"),
		),
		remove: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "remove worktree"),
		),
		moveUp: key.NewBinding(
			key.WithKeys("up", "shift+tab"),
			key.WithHelp("up/shift+tab", "move up"),
		),
		moveDown: key.NewBinding(
			key.WithKeys("down", "tab"),
			key.WithHelp("down/tab", "move down"),
		),
		close: key.NewBinding(
			key.WithKeys("esc", "ctrl+a"),
			key.WithHelp("ctrl+a/esc", "close"),
		),
		changeQuit: key.NewBinding(
			key.WithKeys("esc", "ctrl+c"),
			key.WithHelp("esc/ctrl+c", "quit"),
		),
		addQuit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "submit"),
		),
		switchField: key.NewBinding(
			key.WithKeys("tab", "shift+tab", "up", "down"),
			key.WithHelp("tab", "switch field"),
		),
		confirmYes: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "confirm"),
		),
		confirmNo: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "cancel"),
		),
	}
}
