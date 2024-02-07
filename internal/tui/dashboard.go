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
	m := DashboardModel{lndService: service, ctx: context.Background(), nodeData: nodeData, keys: Keymap}
	m.styles = GetDefaultStyles()
	return &m
}

func (m *DashboardModel) initData(width, height int) {

	defaultList := list.New([]list.Item{}, list.NewDefaultDelegate(), width, height/2)
	defaultList.SetShowHelp(true)

	m.lists = []list.Model{defaultList, defaultList}
	m.forms = []*huh.Form{m.generatePaymentToolsForm(), m.generateChannelToolsForm(), m.generateMessageToolsForm()}

	m.lists[channels].Title = "Channels"
	m.lists[channels].SetItems(m.nodeData.GetChannelsAsListItems(false))
	m.lists[channels].AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			m.keys.OfflineChannels,
			m.keys.Refresh,
		}
	}
	m.lists[payments].Title = "Latest Payments"
	m.lists[payments].SetItems(m.nodeData.GetPaymentsAsListItems())

	m.base = *NewBaseModel(m)
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Base model logic
	model, cmd := m.base.Update(msg)
	if cmd != nil {
		return model, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		windowSizeMsg = msg
		v, h := m.styles.BorderedStyle.GetFrameSize()
		m.initData(windowSizeMsg.Width-h, windowSizeMsg.Height-v)
		m.loaded = true

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keymap.Tab):
			m.Next()
			return m, nil
		case key.Matches(msg, Keymap.ReverseTab):
			m.Prev()
			return m, nil
		case key.Matches(msg, Keymap.OfflineChannels):
			m.lists[channels].SetItems(m.nodeData.GetChannelsAsListItems(true))
		case key.Matches(msg, Keymap.Refresh):
			m.lists[channels].SetItems(m.nodeData.GetChannelsAsListItems(false))
		case key.Matches(msg, Keymap.Enter):
			switch m.focused {
			case channels:
				return m.handleChannelClick()
			}
		}
	}

	var cmds []tea.Cmd

	switch m.focused {
	case payments:
		m.lists[m.focused], cmd = m.lists[m.focused].Update(msg)
		cmds = append(cmds, cmd)

	case channels:
		m.lists[m.focused], cmd = m.lists[m.focused].Update(msg)
		cmds = append(cmds, cmd)

	case paymentTools:
		_, cmd := m.forms[0].Update(msg)
		cmds = append(cmds, cmd)
		if m.forms[0].State == huh.StateCompleted {
			return m.handleFormClick(paymentTools)
		}
	case channelTools:
		_, cmd := m.forms[1].Update(msg)
		cmds = append(cmds, cmd)
	case messageTools:
		_, cmd := m.forms[2].Update(msg)
		cmds = append(cmds, cmd)
		if m.forms[2].State == huh.StateCompleted {
			return m.handleFormClick(messageTools)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m DashboardModel) Init() tea.Cmd {
	return nil
}

func (m DashboardModel) View() string {
	s := m.styles

	if m.loaded {
		channelsView := m.lists[channels].View()
		paymentsView := m.lists[payments].View()

		var listsView string
		switch m.focused {
		case channels:

			listsView = lipgloss.JoinHorizontal(
				lipgloss.Center,
				s.FocusedStyle.Render(channelsView),
				s.BorderedStyle.Render(paymentsView),
			)

		case payments:
			listsView = lipgloss.JoinHorizontal(
				lipgloss.Center,
				s.BorderedStyle.Render(channelsView),
				s.FocusedStyle.Render(paymentsView),
			)

		default:
			listsView = lipgloss.JoinHorizontal(
				lipgloss.Center,
				s.BorderedStyle.Render(channelsView),
				s.BorderedStyle.Render(paymentsView),
			)
		}

		nodeInfoView := lipgloss.JoinVertical(lipgloss.Left, s.BorderedStyle.Render(
			s.Keyword(m.nodeData.NodeInfo.Alias)+"\n"+m.nodeData.NodeInfo.PubKey+
				"\nLnd v"+m.nodeData.NodeInfo.Version))

		balanceView := lipgloss.JoinVertical(lipgloss.Left, s.BorderedStyle.Render(
			s.SubKeyword("Lightning Balance ")+m.nodeData.NodeInfo.ChannelBalance+
				"\n"+s.SubKeyword("Lightning Capacity ")+m.nodeData.NodeInfo.TotalCapacity+
				"\n"+s.SubKeyword("Onchain Balance ")+m.nodeData.NodeInfo.OnChainBalance))

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
	style := m.styles.BorderedStyle
	if m.focused == paymentTools {
		style = m.styles.FocusedStyle
	}

	return style.Render(m.forms[0].WithShowHelp(false).View())
}

func (m *DashboardModel) getMessageTools() string {
	style := m.styles.BorderedStyle
	if m.focused == messageTools {
		style = m.styles.FocusedStyle
	}

	return style.Render(m.forms[2].WithShowHelp(false).View())
}

func (m *DashboardModel) getChannelTools() string {
	style := m.styles.BorderedStyle
	if m.focused == channelTools {
		style = m.styles.FocusedStyle
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
			huh.NewOption("Open Channel", OPTION_CHANNEL_OPEN),
			huh.NewOption("Connect to Peer", OPTION_CONNECT_TO_PEER),
		).
		Value(&formSelection)

	return huh.NewForm(huh.NewGroup(s))
}

func (m *DashboardModel) generateMessageToolsForm() *huh.Form {

	s := huh.NewSelect[string]().
		Title("Messages\n").
		Key("messages").
		Options(
			huh.NewOption("Sign Message", OPTION_MESSAGE_SIGN),
			huh.NewOption("Verify Message", OPTION_MESSAGE_VERIFY),
		).
		Value(&formSelection)

	return huh.NewForm(huh.NewGroup(s))
}

func (m *DashboardModel) handleChannelClick() (tea.Model, tea.Cmd) {
	selectedChannel := m.lists[m.focused].SelectedItem().(lnd.Channel)
	return NewChannelModel(m.lndService, selectedChannel, &m.base).Update(windowSizeMsg)
}

func (m *DashboardModel) handleFormClick(component dashboardComponent) (tea.Model, tea.Cmd) {
	var i tea.Model
	switch component {
	case paymentTools:
		if m.forms[0].GetString("payments") == OPTION_PAYMENT_RECEIVE {
			i = newInvoiceModel(m.ctx, &m.base, m.lndService, StateNone)
		} else {
			i = newPayInvoiceModel(m.lndService, &m.base)
		}
		m.forms[0] = m.generatePaymentToolsForm()
	case messageTools:
		if m.forms[2].GetString("messages") == OPTION_MESSAGE_SIGN {
			
			i = newSignMessageModel(m.lndService, &m.base)
		} else {
			
			i = newVerifyMessageModel(m.lndService, &m.base)
		}
		m.forms[2] = m.generateMessageToolsForm()
	}

	return i.Update(windowSizeMsg)

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
	styles     *Styles
	focused    dashboardComponent
	lists      []list.Model
	forms      []*huh.Form
	lndService *lndclient.GrpcLndServices
	nodeData   lnd.NodeData
	ctx        context.Context
	loaded     bool
	base       BaseModel
	keys       keyMap
}
