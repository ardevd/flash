package tui

import (
	"context"
	"fmt"

	"github.com/ardevd/flash/internal/lnd"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/lnrpc/walletrpc"
)

// Model for the Channel view
type ChannelModel struct {
	styles     *Styles
	channel    lnd.Channel
	lndService *lndclient.GrpcLndServices
	ctx        context.Context
	htlcTable  table.Model
	base       *BaseModel
	help       help.Model
	keys       keyMap
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Back, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Update, k.Close, k.ForceClose}, // first column
		{k.Back, k.Quit},                  // second column
	}
}

// NewChannelModel returns a new Channel Model.
func NewChannelModel(service *lndclient.GrpcLndServices, channel lnd.Channel, backModel tea.Model, base *BaseModel) *ChannelModel {
	m := ChannelModel{lndService: service, ctx: context.Background(), channel: channel, base: base, help: help.New(), keys: Keymap}
	m.styles = GetDefaultStyles()

	m.base.pushView(&m)
	return &m
}

// Load data from API.
func (m *ChannelModel) initData(width, height int) {
	m.initHtlcsTable(width, height)
}

// Model Update logic
func (m *ChannelModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle Base model logic
	model, cmd := m.base.Update(msg)
	if model != nil {
		return model, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		windowSizeMsg = msg
		m.help.Width = msg.Width
		v, h := m.styles.BorderedStyle.GetFrameSize()
		m.initData(windowSizeMsg.Width-h, windowSizeMsg.Height-v)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keymap.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, Keymap.ForceClose):
			// TODO: Force close channel
			m.closeChannel(true)
		}
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
		stateText = fmt.Sprintf("%s\n%s %v%%", m.styles.PositiveString("ONLINE"), m.styles.SubKeyword("Uptime"), m.channel.UptimePct())
	} else {
		stateText = fmt.Sprintf("%s\n%s %v%%", m.styles.NegativeString("OFFLINE"), m.styles.SubKeyword("Uptime"), m.channel.UptimePct())
	}

	return stateText + "\n" + m.styles.SubKeyword("Pending HTLCs ") + fmt.Sprintf("%d", m.channel.Info.NumPendingHtlcs)
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

func (m ChannelModel) getChannelBalanceView() string {
	return fmt.Sprintf("%s\n\n%s", m.styles.Keyword("Balance"), m.channel.Description())
}

func (m ChannelModel) closeChannel(force bool) {

	address, err := m.lndService.WalletKit.NextAddr(m.ctx, "default", walletrpc.AddressType_TAPROOT_PUBKEY, true)

	if err != nil {
		fmt.Println("Error generating receive address")
	}

	outPoint, err := lndclient.NewOutpointFromStr(m.channel.Info.ChannelPoint)
	if err != nil {
		fmt.Println("Unable to parse channel outpoint")
	}

	closeUpdates, closeErrors, err := m.lndService.Client.CloseChannel(m.ctx, outPoint, force, 5, nil)

	if err != nil {
		fmt.Println("Unable to close channel", err)
	}

	// Start goroutine to listen for close updates
	go func() {
		defer close(closeUpdates)
		defer close(closeErrors)

		for {
			select {
			case update := <-closeUpdates:
				// Handle close updates received from the channel
				fmt.Println("Received close update:", update)
			case errorUpdate := <-closeErrors:
				// The closing process is complete
				fmt.Println("Recieved close error:", errorUpdate)
				return
			}
		}
	}()

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

	channelBalanceView := lipgloss.JoinHorizontal(lipgloss.Left, s.BorderedStyle.Render(m.getChannelBalanceView()))

	htlcTableView := lipgloss.JoinVertical(lipgloss.Left, s.BorderedStyle.Render(s.Keyword("Pending HTLCs\n\n")+m.htlcTable.View()))

	helpView := s.Base.Render(m.help.View(m.keys))

	return lipgloss.JoinVertical(lipgloss.Left,
		topView,
		statsView,
		channelBalanceView,
		htlcTableView,
		helpView)
}
