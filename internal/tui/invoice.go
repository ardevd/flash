package tui

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ardevd/flash/internal/lnd"
	"github.com/ardevd/flash/internal/util"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/lightninglabs/lndclient"
	invpkg "github.com/lightningnetwork/lnd/invoices"
	"github.com/skip2/go-qrcode"
)

// Invoice status
type InvoiceState int

const (
	// StateNone is when the invoice is yet to be generated
	StateNone InvoiceState = iota

	// StateGenerated is when the invoice has been generated, but not paid
	StateGenerated

	// StateSettled is when the invoice has been settled
	StateSettled

	// StateExpired is when the invoice has expired
	StateExpired

	// StateError is when the invoice subscription fails
	StateError
)

// Styling
const maxWidth = 80

var (
	red    = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}
	indigo = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	green  = lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}
)

// InvoiceModel Model struct
type InvoiceModel struct {
	lg           *lipgloss.Renderer
	styles       *Styles
	form         *huh.Form
	width        int
	lndService   *lndclient.GrpcLndServices
	ctx          context.Context
	invoiceState InvoiceState
	dashboard    *DashboardModel
}

// Variables for form value reference
var (
	amount     string = "100"
	memo       string
	expiration string = "3600"
)

// Invoice value
var invoiceVal string

func isFormReady(v bool) error {
	if !v {
		return errors.New("cancelled")
	}

	if len(amount) == 0 {
		return errors.New("invalid amount")
	}

	if len(expiration) == 0 {
		return errors.New("no expiration provided")
	}

	return nil
}

// Invoice generation form
func NewInvoiceModel(service *lndclient.GrpcLndServices, context context.Context, state InvoiceState, dashboard *DashboardModel) InvoiceModel {
	m := InvoiceModel{width: maxWidth, lndService: service, ctx: context, invoiceState: state, dashboard: dashboard}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Amount").
				Prompt("?").
				Validate(util.IsAmount).
				Value(&amount),
			huh.NewInput().
				Title("Memo").
				Prompt("?").
				Validate(util.IsMemo).
				Value(&memo),
			huh.NewInput().
				Title("Expiration (seconds)").
				Prompt("?").
				Validate(util.IsMemo).
				Value(&expiration),
			huh.NewConfirm().
				Key("done").
				Title("Ready?").
				Validate(isFormReady).
				Affirmative("Submit").
				Negative("Cancel"),
		),
	).
		WithShowHelp(false).
		WithShowErrors(false)
	return m
}

// BubbleTea init
func (m InvoiceModel) Init() tea.Cmd {
	return m.form.Init()
}

// Navigate back to the dashboard model
func (m InvoiceModel) backToDashboard() (tea.Model, tea.Cmd) {
	return m.dashboard.Update(windowSizeMsg)
}

// Handle update messages for the model
func (m InvoiceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		windowSizeMsg = msg
		m.width = min(msg.Width, maxWidth) - m.styles.Base.GetHorizontalFrameSize()

		// Hack for getting correct focus on amount form field.
		// TODO: Must be a better way
		m.form.NextField()
		m.form.PrevField()

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keymap.Quit):
			return m, tea.Quit
		case key.Matches(msg, Keymap.Back):
			return m.backToDashboard()
		case key.Matches(msg, Keymap.Enter):
			if m.invoiceState == StateSettled || m.invoiceState == StateExpired || m.invoiceState == StateError {
				return m.backToDashboard()
			}
		}

	case paymentSettled:
		// Payment settled
		m.invoiceState = StateSettled

	case paymentExpired:
		// Payment expired
		m.invoiceState = StateExpired

	case paymentCreated:
		return m, m.subscribeToInvoices
	}

	var cmds []tea.Cmd

	// Process the form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted && len(invoiceVal) == 0 {
		// Form is ready, generate invoice
		generatedInvoice, err := m.generateInvoice()
		if err != nil {
			generatedInvoice = err.Error()
		}

		invoiceVal = generatedInvoice
		m.invoiceState = StateGenerated
		return m, paymentCreatedMsg
	}

	return m, tea.Batch(cmds...)
}

func paymentCreatedMsg() tea.Msg {
	return paymentCreated{}
}

func (m InvoiceModel) subscribeToInvoices() tea.Msg {
	// Create a context with cancellation after some time
	e, err := strconv.Atoi(expiration)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e)*time.Second)

	defer cancel()

	// Subscribe to invoices
	invoiceUpdates, streamErr, err := m.lndService.Client.SubscribeInvoices(ctx, lndclient.InvoiceSubscriptionRequest{})
	if err != nil {
		return err
	}

	// TODO: Check this
	// defer close(ch)

	// Handle invoice updates and errors
	for {
		select {
		case invoice := <-invoiceUpdates:
			// Invoice update received
			if invoice.PaymentRequest == invoiceVal {
				// Invoice match
				if invoice.State == invpkg.ContractSettled {
					// Invoice settled
					fmt.Println("Invoice settled")
					return paymentSettled{}
				}
			}

		case err := <-streamErr:
			// TODO: Format this nicely
			fmt.Println("Received stream error:", err)
			return (errors.New("stream error"))
			// Handle the received stream error

		case <-ctx.Done():
			fmt.Println("Invoice expired")
			return paymentExpired{}
		}
	}
}

// Sanity check invoice arguments and generate the LND lightning invoice
func (m InvoiceModel) generateInvoice() (string, error) {

	parsedAmount, err := strconv.ParseUint(amount, 10, 64)
	if err != nil {
		return "", errors.New("invalid amount")
	}

	parsedExpiration, err := strconv.ParseInt(expiration, 10, 64)

	if err != nil {
		return "", errors.New("invalid expiration")
	}

	_, invoice, err := lnd.GeneratePaymentInvoice(m.lndService, m.ctx, memo, parsedAmount, parsedExpiration)
	return invoice, err
}

func (m InvoiceModel) printQrCode() string {
	// Generate the QR code as ASCII art
	qr, err := qrcode.New(invoiceVal, qrcode.Medium)
	if err != nil {
		return err.Error()
	}
	return qr.ToSmallString(true)
}

func (m InvoiceModel) getInvoiceCancelView() string {
	s := m.styles
	view := lipgloss.JoinVertical(lipgloss.Left, borderedStyle.Render(fmt.Sprintf("\n%s\n", s.HeaderText.Render("Invoice Generation Cancelled")) +
		fmt.Sprintf("\n%s\n", "The invoice generation was cancelled. No invoice data committed.") +
		fmt.Sprintf("\n\n%s\n", "Press Esc to return")))

	return view
}

func (m InvoiceModel) View() string {
	_, h := borderedStyle.GetFrameSize()
	s := m.styles

	switch m.invoiceState {
	case StateSettled:
		view := lipgloss.JoinVertical(lipgloss.Left, borderedStyle.Render(fmt.Sprintf("\n%s\n", s.HeaderText.Render("Invoice Settled")) +
			fmt.Sprintf("\n%s\n", "The invoice was settled. Payment received") +
			fmt.Sprintf("\n\n%s\n", "Press Enter to return")))

		return view

	case StateExpired:
		view := lipgloss.JoinVertical(lipgloss.Left, borderedStyle.Render(fmt.Sprintf("\n%s\n", s.HeaderText.Render("Invoice Expired")) +
			fmt.Sprintf("\n%s\n", "The invoice was expired. No payment settled.") +
			fmt.Sprintf("\n\n%s\n", "Press Enter to return")))
		return view
	}

	switch m.form.State {

	case huh.StateCompleted:
		b := strings.Builder{}
		fmt.Fprintf(&b, "\n%s\n", s.HeaderText.Render("Invoice Ready"))
		fmt.Fprintf(&b, "\n%s\n\n", invoiceVal)
		fmt.Fprintf(&b, "%s\n", m.printQrCode())

		return s.Status.Copy().Margin(0, 1).Padding(1, 2).Width(windowSizeMsg.Width-h).Render(b.String()) + "\n\n"
	default:
		v := strings.TrimSuffix(m.form.View(), "\n\n")
		form := m.lg.NewStyle().Margin(1, 0).Render(v)

		errors := m.form.Errors()
		header := m.appBoundaryView("Invoice Generation")
		if len(errors) > 0 {
			header = m.appErrorBoundaryView(m.errorView())
		}
		body := lipgloss.JoinHorizontal(lipgloss.Top, form)

		footer := m.appBoundaryView(m.form.Help().ShortHelpView(m.form.KeyBinds()))
		if len(errors) > 0 {
			if errors[0].Error() == "cancelled" {
				return m.getInvoiceCancelView()
			}
			footer = m.appErrorBoundaryView("")
		}

		return s.Base.Render(header + "\n" + body + "\n\n" + footer)
	}
}

func (m InvoiceModel) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	return s
}

func (m InvoiceModel) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(indigo),
	)
}

func (m InvoiceModel) appErrorBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.ErrorHeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(red),
	)
}
