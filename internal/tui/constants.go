package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

type keymap struct {
	Create     key.Binding
	Enter      key.Binding
	Rename     key.Binding
	Delete     key.Binding
	Back       key.Binding
	Quit       key.Binding
	Left       key.Binding
	Right      key.Binding
	Tab        key.Binding
	ReverseTab key.Binding
}

// Keymap reusable key mappings shared across models
var Keymap = keymap{
	Create: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "create"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next component"),
	),
	ReverseTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "previous component"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Rename: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "rename"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "q"),
		key.WithHelp("ctrl+c/q", "quit"),
	),
	Left: key.NewBinding(
		key.WithKeys("left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right"),
	),
}
