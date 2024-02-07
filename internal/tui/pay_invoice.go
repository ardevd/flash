package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcutil"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/routing/route"
)

type PayInvoiceModel struct {
	styles     *Styles
	lndService *lndclient.GrpcLndServices
	ctx        context.Context
	base       *BaseModel
	keys       keyMap
	form       *huh.Form
}

// Value container
var invoiceString string

// Instantiate model
func newPayInvoiceModel(service *lndclient.GrpcLndServices, base *BaseModel) *PayInvoiceModel {
	m := PayInvoiceModel{lndService: service, base: base, ctx: context.Background(), keys: Keymap}
	m.styles = GetDefaultStyles()
	m.base.pushView(&m)
	m.form = getInvoicePaymentForm()

	return &m
}

// Model update logic
func (m *PayInvoiceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle base model events
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

func getInvoicePaymentForm() *huh.Form {
	form := huh.NewForm(
		huh.NewGroup(huh.NewNote().
			Title("Pay Invoice").
			Description("Pay the provided invoice"),
			huh.NewInput().
				Title("BOLT11 Invoice").
				Prompt(">").
				Value(&invoiceString)))

	form.NextField()
	return form
}

func (m PayInvoiceModel) Init() tea.Cmd {
	return nil
}

func (m PayInvoiceModel) View() string {
	s := m.styles
	v := strings.TrimSuffix(m.form.View(), "\n")
	form := lipgloss.DefaultRenderer().NewStyle().Margin(1, 0).Render(v)
	if m.form.State == huh.StateCompleted {
		return lipgloss.JoinVertical(lipgloss.Left, s.BorderedStyle.Render(fmt.Sprintf("\n%s\n", s.HeaderText.Render("Pay Invoice?"))+
		"\n" + m.decodeInvoice()))
	}
	return lipgloss.JoinVertical(lipgloss.Left, form)
}

func (m PayInvoiceModel) getNodeName(pubkey route.Vertex) string {
	nodeInfo, err := m.lndService.Client.GetNodeInfo(m.ctx, pubkey, false)
	if err != nil {
		return ""
	}

	return nodeInfo.Alias
}

func (m PayInvoiceModel) decodeInvoice() string {
	// Decode the invoice string
	decodedInvoice, err := m.lndService.Client.DecodePaymentRequest(m.ctx, invoiceString)
	if err != nil {
		return "Error decoding invoice: " + err.Error()
	}

	amountInSats := decodedInvoice.Value.ToSatoshis()

	s := m.styles
	return s.Keyword("Amount: ") + amountInSats.String() + "\n" +
		s.Keyword("To: ") + decodedInvoice.Destination.String() + "\n" +
		s.Keyword("Node: ") + m.getNodeName(decodedInvoice.Destination) + "\n" +
		s.Keyword("Description: ") + decodedInvoice.Description + "\n\n" +
		s.SubKeyword("Press Enter to accept, Esc to cancel")
}

func (m PayInvoiceModel) payInvoice() {
	result := m.lndService.Client.PayInvoice(m.ctx, invoiceString, btcutil.Amount(10), nil)
	defer close(result)
	for update := range result {
		if update.Err != nil {
			fmt.Println(update.Err.Error())
			break
		} else {
			fmt.Println("Payment preimage: " + update.Preimage.Hash().String())
			break
		}

	}
}
