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

func (m ChannelModel) getChannelParameters() string {
	edge, err := m.lndService.Client.GetChanInfo(m.ctx, m.channel.Info.ChannelID)
	if err != nil {
		return "Error retrieving channel edge info"
	}

	// Figure out which node is local and remote
	var localNodePolicy, remoteNodePolicy *lndclient.RoutingPolicy
	if edge.Node1.String() == m.channel.Info.PubKeyBytes.String() {
		localNodePolicy = edge.Node2Policy
		remoteNodePolicy = edge.Node1Policy
	} else {
		localNodePolicy = edge.Node1Policy
		remoteNodePolicy = edge.Node2Policy
	}

	localView := fmt.Sprintf("%s\n%s %v\n%s %v\n%s %v\n%s %v\n%s %v", m.styles.Keyword("Local"),
		m.styles.SubKeyword("Base"),
		localNodePolicy.FeeBaseMsat, m.styles.SubKeyword("Rate"), localNodePolicy.FeeRateMilliMsat,
		m.styles.SubKeyword("CLTV Delta"), localNodePolicy.TimeLockDelta,
		m.styles.SubKeyword("Max HTLC"), localNodePolicy.MaxHtlcMsat/1000,
		m.styles.SubKeyword("Min HTLC"), localNodePolicy.MinHtlcMsat/1000)

	remoteView := fmt.Sprintf("%s\n%s %v\n%s %v", m.styles.Keyword("Remote"), m.styles.SubKeyword("Base"), remoteNodePolicy.FeeBaseMsat,
		m.styles.SubKeyword("Rate"), remoteNodePolicy.FeeRateMilliMsat)

	return lipgloss.JoinHorizontal(lipgloss.Left, localView, remoteView)
}

func (m ChannelModel) getChannelStats() string {
	totalReceived := m.channel.Info.TotalReceived
	totalSent := m.channel.Info.TotalSent
	totalSum := totalReceived + totalSent
	unsettledBalance := m.channel.Info.UnsettledBalance

	var (
		channelType,
		openType string
	)
	if m.channel.Info.Private {
		channelType = "Private"
	} else {
		channelType = "Public"
	}

	if m.channel.Info.Initiator {
		openType = "Local"
	} else {
		openType = "Remote"
	}

	return m.styles.Keyword("Stats\n") + m.styles.SubKeyword("Activity (total/sent/received): ") + fmt.Sprintf("%v/%v/%v BTC", totalSum.ToBTC(), totalSent.ToBTC(), totalReceived.ToBTC()) + "\n" +
		m.styles.SubKeyword("Unsettled Balance: ") + unsettledBalance.String() + "\n" +
		m.styles.SubKeyword("Channel Type: ") + channelType + "\n" +
		m.styles.SubKeyword("Channel Opener: ") + openType + "\n" +
		m.styles.SubKeyword("Current Commit Fee: ") + fmt.Sprintf("%v sats", m.channel.Info.CommitFee.ToUnit(btcutil.AmountSatoshi))

}

func (m ChannelModel) View() string {
	s := m.styles
	channelInfoView := lipgloss.JoinVertical(lipgloss.Left, s.BorderedStyle.Render(
		s.Keyword(m.channel.Alias)+
			"\n"+s.SubKeyword("pubkey:")+m.channel.Info.PubKeyBytes.String()+
			"\n"+s.SubKeyword("chanpoint: ")+m.channel.Info.ChannelPoint))

	channelStateView := lipgloss.JoinVertical(lipgloss.Center, s.BorderedStyle.Render(m.getChannelStateView()))

	topView := lipgloss.JoinHorizontal(lipgloss.Left, channelInfoView, channelStateView)

	statsView := lipgloss.JoinHorizontal(lipgloss.Left, s.BorderedStyle.Render(m.getChannelStats()),
		s.BorderedStyle.Render(m.getChannelParameters()))

	htlcTableView := lipgloss.JoinVertical(lipgloss.Left, s.BorderedStyle.Render(s.Keyword("Pending HTLCs\n\n")+m.htlcTable.View()))

	return lipgloss.JoinVertical(lipgloss.Left,
		topView,
		statsView,
		htlcTableView)
}
