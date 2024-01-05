package tui

import (
	"context"
	"fmt"

	"github.com/ardevd/flash/internal/lnd"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lightninglabs/lndclient"
)

// Model for the Channel view
type ChannelModel struct {
	styles     *Styles
	channel    lnd.Channel
	lndService *lndclient.GrpcLndServices
	ctx        context.Context
	htlcTable  table.Model
	base       *BaseModel
}

func NewChannelModel(service *lndclient.GrpcLndServices, channel lnd.Channel, backModel tea.Model) *ChannelModel {
	m := ChannelModel{lndService: service, ctx: context.Background(), channel: channel}
	m.styles = GetDefaultStyles()
	return &m
}

func (m *ChannelModel) initData(width, height int) {
	m.initHtlcsTable(width, height)
}

func (m *ChannelModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	}
	return m, cmd
}

func (m ChannelModel) Init() tea.Cmd {
	return nil
}

func getDirectionString(incoming bool) string {
	if incoming {
		return "IN"
	} else {
		return "OUT"
	}
}

func (m *ChannelModel) initHtlcsTable(width, height int) {
	columns := []table.Column{
		{Title: "Amount", Width: 20},
		{Title: "Expiry (seconds)", Width: 20},
		{Title: "Hash", Width: 20},
		{Title: "Direction", Width: 10},
	}

	rows := []table.Row{}

	// Populate the HTLC table rows
	for _, htlc := range m.channel.Info.PendingHtlcs {
		row := table.Row{fmt.Sprintf("%d", int(htlc.Amount.ToUnit(btcutil.AmountSatoshi))),
			fmt.Sprintf("%d", htlc.Expiry), htlc.Hash.String(), getDirectionString(htlc.Incoming)}
		rows = append(rows, row)
	}

	m.htlcTable = table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithWidth(width),
		table.WithHeight(height/4),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	m.htlcTable.SetStyles(s)
}

func (m ChannelModel) getChannelStateView() string {
	active := m.channel.Info.Active
	var stateText string
	if active {
		stateText = m.styles.PositiveString("ONLINE") + "\n" + m.channel.Info.Uptime.String()
	} else {
		stateText = m.styles.NegativeString("OFFLINE\n")
	}

	return stateText + "\n" + m.styles.SubKeyword("Pending HTLCs: ") + fmt.Sprintf("%d", m.channel.Info.NumPendingHtlcs)
}

func (m ChannelModel) View() string {
	s := m.styles
	channelInfoView := lipgloss.JoinVertical(lipgloss.Left, s.BorderedStyle.Render(
		s.Keyword(m.channel.Alias)+
			"\n"+s.SubKeyword("pubkey:")+m.channel.Info.PubKeyBytes.String()+
			"\n"+s.SubKeyword("chanpoint: ")+m.channel.Info.ChannelPoint))

	channelStateView := lipgloss.JoinVertical(lipgloss.Center, s.BorderedStyle.Render(m.getChannelStateView()))

	topView := lipgloss.JoinHorizontal(lipgloss.Left, channelInfoView, channelStateView)

	htlcTableView := s.BorderedStyle.Render(m.htlcTable.View())

	return lipgloss.JoinVertical(lipgloss.Left,
		topView,
		htlcTableView)
}
