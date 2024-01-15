package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/ardevd/flash/internal/lnd"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/lightninglabs/lndclient"
)

// Model for the Channel view
type ChannelModel struct {
	styles      *Styles
	channel     lnd.Channel
	state       ChannelModelState
	lndService  *lndclient.GrpcLndServices
	ctx         context.Context
	htlcTable   table.Model
	base        *BaseModel
	help        help.Model
	keys        keyMap
	form        *huh.Form
	messageChan chan channelStatusMsg
	messages    []channelStatusMsg
}

// ChannelState indicates the state of the selected Channel model
type ChannelModelState int

const (
	// Default state
	ChannelStateNone ChannelModelState = iota

	// User has initated the channel close flow
	ChannelStateWantClose

	// User has initated the channel force close flow
	ChannelStateWantForceClose
)

// Internal message structure used for passing messages from lndclient to the UI
type channelStatusMsg struct {
	// the actual status message
	message string
}

// Extension function for representing channelStatusMsg objects
func (c channelStatusMsg) String() string {
	s := NewStyles(lipgloss.DefaultRenderer())
	return fmt.Sprint(s.Highlight.Render(c.message))
}

// Variable used for asserting confirmation of channel operations
var operationConfirmed bool

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
func NewChannelModel(service *lndclient.GrpcLndServices, channel lnd.Channel, base *BaseModel) *ChannelModel {
	const numStatusMessages = 1
	m := ChannelModel{lndService: service, ctx: context.Background(), channel: channel, base: base, help: help.New(), keys: Keymap,
		messages: make([]channelStatusMsg, numStatusMessages), messageChan: make(chan channelStatusMsg)}

	m.styles = GetDefaultStyles()
	m.base.pushView(&m)
	m.state = ChannelStateNone

	return &m
}

// Load data from lnd API.
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

	var cmds []tea.Cmd

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
			// Force close channel
			m.form = m.getChannelOperationForm("Force Close Channel", "Latest commitment transaction will be broadcast. Are you sure?")
			m.state = ChannelStateWantClose
		case key.Matches(msg, Keymap.Close):
			// Close channel
			m.form = m.getChannelOperationForm("Close Channel", "A cooperative close will be issued. Are you sure?")
			m.state = ChannelStateWantClose
		}

	// ChannelStatus update receive
	case channelStatusMsg:
		m.messages = append(m.messages, msg)
		// Handle next message
		return m, handleChannelUpdateMessages(m.messageChan)
	}

	// Process the form
	if m.form != nil {
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
			cmds = append(cmds, cmd)
		}

		if m.form.State == huh.StateCompleted {
			// User finished form, figure out what operation and if user confirmed
			if m.state == ChannelStateWantForceClose || m.state == ChannelStateWantClose {
				if operationConfirmed {
					// Initiate channel closure
					go m.closeChannel()
					// Start receiving channel update messages
					cmds = append(cmds, handleChannelUpdateMessages(m.messageChan))
				}
			}

			// Revert to default state
			m.state = ChannelStateNone
			operationConfirmed = false
		}

	}

	return m, tea.Batch(cmds...)
}

func (m ChannelModel) Init() tea.Cmd {
	defer close(m.messageChan)
	return nil
}

// Get string indicating whether HTLC is inbound or outbound
func getDirectionString(incoming bool) string {
	if incoming {
		return "IN"
	} else {
		return "OUT"
	}
}

// Initialize the table of pending HTLCs
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

// Get the current channel state view
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

// Get current channel parameters view
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

// Receive messages from the internal messaging channel and pass it on to Update()
func handleChannelUpdateMessages(sub chan channelStatusMsg) tea.Cmd {
	return func() tea.Msg {
		return channelStatusMsg(<-sub)
	}
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

// Force Close/Close the channel.
func (m ChannelModel) closeChannel() {
	outPoint, err := lndclient.NewOutpointFromStr(m.channel.Info.ChannelPoint)
	if err != nil {
		fmt.Println("Unable to parse channel outpoint")
	}

	forceClose := m.state == ChannelStateWantForceClose

	var targetBlocks int32 = 10
	if forceClose {
		// A force close can't include custom fee
		targetBlocks = 0
	}

	updateChan, errorsChan, err := m.lndService.Client.CloseChannel(m.ctx, outPoint, forceClose, targetBlocks, nil)

	if err != nil {
		fmt.Println("Unable to close channel", err)
	}
	// // Use a separate goroutine to receive updates and errors from the lndService
	go func() {
		for {
			select {
			case update, ok := <-updateChan:
				if !ok {
					// Channel closed
					m.messageChan <- channelStatusMsg{message: "Channel closed"}
					return
				}

				m.messageChan <- channelStatusMsg{message: "Broadasting closing transaction: " + update.CloseTxid().String()}
			case errorUpdate, ok := <-errorsChan:
				if !ok {
					m.messageChan <- channelStatusMsg{message: "Could not close channel: " + errorUpdate.Error()}
					return
				}
				m.messageChan <- channelStatusMsg{message: "Error: " + errorUpdate.Error()}
			}
		}
	}()

}

func (m ChannelModel) getChannelOperationForm(title string, description string) *huh.Form {

	form := huh.NewForm(
		huh.NewGroup(huh.NewNote().
			Title(title).
			Description(description),
			huh.NewConfirm().
				Title("Proceed?").
				Value(&operationConfirmed).
				Affirmative("Yes!").
				Negative("No.")))
	form.NextField()
	return form
}

func (m ChannelModel) getStatusMessages() string {
	s := ""
	if len(m.messages) > 0 {
		s += m.styles.SubKeyword(m.messages[len(m.messages)-1].message)
	}

	return s
}

func (m ChannelModel) View() string {
	s := m.styles

	if m.state == ChannelStateNone {
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

		bottomView := lipgloss.JoinVertical(lipgloss.Left, s.Base.Render(m.getStatusMessages()))

		helpView := s.Base.Render(m.help.View(m.keys))

		return lipgloss.JoinVertical(lipgloss.Left,
			topView,
			statsView,
			channelBalanceView,
			htlcTableView,
			bottomView,
			helpView)
	} else if m.state == ChannelStateWantForceClose {
		v := strings.TrimSuffix(m.form.View(), "\n\n")
		form := lipgloss.DefaultRenderer().NewStyle().Margin(1, 0).Render(v)
		return lipgloss.JoinVertical(lipgloss.Left, form)
	} else if m.state == ChannelStateWantClose {
		v := strings.TrimSuffix(m.form.View(), "\n\n")
		form := lipgloss.DefaultRenderer().NewStyle().Margin(1, 0).Render(v)
		return lipgloss.JoinVertical(lipgloss.Left, form)
	}

	return ""
}
