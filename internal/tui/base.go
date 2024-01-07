package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Base model that handles logic common to all views
type BaseModel struct {
	NavStack []tea.Model // Navigation stack to store views
}

func NewBaseModel(m tea.Model) *BaseModel {
	initialNavStack := []tea.Model{m}
	return &BaseModel{NavStack: initialNavStack}
}

func (b *BaseModel) pushView(m tea.Model) {
	// Push a new view onto the stack
	b.NavStack = append(b.NavStack, m)
}

func (b *BaseModel) popView() tea.Model {
	// Don't pop the last view from the stack
	if len(b.NavStack) > 1 {
		b.NavStack = b.NavStack[:len(b.NavStack)-1]
	}
	return b.NavStack[len(b.NavStack)-1]
}

func (m *BaseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		windowSizeMsg = msg
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keymap.Quit):
			return m, tea.Quit
		case key.Matches(msg, Keymap.Back):
			newModel := m.popView()
			// Trigger the Update method of the current submodel
			// Replace the current submodel on the stack with the updated submodel
			m.pushView(newModel)
			return newModel, nil
		}
	}

	return nil, nil
}

func (m BaseModel) Init() tea.Cmd {
	return nil
}

func (m BaseModel) View() string {
	// Get the current submodel from the top of the stack
	currentModel := m.NavStack[len(m.NavStack)-1]

	// Return the view of the current submodel
	return currentModel.View()
}
