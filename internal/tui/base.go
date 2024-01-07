package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Base model that handles logic common to all views

type BaseModel struct {
}

func NewBaseModel() *BaseModel {
	return &BaseModel{}
}

func (m *BaseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keymap.Quit):
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m BaseModel) Init() tea.Cmd {
	return nil
}

func (m BaseModel) View() string {
	return ""
}
