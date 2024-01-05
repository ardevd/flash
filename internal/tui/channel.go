package tui

import (
	"context"
	"fmt"

	"github.com/ardevd/flash/internal/lnd"
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
	dashboard  *DashboardModel
	base       BaseModel
}

func NewChannelModel(service *lndclient.GrpcLndServices, channel lnd.Channel, dashboard *DashboardModel) *ChannelModel {
	m := ChannelModel{lndService: service, ctx: context.Background(), channel: channel, base: *NewBaseModel(), dashboard: dashboard}
	m.styles = GetDefaultStyles()
	return &m
}

func (m *ChannelModel) initData(width, height int) {
	// TODO: Init list data

}

func (m ChannelModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	return lipgloss.JoinVertical(lipgloss.Left,
		topView)
}
