package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/ardevd/flash/internal/lnd"
	"github.com/ardevd/flash/internal/util"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/routing/route"
)

// Model
type PayInvoiceModel struct {
	styles       *Styles
	lndService   *lndclient.GrpcLndServices
	ctx          context.Context
	base         *BaseModel
	keys         keyMap
	form         *huh.Form
	invoiceState PaymentState
	spinner      spinner.Model
}

// PaymentState indicates the state of a Bolt 11 invoice payment
type PaymentState int

const (
	// PaymentStateNone is when the invoice is yet to be generated
	PaymentStateNone PaymentState = iota

	// PaymentStateDecoded is when the invoice has been and parsed.
	PaymentStateDecoded

	// PaymentStateDecodeError is when the invoice is invalid
	PaymentStateDecodeError

	// PaymentStateSending is when the invoice payment is sending
	PaymentStateSending

	// PaymentStateSettled is when the invoice has settled
	PaymentStateSettled
)

// Value container
var invoiceString string
var maxFee string = "10"

// Instantiate model
func newPayInvoiceModel(service *lndclient.GrpcLndServices, base *BaseModel) *PayInvoiceModel {
	m := PayInvoiceModel{lndService: service, base: base, ctx: context.Background(), keys: Keymap}
	m.styles = GetDefaultStyles()
	m.base.pushView(&m)
	m.form = getInvoicePaymentForm()
	m.invoiceState = PaymentStateNone
	m.spinner = getSpinner()

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
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		windowSizeMsg = msg

	case tea.KeyMsg:
		switch {
		// Enter will pay the invoice is model is in appropriate state
		case key.Matches(msg, Keymap.Enter):
			if m.invoiceState == PaymentStateDecoded {
				m.invoiceState = PaymentStateSending
				return m, paymentCreatedMsg
			}
		}
	// Payment has been decoded and issued.
	case paymentCreated:
		// Animate the spinner and pay the invoice
		cmds = append(cmds, m.spinner.Tick, m.payInvoice)
	// Payment failed
	case paymentError:
		m.invoiceState = PaymentStateNone
	// Payment has been settled
	case paymentSettled:
		m.invoiceState = PaymentStateSettled
	}

	// Process the invoice form
	if m.form != nil {
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
			cmds = append(cmds, cmd)
		}
	}

	if m.form.State == huh.StateCompleted && m.invoiceState == PaymentStateNone {
		// Form is ready, decode the invoice
		_, err := m.lndService.Client.DecodePaymentRequest(m.ctx, invoiceString)
		if err == nil {
			m.invoiceState = PaymentStateDecoded
		} else {
			m.invoiceState = PaymentStateDecodeError
		}
	}

	// Update the spinner
	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// Get the invoice payment form
func getInvoicePaymentForm() *huh.Form {
	form := huh.NewForm(
		huh.NewGroup(huh.NewNote().
			Title("Pay Invoice").
			Description("Pay the provided invoice"),
			huh.NewInput().
				Title("BOLT11 Invoice").
				Prompt(">").
				Value(&invoiceString),
			huh.NewInput().
				Title("Max Fee (sats)").
				Prompt("$").
				Validate(util.IsAmount).
				Value(&maxFee)))

	form.NextField()
	return form
}

// Init
func (m PayInvoiceModel) Init() tea.Cmd {
	return nil
}

// Model view logic
func (m PayInvoiceModel) View() string {
	s := m.styles
	v := strings.TrimSuffix(m.form.View(), "\n")
	form := lipgloss.DefaultRenderer().NewStyle().Margin(1, 0).Render(v)
	switch m.invoiceState {
	case PaymentStateSending:
		return lipgloss.JoinVertical(lipgloss.Left, s.BorderedStyle.Render(m.getPaymentPendingView()))
	case PaymentStateSettled:
		return lipgloss.JoinVertical(lipgloss.Left, s.BorderedStyle.Render(m.getPaymentSettledView()))
	case PaymentStateDecoded:
		return lipgloss.JoinVertical(lipgloss.Left, s.BorderedStyle.Render(fmt.Sprintf("\n%s\n", s.HeaderText.Render("Pay Invoice?"))+
			"\n"+m.getDecodeInvoiceView()))
	case PaymentStateDecodeError:
		return lipgloss.JoinVertical(lipgloss.Left, s.BorderedStyle.Render(fmt.Sprintf("\n%s\n", s.HeaderText.Render("Invalid Invoice"))+
			"\n"+m.getDecodeInvoiceView()))
	}

	return lipgloss.JoinVertical(lipgloss.Left, form)
}

// Get the Payment pending view
func (m PayInvoiceModel) getPaymentPendingView() string {

	return m.styles.HeaderText.Render("Invoice in flight") + "\n\n" +
		fmt.Sprintf("\n\n   %s Sending payment\n\n", m.spinner.View())
}

// Get the Payment settled view
func (m PayInvoiceModel) getPaymentSettledView() string {
	s := m.styles
	return s.HeaderText.Render("Invoice settled") + "\n\n" +
		s.PositiveString("The invoice was successfully settled") + "\n" +
		"Press Esc to return"
}

// Get node name for a given public key. Returns empty string if we can't find a match
func (m PayInvoiceModel) getNodeName(pubkey route.Vertex) string {
	nodeInfo, err := m.lndService.Client.GetNodeInfo(m.ctx, pubkey, false)
	if err != nil {
		return ""
	}

	return nodeInfo.Alias
}

// Decode an invoice string
func (m PayInvoiceModel) getDecodeInvoiceView() string {
	// Decode the invoice string
	invoiceString = lnd.SantizeBoltInvoice(invoiceString)
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

// Pay the invoice
func (m *PayInvoiceModel) payInvoice() tea.Msg {
	fee, err := strconv.Atoi(maxFee)
	if err != nil {
		return paymentError{}
	}
	result := m.lndService.Client.PayInvoice(m.ctx, invoiceString, btcutil.Amount(fee), nil)
	defer close(result)
	completion := make(chan tea.Msg)

	defer close(completion)
	go func() {
		for update := range result {
			if update.Err != nil {
				fmt.Println(update.Err.Error())
				completion <- paymentError{}
			} else {
				completion <- paymentSettled{}
			}
		}
	}()

	return <-completion
}
