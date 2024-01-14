package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/keychain"
	"github.com/tv42/zbase32"
)

// Model for the message signign view
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

func (m SignMessageModel) signMessage() string {

	keyLocator := &keychain.KeyLocator{
		Family: keychain.KeyFamilyNodeKey,
	}

	// Call the SignMessage function
	signature, err := m.lndService.Signer.SignMessage(m.ctx, []byte(messageToSign),
		*keyLocator, lndclient.SignCompact())

	if err != nil {
		fmt.Println("Error: ", err)
	}

	return zbase32.EncodeToString(signature)
}

func (m SignMessageModel) getSignedMessageView() string {
	return fmt.Sprintf("%s\n\n%s", m.styles.Keyword("Signed Message"), m.signMessage())
}

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

func (m SignMessageModel) Init() tea.Cmd {
	return nil
}

func (m SignMessageModel) View() string {
	s := m.styles
	v := strings.TrimSuffix(m.form.View(), "\n\n")
	form := lipgloss.DefaultRenderer().NewStyle().Margin(1, 0).Render(v)

	if m.form.State == huh.StateCompleted && len(messageToSign) > 0 {
		return lipgloss.JoinHorizontal(lipgloss.Left, s.BorderedStyle.Render(m.getSignedMessageView()))
	}

	return lipgloss.JoinVertical(lipgloss.Left, form)

}
