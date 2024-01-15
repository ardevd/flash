package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/lightninglabs/lndclient"
)

// Model for the message signing view model
type SignMessageModel struct {
	styles     *Styles
	lndService *lndclient.GrpcLndServices
	ctx        context.Context
	base       *BaseModel
	keys       keyMap
	form       *huh.Form
}

// Value container
var messageToSign string

// Instantiate a new model
func newSignMessageModel(service *lndclient.GrpcLndServices, base *BaseModel) *SignMessageModel {
	m := SignMessageModel{lndService: service, base: base, ctx: context.Background(), keys: Keymap}
	m.styles = GetDefaultStyles()
	m.base.pushView(&m)
	m.form = getMessageSigningForm()

	return &m
}

// Model Update logic
func (m *SignMessageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

// Sign the message from the form
func (m SignMessageModel) signMessage() string {

	// Call the SignMessage function
	signature, err := m.lndService.Client.SignMessage(m.ctx, []byte(messageToSign))

	if err != nil {
		fmt.Println("Error: ", err)
	}

	return signature
}

// Get the UI element of signed message
func (m SignMessageModel) getSignedMessageView() string {
	return fmt.Sprintf("%s\n\n%s", m.styles.Keyword("Signed Message"), m.signMessage())
}

// getMessageSigningForm returns a new huh.form for signing a message
func getMessageSigningForm() *huh.Form {
	form := huh.NewForm(
		huh.NewGroup(huh.NewNote().
			Title("Sign Message").
			Description("Sign a message with your node's private key"),
			huh.NewInput().
				Title("Message to sign").
				Prompt("?").
				Value(&messageToSign)))
	form.NextField()
	return form
}

// Init the model
func (m SignMessageModel) Init() tea.Cmd {
	return nil
}

// View returns the model view
func (m SignMessageModel) View() string {
	s := m.styles
	v := strings.TrimSuffix(m.form.View(), "\n\n")
	form := lipgloss.DefaultRenderer().NewStyle().Margin(1, 0).Render(v)

	if m.form.State == huh.StateCompleted && len(messageToSign) > 0 {
		return lipgloss.JoinHorizontal(lipgloss.Left, s.BorderedStyle.Render(m.getSignedMessageView()))
	}

	return lipgloss.JoinVertical(lipgloss.Left, form)

}
