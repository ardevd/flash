package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Close           key.Binding
	ForceClose      key.Binding
	Update          key.Binding
	Enter           key.Binding
	Refresh         key.Binding
	Delete          key.Binding
	Back            key.Binding
	Quit            key.Binding
	Left            key.Binding
	Right           key.Binding
	Tab             key.Binding
	ReverseTab      key.Binding
	Help            key.Binding
	OfflineChannels key.Binding
}

// Keymap reusable key mappings shared across models
var Keymap = keyMap{
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Close: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "close"),
	),
	ForceClose: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "force close"),
	),

	OfflineChannels: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "offline channels"),
	),
	Update: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "update"),
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
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
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
