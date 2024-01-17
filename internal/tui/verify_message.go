package tui

import (
	"context"
	"strings"

	"github.com/ardevd/flash/internal/util"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/lightninglabs/lndclient"
)

// Value containers
var signedMessage string
var signature string

type VerifyMessageModel struct {
	styles     *Styles
	lndService *lndclient.GrpcLndServices
	ctx        context.Context
	base       *BaseModel
	keys       keyMap
	form       *huh.Form
}

func newVerifyMessageModel(service *lndclient.GrpcLndServices, base *BaseModel) *VerifyMessageModel {
	m := VerifyMessageModel{lndService: service, base: base, ctx: context.Background(), keys: Keymap}
	m.styles = GetDefaultStyles()
	m.base.pushView(&m)
	m.form = getMessageVerificationForm()
	return &m
}

// Model Update logic
func (m *VerifyMessageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle Base model logic
	model, cmd := m.base.Update(msg)
	if model != nil {
		return model, cmd
	}

	var cmds []tea.Cmd

	// Process the form
	if m.form != nil {
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// getMessageSigningForm returns a new huh.form for signing a message
func getMessageVerificationForm() *huh.Form {
	form := huh.NewForm(
		huh.NewGroup(huh.NewNote().
			Title("Verify Message").
			Description("Verify a message and associated signature"),
			huh.NewInput().
				Title("Message").
				Prompt(">").
				Validate(util.IsMessage).
				Value(&signedMessage),
			huh.NewInput().
				Title("Signature").
				Prompt(">").
				Validate(util.IsMessage).
				Value(&signature)))
	form.NextField()
	return form
}

// Init the model
func (m VerifyMessageModel) Init() tea.Cmd {
	return nil
}

// Get the UI element of the message verification summary
func (m VerifyMessageModel) getMessageVerificationView() string {
	return ""
}

// View returns the model view
func (m VerifyMessageModel) View() string {
	s := m.styles

	v := strings.TrimSuffix(m.form.View(), "\n")
	form := lipgloss.DefaultRenderer().NewStyle().Margin(1, 0).Render(v)

	if m.form.State == huh.StateCompleted && len(messageToSign) > 0 {
		return lipgloss.JoinHorizontal(lipgloss.Left, s.BorderedStyle.Render(m.getMessageVerificationView()))
	}

	return lipgloss.JoinVertical(lipgloss.Left, form)

}
