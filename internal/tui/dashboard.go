package tui

import (
	"context"

	"github.com/ardevd/flash/internal/lnd"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/lightninglabs/lndclient"
)

type dashboardComponent int

const (
	channels dashboardComponent = iota
	payments
	nodeinfo
	messageTools
	channelTools
	paymentTools
)

var formSelection string

func InitDashboard(service *lndclient.GrpcLndServices, nodeData lnd.NodeData) *DashboardModel {
	m := DashboardModel{lndService: service, ctx: context.Background(), nodeData: nodeData}

	return &m
}

func (m *DashboardModel) initData(width, height int) {

	defaultList := list.New([]list.Item{}, list.NewDefaultDelegate(), width, height/2)
	defaultList.SetShowHelp(true)

	m.lists = []list.Model{defaultList, defaultList}
	m.forms = []*huh.Form{m.generatePaymentToolsForm(), m.generateChannelToolsForm(), m.generateMessageToolsForm()}

	m.lists[channels].Title = "Channels"
	m.lists[channels].SetItems(m.nodeData.GetChannelsAsListItems())
	m.lists[payments].Title = "Latest Payments"
	m.lists[payments].SetItems(m.nodeData.GetPaymentsAsListItems())
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		windowSizeMsg = msg

		v, h := borderedStyle.GetFrameSize()
		m.initData(windowSizeMsg.Width-h, windowSizeMsg.Height-v)
		m.loaded = true

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keymap.Tab):
			m.Next()
		case key.Matches(msg, Keymap.ReverseTab):
			m.Prev()
		case key.Matches(msg, Keymap.Enter):
			switch m.focused {
			case paymentTools, messageTools, channelTools:
				return m.handleFormClick()
			}
		case key.Matches(msg, Keymap.Quit):
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	switch m.focused {
	case payments:
		m.lists[m.focused], cmd = m.lists[m.focused].Update(msg)

	case channels:
		m.lists[m.focused], cmd = m.lists[m.focused].Update(msg)

	case paymentTools:
		m.forms[0].Update(msg)
	case channelTools:
		m.forms[1].Update(msg)
	case messageTools:
		m.forms[2].Update(msg)
	}

	return m, cmd
}

func (m DashboardModel) Init() tea.Cmd {
	return nil
}

func (m DashboardModel) View() string {
	if m.quitting {
		return ""
	}

	if m.loaded {
		channelsView := m.lists[channels].View()
		paymentsView := m.lists[payments].View()

		var listsView string
		switch m.focused {
		case channels:

			listsView = lipgloss.JoinHorizontal(
				lipgloss.Center,
				focusedStyle.Render(channelsView),
				borderedStyle.Render(paymentsView),
			)

		case payments:
			listsView = lipgloss.JoinHorizontal(
				lipgloss.Center,
				borderedStyle.Render(channelsView),
				focusedStyle.Render(paymentsView),
			)

		default:
			listsView = lipgloss.JoinHorizontal(
				lipgloss.Center,
				borderedStyle.Render(channelsView),
				borderedStyle.Render(paymentsView),
			)
		}

		nodeInfoView := lipgloss.JoinVertical(lipgloss.Left, borderedStyle.Render(
			keyword(m.nodeData.NodeInfo.Alias)+"\n"+m.nodeData.NodeInfo.PubKey+
				"\nLnd v"+m.nodeData.NodeInfo.Version))

		balanceView := lipgloss.JoinVertical(lipgloss.Left, borderedStyle.Render(
			subKeyword("Lightning Balance ")+m.nodeData.NodeInfo.ChannelBalance+
				"\n"+subKeyword("Lightning Capacity ")+m.nodeData.NodeInfo.TotalCapacity+
				"\n"+subKeyword("Onchain Balance ")+m.nodeData.NodeInfo.OnChainBalance))

		topView := lipgloss.JoinHorizontal(lipgloss.Left,
			nodeInfoView, balanceView)

		toolsView := lipgloss.JoinHorizontal(lipgloss.Left,
			m.getPaymentTools(), m.getChannelTools(), m.getMessageTools())

		return lipgloss.JoinVertical(
			lipgloss.Left,
			topView,
			listsView,
			toolsView)

	}

	return "Loading..."
}

func (m *DashboardModel) getPaymentTools() string {
	style := borderedStyle
	if m.focused == paymentTools {
		style = focusedStyle
	}

	return style.Render(m.forms[0].WithShowHelp(false).View())
}

func (m *DashboardModel) getMessageTools() string {
	style := borderedStyle
	if m.focused == messageTools {
		style = focusedStyle
	}

	return style.Render(m.forms[2].WithShowHelp(false).View())
}

func (m *DashboardModel) getChannelTools() string {
	style := borderedStyle
	if m.focused == channelTools {
		style = focusedStyle
	}

	return style.Render(m.forms[1].WithShowHelp(false).View())
}

func (m *DashboardModel) generatePaymentToolsForm() *huh.Form {
	s := huh.NewSelect[string]().
		Key("payments").
		Title("Payments\n").
		Options(
			huh.NewOption("Send Payment", OPTION_PAYMENT_SEND),
			huh.NewOption("Generate Invoice", OPTION_PAYMENT_RECEIVE),
		).
		Value(&formSelection)

	return huh.NewForm(huh.NewGroup(s))
}

func (m *DashboardModel) generateChannelToolsForm() *huh.Form {
	s := huh.NewSelect[string]().
		Title("Channels and Peers\n").
		Key("channels").
		Options(
			huh.NewOption("Open Channel", "open"),
			huh.NewOption("Connect to Peer", "connect"),
		).
		Value(&formSelection)

	return huh.NewForm(huh.NewGroup(s))
}

func (m *DashboardModel) generateMessageToolsForm() *huh.Form {

	s := huh.NewSelect[string]().
		Title("Messages\n").
		Key("messages").
		Options(
			huh.NewOption("Sign Message", "sign"),
			huh.NewOption("Verify Message", "verify"),
		).
		Value(&formSelection)

	return huh.NewForm(huh.NewGroup(s))
}

func (m *DashboardModel) handleFormClick() (tea.Model, tea.Cmd) {
	i := m.getInvoiceModel()
	return i.Update(windowSizeMsg)
}

func (m *DashboardModel) getInvoiceModel() InvoiceModel {
	return NewInvoiceModel(m.ctx, m.lndService, StateNone, m)
}

// Navigation
func (m *DashboardModel) Prev() {
	switch m.focused {

	case channels:
		m.focused = messageTools
	case payments:
		m.focused = channels
	case paymentTools:
		m.focused = payments
	case channelTools:
		m.focused = paymentTools
	case messageTools:
		m.focused = channelTools
	}
}

func (m *DashboardModel) Next() {
	switch m.focused {
	case channels:
		m.focused = payments
	case payments:
		m.focused = paymentTools
	case paymentTools:
		m.focused = channelTools
	case channelTools:
		m.focused = messageTools
	case messageTools:
		m.focused = channels
	}
}

// Model for the Dashboard view
type DashboardModel struct {
	focused    dashboardComponent
	lists      []list.Model
	forms      []*huh.Form
	lndService *lndclient.GrpcLndServices
	nodeData   lnd.NodeData
	ctx        context.Context
	loaded     bool
	quitting   bool
}
